/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package app

import (
	"fmt"
	epapiplugins "github.com/intel/edge-conductor/pkg/api/plugins"
	eputils "github.com/intel/edge-conductor/pkg/eputils"
	docker "github.com/intel/edge-conductor/pkg/eputils/docker"
	orasutils "github.com/intel/edge-conductor/pkg/eputils/orasutils"
	plugin "github.com/intel/edge-conductor/pkg/plugin"
	wf "github.com/intel/edge-conductor/pkg/workflow"
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

func EpWfStart(epParams *epapiplugins.EpParams, name string) error {
	kitcfg := GetRuntimeTopConfig(epParams)
	if kitcfg == nil {
		return eputils.GetError("errKitConfig")
	}
	address := fmt.Sprintf("%s:%s", kitcfg.Parameters.GlobalSettings.ProviderIP, kitcfg.Parameters.GlobalSettings.WorkflowPort)
	plugin.Address = address

	addonbin := "addon/bin/conductor-plugin"

	var logLevel string
	if log.GetLevel() == log.DebugLevel {
		logLevel = "Debug"
	} else {
		logLevel = "Info"
	}

	if eputils.FileExists(addonbin) {

		finished := make(chan bool)
		go func() {

			cmd := exec.Command(addonbin, address, logLevel)
			_, err := eputils.RunCMDEx(cmd, true)
			if err != nil {
				log.Errorf("Failed to run addon %s on %s", addonbin, address)
			}
			finished <- true
		}()
		err := wf.Start(name, address, WfConfig)
		<-finished
		return err

	} else {
		return wf.Start(name, address, WfConfig)

	}
}

func setHostIptoNoProxy(input_ep_params *epapiplugins.EpParams) error {
	no_proxy := os.Getenv("no_proxy")
	if input_ep_params == nil {
		return eputils.GetError("errHost")
	}

	epHost := input_ep_params.Kitconfig.Parameters.GlobalSettings.ProviderIP
	no_proxy = fmt.Sprintf("%s,%s", no_proxy, epHost)
	if err := os.Setenv("no_proxy", no_proxy); err != nil {
		return err
	}
	return nil
}

func EpUtilsInit(epParams *epapiplugins.EpParams) error {
	regcacert := epParams.Registrycert.Ca.Cert
	auth, err := docker.GetAuthConf(epParams.Kitconfig.Parameters.GlobalSettings.ProviderIP,
		epParams.Kitconfig.Parameters.GlobalSettings.RegistryPort,
		epParams.Kitconfig.Parameters.Customconfig.Registry.User,
		epParams.Kitconfig.Parameters.Customconfig.Registry.Password)
	if err != nil {
		return err
	}

	err = orasutils.OrasNewClient(auth, regcacert)
	if err != nil {
		log.Errorln("Failed to create an OrasClient", err)
		return err
	}
	eputils.SetTemplateParams(epParams)
	eputils.SetTemplateFuncs(funcs)
	return nil
}

func EpWfPreInit(epPms *epapiplugins.EpParams, p map[string]string) (*epapiplugins.EpParams, error) {
	var rfile, kitcfgPath string
	var err error
	epParams := epPms
	if epParams == nil {
		rfile, err = FileNameofRuntime(fnRuntimeInitParams)
		if err != nil {
			log.Errorln("Failed to get runtime file path:", err)
			return nil, eputils.GetError("errParameter")
		}
		if _, err := os.Stat(rfile); os.IsNotExist(err) {
			log.Errorln("Failed to open", rfile, err)
			return nil, eputils.GetError("errParameter")
		}
		epParams = new(epapiplugins.EpParams)
		if err := eputils.LoadSchemaStructFromYamlFile(epParams, rfile); err != nil {
			log.Error(err)
			return nil, err
		}
	}
	epParams.Cmdline = ""
	epParams.Kubeconfig = ""
	for k, v := range p {
		switch k {
		case Epcmdline:
			epParams.Cmdline = v
		case Epkubeconfig:
			epParams.Kubeconfig = v
		case KitConfigPath:
			kitcfgPath = v
		default:
			log.Warnf("%s is not defined in Epparams", k)
		}
	}
	if kitcfgPath == "" {
		if epParams.Kitconfigpath != "" {
			kitcfgPath = epParams.Kitconfigpath
		} else {
			return nil, eputils.GetError("errConfigPath")
		}
	}
	err = setupCustomConfig(kitcfgPath, epParams)
	if err != nil {
		log.Errorln("Failed to setup custom config:", err)
		return nil, err
	}
	if err = setHostIptoNoProxy(epParams); err != nil {
		log.Errorln("Failed to set HostIp to no proxy env", err)
		return nil, err
	}

	if err = EpUtilsInit(epParams); err != nil {
		return nil, err
	}

	return epParams, nil
}

func EpWfTearDown(epParams *epapiplugins.EpParams, rfile string) error {
	teardownCustomConfig(epParams)
	err := eputils.SaveSchemaStructToYamlFile(epParams, rfile)
	if err != nil {
		log.Errorln("Failed to save ep-params:", err)
		return err
	}
	return nil
}

func EpwfLoadServices(epParams *epapiplugins.EpParams) error {
	KitcfgComponentsSelector := &(epParams.Kitconfig.Components.Selector)
	index := len(*KitcfgComponentsSelector)
	*KitcfgComponentsSelector = append((*KitcfgComponentsSelector)[:0], (*KitcfgComponentsSelector)[index:]...)

	err := load_kit_services(epParams.Kitconfig, epParams.Kitconfigpath)
	if err != nil {
		return err
	}

	if KitcfgComponentsSelector == nil {
		log.Warnf("Components Selector not specified, use default value %s", DefaultComponentsSelector)
	}
	return nil
}
