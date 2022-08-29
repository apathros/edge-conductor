/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package eputils

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestEputils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Eputils Suite")
}
