/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package servicelist

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/kube"

	log "github.com/sirupsen/logrus"
)

var gActionConfig *action.Configuration

// initHelm to initialize helm actionConfig.
func initHelm(kubeconfig, namspace string) error {
	if gActionConfig == nil {
		// Get kubeconfig
		kubeconfig_abs, err := filepath.Abs(kubeconfig)
		if err != nil {
			log.Errorln("ERROR: Failed to find kubeconfig.", err)
			return err
		}
		// Init action configuration
		gActionConfig = new(action.Configuration)
		if err := gActionConfig.Init(
			kube.GetConfig(kubeconfig_abs, "", namspace),
			namspace,
			os.Getenv("HELM_DRIVER"),
			func(format string, v ...interface{}) {
				fmt.Printf(format, v)
			}); err != nil {
			return err
		}
	}
	return nil
}

func getHelmStatus(loc_kubeconfig, name, namespace string) string {
	// Init Helm Configurations
	if err := initHelm(loc_kubeconfig, namespace); err != nil {
		log.Errorln("Failed to init Helm Configuration.", err)
		return "Unknown"
	}

	// New Client
	cli_status := action.NewStatus(gActionConfig)
	res, err := cli_status.Run(name)
	if err != nil {
		if err.Error() == "release: not found" {
			return "Not Deployed"
		}
		return "Unknown"
	} else {
		return strings.Title(res.Info.Status.String())
	}
}
