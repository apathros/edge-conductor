/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package eputils

import (
	"testing"
)

func TestErrorcode(t *testing.T) {
	var (
		code        = "E001.999"
		errormes    = "Test error code funtion"
		linkaddress = ""
	)
	ecerror := &EC_errors{code, errormes, linkaddress}
	if ecerror == nil {
		t.Fatalf("Create New EC error failed.")
	}
	if ecerror.Error() == "" {
		t.Fatalf("Get EC error failed.")
	}
	if ecerror.Code() == "" {
		t.Fatalf("Get Error code failed.")
	}
	if ecerror.Msg() == "" {
		t.Fatalf("Get Error message failed.")
	}
}
