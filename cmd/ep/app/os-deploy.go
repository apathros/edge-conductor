/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package app

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// deployCmd represents deploy command
var osDeployCmd = &cobra.Command{
	Use:   "os-deploy",
	Short: "OS Deployment operations.",
	Long:  `OS Deployment operations.`,
}

//nolint: dupl
var osDeployBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build OS Deployment Services.",
	Long:  `Download OS deployer and build the service (binary files and docker images).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Infoln(PROJECTNAME, "- Build OS Deployment Services")
		log.Infoln("==")

		epParams, err := EpWfPreInit(nil, nil)
		if err != nil {
			log.Errorln("Failed to init workflow:", err)
			return err
		}

		if err := EpWfStart(epParams, "os-deploy-build"); err != nil {
			log.Errorln("Failed to start workflow:", err)
			return err
		}

		log.Infoln("==")
		log.Infoln("Done")
		return nil
	},
}

//nolint: dupl
var osDeployStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start OS Deployment Services.",
	Long:  `Start OS deployer services`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Infoln(PROJECTNAME, "- Start OS Deployment Services")
		log.Infoln("==")

		epParams, err := EpWfPreInit(nil, nil)
		if err != nil {
			log.Errorln("Failed to init workflow:", err)
			return err
		}

		if err := EpWfStart(epParams, "os-deploy-start"); err != nil {
			log.Errorln("Failed to start workflow:", err)
			return err
		}

		log.Infoln("==")
		log.Infoln("Done")
		return nil
	},
}

//nolint: dupl
var osDeployStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop OS Deployment Services.",
	Long:  `Stop OS deployer services`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Infoln(PROJECTNAME, "- Stop OS Deployment Services")
		log.Infoln("==")

		epParams, err := EpWfPreInit(nil, nil)
		if err != nil {
			log.Errorln("Failed to init workflow:", err)
			return err
		}

		if err := EpWfStart(epParams, "os-deploy-stop"); err != nil {
			log.Errorln("Failed to start workflow:", err)
			return err
		}

		log.Infoln("==")
		log.Infoln("Done")
		return nil
	},
}

//nolint: dupl
var osDeployCleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Cleanup OS Deployment codebase.",
	Long:  `Cleanup OS Deployment codebase.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Infoln(PROJECTNAME, "- Cleanup OS Deployment codebase")
		log.Infoln("==")

		epParams, err := EpWfPreInit(nil, nil)
		if err != nil {
			log.Errorln("Failed to init workflow:", err)
			return err
		}

		if err := EpWfStart(epParams, "os-deploy-cleanup"); err != nil {
			log.Errorln("Failed to start workflow:", err)
			return err
		}

		log.Infoln("==")
		log.Infoln("Done")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(osDeployCmd)
	osDeployCmd.AddCommand(osDeployBuildCmd)
	osDeployCmd.AddCommand(osDeployStartCmd)
	osDeployCmd.AddCommand(osDeployStopCmd)
	osDeployCmd.AddCommand(osDeployCleanupCmd)
}
