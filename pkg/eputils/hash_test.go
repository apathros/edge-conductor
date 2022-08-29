/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package eputils

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestCheckHash(t *testing.T) {
	_, cf, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("Failed to get current test file.")
	}
	cwd := filepath.Join(filepath.Dir(cf), "..", "..")

	if err := CheckHash(cwd); err != nil {
		t.Fatalf("Failed to CheckHash.")
	}
	if err := loadHashCode("nothing"); err == nil {
		t.Fatalf("Expect error but found nothing.")
	}

	fn, _ := getFileRealPath("test1", "cwd")
	if fn != filepath.Join("cwd", "config", "test1") {
		t.Fatalf("Wrong file name generated %s", fn)
	}
	fn, _ = getFileRealPath("workflow", "cwd")
	if fn != filepath.Join("cwd", "workflow") {
		t.Fatalf("Wrong file name generated %s", fn)
	}
}
