/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package main

import (
	"github.com/intel/edge-conductor/cmd/ep/app"
	_ "github.com/intel/edge-conductor/pkg/epplugins"
	"github.com/intel/edge-conductor/pkg/eputils"

	log "github.com/sirupsen/logrus"
)

func main() {
	if err := eputils.CheckHash(""); err != nil && err.Error() != eputils.ERRORCODE_CHECK_HASH_FAIL {
		log.Errorln("Failed to check file hash:", err)
		return
	}

	app.Execute()
}
