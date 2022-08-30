/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package executor

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"

	api "github.com/intel/edge-conductor/pkg/api/ep"
	pluginapi "github.com/intel/edge-conductor/pkg/api/plugins"
	eputils "github.com/intel/edge-conductor/pkg/eputils"
)

type client interface {
	Connect() error
	CmdWithAttachIO(ctx context.Context, cmd []string, stdin io.Reader, stdout, stderr io.Writer, tty bool) error
	Disconnect() error
}

type nodeInfo struct {
	name   string
	ip     string
	client client
}

type tempParameter struct {
	pluginapi.EpParams
	Value interface{}
	Node  interface{}
}

type Executor struct {
	api.Execspec
	tempParams  tempParameter
	nodesByIP   map[string]*nodeInfo
	nodesByRole map[string][]*nodeInfo
}

func New() *Executor {
	return &Executor{
		nodesByIP:   map[string]*nodeInfo{},
		nodesByRole: map[string][]*nodeInfo{},
	}
}

func (e *Executor) StringOverrideWithNode(s string, ni *nodeInfo) (string, error) {
	p := e.tempParams
	if !strings.Contains(s, `\{\{`) {
		return s, nil
	}
	c := s
	c = strings.ReplaceAll(c, `\{\{`, `{{`)
	c = strings.ReplaceAll(c, `\}\}`, `}}`)

	user, err := user.Current()
	if err != nil {
		return "", err
	}
	for _, n := range e.tempParams.Kitconfig.Parameters.Nodes {
		if n.IP == ni.ip {
			if n.User == "" {
				n.User = user.Username
			}
			p.Node = n
			break
		}
	}
	if p.Node == nil && ni.ip == "127.0.0.1" {
		p.Node = &pluginapi.Node{
			User: user.Username,
			IP:   "127.0.0.1",
		}
	}
	c, err = eputils.StringTemplateConvertWithParams(c, p)
	if err != nil {
		log.Warningf("StringOverrideWithNode error, node: %v, s: %v, err: %v", ni.ip, s, err)
		return "", eputils.GetError("errStringOverrideWithNode")
	}
	log.Debugf("StringOverrideWithNode node: %v, s: %v, new: %v", ni.ip, s, c)
	return c, nil
}

func (e *Executor) NodeListUpdate(params *pluginapi.KitconfigParameters) error {
	if _, has := e.nodesByRole["day-0"]; !has {
		e.nodesByRole["day-0"] = []*nodeInfo{}
		e.nodesByRole["day-0"] = append(e.nodesByRole["day-0"], &nodeInfo{
			name:   "day-0",
			ip:     "127.0.0.1",
			client: &day0Client{},
		})
	}
	for _, n := range params.Nodes {
		if n.IP == "" {
			log.Infof("node [%v] has no IP, ignore\n", n.Name)
			continue
		}
		if n.IP == "127.0.0.1" {
			continue
		}
		if _, has := e.nodesByIP[n.IP]; !has {
			log.Debugf("Add new node [%v], IP: %v\n", n.Name, n.IP)
			key := n.SSHKey
			if key == "" {
				homeDir, err := os.UserHomeDir()
				if err == nil {
					n.SSHKeyPath = strings.Replace(n.SSHKeyPath, "~", homeDir, 1)
					keyBytes, err := ioutil.ReadFile(n.SSHKeyPath)
					if err == nil {
						key = string(keyBytes)
					}
				}
			}
			if n.User == "" {
				log.Warningf("Node [%v] login username missing\n", n.IP)
				return eputils.GetError("errNodeLogin")
			}
			if key == "" && n.SSHPasswd == "" {
				log.Warningf("Node [%v] login password and ssh key missing\n", n.IP)
				return eputils.GetError("errNodeLoginPassword")
			}
			port := int(n.SSHPort)
			if port == 0 {
				port = 22
			}
			node := &nodeInfo{
				name: n.Name,
				ip:   n.IP,
				client: &sshClient{
					host:     n.IP,
					port:     port,
					user:     n.User,
					password: n.SSHPasswd,
					key:      key,
				},
			}
			e.nodesByIP[n.IP] = node
			if len(n.Role) == 0 {
				e.nodesByRole["unknownRole"] = append(e.nodesByRole["unknownRole"], node)
			}
			for _, r := range n.Role {
				if _, hasR := e.nodesByRole[r]; !hasR {
					e.nodesByRole[r] = []*nodeInfo{}
				}
				e.nodesByRole[r] = append(e.nodesByRole[r], node)
			}
		}
	}
	return nil
}

func (e *Executor) SetECParams(epparams *pluginapi.EpParams) error {
	err := eputils.ConvertSchemaStruct(epparams, &e.tempParams)
	if err != nil {
		return err
	}
	return e.NodeListUpdate(epparams.Kitconfig.Parameters)
}

func (e *Executor) SetTempValue(value interface{}) error {
	e.tempParams.Value = value
	return nil
}

func (e *Executor) LoadSpecFromString(specStr string) error {
	log.Debugf("spec: %v\n", specStr)
	specStr, err := eputils.StringTemplateConvertWithParams(specStr, e.tempParams)
	log.Debugf("temp convert: err: %v, spec: %v\n", err, specStr)
	if err != nil {
		return err
	}
	json, err := yaml.YAMLToJSON([]byte(specStr))
	if err != nil {
		return err
	}
	err = e.UnmarshalBinary(json)
	if err != nil {
		return err
	}
	return err
}

func (e *Executor) LoadSpecFromFile(specFile string) error {
	specByte, err := ioutil.ReadFile(specFile)
	if err != nil {
		return err
	}
	err = eputils.CheckHashForContent(specByte, specFile, "")
	if err != nil {
		if err.Error() == eputils.ERRORCODE_CHECK_HASH_FAIL {
			log.Errorln("ERROR: Failed to check executor spec hash:", specFile)
			log.Errorln("ERROR: A user defined executor spec will not be executed by the tool to avoid any security risk.")
		}
		return err
	}
	return e.LoadSpecFromString(string(specByte))
}

func (e *Executor) RunWithAttachIO(ctx context.Context, stdin io.Reader, stdout, stderr io.Writer) error {
	for _, step := range e.Spec.Steps {
		log.Debugf("step: %v\n", step.Name)
		nodes := map[string]*nodeInfo{}
		for _, r := range step.Nodes.AnyOf {
			if nn, has := e.nodesByRole[r]; has {
				nodes[nn[0].ip] = nn[0]
				break
			}
		}
		for _, r := range step.Nodes.AllOf {
			if nn, has := e.nodesByRole[r]; has {
				for _, nnn := range nn {
					nodes[nnn.ip] = nnn
				}
			}
		}
		if len(step.Nodes.NoneOf) != 0 {
			for r := range e.nodesByRole {
				found := false
				for _, selector := range step.Nodes.NoneOf {
					if r == selector {
						found = true
						break
					}
				}
				if found {
					continue
				}
				if nn, has := e.nodesByRole[r]; has {
					for _, nnn := range nn {
						nodes[nnn.ip] = nnn
					}
				}
			}
		}
		for _, command := range step.Commands {
			err := error(nil)
			cnodes := map[string]*nodeInfo{}

			for k, n := range nodes {
				whenString, err := e.StringOverrideWithNode(command.When, n)
				if err != nil {
					log.Warningf("Ignore override error, %v, err: %v\n", command.When, err)
					return eputils.GetError("errIgnoreOverride")
				}
				if whenString == "" {
					whenString = "true"
				}
				when, err := strconv.ParseBool(whenString)
				if err != nil {
					log.Warningf("Ignore format error, %v, err: %v\n", command.When, err)
					return eputils.GetError("errIgnoreFormat")
				}
				if when {
					cnodes[k] = n
				}
			}

			if len(cnodes) == 0 {
				log.Debugf("nodes are empty, ignore cmd [%v]", command)
				continue
			}

			if command.Type == "shell" {
				err = e.helperShell(ctx, cnodes, command.Cmd)
			} else if command.Type == "copyFromDay0" {
				err = e.helperCopyFromDay0(ctx, cnodes, command.Cmd)
			} else if command.Type == "copyToDay0" {
				err = e.helperCopyToDay0(ctx, cnodes, command.Cmd)
			} else if command.Type == "pushImage" {
				err = e.helperPushImage(ctx, cnodes, command.Cmd)
			} else if command.Type == "pushFile" {
				err = e.helperPushFile(ctx, cnodes, command.Cmd)
			} else if command.Type == "pullFile" {
				err = e.helperPullFile(ctx, cnodes, command.Cmd)
			} else if command.Type == "createHarborProject" {
				err = e.helperCreateProjectOnHarbor(ctx, cnodes, command.Cmd)
			} else {
				log.Errorf("Unknown command type: %v\n", command.Type)
				err = eputils.GetError("errUnknownCmdType")
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *Executor) Run(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	return e.RunWithAttachIO(ctx, nil, os.Stdout, os.Stderr)
}
