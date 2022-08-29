/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package servicelist

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func getYamlStatus(loc_kubeconfig, yaml string) string {
	cmd := exec.Command(
		"./kubectl",
		"describe",
		"-f", yaml)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	cmd.Env = append(os.Environ(),
		fmt.Sprintf("KUBECONFIG=%s", loc_kubeconfig),
	)

	err := cmd.Run()

	if err != nil {
		if strings.Contains(stdout.String()+stderr.String(), "not found") {
			return "Not Deployed"
		}
		return "Unknown"
	}

	return "Deployed"
}
