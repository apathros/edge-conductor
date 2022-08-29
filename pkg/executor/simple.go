/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package executor

//go:generate mockgen -destination=./mock/simple_mock.go -package=mock -copyright_file=../../api/schemas/license-header.txt ep/pkg/executor ExecutorWrapper

import (
	"context"
	pluginapi "ep/pkg/api/plugins"
)

type ExecutorWrapper interface {
	SimpleShell(s *pluginapi.ExecSimpleShell, epparams *pluginapi.EpParams) error
	Run(specFile string, epparams *pluginapi.EpParams, value interface{}) error
}

func SimpleShell(s *pluginapi.ExecSimpleShell, epparams *pluginapi.EpParams) error {
	e := New()
	err := e.SetECParams(epparams)
	if err != nil {
		return err
	}
	err = e.SetTempValue(s.Value)
	if err != nil {
		return err
	}
	err = e.LoadSpecFromFile(s.Spec)
	if err != nil {
		return err
	}
	err = e.Run(context.Background())
	return err
}

func Run(specFile string, epparams *pluginapi.EpParams, value interface{}) error {
	e := New()
	err := e.SetECParams(epparams)
	if err != nil {
		return err
	}
	err = e.SetTempValue(value)
	if err != nil {
		return err
	}
	err = e.LoadSpecFromFile(specFile)
	if err != nil {
		return err
	}
	err = e.Run(context.Background())
	return err
}
