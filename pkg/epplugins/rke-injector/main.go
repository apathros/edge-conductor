/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
// Template auto-generated once, maintained by plugin owner.

package rkeinjector

import (
	papi "ep/pkg/api/plugins"
	eputils "ep/pkg/eputils"
	repoutils "ep/pkg/eputils/repoutils"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

var (
	errPullFile = errors.New("Pulling file failure!")
	errRunRKE   = errors.New("Failed to run rke command!")
)

func getDefaultRkeSystemImages(str string) []string {
	var imageList []string
	str_arr := strings.Split(str, "\n")
	flag := false
	for _, str := range str_arr {
		if !flag {
			flag = strings.Contains(str, "Generating images list")
		} else if strings.Contains(str, "/") {
			log.Debugf(str)
			imageList = append(imageList, str)
		}
	}
	return imageList
}

func PluginMain(in eputils.SchemaMapData, outp *eputils.SchemaMapData) error {
	input_ep_params := input_ep_params(in)
	input_eptopcfg := input_ep_params.Kitconfig
	input_docker_images := input_docker_images(in)
	input_files := input_files(in)
	output_rkeconfig := output_rkeconfig(outp)
	output_docker_images := output_docker_images(outp)

	rkeCfgSrc := input_eptopcfg.Cluster.Config

	*output_docker_images = *input_docker_images

	// Get default image list for RKE.
	rkeBin := filepath.Join(input_ep_params.Runtimebin, "rke")
	err := repoutils.PullFileFromRepo(rkeBin, input_files.Files[0].Mirrorurl)
	if err != nil {
		log.Errorf("%s", err)
		return errPullFile
	}
	err = os.Chmod(rkeBin, 0700)
	if err != nil {
		return err
	}
	log.Debugf("Get system image list from the RKE binary.")
	cmd := exec.Command(rkeBin, "-d", "config", "--system-images")
	out, err := eputils.RunCMD(cmd)
	if err != nil {
		log.Errorf("%s", err)
		return errRunRKE
	}

	log.Debugf("%s", out)
	log.Infof("Get default RKE system images...")
	imageList := getDefaultRkeSystemImages(out)
	for _, image := range imageList {
		output_docker_images.Images = append(output_docker_images.Images,
			&papi.ImagesItems0{Name: "", URL: image})
	}

	log.Infof("Read cluster config file from %v", rkeCfgSrc)
	content, err := eputils.LoadJsonFile(rkeCfgSrc)
	if err != nil {
		log.Errorf("%s", err)
		return err
	}

	output_rkeconfig.Content = string(content)
	log.Debugf("%v", output_rkeconfig)

	return nil
}
