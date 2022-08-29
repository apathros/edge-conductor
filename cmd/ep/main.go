/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package main

import (
	"ep/cmd/ep/app"
	_ "ep/pkg/epplugins"
	"ep/pkg/eputils"

	log "github.com/sirupsen/logrus"
)

func main() {
	if err := eputils.CheckHash(""); err != nil && err.Error() != eputils.ERRORCODE_CHECK_HASH_FAIL {
		log.Errorln("Failed to check file hash:", err)
		return
	}

	app.Execute()
}
