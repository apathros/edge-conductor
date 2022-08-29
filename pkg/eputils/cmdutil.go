/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package eputils

import (
	"fmt"
	"strings"
)

var CMD_SPLIT = "\n"

func AddCmdline(s, t string) string {
	return fmt.Sprintf("%s%s%s", s, CMD_SPLIT, t)
}

func CheckCmdline(cmdline, cmd string) bool {
	aryTmp := strings.Split(cmdline, CMD_SPLIT)
	for _, k := range aryTmp {
		if k == cmd {
			return true
		}
	}
	return false
}
