/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package workflow

import (
	"context"
	"fmt"
	wfapi "github.com/intel/edge-conductor/pkg/api/workflow"
	certmgr "github.com/intel/edge-conductor/pkg/certmgr"
	"github.com/intel/edge-conductor/pkg/eputils"
	"net"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func (s *server) serve(address string) error {
	log.Infof("listen on %v", address)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	serverTLSConfig, err := certmgr.GetTLSConfigByName("workflow", "server", lis.Addr().String())
	if err != nil {
		return err
	}
	serverCreds := credentials.NewTLS(serverTLSConfig)

	grpcServer := grpc.NewServer(grpc.Creds(serverCreds))
	wfapi.RegisterWorkflowServer(grpcServer, s)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Errorf("failed to serve: %v", err)
			s.errch <- err
			s.finished <- true
		}
	}()
	return err
}

func (s *server) PluginConnect(ctx context.Context, req *wfapi.PluginConnectRequest) (*wfapi.PluginConnectResponse, error) {
	log.Infof("PluginConnect: plugin %v\n", req.Plugin.Name)
	res := &wfapi.PluginConnectResponse{Result: &wfapi.ConnectResult{Return: wfapi.ConnectResult_Connected}}
	if req.Plugin.Name == "__init__" {
		log.Infof("PluginConnect: __init__\n")
		res.WorkflowData = s.data
		return res, nil
	}
	st := s.getPendingStep(req.Plugin.Name)
	if st == nil {
		log.Infof("PluginConnect: no pending step needs this plugin\n")
		res.Result.Return = wfapi.ConnectResult_Completed
		return res, nil
	}
	log.Infof("PluginConnect: wait\n")
	<-st.started
	s.current.pending = false
	log.Infof("PluginConnect: plugin %v is connected", req.Plugin.Name)
	res.WorkflowData = s.data
	return res, nil
}

func (s *server) PluginPutLog(logstream wfapi.Workflow_PluginPutLogServer) error {
	for {
		if l, err := logstream.Recv(); err == nil {
			fmt.Printf("%s", l.Log)
		} else {
			//cancelled ?
			log.Debugf("logstream, err :%v\n", err)
			break
		}
	}
	return nil
}

func (s *server) PluginComplete(ctx context.Context, req *wfapi.PluginCompleteRequest) (*wfapi.Result, error) {
	log.Infof("PluginComplete: plugin %v, res %v", req.Plugin.Name, req.Result.Return)
	if req.Result.Return != wfapi.Result_Success {
		log.Errorf("PluginComplete error: plugin %v, res %v", req.Plugin.Name, req.Result.Return)
		s.errch <- eputils.GetError("errPluginComplete")
		s.finished <- true
	}
	s.data.PluginData = req.WorkflowData.PluginData
	r := &wfapi.Result{Return: wfapi.Result_Success}
	s.current.finished <- true
	return r, nil
}
