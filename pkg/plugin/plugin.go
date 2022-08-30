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
	eputils "github.com/intel/edge-conductor/pkg/eputils"
	"sync"

	log "github.com/sirupsen/logrus"
)

var Address = "localhost:50088"

type PluginMainFuncs struct {
	name        string
	data        eputils.SchemaStruct
	in          *eputils.SchemaMapData
	out         *eputils.SchemaMapData
	plugin_data *eputils.SchemaMapData
	mainFunc    func(eputils.SchemaMapData, *eputils.SchemaMapData) error
	wg          sync.WaitGroup
	remoteLog   bool
	started     bool
	err         error
}

var (
	mains []*PluginMainFuncs
)

type LogHook struct {
	writer func(s string)
}

func (hook *LogHook) Fire(entry *log.Entry) error {
	line, err := entry.Bytes()
	if err != nil {
		return err
	}
	hook.writer(string(line))
	return err
}

func (hook *LogHook) Levels() []log.Level {
	return log.AllLevels
}

func RegisterPlugin(name string, in *eputils.SchemaMapData, out *eputils.SchemaMapData,
	mainFunc func(eputils.SchemaMapData, *eputils.SchemaMapData) error) {
	m := &PluginMainFuncs{
		name:        name,
		in:          in,
		out:         out,
		plugin_data: &eputils.SchemaMapData{},
		mainFunc:    mainFunc,
		remoteLog:   false,
		started:     false,
		err:         nil,
	}
	mains = append(mains, m)
}

func run(m *PluginMainFuncs) error {
	logctx, logcancel := context.WithCancel(context.Background())
	defer logcancel()
	log.Infof("Start Plugin %v\n", m.name)
	p := New(m.name, m.data, m.plugin_data)

	for {
		log.Infof("Connecting Plugin %v\n", m.name)
		if err := p.Connect(Address); err != nil {
			log.Warningf("Plugin connection error: %v\n", err)
			return eputils.GetError("errPluginConnect")
		}
		if p.finished {
			log.Infof("Plugin is finished\n")
			return nil
		}
		log.Infof("Connected Plugin %v\n", m.name)
		if m.remoteLog {
			log.Debugf("get remote console\n")
			logstream, err := p.client.PluginPutLog(logctx)
			if err != nil {
				log.Warningf("Get log stream error, %v", err)
				return eputils.GetError("errGetLogStream")
			}
			logexit := false
			log.AddHook(&LogHook{
				writer: func(s string) {
					if logexit {
						return
					}
					if err := logstream.Send(&wfapi.Log{Log: s}); err != nil {
						logexit = true
					}
				},
			})
		}
		log.Infof("Exec Plugin %v\n", m.name)
		for k := range *m.in {
			if _, has := (*m.plugin_data)[k]; !has {
				log.Warningf("Cannot find input schema: %s\n", k)
				return eputils.GetError("errInputSchema")
			}
			(*m.in)[k] = (*m.plugin_data)[k]
			log.Debugf("set input[%s] = %v\n", k, (*m.in)[k])
		}
		for k := range *m.out {
			log.Debugf("empty output[%s]\n", k)
			(*m.out)[k] = eputils.SchemaStructNew(k)
		}
		err := m.mainFunc(*m.in, m.out)
		if err == nil {
			*m.plugin_data = *m.out
		} else {
			log.Errorf("Plugin error: name: %v, err: %v", m.name, err)
		}
		log.Infof("Complete Plugin %v\n", m.name)
		err = p.Complete(err)
		if err != nil {
			return err
		}
	}
}

func StartPlugin(name string, errch chan error) error {
	for _, m := range mains {
		if m.name == name {
			if m.started {
				log.Infof("Start Plugin %s, already started\n", m.name)
				return nil
			}
			m.started = true
			m.wg.Add(1)
			go func(m *PluginMainFuncs) {
				m.err = run(m)
				if errch != nil {
					errch <- m.err
				}
				m.wg.Done()
			}(m)
			return nil
		}
	}
	log.Warningf("Cannot find %v", name)
	return eputils.GetError("errFind")
}

func WaitPluginFinished(name string) error {
	for _, m := range mains {
		if m.name == name {
			m.wg.Wait()
			return m.err
		}
	}
	log.Errorf("Cannot find %v", name)
	return eputils.GetError("errFind")
}

func EnablePluginRemoteLog(name string) error {
	for _, m := range mains {
		if m.name == name {
			m.remoteLog = true
			return nil
		}
	}
	log.Warningf("Cannot find %v", name)
	return eputils.GetError("errFind")
}
