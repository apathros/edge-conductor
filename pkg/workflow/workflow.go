/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package workflow

import (
	pluginapi "github.com/intel/edge-conductor/pkg/api/plugins"
	wfapi "github.com/intel/edge-conductor/pkg/api/workflow"
	epplugins "github.com/intel/edge-conductor/pkg/epplugins"
	eputils "github.com/intel/edge-conductor/pkg/eputils"
	docker "github.com/intel/edge-conductor/pkg/eputils/docker"
	plugin "github.com/intel/edge-conductor/pkg/plugin"
	"io/ioutil"
	"os"
	fpath "path/filepath"

	log "github.com/sirupsen/logrus"
)

const (
	RUNTIME_DATA_DIR = "runtime/data"
)

type dataAttr struct {
	name         string
	value        string
	confidential bool
	filepath     string
	filemode     os.FileMode
}

type io struct {
	name       string
	schemaName string
}

type step struct {
	plugin    string
	container string
	pending   bool
	started   chan bool
	finished  chan bool
	inputs    []io
	outputs   []io
}

type server struct {
	wfapi.UnimplementedWorkflowServer
	name             string
	workflow         *wfapi.Workflow
	steps            []step
	current          *step
	plugin_data      eputils.SchemaMapData
	plugin_dataattrs map[string]dataAttr
	containers       wfapi.Containers
	finished         chan bool
	data             *wfapi.WorkflowData
	errch            chan error
}

func isBuiltInPlugin(name string) bool {
	for _, p := range epplugins.PluginList {
		if p == name {
			return true
		}
	}
	return false
}

func findNameFromWf(workflows []*wfapi.WorkflowSpecWorkflowsItems0, name string) *wfapi.WorkflowSpecWorkflowsItems0 {
	for _, wf := range workflows {
		if wf.Name == name {
			return wf
		}
	}
	return nil
}

func (s *server) setInitData() error {
	name := "ep-params"
	initData := eputils.SchemaStructNew(name)
	if data, has := s.plugin_dataattrs[name]; has {
		if err := eputils.LoadSchemaStructFromYaml(initData, data.value); err != nil {
			return eputils.GetError("errLoadData")
		}
	} else {
		return eputils.GetError("errSchemaInitData")
	}
	json, err := initData.MarshalBinary()
	if err != nil {
		return eputils.GetError("errMarshalInitData")
	}
	s.data.Data = json
	return nil
}

func (s *server) getPendingStep(name string) *step {
	for k := range s.steps {
		if s.steps[k].plugin == name && s.steps[k].pending {
			return &s.steps[k]
		}
	}
	return nil
}

func (s *server) IsConfidentialData(name string) bool {
	if data, has := s.plugin_dataattrs[name]; has {
		return data.confidential
	}
	return false
}

func (s *server) loadPluginData() error {
	if err := os.MkdirAll(RUNTIME_DATA_DIR, os.FileMode(0700)); err != nil {
		return err
	}
	for _, data := range s.workflow.Spec.Data {
		yml := data.Value
		filemode := os.FileMode(0600)
		filepath := ""
		if !data.Confidential {
			filepath = fpath.Join(RUNTIME_DATA_DIR, data.Name)
			if _, err := os.Stat(filepath); !os.IsNotExist(err) {
				buf, err := ioutil.ReadFile(filepath)
				if err != nil {
					return err
				}
				log.Infof("Init data[%s] from file: %s\n", data.Name, filepath)
				yml = string(buf)
			} else {
				log.Infof("%s not exists, Init data[%s] from value\n", filepath, data.Name)
			}
		} else {
			log.Infof("Confidential data, init plugin data[%s] from value\n", data.Name)
		}
		s.plugin_dataattrs[data.Name] = dataAttr{
			name:         data.Name,
			filepath:     filepath,
			filemode:     filemode,
			confidential: data.Confidential,
			value:        yml,
		}
	}

	return nil
}

func (s *server) loadPluginConfig() {
	for _, p := range s.workflow.Spec.Plugins {
		for k := range s.steps {
			if s.steps[k].plugin == p.Name {
				s.steps[k].container = p.Container
			}
		}
	}
}

func (s *server) loadContainers() {
	for _, ctn := range s.workflow.Spec.Containers {
		needToRun := false
		for _, st := range s.steps {
			if ctn.Name == st.container {
				ctn.Args = append(ctn.Args, st.plugin)
				needToRun = true
			}
		}
		if needToRun {
			s.containers = append(s.containers, ctn)
		}
	}
}

func run_container(ctn *wfapi.ContainersItems0) error {
	c := &pluginapi.ContainersItems0{}
	if err := eputils.ConvertSchemaStruct(&ctn, c); err != nil {
		log.Errorf("Convert containers data error: %v\n", err)
		return eputils.GetError("errConvContainers")
	}
	log.Infof("remove container %s\n", c.Name)
	if err := docker.DockerRemove(c); err != nil {
		log.Debug(err)
	}
	log.Infof("start container %s\n", c.Name)
	if err := docker.DockerRun(c); err != nil {
		return err
	}
	return nil
}

func (s *server) startPlugins() error {
	for _, st := range s.steps {
		if len(st.container) == 0 && isBuiltInPlugin(st.plugin) {
			log.Debugf("start plugin: %v\n", st.plugin)
			if err := plugin.StartPlugin(st.plugin, s.errch); err != nil {
				return err
			}
		}
	}

	log.Infof("Start Plugin Containers: %v\n", s.containers)
	for _, ctn := range s.containers {
		if err := run_container(ctn); err != nil {
			return err
		}
	}
	return nil
}

func (s *server) loadSteps() error {
	if len(s.workflow.Spec.Workflows) <= 0 {
		return eputils.GetError("errWorkflow")
	}
	wf := findNameFromWf(s.workflow.Spec.Workflows, s.name)
	if wf == nil {
		log.Errorf("This command %s is not supported for current config.", s.name)
		return eputils.GetError("errCmdNotSupported")
	}
	log.Infof("Current workflow: %s\n", wf.Name)
	for _, st := range wf.Steps {
		inputs := []io{}
		for _, in := range st.Input {
			inputs = append(inputs, io{
				name:       in.Name,
				schemaName: st.Name + "." + in.Schema,
			})
		}

		outputs := []io{}
		for _, out := range st.Output {
			outputs = append(outputs, io{
				name:       out.Name,
				schemaName: st.Name + "." + out.Schema,
			})
		}
		s.steps = append(s.steps, step{
			plugin:   st.Name,
			inputs:   inputs,
			outputs:  outputs,
			pending:  true,
			started:  make(chan bool),
			finished: make(chan bool),
		})
	}
	if len(s.steps) > 0 {
		s.current = &s.steps[0]
	}
	log.Debugf("steps: %v\n", s.steps)
	return nil
}

func (s *server) run() error {
	for k, st := range s.steps {
		pdata := eputils.SchemaMapData{}
		log.Debugf("Prepare plugin data\n")
		for _, in := range st.inputs {
			if v, has := s.plugin_data[in.name]; has {
				if s.IsConfidentialData(in.name) {
					log.Debugf("Input %s, schema: %s, value: *confidential data*\n", in.name, in.schemaName)
				} else {
					log.Debugf("Input %s, schema: %s, value: %v\n", in.name, in.schemaName, v)
				}
				pdata[in.schemaName] = v
			} else {
				pdata[in.schemaName] = eputils.SchemaStructNew(in.schemaName)
				if data, has := s.plugin_dataattrs[in.name]; has {
					if err := eputils.LoadSchemaStructFromYaml(pdata[in.schemaName], data.value); err != nil {
						log.Warningf("Load plugin data [%s] from init data error, %v\n", in.name, err)
						return eputils.GetError("errPluginData")
					}
					if s.IsConfidentialData(in.name) {
						log.Debugf("init plugin_data: [%s]: *confidential data*", in.name)
					} else {
						log.Debugf("init plugin_data: [%s]: %s", in.name, data.value)
					}
				} else {
					log.Errorf("Cannot find schema %s in previous step's output or init data \n", in.name)
					return eputils.GetError("errPreviousSchema")
				}
			}
		}
		json, err := pdata.MarshalBinary()
		if err != nil {
			log.Errorf("pdata marshal error: %v", err)
			return eputils.GetError("errMarshalPdata")
		}
		s.data.PluginData = json

		log.Infof("kickoff plugin: %v", st.plugin)
		s.current = &s.steps[k]
		s.current.started <- true
		<-s.current.finished

		if err := pdata.UnmarshalBinary(s.data.PluginData); err != nil {
			log.Errorf("Plugin return pdata error, err: %v, json dump: %v", err, string(s.data.PluginData))
			return eputils.GetError("errPluginReturn")
		}
		for _, out := range st.outputs {
			if v, has := pdata[out.schemaName]; has {
				if s.IsConfidentialData(out.name) {
					log.Debugf("Output %s as %s, value: *confidential data*\n", out.name, out.schemaName)
				} else {
					log.Debugf("Output %s as %s, value: %v\n", out.name, out.schemaName, v)
				}
				s.plugin_data[out.name] = v
			} else {
				log.Errorf("Cannot find schema %s in output\n", out.schemaName)
				return eputils.GetError("errSchemaOutData")
			}
			if data, has := s.plugin_dataattrs[out.name]; has {
				if !data.confidential {
					log.Infof("Update data to file: %s\n", data.filepath)
					yml, err := eputils.SchemaStructToYaml(pdata[out.schemaName])
					if err != nil {
						return err
					}
					if err := ioutil.WriteFile(data.filepath, []byte(yml), data.filemode); err != nil {
						return err
					}
				}
			}
		}
		log.Debugf("PluginComplete: plugin_data: %v", s.plugin_data)
	}
	log.Infof("workflow finished")
	return nil
}

func Start(name string, address string, configFile string) error {
	log.Infof("load workflow config file %v", configFile)
	wf := wfapi.Workflow{}
	err := eputils.LoadSchemaStructFromYamlFile(&wf, configFile)
	if err != nil {
		return err
	}

	s := &server{
		name:             name,
		workflow:         &wf,
		steps:            []step{},
		plugin_data:      eputils.SchemaMapData{},
		plugin_dataattrs: map[string]dataAttr{},
		finished:         make(chan bool),
		data:             &wfapi.WorkflowData{},
		errch:            make(chan error),
	}

	if err := s.loadSteps(); err != nil {
		return err
	}
	if len(s.steps) <= 0 {
		log.Infof("No step in this workflow to run.")
		return nil
	}
	s.loadPluginConfig()
	if err := s.loadPluginData(); err != nil {
		return err
	}
	s.loadContainers()
	if err := s.setInitData(); err != nil {
		return err
	}
	if err := s.serve(address); err != nil {
		return err
	}
	if err := s.startPlugins(); err != nil {
		return err
	}
	go func() {
		err := s.run()
		if err != nil {
			log.Errorln("Server run fail")
		}
		s.errch <- err
		s.finished <- true
	}()
	for err == nil {
		select {
		case <-s.finished:
			return err
		case err = <-s.errch:
			break
		}
	}
	return err
}
