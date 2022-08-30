/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package main

import (
	"flag"
	epapp "github.com/intel/edge-conductor/cmd/ep/app"
	epapiplugins "github.com/intel/edge-conductor/pkg/api/plugins"
	"github.com/intel/edge-conductor/pkg/epplugins"
	plugin "github.com/intel/edge-conductor/pkg/plugin"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

const (
	CONNECT_RETRY = 6
)

func main() {
	flag.Parse()

	if len(flag.Args()) <= 0 {
		os.Exit(1)
	}
	plugin.Address = flag.Args()[0]

	if len(flag.Args()) >= 2 {
		logLevel := flag.Args()[1]
		if logLevel == "Debug" {
			log.SetLevel(log.DebugLevel)
		} else {
			log.SetLevel(log.InfoLevel)
		}
	} else {
		log.SetLevel(log.InfoLevel)
	}
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: false,
	})

	if err := os.Chdir(os.Getenv("PWD")); err != nil {
		log.Errorln(err)
		os.Exit(1)
	}

	epparams := &epapiplugins.EpParams{}
	pInit := plugin.New("__init__", epparams, nil)
	for i := 0; i < CONNECT_RETRY; i++ {
		err := pInit.Connect(plugin.Address)
		if err != nil {
			log.Debugf("Connect Server (retry #%d) : %v", i, err)
			time.Sleep(1 * time.Second)
		} else {
			log.Info("Connected to Server")
			break
		}
		if i == CONNECT_RETRY-1 {
			log.Errorf("Connect Server Err and exit: %v", err)
			os.Exit(1)
		}
	}

	log.Debugf("__init__ return: %v", epparams)
	if err := epapp.EpUtilsInit(epparams); err != nil {
		log.Errorln(err)
		os.Exit(1)
	}

	log.Infof("plug list: %v", epplugins.PluginList)
	for _, p := range epplugins.PluginList {
		log.Infof("Enable plugin remote Log: %v\n", p)
		if err := plugin.EnablePluginRemoteLog(p); err != nil {
			log.Fatal(err)
		}
		log.Infof("Start Plugin: %v\n", p)
		if err := plugin.StartPlugin(p, nil); err != nil {
			log.Errorln(err)
			os.Exit(1)
		}
	}
	for _, p := range epplugins.PluginList {
		if err := plugin.WaitPluginFinished(p); err != nil {
			log.Errorln(err)
			os.Exit(1)
		}
		log.Infof("Plugin Finished: %v\n", p)
	}
}
