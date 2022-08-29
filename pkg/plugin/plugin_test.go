/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package plugin

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

func Test_Fire(t *testing.T) {
	hook := &LogHook{
		writer: func(s string) {
		},
	}
	log.AddHook(hook)
	log.Infof("%s", hook.Levels())
	t.Log("Done")
}
