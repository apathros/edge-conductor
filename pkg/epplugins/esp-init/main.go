/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Template auto-generated once, maintained by plugin owner.

package espinit

import (
	"bufio"
	"bytes"
	"fmt"
	eputils "github.com/intel/edge-conductor/pkg/eputils"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	maxStaicIPNum = 1000
	dnsmasqcnf    = "template/dnsmasq/dnsmasq.conf"
)

func appendStringToFile(content, filename string) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Errorln("Failed to open", filename)
		return err
	}
	defer f.Close()

	_, err = f.WriteString(content)
	if err != nil {
		log.Errorln("Failed to write", filename)
		return err
	}

	return nil
}

func createCleanDir(dirname string) error {
	if eputils.FileExists(dirname) {
		if err := os.RemoveAll(dirname); err != nil {
			log.Errorln("Failed to remove", dirname, err)
			log.Errorln("It is possible that old ESP codebase is not cleaned up or still using by ESP containers.")
			log.Errorln("Please run 'os-deploy cleanup' to remove existing ESP codebase first.")
			return err
		}
	}
	if err := eputils.MakeDir(dirname); err != nil {
		log.Errorln("Failed to create", dirname, err)
		return err
	}
	return nil
}

func checkESPCodebase(codebase string) error {
	buildscript := filepath.Join(codebase, "build.sh")
	runscript := filepath.Join(codebase, "run.sh")

	if !eputils.FileExists(codebase) {
		log.Errorln("Failed to prepare codebase", codebase)
		return eputils.GetError("errESPCodebase")
	}
	if !eputils.FileExists(buildscript) {
		log.Errorf("ESP build.sh script not found under %s", codebase)
		return eputils.GetError("errESPBuild")
	}
	if !eputils.FileExists(runscript) {
		log.Errorf("ESP run.sh script not found under %s", codebase)
		return eputils.GetError("errESPRun")
	}
	log.Infoln("ESP codebase ready at", codebase)
	return nil
}

func initESPcode(local_codebase, local_esp_dir, local_esp_tarball, upstream_rel_tarball, upstream_rel_tarball_SHA256 string) error {
	// Download ESP Release Tarball and Prepare Codebase
	if err := createCleanDir(local_esp_dir); err != nil {
		log.Errorf("Failed to create ESP folder")
		return err
	}
	if err := eputils.DownloadFile(local_esp_tarball, upstream_rel_tarball); err != nil {
		log.Errorln("Failed to download", upstream_rel_tarball, err)
		return eputils.GetError("errDownload")
	} else {
		if err := eputils.CheckFileSHA256(local_esp_tarball, upstream_rel_tarball_SHA256); err != nil {
			log.Errorln("Download failed: SHA256 check failed for", upstream_rel_tarball)
			if err := eputils.RemoveFile(local_esp_tarball); err != nil {
				log.Errorln(err)
				return err
			}
			return err
		} else {
			log.Infoln("SHA256 check PASSED for", local_esp_tarball)
		}
		log.Infoln("Successfully downloaded", local_esp_tarball)
	}
	if err := eputils.UncompressTgz(local_esp_tarball, local_esp_dir); err != nil {
		log.Errorln("Failed to extract", local_esp_tarball, err)
		return err
	}

	log.Infoln("Successfully uncompressed", local_esp_tarball)
	return nil
}

// deleteStrfromFile: delete multiple strings from the target file.
// Todo: will do optimization for big size file if necessary.
// Parameters:
//   filename:     File to be operated.
//   keywords:  key words of multiple strings.
//
func deleteStrfromFile(filename string, keywords []string) error {
	if eputils.FileExists(filename) {
		validfile := eputils.IsValidFile(filename)
		if !validfile {
			return eputils.GetError("errInvalidFile")
		}
	}

	file, err := os.Open(filename)
	if err != nil {
		log.Errorf("Fail to write file %s, %v", filename, err)
		return err
	}

	defer file.Close()

	ReadFileStr := ""
	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		for _, keyword := range keywords {
			if !(strings.Contains(fileScanner.Text(), keyword)) {
				ReadFileStr = ReadFileStr + fileScanner.Text() + "\n"
			}
		}
	}

	err = eputils.WriteStringToFile(ReadFileStr, filename)
	if err != nil {
		log.Errorf("Fail to write file %s, %v", filename, err)
		return err
	}

	return nil
}

func PluginMain(in eputils.SchemaMapData, outp *eputils.SchemaMapData) error {
	input_ep_params := input_ep_params(in)
	input_os_provider_manifest := input_os_provider_manifest(in)

	log.Infof("Plugin: esp-init")

	if input_ep_params.Kitconfig.OS == nil {
		log.Errorf("No OS session in top config.")
		return eputils.GetError("errOSSession")
	}

	if input_ep_params.Kitconfig.OS.Provider != "esp" {
		log.Errorf("Wrong OS Provider: %s", input_ep_params.Kitconfig.OS.Provider)
		return eputils.GetError("errOSProvider")
	}

	if input_os_provider_manifest.Esp == nil {
		log.Errorln("Failed to find manifest for ESP os deployer.")
		return eputils.GetError("errESPManifest")
	}

	// Get ESP Config from customcfg.yml
	userconfig := input_ep_params.Kitconfig.OS.Config
	// Get ESP Release Tarball from OS manifest
	rel_tarball := input_os_provider_manifest.Esp.RelURL
	rel_tarball_SHA256 := input_os_provider_manifest.Esp.RelSha256
	rel_version := input_os_provider_manifest.Esp.RelVersion
	// Workspace Dir for ESP
	esp_dir := filepath.Join(input_ep_params.Workspace, "esp")
	esp_tarball := filepath.Join(esp_dir, fmt.Sprintf("esp-%s.tgz", rel_version))
	codebase := filepath.Join(esp_dir, fmt.Sprintf("Edge-Software-Provisioner-%s", rel_version))
	codebase_new := filepath.Join(esp_dir, "esp")
	codeInitFlag := false

	if !eputils.FileExists(esp_tarball) {
		codeInitFlag = true
	} else {
		if err := eputils.CheckFileSHA256(esp_tarball, rel_tarball_SHA256); err != nil {
			log.Infoln("SHA256 value changed for", esp_tarball)
			codeInitFlag = true
		} else {
			log.Infoln("code tarball already exists", esp_tarball)
			codebase = codebase_new
		}
	}

	if codeInitFlag {
		if err := initESPcode(codebase, esp_dir, esp_tarball, rel_tarball, rel_tarball_SHA256); err != nil {
			return err
		}
	}

	// Check if codebase is ready
	if err := checkESPCodebase(codebase); err != nil {
		return err
	}

	espconfig := filepath.Join(codebase, "conf", "config.yml")
	if err := os.RemoveAll(espconfig); err != nil {
		log.Errorf("Failed to remove old ESP config file.")
		return err
	}

	espDNSmasqConfig := filepath.Join(codebase, dnsmasqcnf)
	userconfigContent, err := eputils.LoadJsonFile(userconfig)
	if err != nil {
		log.Errorf("%v", err)
		return err
	}

	if err := eputils.WriteStringToFile(string(userconfigContent), espconfig); err != nil {
		log.Errorf("Fail to write file %s , %v", espconfig, err)
		return err
	}

	log.Infoln("User defined config ready at", espconfig)

	espCfgViper := viper.New()
	espCfgViper.SetConfigType("yaml")
	if err := espCfgViper.ReadConfig(bytes.NewBuffer(userconfigContent)); err != nil {
		log.Errorf("Could not read config file from viper! %v", err)
		return err
	}

	StaticIPValuestring := ""
	for i := 0; i < maxStaicIPNum; i++ {
		viperStaticIPKey := "dhcp-host" + fmt.Sprint(i)
		viperStaticIPValue := espCfgViper.Get(viperStaticIPKey)
		if viperStaticIPValue != nil {
			StaticIPValuestring = StaticIPValuestring + "dhcp-host" + "=" + viperStaticIPValue.(string) + "\n"
		} else {
			break
		}
	}

	keywordsArray := []string{
		"dhcp-host=",
	}

	if err := deleteStrfromFile(espDNSmasqConfig, keywordsArray); err != nil {
		log.Errorf("Failed to write registry config %v", err)
		return err
	}

	if err := appendStringToFile(StaticIPValuestring, espDNSmasqConfig); err != nil {
		return err
	}
	if codebase != codebase_new {
		// Add code to launch esp
		log.Infof("move code base to new code base")
		cmd := exec.Command("mv", codebase, codebase_new)
		if _, err := eputils.RunCMD(cmd); err != nil {
			log.Errorf("Failed to move code base.")
			return err
		}
	}

	log.Infoln("ESP ready to go!!", codebase_new)

	return nil
}
