/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package eputils

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/undefinedlabs/go-mpatch"
)

var (
	errFileDir = fmt.Errorf("no such file or directory")
)

func patchIOCopyWithDummyData(t *testing.T, srcStr string) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(io.Copy, func(w io.Writer, _ io.Reader) (int64, error) {
		unpatch(t, patch)
		return io.Copy(w, strings.NewReader(srcStr))
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func TestCheckContentSHA256(t *testing.T) {
	cases := []struct {
		name           string
		expectedError  error
		sha256expected string
		genFileContent func() []byte
	}{
		{
			name:           "CheckContentSHA256 unexpected sha256 string",
			expectedError:  GetError("errShaCheckFailed"),
			sha256expected: "0283da60063abfb3a87f1aed845d17fe2d9ba8c780b478dc4ae048f5ee97a6d5", // abcde\n sha256 string
			genFileContent: func() []byte {
				return []byte("abcd")
			},
		},
		{
			name:           "CheckContentSHA256 ok",
			expectedError:  nil,
			sha256expected: "0283da60063abfb3a87f1aed845d17fe2d9ba8c780b478dc4ae048f5ee97a6d5", // abcde\n sha256 string
			genFileContent: func() []byte {
				return []byte("abcde\n")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			fc := tc.genFileContent()

			err := CheckContentSHA256(fc, tc.sha256expected)
			if !isExpectedError(err, tc.expectedError) {
				t.Errorf("Unexpected error: %v", err)
			}

		})
	}
}

func TestCheckFileDescriptorSHA256(t *testing.T) {
	cases := []struct {
		name              string
		expectedError     error
		sha256expected    string
		funcBeforeTest    func()
		genFileDescriptor func() *os.File
	}{
		{
			name:          "CheckFileDescriptorSHA256 io copy error",
			expectedError: testErr,
			funcBeforeTest: func() {
				patchOsOpen(t, &os.File{}, nil)
				patchIoCopy(t, 0, testErr)
			},
			genFileDescriptor: func() *os.File {
				f, _ := os.Open("")
				return f
			},
		},
		{
			name:           "CheckFileDescriptor unexpected sha256 string",
			expectedError:  GetError("errShaCheckFailed"),
			sha256expected: "0283da60063abfb3a87f1aed845d17fe2d9ba8c780b478dc4ae048f5ee97a6d5", // abcde\n sha256 string
			funcBeforeTest: func() {
				patchOsOpen(t, &os.File{}, nil)
				patchIOCopyWithDummyData(t, "abcd")
			},
			genFileDescriptor: func() *os.File {
				f, _ := os.Open("")
				return f
			},
		},
		{
			name:           "CheckFileDescriptor ok",
			expectedError:  nil,
			sha256expected: "0283da60063abfb3a87f1aed845d17fe2d9ba8c780b478dc4ae048f5ee97a6d5", // abcde\n sha256 string
			funcBeforeTest: func() {
				patchOsOpen(t, &os.File{}, nil)
				patchIOCopyWithDummyData(t, "abcde\n")
			},
			genFileDescriptor: func() *os.File {
				f, _ := os.Open("")
				return f
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				tc.funcBeforeTest()
			}

			fd := tc.genFileDescriptor()

			err := CheckFileDescriptorSHA256(fd, tc.sha256expected)
			if !isExpectedError(err, tc.expectedError) {
				t.Errorf("Unexpected error: %v", err)
			}

			fd.Close()
		})
	}
}

func TestCheckFileSHA256(t *testing.T) {
	cases := []struct {
		name           string
		expectedError  error
		filename       string
		sha256expected string
		funcBeforeTest func()
	}{
		{
			name:          "CheckFileSHA256 open file error",
			expectedError: errFileDir,
		},
		{
			name:          "CheckFileSHA256 io copy error",
			expectedError: testErr,
			funcBeforeTest: func() {
				patchOsOpen(t, &os.File{}, nil)
				patchIoCopy(t, 0, testErr)
			},
		},
		{
			name:           "CheckFileSHA256 unexpected sha256 string",
			expectedError:  GetError("errShaCheckFailed"),
			filename:       "",
			sha256expected: "0283da60063abfb3a87f1aed845d17fe2d9ba8c780b478dc4ae048f5ee97a6d5", // abcde\n sha256 string
			funcBeforeTest: func() {
				patchOsOpen(t, &os.File{}, nil)
				patchIOCopyWithDummyData(t, "abcd")
			},
		},
		{
			name:           "CheckFileSHA256 ok",
			expectedError:  nil,
			filename:       "",
			sha256expected: "0283da60063abfb3a87f1aed845d17fe2d9ba8c780b478dc4ae048f5ee97a6d5", // abcde\n sha256 string
			funcBeforeTest: func() {
				patchOsOpen(t, &os.File{}, nil)
				patchIOCopyWithDummyData(t, "abcde\n")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				tc.funcBeforeTest()
			}

			err := CheckFileSHA256(tc.filename, tc.sha256expected)
			if !isExpectedError(err, tc.expectedError) {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestGenFileSHA256(t *testing.T) {
	cases := []struct {
		name           string
		expectedError  error
		filename       string
		sha256expected string
		funcBeforeTest func()
	}{
		{
			name:          "GenFileSHA256 open file error",
			expectedError: errFileDir,
		},
		{
			name:          "GenFileSHA256 io copy error",
			expectedError: testErr,
			funcBeforeTest: func() {
				patchOsOpen(t, &os.File{}, nil)
				patchIoCopy(t, 0, testErr)
			},
		},
		{
			name:           "GenFileSHA256 ok",
			expectedError:  nil,
			filename:       "",
			sha256expected: "0283da60063abfb3a87f1aed845d17fe2d9ba8c780b478dc4ae048f5ee97a6d5", // abcde\n sha256 string
			funcBeforeTest: func() {
				patchOsOpen(t, &os.File{}, nil)
				patchIOCopyWithDummyData(t, "abcde\n")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				tc.funcBeforeTest()
			}

			sha256Str, err := GenFileSHA256(tc.filename)
			if !isExpectedError(err, tc.expectedError) {
				t.Errorf("Unexpected error: %v", err)
			}
			if sha256Str != tc.sha256expected {
				t.Errorf("Unexpected output: %v, expected: %v", sha256Str, tc.sha256expected)
			}
		})
	}
}
