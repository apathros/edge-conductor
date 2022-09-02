/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package plugin

import (
	"context"
	wfapi "github.com/intel/edge-conductor/pkg/api/workflow"
	certmgr "github.com/intel/edge-conductor/pkg/certmgr"
	eputils "github.com/intel/edge-conductor/pkg/eputils"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	log "github.com/sirupsen/logrus"
)

type Plugin struct {
	name        string
	conn        *grpc.ClientConn
	client      wfapi.WorkflowClient
	data        eputils.SchemaStruct
	plugin_data eputils.SchemaStruct
	finished    bool
}

const (
	CONNECT_TIMEOUT = 3600
	DEFAULT_TIMEOUT = 5
)

func New(name string, data eputils.SchemaStruct, plugin_data eputils.SchemaStruct) *Plugin {
	return &Plugin{name: name, data: data, plugin_data: plugin_data, finished: false}
}

func (p *Plugin) Connect(address string) error {
	clientTLSConfig, err := certmgr.GetTLSConfigByName("workflow", "client", "")
	if err != nil {
		return err
	}
	clientCreds := credentials.NewTLS(clientTLSConfig)
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_TIMEOUT*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(clientCreds))
	if err != nil {
		log.Warningf("grpc connect error: %v", err)
		return eputils.GetError("errGrpcConnect")
	}
	p.client = wfapi.NewWorkflowClient(conn)
	p.conn = conn

	ctx, cancel = context.WithTimeout(context.Background(), CONNECT_TIMEOUT*time.Second)
	defer cancel()
	r, err := p.client.PluginConnect(ctx, &wfapi.PluginConnectRequest{Plugin: &wfapi.Plugin{Name: p.name}})
	if err != nil {
		return err
	}
	if r.Result.Return == wfapi.ConnectResult_Completed {
		p.finished = true
		return nil
	}
	if r.Result.Return == wfapi.ConnectResult_Connected {
		if p.plugin_data != nil && r.WorkflowData.PluginData != nil {
			if err := p.plugin_data.UnmarshalBinary(r.WorkflowData.PluginData); err != nil {
				return eputils.GetError("errUnmarshalPlugin")
			}
		}
		if p.data != nil && r.WorkflowData.Data != nil {
			if err := p.data.UnmarshalBinary(r.WorkflowData.Data); err != nil {
				log.Errorf("Unmarshal data error: %v", err)
				return eputils.GetError("errUnmarshalData")
			}
		}
		return nil
	}
	return eputils.GetError("errUnknown")

}

func (p *Plugin) Complete(err error) error {
	defer p.conn.Close()
	req := &wfapi.PluginCompleteRequest{
		Plugin:       &wfapi.Plugin{Name: p.name},
		Result:       &wfapi.Result{Return: wfapi.Result_Success},
		WorkflowData: &wfapi.WorkflowData{},
	}
	if err != nil {
		req.Result.Return = wfapi.Result_Error
	} else {
		if p.plugin_data != nil {
			if req.WorkflowData.PluginData, err = p.plugin_data.MarshalBinary(); err != nil {
				return eputils.GetError("errUnmarshalPlugin")
			}
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_TIMEOUT*time.Second)
	defer cancel()
	_, err = p.client.PluginComplete(ctx, req)
	if err != nil {
		log.Errorf("Plugin Complete error, %v", err)
		return eputils.GetError("errPluginComplete")
	}
	return err
}
