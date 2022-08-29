/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
// Template auto-generated once, maintained by plugin owner.

package servicelist

import (
	eputils "ep/pkg/eputils"
	repoutils "ep/pkg/eputils/repoutils"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"text/tabwriter"
)

func PluginMain(in eputils.SchemaMapData, outp *eputils.SchemaMapData) error {
	input_ep_params := input_ep_params(in)
	input_serviceconfig := input_serviceconfig(in)

	runtime_kubeconfig := input_ep_params.Kubeconfig

	tmpFolder := filepath.Join(input_ep_params.Runtimedir, "tmp")
	defer func() {
		err := os.RemoveAll(tmpFolder)
		if err != nil {
			log.Errorln("failed to remove", tmpFolder, err)
		}
	}()

	const padding = 3

	w := tabwriter.NewWriter(
		os.Stdout,
		0, 0, padding, ' ',
		tabwriter.FilterHTML)

	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "\tNAME\tTYPE\tSTATUS\tURL\tOVERRIDE\t")
	fmt.Fprintln(w, "\t====\t====\t======\t===\t========\t")
	for _, service := range input_serviceconfig.Components {
		if service.Type == "helm" {
			fmt.Fprintf(w, "\t%s\t%s\t%s\t%s\t%s\t\n",
				service.Name,
				service.Type,
				getHelmStatus(runtime_kubeconfig, service.Name, service.Namespace),
				service.URL,
				service.Chartoverride)
		} else if service.Type == "yaml" {
			targetFile := filepath.Join(tmpFolder, service.Name+".yml")
			if err := repoutils.PullFileFromRepo(targetFile, service.URL); err != nil {
				log.Errorln("Failed to pull file", service.URL)
				return err
			}

			fmt.Fprintf(w, "\t%s\t%s\t%s\t%s\t%s\t\n",
				service.Name,
				service.Type,
				getYamlStatus(runtime_kubeconfig, targetFile),
				service.URL,
				"N/A")
		}
	}
	fmt.Fprintln(w, "")
	if err := w.Flush(); err != nil {
		return err
	}

	return nil
}
