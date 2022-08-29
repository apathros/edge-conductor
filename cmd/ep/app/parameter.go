/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package app

import (
	epapiplugins "ep/pkg/api/plugins"
	wfapi "ep/pkg/api/workflow"
	"ep/pkg/eputils"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"text/template"

	"encoding/base64"

	"github.com/foomo/htpasswd"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"
)

var extPaths = [4]string{
	"config/extensions/",
	"config/extensions/config.d/",
	"addon/config/extensions/",
	"addon/config/extensions/config.d/",
}

var funcs = template.FuncMap{
	"readfile": func(f string) (string, error) {
		b, err := ioutil.ReadFile(f)
		if err != nil {
			return "", nil
		}
		return eputils.StringTemplateConvert(string(b))
	},
	"mergeconfig": func(p interface{}) (string, error) {
		//TODO: merge multiple config files.

		files, ok := p.([]string)
		if !ok {
			return "", nil
		}
		configfile := ""
		if len(files) > 0 {
			configfile = files[0]
		} else {
			return "", nil
		}

		b, err := ioutil.ReadFile(configfile)
		if err != nil {
			return "", nil
		}
		return eputils.StringTemplateConvert(string(b))

	},
	"structtoyaml": func(p interface{}) (string, error) {
		ymlout, err := yaml.Marshal(p)
		if err != nil {
			return "", nil
		}
		return string(ymlout), nil
	},
	"htpasswd": func(usr string, pwd string) (string, error) {
		tmpFile, err := ioutil.TempFile(os.TempDir(), "")
		if err != nil {
			log.Error("Cannot create temporary file", err)
			return "", nil
		}
		defer os.Remove(tmpFile.Name())

		err = htpasswd.SetPassword(tmpFile.Name(), usr, pwd, htpasswd.HashBCrypt)
		if err != nil {
			log.Errorln("Set Password Failed:", err)
		}
		b, err := ioutil.ReadFile(tmpFile.Name())
		if err != nil {
			return "", nil
		}

		if err = tmpFile.Close(); err != nil {
			log.Fatal(err)
		}

		return string(b), nil
	},
	"base64": func(value string) (string, error) {
		data := base64.StdEncoding.EncodeToString([]byte(value))
		return data, nil
	},
	"include_data":       wf_include_data,
	"include_workflows":  wf_include_workflows,
	"include_plugins":    wf_include_plugins,
	"include_containers": wf_include_containers,
}

func get_files(filelocation string) ([]string, error) {
	var files []string

	if !eputils.FileExists(filelocation) {
		return files, nil
	}

	if eputils.IsDirectory(filelocation) {
		if err := filepath.Walk(filelocation,
			func(fpath string, finfo os.FileInfo, err error) error {
				if err == nil && !finfo.IsDir() && strings.ToLower(filepath.Ext(fpath)) == ".yml" {
					files = append(files, fpath)
				}
				return err
			}); err != nil {
			return files, err
		}
	} else {
		files = append(files, filelocation)
	}

	return files, nil
}

func schemastruct_to_yaml_item(v eputils.SchemaStruct) (string, error) {
	outputstr := ""
	result, err := eputils.SchemaStructToYaml(v)
	if err != nil {
		return "", err
	}
	lines := strings.Split(result, "\n")
	for id, l := range lines {
		if id == 0 {
			outputstr = "- " + l + "\n"
		} else {
			outputstr = outputstr + "  " + l + "\n"
		}
	}
	return outputstr, nil
}

func wf_include(filelocation, field string) (string, error) {
	outputstr := ""

	files, err := get_files(filelocation)
	if err != nil {
		return "", err
	}
	log.Debugf("Files: %s", files)

	for _, fn := range files {
		wf := wfapi.Workflow{}
		if err := eputils.LoadSchemaStructFromYamlFile(&wf, fn); err != nil {
			return "", err
		}
		log.Debugf("Load from File: %s", fn)
		if wf.Spec == nil {
			continue
		}
		if field == "workflows" {
			for _, item := range wf.Spec.Workflows {
				result, err := schemastruct_to_yaml_item(item)
				if err != nil {
					return "", err
				}
				log.Debugf("Load workflow:\n%s", result)
				outputstr = outputstr + result
			}

		} else if field == "data" {
			for _, item := range wf.Spec.Data {
				result, err := schemastruct_to_yaml_item(item)
				if err != nil {
					return "", err
				}
				log.Debugf("Load data:\n%s", result)
				outputstr = outputstr + result
			}

		} else if field == "containers" {
			for _, item := range wf.Spec.Containers {
				result, err := schemastruct_to_yaml_item(item)
				if err != nil {
					return "", err
				}
				log.Debugf("Load container:\n%s", result)
				outputstr = outputstr + result
			}

		} else if field == "plugins" {
			for _, item := range wf.Spec.Plugins {
				result, err := schemastruct_to_yaml_item(item)
				if err != nil {
					return "", err
				}
				log.Debugf("Load plugin:\n%s", result)
				outputstr = outputstr + result
			}

		}
	}
	return outputstr, nil
}

func wf_include_data(filelocation string) (string, error) {
	return wf_include(filelocation, "data")
}

func wf_include_workflows(filelocation string) (string, error) {
	return wf_include(filelocation, "workflows")
}

func wf_include_plugins(filelocation string) (string, error) {
	return wf_include(filelocation, "plugins")
}

func wf_include_containers(filelocation string) (string, error) {
	return wf_include(filelocation, "containers")
}

func setupCustomConfig(kitcfgPath string, epp *epapiplugins.EpParams) error {
	var kitcfg epapiplugins.Kitconfig
	if epp == nil {
		return eputils.GetError("errParameter")
	} else if epp.Kitconfig == nil {
		return eputils.GetError("errKitConfig")
	}
	if kitcfgPath == "" {
		if epp.Kitconfigpath == "" {
			return eputils.GetError("errConfigPath")
		} else {
			kitcfgPath = epp.Kitconfigpath
		}
	} else {
		epp.Kitconfigpath = kitcfgPath
	}
	if _, err := os.Stat(kitcfgPath); os.IsNotExist(err) {
		log.Error("Failed to open", kitcfgPath, err)
		return err
	}
	if err := load_kit_config(&kitcfg, kitcfgPath); err != nil {
		log.Error(err)
		return err
	}

	// Load node list from user config file.
	epp.Kitconfig.Parameters.Nodes = kitcfg.Parameters.Nodes

	// Load Customconfig from user config file.
	epp.Kitconfig.Parameters.Customconfig = kitcfg.Parameters.Customconfig
	ctmcfg := epp.Kitconfig.Parameters.Customconfig
	needUserPassword := false
	if ctmcfg == nil {
		needUserPassword = true
	} else {
		if ctmcfg.Registry != nil {
			if ctmcfg.Registry.Externalurl == "" {
				// Harbor bootstrap user is admin
				if ctmcfg.Registry.User != "admin" {
					needUserPassword = true
				}
				if ctmcfg.Registry.Password == "" {
					needUserPassword = true
				}
			}
		} else {
			needUserPassword = true
		}
	}
	if needUserPassword {
		log.Error("Please provide a registry password with admin user via custom config file, check doc for more detail.")
		return eputils.GetError("errRegistryPw")
	}
	return nil
}

func teardownCustomConfig(epp *epapiplugins.EpParams) {
	epp.Kitconfig.Parameters.Customconfig = &epapiplugins.Customconfig{}
	epp.Kitconfig.Parameters.Nodes = []*epapiplugins.Node{}
}

func InitEpParams(topcfg epapiplugins.Kitconfig) *epapiplugins.EpParams {
	currentUserName := "unknown"
	u, err := user.Current()
	if err != nil {
		log.Error(err)
		return nil
	} else if u != nil {
		currentUserName = u.Username
	}
	epparams := &epapiplugins.EpParams{
		User:         currentUserName,
		Workspace:    GetWorkspacePath(),
		Runtimedir:   GetRuntimeFolder(),
		Runtimebin:   filepath.Join(GetRuntimeFolder(), "bin"),
		Runtimedata:  filepath.Join(GetRuntimeFolder(), "data"),
		Registrycert: &epapiplugins.Certificate{},
	}
	epparams.Kitconfig = &topcfg
	if err := eputils.ConvertSchemaStruct(&registrycerts, epparams.Registrycert); err != nil {
		log.Error(err)
		return nil
	}
	log.Infof("ext cfg is %+v", topcfg.Parameters.Extensions)
	for _, ext := range topcfg.Parameters.Extensions {
		var extcfg epapiplugins.Extension
		var extPath string
		log.Info("extension in kit is ", ext)
		foundExt := false
		// first found , first use
		for _, path := range extPaths {
			extPath = path + ext + ".yml"
			if _, err := os.Stat(extPath); err == nil {
				log.Debugf("Found extPath: %s ", extPath)
				foundExt = true
				break
			} else {
				log.Debugf("Extension is not existing in %s ", extPath)
				continue
			}
		}
		if !foundExt {
			log.Errorf("Can not find extension: %s", ext)
			return nil
		}
		if err := eputils.LoadSchemaStructFromYamlFile(&extcfg, extPath); err != nil {
			log.Error(err)
			return nil
		}
		log.Infof("ext cfg is %+v", extcfg)
		epparams_ext := &epapiplugins.EpParamsExtensionsItems0{
			Name:      ext,
			Extension: &extcfg,
		}
		epparams.Extensions = append(epparams.Extensions, epparams_ext)
	}
	// Init runtime folders
	if err := MakeDir(epparams.Runtimebin); err != nil {
		return nil
	}
	if err := MakeDir(epparams.Runtimedata); err != nil {
		return nil
	}

	return epparams
}
