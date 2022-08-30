/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package executor

import (
	"context"
	"fmt"
	"github.com/intel/edge-conductor/pkg/eputils"
	docker "github.com/intel/edge-conductor/pkg/eputils/docker"
	"github.com/intel/edge-conductor/pkg/eputils/orasutils"
	repoutils "github.com/intel/edge-conductor/pkg/eputils/repoutils"
	restfulcli "github.com/intel/edge-conductor/pkg/eputils/restfulcli"
	"io"
	"os"
	"path"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

var (
	DayZeroCertFilePath = "cert/pki/ca.pem"
)

func (e *Executor) CmdOverrideWithNode(cmd []string, ni *nodeInfo) ([]string, error) {
	newCmd := make([]string, len(cmd))
	copy(newCmd, cmd)
	for k, c := range newCmd {
		c, err := e.StringOverrideWithNode(c, ni)
		if err != nil {
			return newCmd, err
		}
		newCmd[k] = c
		log.Debugf("CmdOverrideWithNode node: %v, newCmd: %v", ni.ip, newCmd[k])
	}
	return newCmd, nil
}

func (e *Executor) runPipeFrom(ctx context.Context, nodes map[string]*nodeInfo, cmd []string, from *nodeInfo, fromCmd []string) error {
	finalErr := error(nil)
	wg := sync.WaitGroup{}
	wg.Add(len(nodes))
	for _, n := range nodes {
		n := n
		cmd, err := e.CmdOverrideWithNode(cmd, n)
		if err != nil {
			return err
		}
		go func() {
			defer wg.Done()
			if err := n.client.Connect(); err != nil {
				finalErr = err
				return
			}
			r, w := io.Pipe()
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := from.client.CmdWithAttachIO(ctx, fromCmd,
					nil, w, os.Stderr, false); err != nil {
					finalErr = err
				}
				if err := w.Close(); err != nil {
					finalErr = err
				}
			}()
			if err := n.client.CmdWithAttachIO(ctx, cmd,
				r, os.Stdout, os.Stderr, false); err != nil {
				finalErr = err
				return
			}
			if err := n.client.Disconnect(); err != nil {
				finalErr = err
				return
			}
		}()
	}
	wg.Wait()
	return finalErr
}

func (e *Executor) runPipeTo(ctx context.Context, nodes map[string]*nodeInfo, cmd []string, to *nodeInfo, toCmd []string) error {
	finalErr := error(nil)
	wg := sync.WaitGroup{}
	wg.Add(len(nodes))
	for _, n := range nodes {
		n := n
		cmd, err := e.CmdOverrideWithNode(cmd, n)
		if err != nil {
			return err
		}
		go func() {
			defer wg.Done()
			if err := n.client.Connect(); err != nil {
				finalErr = err
				return
			}
			r, w := io.Pipe()
			defer func() {
				if err := w.Close(); err != nil {
					log.Errorf("Failed to close io pipe.")
				}
			}()
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := to.client.CmdWithAttachIO(ctx, toCmd,
					r, os.Stdout, os.Stderr, false); err != nil {
					finalErr = err
				}
			}()
			if err := n.client.CmdWithAttachIO(ctx, cmd,
				nil, w, os.Stderr, false); err != nil {
				finalErr = err
				return
			}
			if err := n.client.Disconnect(); err != nil {
				finalErr = err
				return
			}
		}()
	}
	wg.Wait()
	return finalErr
}

func (e *Executor) helperShell(ctx context.Context, nodes map[string]*nodeInfo, cmd []string) error {
	log.Debugf("cmd: %v", strings.Join(cmd, "@"))
	finalErr := error(nil)
	wg := sync.WaitGroup{}
	wg.Add(len(nodes))
	for _, n := range nodes {
		n := n
		cmd, err := e.CmdOverrideWithNode(cmd, n)
		if err != nil {
			return err
		}
		go func() {
			defer wg.Done()
			if err := n.client.Connect(); err != nil {
				finalErr = err
				return
			}
			if err := n.client.CmdWithAttachIO(ctx, cmd, nil, os.Stdout, os.Stderr, true); err != nil {
				finalErr = err
				return
			}
			if err := n.client.Disconnect(); err != nil {
				finalErr = err
				return
			}
		}()
	}
	wg.Wait()
	return finalErr
}

func (e *Executor) helperCopyFromDay0(ctx context.Context, nodes map[string]*nodeInfo, cmd []string) error {
	log.Debugf("CopyFromDay0: from %v to %v\n", cmd[0], cmd[1])
	fromBase := path.Base(cmd[0])
	fromDir := path.Dir(cmd[0])
	toDir := cmd[1]
	if toDir[len(toDir)-1] != '/' {
		log.Errorf("CopyFromDay0: \"%s\" should be end with \"/\"", toDir)
		return eputils.GetError("errCopyFromDay0")
	}
	return e.runPipeFrom(ctx, nodes,
		[]string{"tar", "-x", "-C", toDir},
		e.nodesByRole["day-0"][0],
		[]string{"tar", "-c", "-C", fromDir, fromBase})
}

func (e *Executor) helperCopyToDay0(ctx context.Context, nodes map[string]*nodeInfo, cmd []string) error {
	log.Debugf("CopyToDay0: from %v to %v\n", cmd[0], cmd[1])
	fromBase := path.Base(cmd[0])
	fromDir := path.Dir(cmd[0])
	toDir := cmd[1]
	if toDir[len(toDir)-1] != '/' {
		log.Errorf("CopyToDay0: \"%s\" should be end with \"/\"", toDir)
		return eputils.GetError("errCopyToDay0")
	}
	return e.runPipeTo(ctx, nodes,
		[]string{"tar", "-c", "-C", fromDir, fromBase},
		e.nodesByRole["day-0"][0],
		[]string{"tar", "-x", "-C", toDir})
}

func (e *Executor) helperPushImage(ctx context.Context, nodes map[string]*nodeInfo, cmd []string) error {
	log.Debugf("ctx: %v", ctx)
	d0 := e.nodesByRole["day-0"][0]
	for _, n := range nodes {
		if n.ip != d0.ip {
			return eputils.GetError("errNoDay0")
		}
	}
	cmd, err := e.CmdOverrideWithNode(cmd, d0)
	if err != nil {
		return err
	}
	input_ep_params := e.tempParams
	auth, err := docker.GetAuthConf(input_ep_params.Kitconfig.Parameters.GlobalSettings.ProviderIP,
		input_ep_params.Kitconfig.Parameters.GlobalSettings.RegistryPort,
		input_ep_params.Kitconfig.Parameters.Customconfig.Registry.User,
		input_ep_params.Kitconfig.Parameters.Customconfig.Registry.Password)
	if err != nil {
		return err
	}
	var newImages []string
	if newImages, err = restfulcli.MapImageURLCreateHarborProject(input_ep_params.Kitconfig.Parameters.GlobalSettings.ProviderIP,
		input_ep_params.Kitconfig.Parameters.GlobalSettings.RegistryPort,
		input_ep_params.Kitconfig.Parameters.Customconfig.Registry.User,
		input_ep_params.Kitconfig.Parameters.Customconfig.Registry.Password, cmd); err != nil {
		if err != nil {
			return err
		}
	}

	for _, url := range newImages {
		newTag, err := docker.TagImageToLocal(url, auth.ServerAddress)
		if err != nil {
			return err
		}
		log.Infof("Push %s to %s", url, newTag)
		if err := docker.ImagePush(newTag, auth); err != nil {
			return err
		}
	}

	return nil
}

func (e *Executor) helperCreateProjectOnHarbor(ctx context.Context, nodes map[string]*nodeInfo, image []string) error {
	log.Debugf("image list: %v", image)
	log.Debugf("ctx: %v", ctx)
	log.Debugf("nodes: %v", nodes)
	input_ep_params := e.tempParams
	if input_ep_params.Kitconfig == nil || input_ep_params.Kitconfig.Parameters == nil || input_ep_params.Kitconfig.Parameters.GlobalSettings == nil || input_ep_params.Kitconfig.Parameters.Customconfig == nil {
		err := eputils.GetError("errKitCfgParameter")
		return err
	}
	if _, err := restfulcli.MapImageURLCreateHarborProject(input_ep_params.Kitconfig.Parameters.GlobalSettings.ProviderIP,
		input_ep_params.Kitconfig.Parameters.GlobalSettings.RegistryPort,
		input_ep_params.Kitconfig.Parameters.Customconfig.Registry.User,
		input_ep_params.Kitconfig.Parameters.Customconfig.Registry.Password, image); err != nil {
		return err
	}

	return nil
}

func (e *Executor) helperPushFile(ctx context.Context, nodes map[string]*nodeInfo, cmd []string) error {

	log.Debugf("ctx: %v", ctx)

	d0 := e.nodesByRole["day-0"][0]
	for _, n := range nodes {
		if n.ip != d0.ip {
			log.Errorf("pushFile: only support on day-0")
			return eputils.GetError("errPushOnlyOnDay0")
		}
	}

	if len(cmd) == 0 {
		log.Errorf("pushFile: invalid command")
		return eputils.GetError("errPushInvalidCmd")
	}

	cmd, err := e.CmdOverrideWithNode(cmd, d0)
	if err != nil {
		return err
	}
	/* Check the following parameters in OrasPushFile, then no need check it again */
	fileName := cmd[0]
	subRef := cmd[1]
	rev := cmd[2]

	log.Debugf("helperPushFile fileName:%s subRef:%s rev:%s\r\n", fileName, subRef, rev)

	_, err = repoutils.PushFileToRepo(fileName, subRef, rev)
	if err != nil {
		log.Errorf("helperPushFile failed! fileName:%s\r\n", fileName)
		return err
	}

	return nil
}

func (e *Executor) helperPullFile(ctx context.Context, nodes map[string]*nodeInfo, cmd []string) error {

	log.Debugf("ctx: %v", ctx)

	d0 := e.nodesByRole["day-0"][0]
	for _, n := range nodes {
		if n.ip != d0.ip {
			log.Errorf("pullFile: only support on day-0")
			return eputils.GetError("errPullOnlyOnDay0")
		}
	}

	if len(cmd) == 0 {
		log.Errorf("pullFile: invalid command")
		return eputils.GetError("errPullInvalidCmd")
	}

	cmd, err := e.CmdOverrideWithNode(cmd, d0)
	if err != nil {
		return err
	}
	input_ep_params := e.tempParams
	ip := input_ep_params.Kitconfig.Parameters.GlobalSettings.ProviderIP
	port := input_ep_params.Kitconfig.Parameters.GlobalSettings.RegistryPort

	targetFile := cmd[0]
	subRef := cmd[1]
	rev := cmd[2]

	if rev == "" {
		rev = "0.0.0"
	}

	targetUrl := fmt.Sprintf("oci://%s:%s/%s/%s:%s", ip, port, orasutils.RegProject, subRef, rev)
	log.Debugf("helperPullFile targetUrl: %s\n", targetUrl)

	err = repoutils.PullFileFromRepo(targetFile, targetUrl)
	if err != nil {
		log.Errorf("helperPullFile failed! targetUrl:%s\r\n", targetUrl)
		return err
	}

	return nil
}
