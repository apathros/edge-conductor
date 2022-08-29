/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package app

import (
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/undefinedlabs/go-mpatch"
)

func TestCMDExecute(t *testing.T) {
	_, err := mpatch.PatchMethod(cobra.OnInitialize, func(y ...func()) {
	})
	require.NoError(t, err, "PatchMethod Error:")

	_, err = mpatch.PatchMethod(cobra.CheckErr, func(msg interface{}) {
	})
	require.NoError(t, err, "PatchMethod Error:")

	_, err = mpatch.PatchInstanceMethodByName(reflect.TypeOf(rootCmd), "Execute", func(c *cobra.Command) error {
		return nil
	})
	require.NoError(t, err, "PatchMethod Error:")

	Execute()
}
func TestInitConfig(t *testing.T) {

	verbose = true
	InitConfig()
	if log.GetLevel() != log.DebugLevel {
		t.Errorf("TestInitConfig failed, expected Log DebugLevel\r\n")
	}

	verbose = false
	InitConfig()
	if log.GetLevel() != log.InfoLevel {
		t.Errorf("TestInitConfig failed, expected Log InfoLevel\r\n")
	}
}
