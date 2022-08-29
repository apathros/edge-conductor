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
	"os"
	"path/filepath"
)

var (
	purge bool
)

func ep_deinit() error {
	log.Infoln("Deinit", PROJECTNAME)
	log.Infoln("==")
	var paramsInject map[string]string
	if purge {
		paramsInject = map[string]string{
			Epcmdline: "purge",
		}
	} else {
		paramsInject = map[string]string{
			Epcmdline: "",
		}
	}

	epParams, err := EpWfPreInit(nil, paramsInject)
	if err != nil {
		log.Errorln("Failed to init workflow:", err)
		return err
	}

	if err := EpWfStart(epParams, "deinit"); err != nil {
		log.Errorln("Failed to start workflow:", err)
		return err
	}

	log.Infoln("Remove runtime data files.")
	if err := os.RemoveAll(runtimeDataDir); err != nil {
		log.Errorln("Failed to remove runtime folder:", err)
		return err
	}

	if purge {
		log.Infoln("Remove runtime folder.")
		if err := os.RemoveAll(epParams.Runtimedir); err != nil {
			log.Errorln("Failed to remove runtime folder:", err)
			return err
		}

		certdir := filepath.Join(epParams.Workspace, "cert")
		if err := os.RemoveAll(certdir); err != nil {
			log.Errorln("Failed to remove cert folder:", err)
			return err
		}
	}
	log.Infoln("==")
	log.Infoln("Done")
	return nil
}

// initCmd represents init command
var deinitCmd = &cobra.Command{
	Use:   "deinit",
	Short: "Deinit",
	Long:  `Deinit configurations and base services.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ep_deinit(); err != nil {
			log.Errorln(err)
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deinitCmd)
	deinitCmd.PersistentFlags().BoolVarP(&purge, "purge", "p", false, "clean up runtime configurations and container volume.")
}
