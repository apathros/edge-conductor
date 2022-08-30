/*
* Copyright (c) 2021 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package app

import (
	"fmt"
	epapiplugins "github.com/intel/edge-conductor/pkg/api/plugins"
	"github.com/intel/edge-conductor/pkg/eputils"
	"os"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"
)

type KitBaseConfig struct {
	Use        []string `yaml:"Use"`
	ParamsMap  map[string]interface{}
	File       string
	Parameters interface{}
}

type KitParams struct {
	Kitconfig *KitBaseConfig
	Workspace string
}

func get_item_yaml(name string, yml string) (string, error) {
	lines := strings.Split(yml, "\n")
	if len(lines) <= 0 {
		return "", eputils.GetError("errYml")
	}
	matched := false
	y := ""
	rs, _ := regexp.Compile(fmt.Sprintf("^%s:", name))
	re, _ := regexp.Compile("^[#\\- \t]")
	for _, l := range lines {
		if match := re.MatchString(l); len(l) > 0 && (!match) && matched {
			break
		}
		if match := rs.MatchString(l); match {
			matched = true
		}
		if matched {
			y = fmt.Sprintf("%s\n%s", y, l)
		}
	}
	return y, nil
}

func load_kit_bcfg(file string) (*KitBaseConfig, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	bcfg := KitBaseConfig{}
	useyml, err := get_item_yaml("Use", string(data))
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal([]byte(useyml), &bcfg)
	if err != nil {
		return nil, err
	}
	paramsyml, err := get_item_yaml("Parameters", string(data))
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal([]byte(paramsyml), &bcfg.ParamsMap)
	if err != nil {
		return nil, err
	}
	bcfg.File = file
	return &bcfg, nil
}

func load_kit_bcfg_recursively(file string) ([]*KitBaseConfig, error) {
	bcfgs := []*KitBaseConfig{}
	bcfg, err := load_kit_bcfg(file)
	if err != nil {
		return nil, err
	}
	for _, uf := range bcfg.Use {
		bcs, err := load_kit_bcfg_recursively(uf)
		if err != nil {
			return nil, err
		}
		bcfgs = append(bcfgs, bcs...)
	}
	bcfgs = append(bcfgs, bcfg)
	return bcfgs, nil
}

func merge_component(kitconfig *epapiplugins.Kitconfig, compMap map[string]interface{}) error {
	name, ok := compMap["name"].(string)
	if !ok {
		return eputils.GetError("errSelector")
	}
	overrideyaml := ""
	override, ok := compMap["override"].(map[string]interface{})
	if ok {
		data, err := yaml.Marshal(override)
		if err != nil {
			log.Warningf("%v override err: %v", name, err)
			return eputils.GetError("errSelectOverride")
		}
		overrideyaml = string(data)
	} else {
		log.Debugf("Cannot find override in %v", name)
	}
	found := false
	for _, v := range kitconfig.Components.Selector {
		if name == v.Name {
			found = true
			if len(v.OverrideYaml) > 0 {
				m := map[string]interface{}{}
				err := yaml.Unmarshal([]byte(v.OverrideYaml), &m)
				if err != nil {
					log.Warningf("%v override yaml unmarshal error: %v", name, err)
					return eputils.GetError("errUnmarshalOver")
				}
				mm := eputils.MergeMaps(m, override)
				data, err := yaml.Marshal(mm)
				if err != nil {
					log.Warningf("%v override merge err: %v", name, err)
					return eputils.GetError("errSelectOverride")
				}
				overrideyaml = string(data)
			}
			v.OverrideYaml = overrideyaml
			break
		}
	}
	if !found {
		kitconfig.Components.Selector = append(kitconfig.Components.Selector,
			&epapiplugins.KitconfigComponentsSelectorItems0{
				Name:         name,
				OverrideYaml: overrideyaml,
			})
	}
	comp := epapiplugins.Component{}
	err := yaml.Unmarshal([]byte(overrideyaml), &comp)
	if err != nil {
		log.Warningf("%v override unmarshal to component error: %v", name, err)
		return eputils.GetError("errUnmarshalOver")
	}
	return nil
}

func load_kit_config(kitconfig *epapiplugins.Kitconfig, kitfile string) error {
	bcfgs, err := load_kit_bcfg_recursively(kitfile)
	if err != nil {
		return err
	}

	kitparams := &KitParams{
		Workspace: GetWorkspacePath(),
		Kitconfig: &KitBaseConfig{},
	}

	params := map[string]interface{}{}
	for _, bcfg := range bcfgs {
		params = eputils.MergeMaps(params, bcfg.ParamsMap)
	}
	if err != nil {
		return err
	}
	kitparams.Kitconfig.Parameters = params["Parameters"]
	*kitconfig = epapiplugins.Kitconfig{
		Parameters: &epapiplugins.KitconfigParameters{
			GlobalSettings: &epapiplugins.KitconfigParametersGlobalSettings{},
		},
		OS:         &epapiplugins.KitconfigOS{},
		Cluster:    &epapiplugins.KitconfigCluster{},
		Components: &epapiplugins.KitconfigComponents{},
	}
	err = eputils.ConvertSchemaStruct(kitparams.Kitconfig.Parameters, kitconfig.Parameters)
	if err != nil {
		return err
	}
	if kitconfig.Parameters == nil {
		kitconfig.Parameters = &epapiplugins.KitconfigParameters{}
	}
	if kitconfig.Parameters.GlobalSettings == nil {
		kitconfig.Parameters.GlobalSettings = &epapiplugins.KitconfigParametersGlobalSettings{}
	}
	if kitconfig.Parameters.GlobalSettings.RegistryPort == "" {
		kitconfig.Parameters.GlobalSettings.RegistryPort = DefaultRegistryPort
	}
	if kitconfig.Parameters.GlobalSettings.ProviderIP == "" {
		kitconfig.Parameters.GlobalSettings.ProviderIP = GetHostDefaultIP()
	}
	if kitconfig.Parameters.GlobalSettings.WorkflowPort == "" {
		kitconfig.Parameters.GlobalSettings.WorkflowPort = DefaultWfPort
	}

	for _, bcfg := range bcfgs {
		log.Debugf("load bcfg from: %v", bcfg.File)
		data, err := os.ReadFile(bcfg.File)
		if err != nil {
			return err
		}
		yml, err := eputils.StringTemplateConvertWithParams(string(data), kitparams)
		if err != nil {
			return err
		}
		if err != nil {
			return err
		}

		cfg := epapiplugins.Kitconfig{}
		json, err := yaml.YAMLToJSON([]byte(yml))
		if err != nil {
			return err
		}
		err = cfg.UnmarshalBinary(json)
		if err != nil {
			return err
		}
		if cfg.OS != nil && len(cfg.OS.Provider) > 0 {
			kitconfig.OS = cfg.OS
		}
		if cfg.Cluster != nil && len(cfg.Cluster.Provider) > 0 {
			kitconfig.Cluster = cfg.Cluster
		}

		if cfg.Components == nil {
			cfg.Components = &epapiplugins.KitconfigComponents{}
		}
		for _, m := range cfg.Components.Manifests {
			found := false
			for _, mm := range kitconfig.Components.Manifests {
				if m == mm {
					found = true
					break
				}
			}
			if !found {
				kitconfig.Components.Manifests = append(kitconfig.Components.Manifests, m)
			}
		}
		if err := load_kit_services(kitconfig, kitfile); err != nil {
			return err
		}
	}
	if kitconfig.Cluster == nil {
		kitconfig.Cluster = &epapiplugins.KitconfigCluster{}
	}
	if kitconfig.Cluster.Provider == "" {
		kitconfig.Cluster.Provider = DefaultClusterProvider
	}
	if kitconfig.Cluster.Provider != DefaultClusterProvider && kitconfig.Cluster.Config == "" {
		kitconfig.Cluster.Config = DefaultClusterConfig
	}
	if kitconfig.OS == nil {
		kitconfig.OS = &epapiplugins.KitconfigOS{}
	}
	if kitconfig.OS.Provider == "" {
		kitconfig.OS.Provider = DefaultOSProvider
	}
	if kitconfig.OS.Config == "" {
		kitconfig.OS.Config = DefaultOSConfig
	}
	if kitconfig.OS.Distro == "" {
		kitconfig.OS.Distro = DefaultOSDistro
	}
	if kitconfig.Parameters.DefaultSSHKeyPath == "" && kitconfig.OS.Provider == "esp" {
		return eputils.GetError("errSSHPath")
	}
	eputils.DumpVar(kitconfig)
	if kitconfig.Validate(nil) != nil {
		log.Warningf("Verify kitconfig err: %v", err)
		return eputils.GetError("errKitConfig")
	}

	return nil
}

func load_kit_services(kitconfig *epapiplugins.Kitconfig, kitfile string) error {
	bcfgs, err := load_kit_bcfg_recursively(kitfile)
	if err != nil {
		return err
	}
	kitparams := &KitParams{
		Workspace: GetWorkspacePath(),
		Kitconfig: &KitBaseConfig{},
	}
	for _, bcfg := range bcfgs {
		data, err := os.ReadFile(bcfg.File)
		if err != nil {
			return err
		}
		yml, err := eputils.StringTemplateConvertWithParams(string(data), kitparams)
		if err != nil {
			return err
		}
		componentYml, err := get_item_yaml("Components", yml)
		if err != nil {
			return err
		}
		componentMap := map[string]interface{}{}
		err = yaml.Unmarshal([]byte(componentYml), &componentMap)
		if err != nil {
			return err
		}

		if _, ok := componentMap["Components"]; ok {
			for k, v := range componentMap["Components"].(map[string]interface{}) {
				if k != "selector" {
					continue
				}
				if _, ok := v.([]interface{}); !ok {
					log.Warningf("Component selector error, file: %v", bcfg.File)
					return eputils.GetError("errCompSelect")
				}
				for _, sv := range v.([]interface{}) {
					if _, ok := sv.(map[string]interface{}); !ok {
						log.Warningf("Component selector error, file: %v", bcfg.File)
						return eputils.GetError("errCompSelect")
					}
					err := merge_component(kitconfig, sv.(map[string]interface{}))
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
