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
	forceDownload           bool
	clusterKubeConfig       string
	clusterExportKubeConfig string
)

func check_cluster_cmd() error {
	if _, err := os.Stat(clusterKubeConfig); os.IsNotExist(err) {
		return err
	}
	return nil
}

// deployCmd represents deploy command
var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Cluster operations.",
	Long:  `Cluster operations.`,
}

//nolint: dupl
var deployClusterCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy Cluster.",
	Long:  `Deploy Cluster.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Infoln(PROJECTNAME, "- Deploy Cluster")
		log.Infoln("==")
		paramsInject := map[string]string{
			Epkubeconfig: clusterExportKubeConfig,
		}
		epParams, err := EpWfPreInit(nil, paramsInject)
		if err != nil {
			log.Errorln("Failed to init workflow:", err)
			return err
		}
		if err := EpWfStart(epParams, "cluster-deploy"); err != nil {
			log.Errorln("Failed to start workflow:", err)
			return err
		}
		log.Infoln("==")
		log.Infoln("Done")
		return nil
	},
}

//nolint: dupl
var buildClusterCmd = &cobra.Command{
	Use:   "build",
	Short: "Build Cluster.",
	Long:  `Build Cluster.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Infoln(PROJECTNAME, "- Build Cluster")
		log.Infoln("==")

		Epcmd := ""
		var paramsInject map[string]string
		if forceDownload {
			Epcmd = eputils.AddCmdline(Epcmd, "force-download")
		}
		paramsInject = map[string]string{
			Epcmdline: Epcmd,
		}
		epParams, err := EpWfPreInit(nil, paramsInject)
		if err != nil {
			log.Errorln("Failed to init workflow:", err)
			return err
		}

		if err := EpWfStart(epParams, "cluster-build"); err != nil {
			log.Errorln("Failed to start workflow:", err)
			return err
		}

		log.Infoln("==")
		log.Infoln("Done")
		return nil
	},
}

//nolint: dupl
var removeClusterCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove Cluster - only available for KIND clusters.",
	Long:  `Remove Cluster - only available for KIND clusters.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Infoln(PROJECTNAME, "- Remove Cluster")
		log.Infoln("==")

		epParams, err := EpWfPreInit(nil, nil)
		if err != nil {
			log.Errorln("Failed to init workflow:", err)
			return err
		}

		if err := EpWfStart(epParams, "cluster-remove"); err != nil {
			log.Errorln("Failed to start workflow:", err)
			return err
		}

		log.Infoln("==")
		log.Infoln("Done")
		return nil
	},
}

//nolint: dupl
var getClusterInfoCmd = &cobra.Command{
	Use:   "reconcile",
	Short: "Access existing cluster",
	Long:  `Access existing cluster and get target cluster necessary info `,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Infoln(PROJECTNAME, "- Access existing Cluster")
		log.Infoln("==")
		paramsInject := map[string]string{
			Epkubeconfig: clusterExportKubeConfig,
		}
		epParams, err := EpWfPreInit(nil, paramsInject)
		if err != nil {
			log.Errorln("Failed to init workflow:", err)
			return err
		}

		if err := EpWfStart(epParams, "cluster-reconcile"); err != nil {
			log.Errorln("Failed to start workflow:", err)
			return err
		}

		log.Infoln("==")
		log.Infoln("Done")
		return nil
	},
}

//nolint: dupl
var joinClusterCmd = &cobra.Command{
	Use:   "join",
	Short: "Join Node to existing cluster",
	Long:  `Join Node to existing cluster`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Infoln(PROJECTNAME, "- Access existing Cluster")
		log.Infoln("==")

		if err := check_cluster_cmd(); err != nil {
			log.Errorln("Invalid command line:", err)
			return err
		}
		paramsInject := map[string]string{
			Epkubeconfig: clusterKubeConfig,
		}
		epParams, err := EpWfPreInit(nil, paramsInject)
		if err != nil {
			log.Errorln("Failed to init workflow:", err)
			return err
		}

		if err := EpWfStart(epParams, "node-join"); err != nil {
			log.Errorln("Failed to start workflow:", err)
			return err
		}

		log.Infoln("==")
		log.Infoln("Done")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(clusterCmd)

	clusterCmd.AddCommand(deployClusterCmd)
	clusterCmd.AddCommand(removeClusterCmd)
	clusterCmd.AddCommand(buildClusterCmd)
	clusterCmd.AddCommand(getClusterInfoCmd)
	clusterCmd.PersistentFlags().StringVar(&clusterExportKubeConfig, "export-kubeconfig", GetDefaultKubeConfig(), "export kubeconfig file path")
	clusterCmd.PersistentFlags().StringVar(&clusterKubeConfig, "kubeconfig", GetDefaultKubeConfig(), "kubeconfig file path")
	clusterCmd.AddCommand(joinClusterCmd)

	buildClusterCmd.PersistentFlags().BoolVarP(&forceDownload, "force-download", "f", false, "download images with always policy")
}
