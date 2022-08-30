/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package eputils

//go:generate mockgen -destination=./mock/fileutil_mock.go -package=mock -copyright_file=../../api/schemas/license-header.txt github.com/intel/edge-conductor/pkg/eputils FileWrapper

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	mpatch "github.com/undefinedlabs/go-mpatch"
)

const (
	testdatapath = "testdata"
	RmFile       = "filetobeRemove.yml"
)

var (
	testErr        = fmt.Errorf("test error")
	errNoFileDir   = errors.New("no such file or directory")
	errIsDir       = errors.New("is a directory")
	errInvalidChar = errors.New("invalid character")
	errLstatTest   = errors.New("lstat testdata")
	errEncHeader   = errors.New("archive/tar: cannot encode header")
)

func patchFileExists(t *testing.T, retValue bool, NumOfTimesCalled int) {
	var patch *mpatch.Patch
	var patchErr error
	var times int
	patch, patchErr = mpatch.PatchMethod(FileExists, func(_ string) bool {
		times++
		if times >= NumOfTimesCalled {
			unpatch(t, patch)
		}

		return retValue
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func patchMakeDir(t *testing.T, err error) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(MakeDir, func(_ string) error {
		unpatch(t, patch)
		return err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func patchReadlink(t *testing.T, name string, err error) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(os.Readlink, func(_ string) (string, error) {
		unpatch(t, patch)
		return name, err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func patchOsStat(t *testing.T, fileInfo fs.FileInfo, err error) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(os.Stat, func(_ string) (fs.FileInfo, error) {
		unpatch(t, patch)
		return fileInfo, err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func patchOsChmod(t *testing.T, err error, nextPatch func()) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(os.Chmod, func(_ string, _ fs.FileMode) error {
		unpatch(t, patch)
		if nextPatch != nil {
			nextPatch()
		}
		return err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func patchOsOpenFile(t *testing.T, file *os.File, err error, nextPatch func()) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(os.OpenFile, func(fileName string, _ int, _ fs.FileMode) (*os.File, error) {
		unpatch(t, patch)
		if nextPatch != nil {
			nextPatch()
		}
		return file, err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func patchOsOpen(t *testing.T, file *os.File, err error) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(os.Open, func(_ string) (*os.File, error) {
		unpatch(t, patch)
		return file, err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func patchOsCreate(t *testing.T, file *os.File, err error) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(os.Create, func(_ string) (*os.File, error) {
		unpatch(t, patch)
		return file, err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func patchHttpGet(t *testing.T, resp *http.Response, err error) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(http.Get, func(_ string) (*http.Response, error) {
		unpatch(t, patch)
		return resp, err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

//nolint:unparam
func patchIoCopy(t *testing.T, written int64, err error) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(io.Copy, func(_ io.Writer, _ io.Reader) (int64, error) {
		unpatch(t, patch)
		return written, err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func patchIsValidFile(t *testing.T, valid bool, patchNext func()) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(IsValidFile, func(_ string) bool {
		unpatch(t, patch)
		if patchNext != nil {
			patchNext()
		}
		return valid
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func patchGzipNewReader(t *testing.T, reader *gzip.Reader, err error) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(gzip.NewReader, func(_ io.Reader) (*gzip.Reader, error) {
		unpatch(t, patch)
		return reader, err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func patchTarNewReader(t *testing.T, reader *tar.Reader) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(tar.NewReader, func(_ io.Reader) *tar.Reader {
		unpatch(t, patch)
		return reader
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

//nolint:unparam
func patchTarReaderNext(t *testing.T, header *tar.Header, err error) {
	var patch *mpatch.Patch
	var patchErr error
	reader := &tar.Reader{}
	patch, patchErr = mpatch.PatchInstanceMethodByName(reflect.TypeOf(reader), "Next", func(_ *tar.Reader) (*tar.Header, error) {
		unpatch(t, patch)
		return header, err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func patchTarFileInfoHeader(t *testing.T, header *tar.Header, err error) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(tar.FileInfoHeader, func(fi fs.FileInfo, link string) (*tar.Header, error) {
		unpatch(t, patch)
		return header, err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func patchFileWriteString(t *testing.T, n int, err error) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchInstanceMethodByName(reflect.TypeOf(&os.File{}), "WriteString", func(_ *os.File, _ string) (int, error) {
		unpatch(t, patch)
		return n, err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

type JsonTest struct {
	Cn    string   `json:"CN"`
	Hosts []string `json:"hosts"`
}

func TestIsValidFile(t *testing.T) {
	cases := []struct {
		name        string
		input       []string
		retError    error
		expectedRet bool
	}{
		{
			"No_filename",
			[]string{""},
			nil,
			false,
		},
		{
			"Valid_filename",
			[]string{filepath.Join(testdatapath, "fileutil1.yml")},
			nil,
			true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsValidFile(tc.input[0])
			t.Logf("filename:%s", tc.input[0])
			if tc.expectedRet == result {
				t.Log("Done")
				return
			} else {
				t.Logf("Test case %s failed.", tc.name)
				t.Error(result)
			}
		})
	}

}

func TestCheckFileLink(t *testing.T) {
	cases := []struct {
		name           string
		input          []string
		expectedRet    Filelink
		funcBeforeTest func()
	}{
		{
			name:        "No_filename",
			input:       []string{""},
			expectedRet: Wrongfile,
		},
		{
			name:        "Normal file",
			input:       []string{filepath.Join(testdatapath, "fileutil1.yml")},
			expectedRet: Normalfile,
		},
		{
			name:        "softlink file",
			input:       []string{filepath.Join(testdatapath, "fileutil_symb.yml")},
			expectedRet: "symbol",
		},
		{
			name:        "Readlink error",
			input:       []string{filepath.Join(testdatapath, "fileutil_symb.yml")},
			expectedRet: Wrongfile,
			funcBeforeTest: func() {
				patchReadlink(t, "", testErr)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				tc.funcBeforeTest()
			}
			t.Logf("filename:%s", tc.input[0])
			if filelink, _ := CheckFileLink(tc.input[0]); filelink == tc.expectedRet {
				t.Log("Done")
				return
			} else {
				t.Logf("Test case %s failed. filelink: %s", tc.name, string(filelink))
				t.Error(false)
			}
		})
	}

}

func TestMakeDir(t *testing.T) {
	cases := []struct {
		name        string
		input       []string
		retError    error
		expectError bool
	}{
		{
			"make existed dir",
			[]string{"testdata/fileutil1.yml"},
			nil,
			true,
		},
		{
			"make dir success",
			[]string{filepath.Join(testdatapath, "fileTestFolder")},
			nil,
			false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := MakeDir(tc.input[0])
			if (result != nil && !tc.expectError) ||
				(result == nil && tc.expectError) {
				t.Logf("Test case %s failed.", tc.name)
				t.Error(result)
			} else {
				t.Log("Done")
			}
		})
	}

}

func TestFileExists(t *testing.T) {
	cases := []struct {
		name        string
		input       []string
		retError    error
		expectError bool
	}{
		{
			"test file not exists",
			[]string{filepath.Join(testdatapath, "not_exist.yml")},
			nil,
			true,
		},
		{
			"test file exists",
			[]string{filepath.Join(testdatapath, "fileutil_test.yml")},
			nil,
			false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if result := FileExists(tc.input[0]); !result {
				if tc.expectError {
					t.Log("Done")
					return
				} else {
					t.Logf("Test case %s failed.", tc.name)
					t.Error(result)
				}
			}
		})
	}

}

func TestIsDirectory(t *testing.T) {
	cases := []struct {
		name        string
		input       []string
		retError    error
		expectedRet bool
	}{
		{
			"Directory not exists",
			[]string{filepath.Join(testdatapath, "fileutil_test")},
			nil,
			false,
		},
		{
			"Not a Directory",
			[]string{filepath.Join(testdatapath, "fileutil_test.yml")},
			nil,
			false,
		},
		{
			"Normal Directory",
			[]string{filepath.Join(testdatapath)},
			nil,
			true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsDirectory(tc.input[0])
			t.Logf("Test input parameter:%s", tc.input[0])
			if tc.expectedRet == result {
				t.Log("Done")
				return
			} else {
				t.Logf("Test case %s failed.", tc.name)
				t.Error(result)
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	cases := []struct {
		name           string
		input          []string
		exceptError    error
		funcBeforeTest func()
	}{
		{
			name:        "src file not exist",
			input:       []string{filepath.Join(testdatapath, ""), filepath.Join(testdatapath, "not_exist.yml")},
			exceptError: errNoFileDir,
		},
		{
			name:        "file not valid",
			input:       []string{filepath.Join(testdatapath, ""), filepath.Join(testdatapath, "fileutil1.yml")},
			exceptError: GetError("errInvalidFile"),
			funcBeforeTest: func() {
				patchIsValidFile(t, false, nil)
			},
		},
		{
			name:        "dst file exists and file type is linkfile",
			input:       []string{filepath.Join(testdatapath, "fileutil_symb.yml"), filepath.Join(testdatapath, "fileutil1.yml")},
			exceptError: GetError("errInvalidFile"),
			funcBeforeTest: func() {
				patchIsValidFile(t, true, func() {
					patchIsValidFile(t, false, nil)
				})
			},
		},
		{
			name:        "os.Open error",
			input:       []string{filepath.Join(testdatapath, "fileutil_symb.yml"), filepath.Join(testdatapath, "fileutil1.yml")},
			exceptError: testErr,
			funcBeforeTest: func() {
				patchOsOpen(t, nil, testErr)
			},
		},
		{
			name:        "os.Stat error",
			input:       []string{filepath.Join(testdatapath, RmFile), filepath.Join(testdatapath, "fileutil1.yml")},
			exceptError: testErr,
			funcBeforeTest: func() {
				patchOsStat(t, nil, testErr)
				patchFileExists(t, false, 1)
			},
		},
		{
			name:        "os.OpenFile error",
			input:       []string{filepath.Join(testdatapath, RmFile), filepath.Join(testdatapath, "fileutil1.yml")},
			exceptError: testErr,
			funcBeforeTest: func() {
				patchOsOpen(t, &os.File{}, nil)
				patchOsOpenFile(t, nil, testErr, nil)
			},
		},
		{
			name:        "copy file OK",
			input:       []string{filepath.Join(testdatapath, RmFile), filepath.Join(testdatapath, "fileutil1.yml")},
			exceptError: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("start %s\n", tc.name)
			if tc.funcBeforeTest != nil {
				tc.funcBeforeTest()
			}
			_, result := CopyFile(tc.input[0], tc.input[1])
			if !isExpectedError(result, tc.exceptError) {
				t.Errorf("expected error %v, but function returned error: %v", tc.exceptError, result)
			}
			t.Logf("end %s\n", tc.name)
		})
	}

}

func TestRemoveFile(t *testing.T) {
	cases := []struct {
		name           string
		input          []string
		retError       error
		expectError    bool
		funcBeforeTest func()
	}{
		{
			name:        "Remove file fail",
			input:       []string{filepath.Join(testdatapath, "not_exist.yml")},
			retError:    nil,
			expectError: true,
			funcBeforeTest: func() {
				var patch *mpatch.Patch
				var err error
				patch, err = mpatch.PatchMethod(os.Remove, func(_ string) error {
					unpatch(t, patch)
					return testErr
				})
				if err != nil {
					t.Errorf("patch error: %v", err)
				}
			},
		},
		{
			name:        "Remove file OK",
			input:       []string{filepath.Join(testdatapath, RmFile)},
			retError:    nil,
			expectError: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				tc.funcBeforeTest()
			}
			result := RemoveFile(tc.input[0])
			if (result != nil && !tc.expectError) ||
				(result == nil && tc.expectError) {
				t.Logf("Test case %s failed.", tc.name)
				t.Error(result)
			} else {
				t.Log("Done")
			}
		})
	}
}

func TestWriteStringToFile(t *testing.T) {
	cases := []struct {
		name           string
		input          []string
		expectError    error
		funcBeforeTest func()
	}{
		{
			name:        "Create file fail",
			input:       []string{"test string", filepath.Join(testdatapath, "")},
			expectError: errIsDir,
		},
		{
			name:        "stat file fail",
			input:       []string{"test string", filepath.Join(testdatapath, "testfile")},
			expectError: testErr,
			funcBeforeTest: func() {
				patchOsStat(t, nil, testErr)
			},
		},
		{
			name:        "Chmod to 0600 fail",
			input:       []string{"test string", filepath.Join(testdatapath, "testfile")},
			expectError: testErr,
			funcBeforeTest: func() {
				patchOsChmod(t, testErr, nil)
			},
		},
		{
			name:        "WriteString fail",
			input:       []string{"test string", filepath.Join(testdatapath, "testfile")},
			expectError: testErr,
			funcBeforeTest: func() {
				patchFileWriteString(t, -1, testErr)
			},
		},
		{
			name:        "Chmod to original fail",
			input:       []string{"test string", filepath.Join(testdatapath, "testfile")},
			expectError: testErr,
			funcBeforeTest: func() {
				patchOsChmod(t, nil, func() {
					patchOsChmod(t, testErr, nil)
				})
			},
		},
		{
			name:        "Write string to file OK",
			input:       []string{"", filepath.Join(testdatapath, "testfile")},
			expectError: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				tc.funcBeforeTest()
			}
			result := WriteStringToFile(tc.input[0], tc.input[1])
			if !isExpectedError(result, tc.expectError) {
				t.Errorf("expected error %v, but function returned error: %v", tc.expectError, result)
			}
		})
	}
}

func TestDownloadFile(t *testing.T) {
	func_download_ok := func() []*mpatch.Patch {
		resp := &http.Response{Body: io.NopCloser(strings.NewReader("test reader"))}
		p1, err := mpatch.PatchMethod(http.Get, func(string) (*http.Response, error) { return resp, nil })
		if err != nil {
			t.Fatal(err)
		}
		return []*mpatch.Patch{p1}
	}

	cases := []struct {
		name        string
		input       []string
		retError    error
		expectError bool
		funcPatch   func() []*mpatch.Patch
	}{
		{
			"Download file fail",
			[]string{"", filepath.Join(testdatapath, "")},
			nil,
			true,
			nil,
		},
		{
			"Download http get error",
			[]string{filepath.Join(testdatapath, "kind"), "https://kind.sigs.k8s.io/dl/v0.12.0/kind-linux-amd64"},
			nil,
			true,
			func() []*mpatch.Patch {
				patchHttpGet(t, nil, testErr)
				return nil
			},
		},
		{
			"Download http create file error",
			[]string{filepath.Join(testdatapath, "/kind/kind"), "https://kind.sigs.k8s.io/dl/v0.12.0/kind-linux-amd64"},
			nil,
			true,
			func() []*mpatch.Patch {
				resp := &http.Response{}
				resp.Body = io.NopCloser(strings.NewReader(""))
				patchHttpGet(t, resp, nil)
				return nil
			},
		},
		{
			"Download http write file error",
			[]string{filepath.Join(testdatapath, "/kind"), "https://kind.sigs.k8s.io/dl/v0.12.0/kind-linux-amd64"},
			nil,
			true,
			func() []*mpatch.Patch {
				resp := &http.Response{}
				resp.Body = io.NopCloser(strings.NewReader(""))
				resp.Body.Close()
				patchHttpGet(t, resp, nil)
				patchIoCopy(t, 0, testErr)
				return nil
			},
		},
		{
			"Download file from file",
			[]string{filepath.Join(testdatapath, "destfile.yml"), filepath.Join(testdatapath, "fileutil1.yml")},
			nil,
			true,
			nil,
		},
		{
			"Download file from file error",
			[]string{filepath.Join(testdatapath, "destfile.yml"), fmt.Sprintf("file://%s", filepath.Join(testdatapath, "aaa.yml"))},
			nil,
			true,
			nil,
		},
		{
			"Download file from https OK",
			[]string{filepath.Join(testdatapath, "kind"), "https://kind.sigs.k8s.io/dl/v0.12.0/kind-linux-amd64"},
			nil,
			false,
			func_download_ok,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcPatch != nil {
				plist := tc.funcPatch()
				for _, p := range plist {
					defer unpatch(t, p)
				}
			}

			result := DownloadFile(tc.input[0], tc.input[1])
			if (result != nil && !tc.expectError) ||
				(result == nil && tc.expectError) {
				t.Logf("Test case %s failed.", tc.name)
				t.Error(result)
			} else {
				t.Log("Done")
			}
		})
	}
	//Remove downloaded files after download test
	err1 := RemoveFile(filepath.Join(testdatapath, "destfile.yml"))
	err2 := RemoveFile(filepath.Join(testdatapath, "kind"))
	if err1 != nil || err2 != nil {
		t.Error("Download file test not finished!")
	}
}

func TestLoadJsonFromFile(t *testing.T) {
	jsonTest := &JsonTest{}
	cases := []struct {
		name        string
		input1      string
		input2      *JsonTest
		expectError error
	}{
		{
			"Write string to file fail",
			"",
			jsonTest,
			errNoFileDir,
		},
		{
			"Load json from file fail",
			filepath.Join(testdatapath, "fileutil_test.yml"),
			jsonTest,
			errInvalidChar,
		},
		{
			"Load json from file OK",
			filepath.Join(testdatapath, "fileutil_test.json"),
			jsonTest,
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := LoadJsonFromFile(tc.input1, tc.input2)
			if !isExpectedError(result, tc.expectError) {
				t.Errorf("expected error %v, but function returned error: %v", tc.expectError, result)
			}
		})
	}
}

func TestCreateFolderIfNotExist(t *testing.T) {
	cases := []struct {
		name           string
		input          []string
		expectError    error
		funcBeforeTest func()
	}{
		{
			"CreateFolderIfNotExist1 fail",
			[]string{filepath.Join(testdatapath, "abcdefghijklmnopqrstuvwxyz")},
			testErr,
			func() {
				patchMakeDir(t, testErr)
			},
		},
		{
			"CreateFolderIfNotExist2 fail",
			[]string{filepath.Join(testdatapath, "fileutil1.yml")},
			nil,
			nil,
		},
		{
			"CreateFolderIfNotExist OK",
			[]string{filepath.Join(testdatapath, "fileutil3.yml")},
			nil,
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				tc.funcBeforeTest()
			}
			result := CreateFolderIfNotExist(tc.input[0])
			if !isExpectedError(result, tc.expectError) {
				t.Errorf("expected error %v, but function returned error: %v", tc.expectError, result)
			}
		})
	}
	//Remove created file after download test
	err := RemoveFile(filepath.Join(testdatapath, "fileutil3.yml"))
	if err != nil {
		t.Error("CreateFolderIfNotExist test not finished!")
	}
}

func TestCompressTar(t *testing.T) {
	cases := []struct {
		name           string
		input          []string
		expectError    error
		funcBeforeTest func()
	}{
		{
			name:        "CompressTar1 fail",
			input:       []string{filepath.Join(testdatapath, "not_exist"), ""},
			expectError: errNoFileDir,
		},
		{
			name:        "CompressTar2 fail",
			input:       []string{filepath.Join(testdatapath, "fileTestFolder"), ""},
			expectError: errNoFileDir,
		},
		{
			name:        "CompressTar3 file is not valid",
			input:       []string{filepath.Join(testdatapath, "fileTestFolder"), filepath.Join(testdatapath, "fileutil.tar")},
			expectError: GetError("errInvalidFile"),
			funcBeforeTest: func() {
				patchIsValidFile(t, false, nil)
			},
		},
		{
			name:        "CompressTar4 filepath walk fail",
			input:       []string{filepath.Join("testdatapath", "fileTestFolder/fileTestFolder/fileTestFolder"), filepath.Join(testdatapath, "fileutil.tar")},
			expectError: errLstatTest,
			funcBeforeTest: func() {
				fileInfo, _ := os.Stat("test")
				patchOsStat(t, fileInfo, nil)
			},
		},
		{
			name:        "CompressTar5 FileInfoHeader fail",
			input:       []string{filepath.Join(testdatapath, "fileTestFolder"), filepath.Join(testdatapath, "fileutil.tar")},
			expectError: testErr,
			funcBeforeTest: func() {
				patchTarFileInfoHeader(t, &tar.Header{}, testErr)
			},
		},
		{
			name:        "CompressTar6 WriteHeader fail",
			input:       []string{filepath.Join(testdatapath, "fileTestFolder"), filepath.Join(testdatapath, "fileutil.tar")},
			expectError: errEncHeader,
			funcBeforeTest: func() {
				patchTarFileInfoHeader(t, &tar.Header{
					Typeflag: tar.TypeXHeader,
				}, nil)
			},
		},
		{
			name:        "CompressTar7 Open file fail",
			input:       []string{filepath.Join(testdatapath, "fileTestFolder"), filepath.Join(testdatapath, "fileutil.tar")},
			expectError: testErr,
			funcBeforeTest: func() {
				file, err := os.OpenFile(filepath.Join(testdatapath, "fileutil.tar"), os.O_RDWR|os.O_CREATE, 0600)
				if err != nil {
					t.Errorf("open file error %v", err)
				}
				patchOsOpenFile(t, file, nil, func() {
					dir, err := os.OpenFile(filepath.Join(testdatapath, "fileTestFolder"), os.O_RDONLY, 0)
					if err != nil {
						t.Errorf("open file error %v", err)
					}
					patchOsOpenFile(t, dir, nil, func() {
						patchOsOpenFile(t, nil, testErr, nil)
					})
				})
			},
		},
		{
			name:        "CompressTar8 tar file in the path directory",
			input:       []string{filepath.Join(testdatapath), filepath.Join(testdatapath, "fileutil.tar")},
			expectError: GetError("errTarPath"),
		},
		{
			name:        "CompressTar9 tar file path is same as the path",
			input:       []string{filepath.Join(testdatapath, "testfile"), filepath.Join(testdatapath, "testfile")},
			expectError: GetError("errTarPath"),
		},
		{
			name:        "CompressTar OK directory",
			input:       []string{filepath.Join(testdatapath, "fileTestFolder"), filepath.Join(testdatapath, "fileutil.tar")},
			expectError: nil,
		},
		{
			name:        "CompressTar OK file",
			input:       []string{filepath.Join(testdatapath, "testfile"), filepath.Join(testdatapath, "fileutil.tar")},
			expectError: nil,
		},
	}

	err := MakeDir(filepath.Join(testdatapath, "fileTestFolder"))
	require.NoError(t, err, "Problem creating fileTestFolder:")

	file, _ := os.Create(filepath.Join(testdatapath, "fileTestFolder", "testfile"))
	file.Close()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				tc.funcBeforeTest()
			}
			result := CompressTar(tc.input[0], tc.input[1], 0600)
			if !isExpectedError(result, tc.expectError) {
				t.Errorf("expected error %v, but function returned error: %v", tc.expectError, result)
			}
		})
	}

	//remove test folder and tar file after test.
	if err := os.RemoveAll(filepath.Join(testdatapath, "fileTestFolder")); err != nil {
		t.Errorf("UncompressTgz test not finished!. os.RemoveAll error: %v", err)
	}
	if err := os.Remove(filepath.Join(testdatapath, "fileutil.tar")); err != nil {
		t.Errorf("UncompressTgz test not finished!. os.Remove error: %v", err)
	}
}

func TestUncompressTgz(t *testing.T) {
	cases := []struct {
		name           string
		input          []string
		retError       error
		expectError    bool
		funcBeforeTest func()
	}{
		{
			name:        "UncompressTgz1 fail",
			input:       []string{filepath.Join(testdatapath, "fileutil.tar"), ""},
			retError:    nil,
			expectError: true,
		},
		{
			name:        "UncompressTgz2 fail",
			input:       []string{filepath.Join(testdatapath, ""), ""},
			retError:    nil,
			expectError: true,
		},
		{
			name:        "UncompressTgz3 OpenFile fail",
			input:       []string{filepath.Join(testdatapath, "fileutil_test.tar.gz"), filepath.Join(testdatapath, "fileTestFolder")},
			retError:    nil,
			expectError: true,
			funcBeforeTest: func() {
				patchOsOpenFile(t, nil, testErr, nil)
			},
		},
		{
			name:        "UncompressTgz4 IsValidFile fail",
			input:       []string{filepath.Join(testdatapath, "fileutil_test.tar.gz"), filepath.Join(testdatapath, "fileTestFolder")},
			retError:    nil,
			expectError: true,
			funcBeforeTest: func() {
				patchIsValidFile(t, false, nil)
			},
		},
		{
			name:        "UncompressTgz5 new gzip reader fail",
			input:       []string{filepath.Join(testdatapath, "fileutil_test.tar.gz"), filepath.Join(testdatapath, "fileTestFolder")},
			retError:    nil,
			expectError: true,
			funcBeforeTest: func() {
				patchGzipNewReader(t, nil, testErr)
			},
		},
		{
			name:        "UncompressTgz6 new tar reader fail",
			input:       []string{filepath.Join(testdatapath, "fileutil_test.tar.gz"), filepath.Join(testdatapath, "fileTestFolder")},
			retError:    nil,
			expectError: true,
			funcBeforeTest: func() {
				patchTarNewReader(t, tar.NewReader(strings.NewReader("asdasdf")))
			},
		},
		{
			name:        "UncompressTgz7 tar type dir mkdir fail",
			input:       []string{filepath.Join(testdatapath, "fileutil_test.tar.gz"), filepath.Join(testdatapath, "fileTestFolder")},
			retError:    nil,
			expectError: true,
			funcBeforeTest: func() {
				patchTarReaderNext(t, &tar.Header{
					Typeflag: tar.TypeDir,
					Name:     "",
				}, nil)
			},
		},
		{
			name:        "UncompressTgz8 tar type reg not valid file",
			input:       []string{filepath.Join(testdatapath, "fileutil_test.tar.gz"), filepath.Join(testdatapath, "fileTestFolder")},
			retError:    nil,
			expectError: true,
			funcBeforeTest: func() {
				patchTarReaderNext(t, &tar.Header{
					Typeflag: tar.TypeReg,
					Name:     "",
				}, nil)
				patchFileExists(t, true, 2)
			},
		},
		{
			name:        "UncompressTgz9 tar type reg open file fail",
			input:       []string{filepath.Join(testdatapath, "fileutil_test.tar.gz"), filepath.Join(testdatapath, "fileTestFolder")},
			retError:    nil,
			expectError: true,
			funcBeforeTest: func() {
				patchTarReaderNext(t, &tar.Header{
					Typeflag: tar.TypeReg,
					Name:     "\\//--::.\\/",
				}, nil)
			},
		},
		{
			name:        "UncompressTgz10 tar type reg io copy fail",
			input:       []string{filepath.Join(testdatapath, "fileutil_test.tar.gz"), filepath.Join(testdatapath, "fileTestFolder")},
			retError:    nil,
			expectError: true,
			funcBeforeTest: func() {
				patchIoCopy(t, 0, testErr)
			},
		},
		{
			name:        "UncompressTgz11 tar type TypeXHeader",
			input:       []string{filepath.Join(testdatapath, "fileutil_test.tar.gz"), filepath.Join(testdatapath, "fileTestFolder")},
			retError:    nil,
			expectError: false,
			funcBeforeTest: func() {
				patchTarReaderNext(t, &tar.Header{
					Typeflag: tar.TypeXHeader,
				}, nil)
			},
		},
		{
			name:        "UncompressTgz12 tar type TypeXGlobalHeader",
			input:       []string{filepath.Join(testdatapath, "fileutil_test.tar.gz"), filepath.Join(testdatapath, "fileTestFolder")},
			retError:    nil,
			expectError: false,
			funcBeforeTest: func() {
				patchTarReaderNext(t, &tar.Header{
					Typeflag: tar.TypeXGlobalHeader,
				}, nil)
			},
		},
		{
			name:        "UncompressTgz13 tar type invalid",
			input:       []string{filepath.Join(testdatapath, "fileutil_test.tar.gz"), filepath.Join(testdatapath, "fileTestFolder")},
			retError:    nil,
			expectError: true,
			funcBeforeTest: func() {
				patchTarReaderNext(t, &tar.Header{
					Typeflag: tar.TypeChar,
				}, nil)
			},
		},
		{
			name:        "UncompressTgz OK",
			input:       []string{filepath.Join(testdatapath, "fileutil_test.tar.gz"), filepath.Join(testdatapath, "fileTestFolder")},
			retError:    nil,
			expectError: false,
		},
	}
	for _, tc := range cases {
		err := MakeDir(filepath.Join(testdatapath, "fileTestFolder"))
		require.NoError(t, err, "Problem creating fileTestFolder:")

		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				tc.funcBeforeTest()
			}
			result := UncompressTgz(tc.input[0], tc.input[1])
			if (result != nil && !tc.expectError) ||
				(result == nil && tc.expectError) {
				t.Logf("Test case %s failed.", tc.name)
				t.Error(result)
			} else {
				t.Log("Done")
			}
		})
		//remove test folder and tar file after test.
		if err := os.RemoveAll(filepath.Join(testdatapath, "fileTestFolder")); err != nil {
			t.Errorf("UncompressTgz test not finished!. os.RemoveAll error: %v", err)
		}
	}
}

func TestGzipCompress(t *testing.T) {
	cases := []struct {
		name           string
		input          []string
		expectedError  error
		funcBeforeTest func()
	}{
		{
			name:          "GzipCompress failed to open",
			input:         []string{filepath.Join(testdatapath, "src.tar"), filepath.Join(testdatapath)},
			expectedError: testErr,
			funcBeforeTest: func() {
				patchOsOpen(t, &os.File{}, testErr)
			},
		},
		{
			name:          "GzipCompress File already exist",
			input:         []string{filepath.Join(testdatapath, "src.tar"), filepath.Join(testdatapath)},
			expectedError: GetError("errFileExist"),
			funcBeforeTest: func() {
				patchOsOpen(t, &os.File{}, nil)
				patchFileExists(t, true, 1)
			},
		},
		{
			name:          "GzipCompress failed to create",
			input:         []string{filepath.Join(testdatapath, "src.tar"), filepath.Join(testdatapath)},
			expectedError: testErr,
			funcBeforeTest: func() {
				patchOsOpen(t, &os.File{}, nil)
				patchFileExists(t, false, 1)
				patchOsCreate(t, nil, testErr)
			},
		},
		{
			name:          "GzipCompress normal",
			input:         []string{filepath.Join(testdatapath, "src.tar"), filepath.Join(testdatapath)},
			expectedError: nil,
			funcBeforeTest: func() {
				patchOsOpen(t, &os.File{}, nil)
				patchFileExists(t, false, 1)
				patchOsCreate(t, nil, nil)
				patchIoCopy(t, 0, nil)
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				tc.funcBeforeTest()
			}
			result := GzipCompress(tc.input[0], tc.input[1])
			if !isExpectedError(result, tc.expectedError) {
				t.Errorf("expected error %v, but function returned error: %v", tc.expectedError, result)
			}
		})
	}
}
