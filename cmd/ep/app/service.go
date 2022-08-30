/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package app

import (
	"github.com/intel/edge-conductor/pkg/eputils"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	serviceKubeConfig string
)

func check_service_cmd() error {
	if _, err := os.Stat(serviceKubeConfig); os.IsNotExist(err) {
		return err
	}
	return nil
}

// deployCmd represents deploy command
var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Service operations.",
	Long:  `Service operations.`,
}

//nolint: dupl
var buildServiceCmd = &cobra.Command{
	Use:   "build",
	Short: "Build Services.",
	Long:  `Parse service config, download service files and inject service config with local URLs.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Infoln(PROJECTNAME, "- Build Services")
		log.Infoln("==")

		Epcmd := ""
		var paramsInject map[string]string
		if forceDownload {
			Epcmd = eputils.AddCmdline(Epcmd, "force-download")
		}
		paramsInject = map[string]string{
			Epkubeconfig: serviceKubeConfig,
			Epcmdline:    Epcmd,
		}

		epParams, err := EpWfPreInit(nil, paramsInject)
		if err != nil {
			log.Errorln("Failed to init workflow:", err)
			return err
		}
		if err := EpwfLoadServices(epParams); err != nil {
			log.Errorln("Failed to load services:", err)
			return err
		}
		defer func() {
			epparams_runtime_file, err := FileNameofRuntime(fnRuntimeInitParams)
			if err != nil {
				log.Errorln("Failed to get runtime file path:", err)
			}
			err = EpWfTearDown(epParams, epparams_runtime_file)
			if err != nil {
				log.Errorln("Workflow Tear Down Error:", err)
			}
		}()

		if err := EpWfStart(epParams, "service-build"); err != nil {
			log.Errorln("Failed to start workflow:", err)
			return err
		}
		log.Infoln("==")
		log.Infoln("Done")
		return nil
	},
}

//nolint: dupl
var deployServiceCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy Services.",
	Long:  `Deploy Services on K8s cluster according to the service config.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Infoln(PROJECTNAME, "- Deploy Services")
		log.Infoln("==")

		if err := check_service_cmd(); err != nil {
			log.Errorln("Invalid command line:", err)
			return err
		}
		paramsInject := map[string]string{
			Epkubeconfig: serviceKubeConfig,
		}
		epParams, err := EpWfPreInit(nil, paramsInject)
		if err != nil {
			log.Errorln("Failed to init workflow:", err)
			return err
		}

		if err := EpWfStart(epParams, "service-deploy"); err != nil {
			log.Errorln("Failed to start workflow:", err)
			return err
		}

		log.Infoln("==")
		log.Infoln("Done")
		return nil
	},
}

//nolint: dupl
var listServiceCmd = &cobra.Command{
	Use:   "list",
	Short: "List Planned Services.",
	Long:  `Show the list of services planned to be deployed on the cluster.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Infoln(PROJECTNAME, "- List Services")
		log.Infoln("==")

		if err := check_service_cmd(); err != nil {
			log.Errorln("Invalid command line:", err)
			return err
		}
		paramsInject := map[string]string{
			Epkubeconfig: serviceKubeConfig,
		}
		epParams, err := EpWfPreInit(nil, paramsInject)
		if err != nil {
			log.Errorln("Failed to init workflow:", err)
			return err
		}

		if err := EpWfStart(epParams, "service-list"); err != nil {
			log.Errorln("Failed to start workflow:", err)
			return err
		}

		log.Infoln("==")
		log.Infoln("Done")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serviceCmd)
	serviceCmd.AddCommand(buildServiceCmd)
	serviceCmd.AddCommand(deployServiceCmd)
	serviceCmd.AddCommand(listServiceCmd)
	serviceCmd.PersistentFlags().StringVar(&serviceKubeConfig, "kubeconfig", GetDefaultKubeConfig(), "kubeconfig file path")

	buildServiceCmd.PersistentFlags().BoolVarP(&forceDownload, "force-download", "f", false, "download images with always policy")
}
