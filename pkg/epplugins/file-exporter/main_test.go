/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Template auto-generated once, maintained by plugin owner.

package fileexporter

import (
	"ep/pkg/eputils"
	"errors"
	"fmt"
	"testing"

	mpatch "github.com/undefinedlabs/go-mpatch"
)

var (
	errCreateFolder = errors.New("Failed to Create Folder")
	errWriteFile    = errors.New("Failed to Write File")
)

func unpatch(t *testing.T, m *mpatch.Patch) {
	err := m.Unpatch()
	if err != nil {
		t.Fatal(err)
	}
}

func TestPluginMainSuccess(t *testing.T) {
	guardcreate, err := mpatch.PatchMethod(eputils.CreateFolderIfNotExist,
		func(a string) error {
			return nil
		})
	if err != nil {
		t.Fatal(err)
	}
	defer unpatch(t, guardcreate)
	guardwrite, err := mpatch.PatchMethod(eputils.WriteStringToFile,
		func(a, b string) error {
			return nil
		})
	if err != nil {
		t.Fatal(err)
	}
	defer unpatch(t, guardwrite)

	input := generateInput(map[string][]byte{
		"exportcontent": []byte(`{"content":"testdata"}`),
		"exportpath":    []byte(`{"path":"testpath"}`),
	})
	inputEmpty := generateInput(map[string][]byte{
		"exportcontent": []byte(`{"content":"testdata"}`),
		"exportpath":    []byte(`{"path":""}`),
	})
	testOutput := generateOutput(nil)

	if err := PluginMain(input, &testOutput); err != nil {
		t.Error("Unexpected error.")
	}
	if err := PluginMain(inputEmpty, &testOutput); err != nil {
		t.Error("Unexpected error.")
	}
}

func TestPluginMainErrorFolder(t *testing.T) {
	input := generateInput(map[string][]byte{
		"exportcontent": []byte(`{"content":"testdata"}`),
		"exportpath":    []byte(`{"path":"/fake/testpath"}`),
	})
	testOutput := generateOutput(nil)

	// Failed to Create Folder
	guardcreate, err := mpatch.PatchMethod(eputils.CreateFolderIfNotExist,
		func(a string) error {
			return errCreateFolder
		})
	if err != nil {
		t.Fatal(err)
	}
	defer unpatch(t, guardcreate)

	// Error Case
	if err := PluginMain(input, &testOutput); err == nil {
		t.Error("Expected error not returned.")
	} else {
		if fmt.Sprint(err) == "Failed to Create Folder" {
			t.Log("Output expected.")
		} else {
			t.Error("Found Unexpected Error:", err)
		}
	}
}

func TestPluginMainErrorWrite(t *testing.T) {
	input := generateInput(map[string][]byte{
		"exportcontent": []byte(`{"content":"testdata"}`),
		"exportpath":    []byte(`{"path":"/fake/testpath"}`),
	})
	testOutput := generateOutput(nil)

	// Failed to Write File
	guardwrite, err := mpatch.PatchMethod(eputils.WriteStringToFile,
		func(a, b string) error {
			return errWriteFile
		})
	if err != nil {
		t.Fatal(err)
	}
	defer unpatch(t, guardwrite)
	// Create Folder
	guardcreate, err := mpatch.PatchMethod(eputils.CreateFolderIfNotExist,
		func(a string) error {
			return nil
		})
	if err != nil {
		t.Fatal(err)
	}
	defer unpatch(t, guardcreate)

	// Error Case
	if err := PluginMain(input, &testOutput); err == nil {
		t.Error("Expected error not returned.")
	} else {
		if fmt.Sprint(err) == "Failed to Write File" {
			t.Log("Output expected.")
		} else {
			t.Error("Found Unexpected Error:", err)
		}
	}
}
