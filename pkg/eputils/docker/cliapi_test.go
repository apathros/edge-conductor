/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

//nolint: dupl
package docker

import (
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	api "github.com/intel/edge-conductor/pkg/api/plugins"
	eputils "github.com/intel/edge-conductor/pkg/eputils"
	clientmock "github.com/intel/edge-conductor/pkg/eputils/docker/mock"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/golang/mock/gomock"
	"github.com/moby/moby/pkg/jsonmessage"
	mpatch "github.com/undefinedlabs/go-mpatch"
)

var (
	testError  = fmt.Errorf("test error")
	errArg     = errors.New("invalid argument")
	errPerm    = fmt.Errorf("permission denied")
	errFileDir = fmt.Errorf("no such file or directory")
	errPortInv = fmt.Errorf("Invalid containerPort")
)

func unpatchAll(t *testing.T, pList []*mpatch.Patch) {
	for _, p := range pList {
		if p != nil {
			if err := p.Unpatch(); err != nil {
				t.Errorf("unpatch error: %v", err)
			}
		}
	}
}

func getDockerClientErrFunc(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
	var patch *mpatch.Patch
	patch, _ = mpatch.PatchMethod(getDockerClient, func() (*client.Client, error) {
		unpatch(t, patch)
		return nil, testError
	})
	return nil
}

func patchIsValidFile(t *testing.T, isValid bool) {
	var patch *mpatch.Patch
	patch, _ = mpatch.PatchMethod(eputils.IsValidFile, func(filename string) bool {
		unpatch(t, patch)
		return isValid
	})
}

func patchOpenFile(t *testing.T, file *os.File, err error) {
	var patch *mpatch.Patch
	patch, _ = mpatch.PatchMethod(os.OpenFile, func(name string, flag int, perm os.FileMode) (*os.File, error) {
		unpatch(t, patch)
		return file, err
	})
}

//nolint:unparam
func patchStat(t *testing.T, fileInfo fs.FileInfo, err error) {
	var patch *mpatch.Patch
	patch, _ = mpatch.PatchMethod(os.Stat, func(name string) (fs.FileInfo, error) {
		unpatch(t, patch)
		return fileInfo, err
	})
}

//nolint:unparam
func patchReadAll(t *testing.T, data []byte, err error) {
	var patch *mpatch.Patch
	patch, _ = mpatch.PatchMethod(ioutil.ReadAll, func(r io.Reader) ([]byte, error) {
		unpatch(t, patch)
		return data, err
	})
}

func patchReadFile(t *testing.T, data []byte, err error) {
	var patch *mpatch.Patch
	patch, _ = mpatch.PatchMethod(os.ReadFile, func(name string) ([]byte, error) {
		unpatch(t, patch)
		return data, err
	})
}

func patchJsonMarshal(t *testing.T, data []byte, err error) {
	var patch *mpatch.Patch
	patch, _ = mpatch.PatchMethod(json.Marshal, func(_ interface{}) ([]byte, error) {
		unpatch(t, patch)
		return data, err
	})
}

func patchDisplayJSONMessagesStream(t *testing.T, err error) {
	var patch *mpatch.Patch
	patch, _ = mpatch.PatchMethod(jsonmessage.DisplayJSONMessagesStream, func(in io.Reader, out io.Writer, terminalFd uintptr, isTerminal bool, auxCallback func(jsonmessage.JSONMessage)) error {
		unpatch(t, patch)
		return err
	})
}

func TestGetDockerClient(t *testing.T) {
	normalFunc := func(ctrl *gomock.Controller, mockOutput error) []*mpatch.Patch {
		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)
		patch, err := mpatch.PatchMethod(client.NewClientWithOpts, mockDockerClientInterface.NewClientWithOpts)
		if err != nil {
			t.Errorf("mpatch error: %v", err)
		}
		mockDockerClientInterface.EXPECT().
			NewClientWithOpts(gomock.Any(), gomock.Any()).
			Return(nil, mockOutput)

		return []*mpatch.Patch{patch}
	}
	cases := []struct {
		mockOutput     error
		wantErr        error
		funcBeforeTest func(*gomock.Controller, error) []*mpatch.Patch
	}{
		{
			mockOutput:     testError,
			wantErr:        testError,
			funcBeforeTest: normalFunc,
		},
		{
			mockOutput: nil,
			wantErr:    nil,
		},
	}

	for n, testCase := range cases {
		t.Logf("TestGetContainerByName case %d start", n)
		func() {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			if testCase.funcBeforeTest != nil {
				pList := testCase.funcBeforeTest(ctrl, testCase.mockOutput)
				defer unpatchAll(t, pList)
			}

			_, err := getDockerClient()
			if !errors.Is(err, testCase.wantErr) &&
				(err == nil || !strings.Contains(err.Error(), testCase.wantErr.Error())) {
				t.Errorf("Unexpected error: %v", err)
			}
		}()
		t.Logf("TestGetContainerByName case %d End", n)
	}

	t.Log("Done")
}

// Test Entry
func TestGetContainerByName(t *testing.T) {
	normalFunc := func(ctrl *gomock.Controller, mockOutput []types.Container) []*mpatch.Patch {
		cli := &client.Client{}
		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)
		patch, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ContainerList", mockDockerClientInterface.ContainerList)
		if err != nil {
			t.Errorf("mpatch error")
		}
		mockDockerClientInterface.EXPECT().
			ContainerList(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(mockOutput, nil)

		return []*mpatch.Patch{patch}
	}
	ownGetDockerClientErrFunc := func(ctrl *gomock.Controller, _ []types.Container) []*mpatch.Patch {
		getDockerClientErrFunc(t, ctrl)
		return nil
	}

	ContainerListErrFunc := func(ctrl *gomock.Controller, mockOutput []types.Container) []*mpatch.Patch {
		cli := &client.Client{}
		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)
		patch, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ContainerList", mockDockerClientInterface.ContainerList)
		if err != nil {
			t.Errorf("mpatch error")
		}
		mockDockerClientInterface.EXPECT().
			ContainerList(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(mockOutput, testError)
		return []*mpatch.Patch{patch}
	}
	cases := []struct {
		mockOutput     []types.Container
		wantOutput     *types.Container
		wantErr        error
		funcBeforeTest func(*gomock.Controller, []types.Container) []*mpatch.Patch
	}{
		{
			mockOutput: []types.Container{
				{ID: "testId1"},
				{ID: "testId2"},
			},
			wantOutput:     &types.Container{ID: "testId1"},
			wantErr:        nil,
			funcBeforeTest: normalFunc,
		},
		{
			mockOutput:     []types.Container{},
			wantOutput:     nil,
			wantErr:        nil,
			funcBeforeTest: normalFunc,
		},
		{
			mockOutput:     nil,
			wantOutput:     nil,
			funcBeforeTest: ownGetDockerClientErrFunc,
			wantErr:        testError,
		},
		{
			mockOutput:     nil,
			wantOutput:     nil,
			funcBeforeTest: ContainerListErrFunc,
			wantErr:        testError,
		},
	}

	for n, testCase := range cases {
		t.Logf("TestGetContainerByName case %d start", n)
		func() {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			if testCase.funcBeforeTest != nil {
				pList := testCase.funcBeforeTest(ctrl, testCase.mockOutput)
				defer unpatchAll(t, pList)
			}

			container, err := GetContainerByName("test")
			if !errors.Is(err, testCase.wantErr) &&
				(err == nil || !strings.Contains(err.Error(), testCase.wantErr.Error())) {
				t.Errorf("Unexpected error: %v", err)
			}

			if container != nil && testCase.wantOutput != nil && container.ID != testCase.wantOutput.ID {
				t.Errorf("Get wrong container: %v, want container: %v", *container, *testCase.wantOutput)
			}
		}()
		t.Logf("TestGetContainerByName case %d End", n)
	}

	t.Log("Done")
}

func TestImagePull(t *testing.T) {
	normalFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		cli := &client.Client{}
		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)
		patch, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ImagePull", mockDockerClientInterface.ImagePull)
		if err != nil {
			t.Errorf("mpatch error")
		}
		readerCloser := io.NopCloser(strings.NewReader("Hello, world!"))
		mockDockerClientInterface.EXPECT().
			ImagePull(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(readerCloser, nil)
		patchDisplayJSONMessagesStream(t, nil)
		return []*mpatch.Patch{patch}
	}
	displayJSONMessagesStreamErrorFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		cli := &client.Client{}
		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)
		patch, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ImagePull", mockDockerClientInterface.ImagePull)
		if err != nil {
			t.Errorf("mpatch error")
		}
		readerCloser := io.NopCloser(strings.NewReader("Hello, world!"))
		mockDockerClientInterface.EXPECT().
			ImagePull(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(readerCloser, nil)
		patchDisplayJSONMessagesStream(t, testError)
		return []*mpatch.Patch{patch}
	}
	jsonMarshalFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		patchJsonMarshal(t, nil, testError)
		return nil
	}
	cases := []struct {
		authConfig     *types.AuthConfig
		wantErr        error
		funcBeforeTest func(*testing.T, *gomock.Controller) []*mpatch.Patch
	}{
		{
			authConfig:     &types.AuthConfig{},
			wantErr:        nil,
			funcBeforeTest: normalFunc,
		},
		{
			authConfig:     nil,
			wantErr:        nil,
			funcBeforeTest: normalFunc,
		},
		{
			authConfig:     nil,
			wantErr:        testError,
			funcBeforeTest: getDockerClientErrFunc,
		},
		{
			authConfig:     &types.AuthConfig{},
			wantErr:        testError,
			funcBeforeTest: jsonMarshalFunc,
		},
		{
			authConfig:     &types.AuthConfig{},
			wantErr:        testError,
			funcBeforeTest: displayJSONMessagesStreamErrorFunc,
		},
	}
	for n, testCase := range cases {
		t.Logf("TestImagePull case %d start", n)
		func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			if testCase.funcBeforeTest != nil {
				pList := testCase.funcBeforeTest(t, ctrl)
				defer unpatchAll(t, pList)
			}

			err := ImagePull("test", &types.AuthConfig{})
			if !errors.Is(err, testCase.wantErr) &&
				(err == nil || !strings.Contains(err.Error(), testCase.wantErr.Error())) {
				t.Errorf("Unexpected error: %v", err)
			}
		}()
		t.Logf("TestImagePull case %d End", n)
	}

	t.Log("Done")
}

func TestImagePush(t *testing.T) {
	normalFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {

		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)
		cli := &client.Client{}
		patch, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ImagePush", mockDockerClientInterface.ImagePush)
		if err != nil {
			t.Errorf("mpatch error")
		}
		readerCloser := io.NopCloser(strings.NewReader("Hello, world!"))
		mockDockerClientInterface.EXPECT().
			ImagePush(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(readerCloser, nil)
		patchDisplayJSONMessagesStream(t, nil)
		return []*mpatch.Patch{patch}
	}
	displayJSONMessagesStreamErrorFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)
		cli := &client.Client{}
		patch, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ImagePush", mockDockerClientInterface.ImagePush)
		if err != nil {
			t.Errorf("mpatch error")
		}
		readerCloser := io.NopCloser(strings.NewReader("Hello, world!"))
		mockDockerClientInterface.EXPECT().
			ImagePush(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(readerCloser, nil)
		patchDisplayJSONMessagesStream(t, testError)
		return []*mpatch.Patch{patch}
	}
	jsonMarshalFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		patchJsonMarshal(t, nil, testError)
		return nil
	}

	cases := []struct {
		authConfig     *types.AuthConfig
		wantErr        error
		funcBeforeTest func(*testing.T, *gomock.Controller) []*mpatch.Patch
	}{
		{
			authConfig:     &types.AuthConfig{},
			wantErr:        nil,
			funcBeforeTest: normalFunc,
		},
		{
			authConfig:     nil,
			wantErr:        nil,
			funcBeforeTest: normalFunc,
		},
		{
			authConfig:     nil,
			wantErr:        testError,
			funcBeforeTest: getDockerClientErrFunc,
		},
		{
			authConfig:     &types.AuthConfig{},
			wantErr:        testError,
			funcBeforeTest: jsonMarshalFunc,
		},
		{
			authConfig:     &types.AuthConfig{},
			wantErr:        testError,
			funcBeforeTest: displayJSONMessagesStreamErrorFunc,
		},
	}
	for n, testCase := range cases {
		t.Logf("TestImagePush case %d start", n)
		func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			if testCase.funcBeforeTest != nil {
				pList := testCase.funcBeforeTest(t, ctrl)
				defer unpatchAll(t, pList)
			}

			err := ImagePush("test", testCase.authConfig)
			if !errors.Is(err, testCase.wantErr) &&
				(err == nil || !strings.Contains(err.Error(), testCase.wantErr.Error())) {
				t.Errorf("Unexpected error: %v", err)
			}
		}()
		t.Logf("TestImagePush case %d end", n)
	}

	t.Log("Done")
}

func TestImagePushToRegistry(t *testing.T) {
	conf := api.Customconfig{
		Registry: &api.CustomconfigRegistry{
			User:     "user",
			Password: "password",
		},
	}

	// CASE1
	patchTagImageToLocal, err := mpatch.PatchMethod(TagImageToLocal, func(string, string) (string, error) { return "", testError })
	if err != nil {
		t.Errorf("mpatch error")
	}
	err = ImagePushToRegistry("testimage", "testregistry", &conf)
	unpatch(t, patchTagImageToLocal)
	if !errors.Is(err, testError) {
		t.Errorf("expected %v, function returned %v", testError, err)
	}

	// CASE2
	patchTagImageToLocal, err = mpatch.PatchMethod(TagImageToLocal, func(string, string) (string, error) { return "newtag", nil })
	if err != nil {
		t.Errorf("mpatch error")
	}
	patchImagePush, err := mpatch.PatchMethod(ImagePush, func(string, *types.AuthConfig) error { return testError })
	if err != nil {
		t.Errorf("mpatch error")
	}
	err = ImagePushToRegistry("testimage", "testregistry", &conf)
	unpatch(t, patchTagImageToLocal)
	unpatch(t, patchImagePush)
	if !errors.Is(err, testError) {
		t.Errorf("expected %v, function returned %v", testError, err)
	}

	// CASE3
	p1, err := mpatch.PatchMethod(TagImageToLocal, func(string, string) (string, error) { return "newtag", nil })
	if err != nil {
		t.Errorf("mpatch error")
	}
	p2, err := mpatch.PatchMethod(ImagePush, func(string, *types.AuthConfig) error { return nil })
	if err != nil {
		t.Errorf("mpatch error")
	}
	defer unpatchAll(t, []*mpatch.Patch{p1, p2})

	if err = ImagePushToRegistry("testimage", "testregistry", &conf); err != nil {
		t.Errorf("Failed")
	}

	t.Log("Done")
}

func TestImageBuild(t *testing.T) {
	normalFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		os.Setenv("http_proxy", "test_http_proxy")
		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)
		cli := &client.Client{}
		patch, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ImageBuild", mockDockerClientInterface.ImageBuild)
		if err != nil {
			t.Errorf("mpatch error")
		}
		patchIsValidFile(t, true)
		patchOpenFile(t, &os.File{}, nil)

		readerCloser := io.NopCloser(strings.NewReader("TestImageBuild Body!"))
		mockDockerClientInterface.EXPECT().
			ImageBuild(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(types.ImageBuildResponse{Body: readerCloser}, nil)
		return []*mpatch.Patch{patch}
	}

	openFileErrorFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		patchIsValidFile(t, true)
		patchOpenFile(t, nil, errPerm)
		return nil
	}
	imageBulildErrorFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)
		cli := &client.Client{}
		patch, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ImageBuild", mockDockerClientInterface.ImageBuild)
		if err != nil {
			t.Errorf("mpatch error")
		}
		patchIsValidFile(t, true)
		patchOpenFile(t, &os.File{}, nil)
		mockDockerClientInterface.EXPECT().
			ImageBuild(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(types.ImageBuildResponse{}, testError)
		return []*mpatch.Patch{patch}
	}
	readBuildResponseErrorFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)
		cli := &client.Client{}
		patch, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ImageBuild", mockDockerClientInterface.ImageBuild)
		if err != nil {
			t.Errorf("mpatch error")
		}
		patchIsValidFile(t, true)
		patchOpenFile(t, &os.File{}, nil)
		patchReadAll(t, nil, testError)
		readerCloser := io.NopCloser(strings.NewReader("TestImageBuild Body!"))
		mockDockerClientInterface.EXPECT().
			ImageBuild(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(types.ImageBuildResponse{Body: readerCloser}, nil)
		return []*mpatch.Patch{patch}
	}
	closeBuildResponseBodyErrorFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)
		cli := &client.Client{}
		patch, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ImageBuild", mockDockerClientInterface.ImageBuild)
		if err != nil {
			t.Errorf("mpatch error")
		}
		patchIsValidFile(t, true)
		patchOpenFile(t, &os.File{}, nil)
		patchReadAll(t, nil, testError)
		mockDockerClientInterface.EXPECT().
			ImageBuild(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(types.ImageBuildResponse{Body: &os.File{}}, nil)
		return []*mpatch.Patch{patch}
	}

	cases := []struct {
		dockerBuildTar      string
		dockerFilePathInTar string
		tag                 string
		wantErr             error
		funcBeforeTest      func(*testing.T, *gomock.Controller) []*mpatch.Patch
		funcAfterTest       func()
	}{
		{
			dockerBuildTar:      "tmp/TestImageBuild.test.data",
			dockerFilePathInTar: "test_file_path_in_tar",
			tag:                 "test_tag",
			wantErr:             nil,
			funcBeforeTest:      normalFunc,
		},
		{
			dockerBuildTar:      "test_build_tar",
			dockerFilePathInTar: "test_file_path_in_tar",
			tag:                 "test_tag",
			wantErr:             testError,
			funcBeforeTest:      getDockerClientErrFunc,
		},
		{
			dockerBuildTar:      "",
			dockerFilePathInTar: "test_file_path_in_tar",
			tag:                 "test_tag",
			wantErr:             eputils.GetError("errInvalidFile"),
			funcBeforeTest:      nil,
		},
		{
			dockerBuildTar:      "tmp/TestImageBuild.test.data",
			dockerFilePathInTar: "test_file_path_in_tar",
			tag:                 "test_tag",
			wantErr:             errPerm,
			funcBeforeTest:      openFileErrorFunc,
		},
		{
			dockerBuildTar:      "tmp/TestImageBuild.test.data",
			dockerFilePathInTar: "test_file_path_in_tar",
			tag:                 "test_tag",
			wantErr:             testError,
			funcBeforeTest:      imageBulildErrorFunc,
		},
		{
			dockerBuildTar:      "tmp/TestImageBuild.test.data",
			dockerFilePathInTar: "test_file_path_in_tar",
			tag:                 "test_tag",
			wantErr:             testError,
			funcBeforeTest:      readBuildResponseErrorFunc,
		},
		{
			dockerBuildTar:      "tmp/TestImageBuild.test.data",
			dockerFilePathInTar: "test_file_path_in_tar",
			tag:                 "test_tag",
			wantErr:             errArg,
			funcBeforeTest:      closeBuildResponseBodyErrorFunc,
		},
	}

	for n, testCase := range cases {
		t.Logf("TestImageBuild case %d start", n)
		func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			if testCase.funcBeforeTest != nil {
				pList := testCase.funcBeforeTest(t, ctrl)
				defer unpatchAll(t, pList)
			}

			err := ImageBuild(testCase.dockerBuildTar, testCase.dockerFilePathInTar, testCase.tag)

			if testCase.funcAfterTest != nil {
				testCase.funcAfterTest()
			}

			if !errors.Is(err, testCase.wantErr) &&
				(err == nil || !strings.Contains(err.Error(), testCase.wantErr.Error())) {
				t.Errorf("Unexpected error: %v", err)
			}
		}()
		t.Logf("TestImageBuild case %d end", n)
	}

	t.Log("Done")
}

func TestImageLoad(t *testing.T) {
	normalFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		patchStat(t, nil, nil)
		patchIsValidFile(t, true)
		patchOpenFile(t, &os.File{}, nil)
		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)
		cli := &client.Client{}
		patch, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ImageLoad", mockDockerClientInterface.ImageLoad)
		if err != nil {
			t.Errorf("mpatch error")
		}
		readerCloser := io.NopCloser(strings.NewReader("TestImageLoad Body!"))
		mockDockerClientInterface.EXPECT().
			ImageLoad(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(types.ImageLoadResponse{Body: readerCloser}, nil)
		return []*mpatch.Patch{patch}
	}

	NotValidFileFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		patchStat(t, nil, nil)
		patchIsValidFile(t, false)
		return nil
	}
	openFileErrorFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		patchStat(t, nil, nil)
		patchIsValidFile(t, true)
		patchOpenFile(t, nil, testError)
		return nil
	}
	ImageLoadErrorFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		patchStat(t, nil, nil)
		patchIsValidFile(t, true)
		patchOpenFile(t, &os.File{}, nil)
		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)
		cli := &client.Client{}
		patch, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ImageLoad", mockDockerClientInterface.ImageLoad)
		if err != nil {
			t.Errorf("mpatch error")
		}
		mockDockerClientInterface.EXPECT().
			ImageLoad(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(types.ImageLoadResponse{}, testError)
		return []*mpatch.Patch{patch}
	}
	readLoadResponseErrorFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		patchStat(t, nil, nil)
		patchIsValidFile(t, true)
		patchOpenFile(t, &os.File{}, nil)
		patchReadAll(t, nil, testError)
		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)
		cli := &client.Client{}
		patch, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ImageLoad", mockDockerClientInterface.ImageLoad)
		if err != nil {
			t.Errorf("mpatch error")
		}

		mockDockerClientInterface.EXPECT().
			ImageLoad(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(types.ImageLoadResponse{Body: io.NopCloser(strings.NewReader(""))}, nil)
		return []*mpatch.Patch{patch}
	}
	closeResponseBodyErrorFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		patchStat(t, nil, nil)
		patchIsValidFile(t, true)
		patchOpenFile(t, &os.File{}, nil)
		patchReadAll(t, nil, testError)
		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)
		cli := &client.Client{}
		patch, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ImageLoad", mockDockerClientInterface.ImageLoad)
		if err != nil {
			t.Errorf("mpatch error")
		}

		mockDockerClientInterface.EXPECT().
			ImageLoad(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(types.ImageLoadResponse{Body: &os.File{}}, nil)
		return []*mpatch.Patch{patch}
	}

	cases := []struct {
		tarball        string
		wantErr        error
		funcBeforeTest func(*testing.T, *gomock.Controller) []*mpatch.Patch
		funcAfterTest  func()
	}{
		{
			tarball:        "tmp/TestImageLoad.test.data",
			wantErr:        nil,
			funcBeforeTest: normalFunc,
		},
		{
			tarball:        "",
			wantErr:        testError,
			funcBeforeTest: getDockerClientErrFunc,
		},
		{
			tarball:        "tmp/TestImageLoad.test.data",
			wantErr:        eputils.GetError("errInvalidFile"),
			funcBeforeTest: NotValidFileFunc,
		},
		{
			tarball:        "test_tarball",
			wantErr:        errFileDir,
			funcBeforeTest: nil,
		},
		{
			tarball:        "tmp/TestImageLoad.test.data",
			wantErr:        testError,
			funcBeforeTest: openFileErrorFunc,
		},
		{
			tarball:        "tmp/TestImageLoad.test.data",
			wantErr:        testError,
			funcBeforeTest: ImageLoadErrorFunc,
		},
		{
			tarball:        "tmp/TestImageLoad.test.data",
			wantErr:        testError,
			funcBeforeTest: readLoadResponseErrorFunc,
		},
		{
			tarball:        "tmp/TestImageLoad.test.data",
			wantErr:        errArg,
			funcBeforeTest: closeResponseBodyErrorFunc,
		},
	}
	for n, testCase := range cases {
		t.Logf("TestImageLoad case %d start", n)
		func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			if testCase.funcBeforeTest != nil {
				pList := testCase.funcBeforeTest(t, ctrl)
				defer unpatchAll(t, pList)
			}

			err := ImageLoad(testCase.tarball)

			if testCase.funcAfterTest != nil {
				testCase.funcAfterTest()
			}

			if !errors.Is(err, testCase.wantErr) &&
				(err == nil || !strings.Contains(err.Error(), testCase.wantErr.Error())) {
				t.Errorf("Unexpected error: %v", err)
			}
		}()
		t.Logf("TestImageLoad case %d end", n)
	}

	t.Log("Done")
}

func TestCreateContainer(t *testing.T) {

	normalFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)
		mockDockerClientWrapperImage := clientmock.NewMockDockerClientWrapperImage(ctrl)
		cli := &client.Client{}
		patchNetworkCreate, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "NetworkCreate", mockDockerClientInterface.NetworkCreate)
		if err != nil {
			t.Errorf("mpatch error")
		}
		patchContainerCreate, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ContainerCreate", mockDockerClientInterface.ContainerCreate)
		if err != nil {
			t.Errorf("mpatch error")
		}
		patchImagePull, err := mpatch.PatchMethod(ImagePull, mockDockerClientWrapperImage.ImagePull)
		if err != nil {
			t.Errorf("mpatch error")
		}
		mockDockerClientWrapperImage.EXPECT().ImagePull(gomock.Any(), gomock.Any()).Return(nil).Times(1)

		mockDockerClientInterface.EXPECT().
			NetworkCreate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(types.NetworkCreateResponse{}, nil)

		mockDockerClientInterface.EXPECT().
			ContainerCreate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(container.ContainerCreateCreatedBody{ID: "test_id"}, nil)

		return []*mpatch.Patch{patchNetworkCreate, patchContainerCreate, patchImagePull}
	}

	ImagePullErrFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		patchImagePull, err := mpatch.PatchMethod(ImagePull, func(imageRef string, authConf *types.AuthConfig) error {
			return testError
		})
		if err != nil {
			t.Errorf("mpatch error")
		}
		return []*mpatch.Patch{patchImagePull}
	}

	NetworkCreateErrorFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)
		cli := &client.Client{}
		patch, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "NetworkCreate", mockDockerClientInterface.NetworkCreate)
		if err != nil {
			t.Errorf("mpatch error")
		}

		mockDockerClientInterface.EXPECT().
			NetworkCreate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(types.NetworkCreateResponse{}, testError)
		return []*mpatch.Patch{patch}
	}

	containerCreateErrorFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		cli := &client.Client{}
		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)

		patch, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ContainerCreate", mockDockerClientInterface.ContainerCreate)
		if err != nil {
			t.Errorf("mpatch error")
		}

		mockDockerClientInterface.EXPECT().
			ContainerCreate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(container.ContainerCreateCreatedBody{}, testError)
		return []*mpatch.Patch{patch}
	}

	cases := []struct {
		imageName       string
		containerName   string
		hostName        string
		networkMode     string
		networkNames    []string
		userInContainer string
		ports           []string
		needPullImage   bool
		wantRespId      string
		wantErr         error
		funcBeforeTest  func(*testing.T, *gomock.Controller) []*mpatch.Patch
		funcAfterTest   func()
	}{
		{
			imageName:       "test_image",
			containerName:   "test_container_name",
			hostName:        "test_host",
			networkMode:     "macvlan",
			networkNames:    []string{"test_network_name"},
			userInContainer: "test_user",
			needPullImage:   true,
			wantRespId:      "test_id",
			wantErr:         nil,
			funcBeforeTest:  normalFunc,
		},
		{
			wantRespId:     "",
			wantErr:        testError,
			funcBeforeTest: getDockerClientErrFunc,
		},
		{
			wantRespId:     "",
			needPullImage:  true,
			wantErr:        testError,
			funcBeforeTest: ImagePullErrFunc,
		},
		{
			networkMode:    "macvlan",
			ports:          []string{"___...|||"},
			needPullImage:  false,
			wantRespId:     "",
			wantErr:        errPortInv,
			funcBeforeTest: nil,
		},
		{
			networkMode:    "macvlan",
			networkNames:   []string{"test_network_name"},
			ports:          []string{"8000:8000"},
			needPullImage:  false,
			wantRespId:     "",
			wantErr:        testError,
			funcBeforeTest: NetworkCreateErrorFunc,
		},
		{
			networkMode:    "macvlan",
			needPullImage:  false,
			wantRespId:     "",
			wantErr:        testError,
			funcBeforeTest: containerCreateErrorFunc,
		},
	}

	for n, testCase := range cases {
		t.Logf("TestCreateContainer case %d start", n)
		func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			if testCase.funcBeforeTest != nil {
				pList := testCase.funcBeforeTest(t, ctrl)
				defer unpatchAll(t, pList)
			}

			respId, err := CreateContainer(
				testCase.imageName,
				testCase.containerName,
				testCase.hostName,
				testCase.networkMode,
				testCase.networkNames,
				testCase.userInContainer,
				true, testCase.needPullImage, true, true,
				nil, nil, nil, nil, nil, testCase.ports, nil, nil, nil,
				"restart",
			)

			if testCase.funcAfterTest != nil {
				testCase.funcAfterTest()
			}
			if !errors.Is(err, testCase.wantErr) &&
				(err == nil || !strings.Contains(err.Error(), testCase.wantErr.Error())) {
				t.Errorf("Unexpected error: %v", err)
			}

			if respId != testCase.wantRespId {
				t.Errorf("Get wrong resp id: %v", respId)
			}

		}()
		t.Logf("TestCreateContainer case %d end", n)
	}

	t.Log("Done")
}

func TestStartContainer(t *testing.T) {

	normalFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		cli := &client.Client{}
		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)
		mockDockerClientWrapperContainer := clientmock.NewMockDockerClientWrapperContainer(ctrl)

		patchContainerStart, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ContainerStart", mockDockerClientInterface.ContainerStart)
		if err != nil {
			t.Errorf("mpatch error")
		}
		patchContainerLogs, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ContainerLogs", mockDockerClientInterface.ContainerLogs)
		if err != nil {
			t.Errorf("mpatch error")
		}
		patchGetContainerByName, err := mpatch.PatchMethod(GetContainerByName, mockDockerClientWrapperContainer.GetContainerByName)
		if err != nil {
			t.Errorf("mpatch error")
		}
		patchContainerInspect, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ContainerInspect", mockDockerClientInterface.ContainerInspect)
		if err != nil {
			t.Errorf("mpatch error")
		}

		mockDockerClientWrapperContainer.EXPECT().
			GetContainerByName(gomock.Any()).
			Return(&types.Container{ID: "test_id"}, nil).Times(1)
		mockDockerClientInterface.EXPECT().
			ContainerStart(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil)
		readerCloser := io.NopCloser(strings.NewReader("Hello, world!"))
		mockDockerClientInterface.EXPECT().
			ContainerLogs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(readerCloser, nil)

		test_ExitCode := types.ContainerState{ExitCode: 0}
		test_ContainerJsonBase := types.ContainerJSONBase{State: &test_ExitCode}
		res := types.ContainerJSON{ContainerJSONBase: &test_ContainerJsonBase, Mounts: nil, Config: nil, NetworkSettings: nil}
		mockDockerClientInterface.EXPECT().
			ContainerInspect(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(res, nil)
		return []*mpatch.Patch{patchContainerStart, patchContainerLogs, patchGetContainerByName, patchContainerInspect}
	}

	getContainerNameErrorFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {

		mockDockerClientWrapperContainer := clientmock.NewMockDockerClientWrapperContainer(ctrl)

		patch, err := mpatch.PatchMethod(GetContainerByName, mockDockerClientWrapperContainer.GetContainerByName)
		if err != nil {
			t.Errorf("mpatch error")
		}

		mockDockerClientWrapperContainer.EXPECT().
			GetContainerByName(gomock.Any()).Return(nil, testError).Times(1)
		return []*mpatch.Patch{patch}
	}

	getContainerNameNilFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {

		mockDockerClientWrapperContainer := clientmock.NewMockDockerClientWrapperContainer(ctrl)

		patch, err := mpatch.PatchMethod(GetContainerByName, mockDockerClientWrapperContainer.GetContainerByName)
		if err != nil {
			t.Errorf("mpatch error")
		}

		mockDockerClientWrapperContainer.EXPECT().
			GetContainerByName(gomock.Any()).Return(nil, nil).Times(1)
		return []*mpatch.Patch{patch}
	}

	startContainerErrorFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		cli := &client.Client{}
		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)
		patch, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ContainerStart", mockDockerClientInterface.ContainerStart)
		if err != nil {
			t.Errorf("mpatch error")
		}
		mockDockerClientInterface.EXPECT().
			ContainerStart(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(testError)
		return []*mpatch.Patch{patch}
	}

	containerLogsErrorFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		cli := &client.Client{}

		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)

		patchContainerStart, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ContainerStart", mockDockerClientInterface.ContainerStart)
		if err != nil {
			t.Errorf("mpatch error")
		}
		patchContainerLogs, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ContainerLogs", mockDockerClientInterface.ContainerLogs)
		if err != nil {
			t.Errorf("mpatch error")
		}

		mockDockerClientInterface.EXPECT().
			ContainerStart(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil)
		mockDockerClientInterface.EXPECT().
			ContainerLogs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, testError)
		return []*mpatch.Patch{patchContainerStart, patchContainerLogs}
	}

	containerInspectErrorFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		cli := &client.Client{}

		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)
		mockDockerClientWrapperContainer := clientmock.NewMockDockerClientWrapperContainer(ctrl)

		patchGetContainerByName, err := mpatch.PatchMethod(GetContainerByName, mockDockerClientWrapperContainer.GetContainerByName)
		if err != nil {
			t.Errorf("mpatch error")
		}
		patchContainerStart, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ContainerStart", mockDockerClientInterface.ContainerStart)
		if err != nil {
			t.Errorf("mpatch error")
		}
		patchContainerLogs, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ContainerLogs", mockDockerClientInterface.ContainerLogs)
		if err != nil {
			t.Errorf("mpatch error")
		}

		patchContainerInspect, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ContainerInspect", mockDockerClientInterface.ContainerInspect)
		if err != nil {
			t.Errorf("mpatch error")
		}

		mockDockerClientWrapperContainer.EXPECT().
			GetContainerByName(gomock.Any()).
			Return(&types.Container{ID: "test_id"}, nil).AnyTimes()
		mockDockerClientInterface.EXPECT().
			ContainerStart(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil)
		readerCloser := io.NopCloser(strings.NewReader("Hello, world!"))
		mockDockerClientInterface.EXPECT().
			ContainerLogs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(readerCloser, nil)
		mockDockerClientInterface.EXPECT().
			ContainerInspect(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(types.ContainerJSON{}, testError)
		return []*mpatch.Patch{patchGetContainerByName, patchContainerStart, patchContainerLogs, patchContainerInspect}
	}

	containerInspectNilFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		cli := &client.Client{}

		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)
		mockDockerClientWrapperContainer := clientmock.NewMockDockerClientWrapperContainer(ctrl)

		patchGetContainerByName, err := mpatch.PatchMethod(GetContainerByName, mockDockerClientWrapperContainer.GetContainerByName)
		if err != nil {
			t.Errorf("mpatch error")
		}
		patchContainerStart, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ContainerStart", mockDockerClientInterface.ContainerStart)
		if err != nil {
			t.Errorf("mpatch error")
		}
		patchContainerLogs, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ContainerLogs", mockDockerClientInterface.ContainerLogs)
		if err != nil {
			t.Errorf("mpatch error")
		}

		patchContainerInspect, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ContainerInspect", mockDockerClientInterface.ContainerInspect)
		if err != nil {
			t.Errorf("mpatch error")
		}

		mockDockerClientWrapperContainer.EXPECT().
			GetContainerByName(gomock.Any()).
			Return(&types.Container{ID: "test_id"}, nil).AnyTimes()
		mockDockerClientInterface.EXPECT().
			ContainerStart(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil)
		readerCloser := io.NopCloser(strings.NewReader("Hello, world!"))
		mockDockerClientInterface.EXPECT().
			ContainerLogs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(readerCloser, nil)
		mockDockerClientInterface.EXPECT().
			ContainerInspect(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(types.ContainerJSON{}, nil)
		return []*mpatch.Patch{patchGetContainerByName, patchContainerStart, patchContainerLogs, patchContainerInspect}
	}

	containerInspectExitCodeErrorFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		cli := &client.Client{}

		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)
		mockDockerClientWrapperContainer := clientmock.NewMockDockerClientWrapperContainer(ctrl)

		patchGetContainerByName, err := mpatch.PatchMethod(GetContainerByName, mockDockerClientWrapperContainer.GetContainerByName)
		if err != nil {
			t.Errorf("mpatch error")
		}
		patchContainerStart, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ContainerStart", mockDockerClientInterface.ContainerStart)
		if err != nil {
			t.Errorf("mpatch error")
		}
		patchContainerLogs, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ContainerLogs", mockDockerClientInterface.ContainerLogs)
		if err != nil {
			t.Errorf("mpatch error")
		}

		patchContainerInspect, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ContainerInspect", mockDockerClientInterface.ContainerInspect)
		if err != nil {
			t.Errorf("mpatch error")
		}

		mockDockerClientWrapperContainer.EXPECT().
			GetContainerByName(gomock.Any()).
			Return(&types.Container{ID: "test_id"}, nil).AnyTimes()
		mockDockerClientInterface.EXPECT().
			ContainerStart(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil)
		readerCloser := io.NopCloser(strings.NewReader("Hello, world!"))
		mockDockerClientInterface.EXPECT().
			ContainerLogs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(readerCloser, nil)
		test_ExitCode := types.ContainerState{ExitCode: 1}
		test_ContainerJsonBase := types.ContainerJSONBase{State: &test_ExitCode}
		res := types.ContainerJSON{ContainerJSONBase: &test_ContainerJsonBase, Mounts: nil, Config: nil, NetworkSettings: nil}
		mockDockerClientInterface.EXPECT().
			ContainerInspect(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(res, nil)
		return []*mpatch.Patch{patchContainerStart, patchContainerLogs, patchGetContainerByName, patchContainerInspect}
	}
	cases := []struct {
		containerID     string
		containerName   string
		runInBackground bool
		wantErr         error
		funcBeforeTest  func(*testing.T, *gomock.Controller) []*mpatch.Patch
		funcAfterTest   func()
	}{
		{
			containerID:     "",
			containerName:   "test_container_name",
			runInBackground: false,
			wantErr:         nil,
			funcBeforeTest:  normalFunc,
		},
		{
			containerID:    "",
			containerName:  "test_container_name",
			wantErr:        testError,
			funcBeforeTest: getDockerClientErrFunc,
		},
		{
			containerID:    "",
			containerName:  "test_container_name",
			wantErr:        testError,
			funcBeforeTest: getContainerNameErrorFunc,
		},
		{
			containerID:    "",
			containerName:  "test_container_name",
			wantErr:        eputils.GetError("errNoContainer"),
			funcBeforeTest: getContainerNameNilFunc,
		},
		{
			containerID:     "test_container_id",
			containerName:   "test_container_name",
			runInBackground: true,
			wantErr:         testError,
			funcBeforeTest:  startContainerErrorFunc,
		},
		{
			containerID:     "test_container_id",
			containerName:   "test_container_name",
			runInBackground: false,
			wantErr:         testError,
			funcBeforeTest:  containerLogsErrorFunc,
		},
		{
			containerID:     "test_container_id",
			containerName:   "test_container_name",
			runInBackground: false,
			wantErr:         testError,
			funcBeforeTest:  containerInspectErrorFunc,
		},
		{
			containerID:     "test_container_id",
			containerName:   "test_container_name",
			runInBackground: false,
			wantErr:         eputils.GetError("errAbnormalExit"),
			funcBeforeTest:  containerInspectNilFunc,
		},
		{
			containerID:     "test_container_id",
			containerName:   "test_container_name",
			runInBackground: false,
			wantErr:         eputils.GetError("errAbnormalExit"),
			funcBeforeTest:  containerInspectExitCodeErrorFunc,
		},
	}

	for n, testCase := range cases {
		t.Logf("TestStartContainer case %d start", n)
		func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			if testCase.funcBeforeTest != nil {
				pList := testCase.funcBeforeTest(t, ctrl)
				defer unpatchAll(t, pList)
			}

			err := StartContainer(testCase.containerID, testCase.containerName, testCase.runInBackground)

			if testCase.funcAfterTest != nil {
				testCase.funcAfterTest()
			}
			if !errors.Is(err, testCase.wantErr) &&
				(err == nil || !strings.Contains(err.Error(), testCase.wantErr.Error())) {
				t.Errorf("Unexpected error: %v", err)
			}
		}()
		t.Logf("TestStartContainer case %d end", n)
	}

	t.Log("Done")
}

func TestRunContainer(t *testing.T) {
	normalFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		mockDockerClientWrapperContainer := clientmock.NewMockDockerClientWrapperContainer(ctrl)

		patchCreateContainer, err := mpatch.PatchMethod(CreateContainer, mockDockerClientWrapperContainer.CreateContainer)
		if err != nil {
			t.Errorf("mpatch error")
		}
		patchStartContainer, err := mpatch.PatchMethod(StartContainer, mockDockerClientWrapperContainer.StartContainer)
		if err != nil {
			t.Errorf("mpatch error")
		}

		mockDockerClientWrapperContainer.EXPECT().
			CreateContainer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return("test_id", nil).Times(1)

		mockDockerClientWrapperContainer.EXPECT().
			StartContainer(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil).Times(1)
		return []*mpatch.Patch{patchCreateContainer, patchStartContainer}
	}

	createContainerErrorFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		mockDockerClientWrapperContainer := clientmock.NewMockDockerClientWrapperContainer(ctrl)

		patchCreateContainer, err := mpatch.PatchMethod(CreateContainer, mockDockerClientWrapperContainer.CreateContainer)
		if err != nil {
			t.Errorf("mpatch error")
		}

		mockDockerClientWrapperContainer.EXPECT().
			CreateContainer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return("", testError).Times(1)
		return []*mpatch.Patch{patchCreateContainer}
	}
	startContainerErrorFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		mockDockerClientWrapperContainer := clientmock.NewMockDockerClientWrapperContainer(ctrl)

		patchCreateContainer, err := mpatch.PatchMethod(CreateContainer, mockDockerClientWrapperContainer.CreateContainer)
		if err != nil {
			t.Errorf("mpatch error")
		}
		patchStartContainer, err := mpatch.PatchMethod(StartContainer, mockDockerClientWrapperContainer.StartContainer)
		if err != nil {
			t.Errorf("mpatch error")
		}

		mockDockerClientWrapperContainer.EXPECT().
			CreateContainer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return("test_id", nil).Times(1)

		mockDockerClientWrapperContainer.EXPECT().
			StartContainer(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(testError).Times(1)
		return []*mpatch.Patch{patchCreateContainer, patchStartContainer}
	}

	cases := []struct {
		wantCntId      string
		wantErr        error
		funcBeforeTest func(*testing.T, *gomock.Controller) []*mpatch.Patch
		funcAfterTest  func()
	}{
		{
			wantCntId:      "test_id",
			wantErr:        nil,
			funcBeforeTest: normalFunc,
		},
		{
			wantCntId:      "",
			wantErr:        testError,
			funcBeforeTest: createContainerErrorFunc,
		},
		{
			wantCntId:      "test_id",
			wantErr:        testError,
			funcBeforeTest: startContainerErrorFunc,
		},
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	for n, testCase := range cases {
		t.Logf("TestRunContainer case %d start", n)
		func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			if testCase.funcBeforeTest != nil {
				pList := testCase.funcBeforeTest(t, ctrl)
				defer unpatchAll(t, pList)
			}

			cntId, err := RunContainer("test_image",
				"test_container_name",
				"test_host",
				"macvlan",
				[]string{"test_network_name"},
				"test_user",
				true, true, true, true,
				nil, nil, nil, nil, nil, nil, nil, nil, nil,
				"restart")

			if testCase.funcAfterTest != nil {
				testCase.funcAfterTest()
			}
			if !errors.Is(err, testCase.wantErr) &&
				(err == nil || !strings.Contains(err.Error(), testCase.wantErr.Error())) {
				t.Errorf("Unexpected error: %v", err)
			}

			if cntId != testCase.wantCntId {
				t.Errorf("Get wrong container id: %v", cntId)
			}

		}()
		t.Logf("TestRunContainer case %d end", n)
	}

	t.Log("Done")
}

func TestStopContainer(t *testing.T) {
	normalFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		cli := &client.Client{}
		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)
		mockDockerClientWrapperContainer := clientmock.NewMockDockerClientWrapperContainer(ctrl)

		patchContainerStop, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ContainerStop", mockDockerClientInterface.ContainerStop)
		if err != nil {
			t.Errorf("mpatch error")
		}
		patchGetContainerByName, err := mpatch.PatchMethod(GetContainerByName, mockDockerClientWrapperContainer.GetContainerByName)
		if err != nil {
			t.Errorf("mpatch error")
		}

		mockDockerClientWrapperContainer.EXPECT().
			GetContainerByName(gomock.Any()).Return(&types.Container{ID: "test_id123456789", State: "running"}, nil).Times(1)

		mockDockerClientInterface.EXPECT().
			ContainerStop(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil)
		return []*mpatch.Patch{patchContainerStop, patchGetContainerByName}
	}

	getContainerByNameErrorFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		patchGetContainerByName, err := mpatch.PatchMethod(GetContainerByName, func(containerName string) (*types.Container, error) {
			return nil, testError
		})
		if err != nil {
			t.Errorf("mpatch error")
		}
		return []*mpatch.Patch{patchGetContainerByName}
	}

	containerStopErrorFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		cli := &client.Client{}
		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)
		mockDockerClientWrapperContainer := clientmock.NewMockDockerClientWrapperContainer(ctrl)

		patchContainerStop, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ContainerStop", mockDockerClientInterface.ContainerStop)
		if err != nil {
			t.Errorf("mpatch error")
		}
		patchGetContainerByName, err := mpatch.PatchMethod(GetContainerByName, mockDockerClientWrapperContainer.GetContainerByName)
		if err != nil {
			t.Errorf("mpatch error")
		}
		mockDockerClientWrapperContainer.EXPECT().
			GetContainerByName(gomock.Any()).Return(&types.Container{ID: "test_id123456789", State: "running"}, nil).Times(1)

		mockDockerClientInterface.EXPECT().
			ContainerStop(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(testError)
		return []*mpatch.Patch{patchContainerStop, patchGetContainerByName}
	}

	cases := []struct {
		containerName  string
		wantErr        error
		funcBeforeTest func(*testing.T, *gomock.Controller) []*mpatch.Patch
		funcAfterTest  func()
	}{
		{
			containerName:  "test_container_name",
			wantErr:        nil,
			funcBeforeTest: normalFunc,
		},
		{
			containerName:  "test_container_name",
			wantErr:        testError,
			funcBeforeTest: getDockerClientErrFunc,
		},
		{
			containerName:  "test_container_name",
			wantErr:        testError,
			funcBeforeTest: getContainerByNameErrorFunc,
		},
		{
			containerName:  "test_container_name",
			wantErr:        testError,
			funcBeforeTest: containerStopErrorFunc,
		},
	}

	for n, testCase := range cases {
		t.Logf("TestStopContainer case %d start", n)
		func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			if testCase.funcBeforeTest != nil {
				pList := testCase.funcBeforeTest(t, ctrl)
				defer unpatchAll(t, pList)
			}

			err := StopContainer(testCase.containerName)

			if testCase.funcAfterTest != nil {
				testCase.funcAfterTest()
			}

			if !errors.Is(err, testCase.wantErr) &&
				(err == nil || !strings.Contains(err.Error(), testCase.wantErr.Error())) {
				t.Errorf("Unexpected error: %v", err)
			}

		}()
		t.Logf("TestStopContainer case %d end", n)
	}

	t.Log("Done")
}

func TestRemoveContainer(t *testing.T) {
	normalFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		cli := &client.Client{}
		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)
		mockDockerClientWrapperContainer := clientmock.NewMockDockerClientWrapperContainer(ctrl)

		patchContainerRemove, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ContainerRemove", mockDockerClientInterface.ContainerRemove)
		if err != nil {
			t.Errorf("mpatch error")
		}

		patchStopContainer, err := mpatch.PatchMethod(StopContainer, mockDockerClientWrapperContainer.StopContainer)
		if err != nil {
			t.Errorf("mpatch error")
		}

		patchGetGetContainerByName, err := mpatch.PatchMethod(GetContainerByName, mockDockerClientWrapperContainer.GetContainerByName)
		if err != nil {
			t.Errorf("mpatch error")
		}

		mockDockerClientWrapperContainer.EXPECT().
			StopContainer(gomock.Any()).Return(nil).Times(1)

		mockDockerClientWrapperContainer.EXPECT().
			GetContainerByName(gomock.Any()).Return(&types.Container{ID: "test_id123456789", State: "running"}, nil).Times(1)

		mockDockerClientInterface.EXPECT().
			ContainerRemove(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil)
		return []*mpatch.Patch{patchContainerRemove, patchStopContainer, patchGetGetContainerByName}
	}

	stopContainerErrorFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		mockDockerClientWrapperContainer := clientmock.NewMockDockerClientWrapperContainer(ctrl)

		patchStopContainer, err := mpatch.PatchMethod(StopContainer, mockDockerClientWrapperContainer.StopContainer)
		if err != nil {
			t.Errorf("mpatch error")
		}

		mockDockerClientWrapperContainer.EXPECT().
			StopContainer(gomock.Any()).Return(testError).Times(1)
		return []*mpatch.Patch{patchStopContainer}
	}

	getContainerByNameErrorFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		mockDockerClientWrapperContainer := clientmock.NewMockDockerClientWrapperContainer(ctrl)

		patchStopContainer, err := mpatch.PatchMethod(StopContainer, mockDockerClientWrapperContainer.StopContainer)
		if err != nil {
			t.Errorf("mpatch error")
		}

		patchGetGetContainerByName, err := mpatch.PatchMethod(GetContainerByName, mockDockerClientWrapperContainer.GetContainerByName)
		if err != nil {
			t.Errorf("mpatch error")
		}

		mockDockerClientWrapperContainer.EXPECT().
			StopContainer(gomock.Any()).Return(nil).Times(1)

		mockDockerClientWrapperContainer.EXPECT().
			GetContainerByName(gomock.Any()).Return(nil, testError).Times(1)
		return []*mpatch.Patch{patchStopContainer, patchGetGetContainerByName}
	}

	containerRemoveErrorFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		cli := &client.Client{}
		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)
		mockDockerClientWrapperContainer := clientmock.NewMockDockerClientWrapperContainer(ctrl)

		patchContainerRemove, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ContainerRemove", mockDockerClientInterface.ContainerRemove)
		if err != nil {
			t.Errorf("mpatch error")
		}

		patchStopContainer, err := mpatch.PatchMethod(StopContainer, mockDockerClientWrapperContainer.StopContainer)
		if err != nil {
			t.Errorf("mpatch error")
		}

		patchGetGetContainerByName, err := mpatch.PatchMethod(GetContainerByName, mockDockerClientWrapperContainer.GetContainerByName)
		if err != nil {
			t.Errorf("mpatch error")
		}

		mockDockerClientWrapperContainer.EXPECT().
			StopContainer(gomock.Any()).Return(nil).Times(1)

		mockDockerClientWrapperContainer.EXPECT().
			GetContainerByName(gomock.Any()).Return(&types.Container{ID: "test_id123456789", State: "running"}, nil).Times(1)

		mockDockerClientInterface.EXPECT().
			ContainerRemove(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(testError)
		return []*mpatch.Patch{patchContainerRemove, patchStopContainer, patchGetGetContainerByName}
	}

	cases := []struct {
		containerName  string
		wantErr        error
		funcBeforeTest func(*testing.T, *gomock.Controller) []*mpatch.Patch
		funcAfterTest  func()
	}{
		{
			containerName:  "test_container_name",
			wantErr:        nil,
			funcBeforeTest: normalFunc,
		},
		{
			containerName:  "test_container_name",
			wantErr:        testError,
			funcBeforeTest: getDockerClientErrFunc,
		},
		{
			containerName:  "test_container_name",
			wantErr:        testError,
			funcBeforeTest: stopContainerErrorFunc,
		},
		{
			containerName:  "test_container_name",
			wantErr:        testError,
			funcBeforeTest: getContainerByNameErrorFunc,
		},
		{
			containerName:  "test_container_name",
			wantErr:        testError,
			funcBeforeTest: containerRemoveErrorFunc,
		},
	}

	for n, testCase := range cases {
		t.Logf("TestRemoveContainer case %d start", n)
		func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			if testCase.funcBeforeTest != nil {
				pList := testCase.funcBeforeTest(t, ctrl)
				defer unpatchAll(t, pList)
			}

			err := RemoveContainer(testCase.containerName)

			if testCase.funcAfterTest != nil {
				testCase.funcAfterTest()
			}

			if !errors.Is(err, testCase.wantErr) &&
				(err == nil || !strings.Contains(err.Error(), testCase.wantErr.Error())) {
				t.Errorf("Unexpected error: %v", err)
			}

		}()
		t.Logf("TestRemoveContainer case %d end", n)
	}

	t.Log("Done")
}

func TestTagImageToLocal(t *testing.T) {
	normalFunc := func(ctrl *gomock.Controller, retErr error) []*mpatch.Patch {
		cli := &client.Client{}
		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)
		patchImageInspectWithRaw, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ImageInspectWithRaw", mockDockerClientInterface.ImageInspectWithRaw)
		if err != nil {
			t.Errorf("mpatch error")
		}

		patchImageTag, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ImageTag", mockDockerClientInterface.ImageTag)
		if err != nil {
			t.Errorf("mpatch error")
		}

		mockDockerClientInterface.EXPECT().
			ImageInspectWithRaw(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(types.ImageInspect{ID: "test_id"}, nil, nil)

		mockDockerClientInterface.EXPECT().
			ImageTag(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(retErr)
		return []*mpatch.Patch{patchImageInspectWithRaw, patchImageTag}
	}

	imageInspectWithRawErrorFunc := func(ctrl *gomock.Controller, retErr error) []*mpatch.Patch {
		cli := &client.Client{}
		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)

		patchImageInspectWithRaw, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ImageInspectWithRaw", mockDockerClientInterface.ImageInspectWithRaw)
		if err != nil {
			t.Errorf("mpatch error")
		}

		mockDockerClientInterface.EXPECT().
			ImageInspectWithRaw(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(types.ImageInspect{}, nil, retErr)
		return []*mpatch.Patch{patchImageInspectWithRaw}
	}

	cases := []struct {
		imageTag       string
		registryURL    string
		wantNewTag     string
		wantErr        error
		funcBeforeTest func(*testing.T, *gomock.Controller) []*mpatch.Patch
		funcAfterTest  func()
	}{
		{
			imageTag:    "test_image_tag",
			registryURL: "test_registry_url",
			wantNewTag:  "test_registry_url/test_image_tag",
			wantErr:     nil,
			funcBeforeTest: func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
				return normalFunc(ctrl, nil)
			},
		},
		{
			imageTag:       "test_image_tag",
			registryURL:    "test_registry_url",
			wantNewTag:     "",
			wantErr:        testError,
			funcBeforeTest: getDockerClientErrFunc,
		},
		{
			imageTag:    "test_image_tag",
			registryURL: "test_registry_url",
			wantNewTag:  "",
			wantErr:     testError,
			funcBeforeTest: func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
				return imageInspectWithRawErrorFunc(ctrl, testError)
			},
		},
		{
			imageTag:    "test_image_tag",
			registryURL: "test_registry_url",
			wantNewTag:  "",
			wantErr:     testError,
			funcBeforeTest: func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
				return normalFunc(ctrl, testError)
			},
		},
	}

	for n, testCase := range cases {
		t.Logf("TestTagImageToLocal case %d start", n)
		func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			if testCase.funcBeforeTest != nil {
				pList := testCase.funcBeforeTest(t, ctrl)
				defer unpatchAll(t, pList)
			}

			newTag, err := TagImageToLocal(testCase.imageTag, testCase.registryURL)

			if testCase.funcAfterTest != nil {
				testCase.funcAfterTest()
			}

			if !errors.Is(err, testCase.wantErr) &&
				(err == nil || !strings.Contains(err.Error(), testCase.wantErr.Error())) {
				t.Errorf("Unexpected error: %v", err)
			}

			if newTag != testCase.wantNewTag {
				t.Errorf("Get wrong tag: %v", newTag)
			}

		}()
		t.Logf("TestTagImageToLocal case %d end", n)
	}

	t.Log("Done")
}

func TestTagImage(t *testing.T) {
	normalFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		cli := &client.Client{}
		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)

		patchImageInspectWithRaw, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ImageInspectWithRaw", mockDockerClientInterface.ImageInspectWithRaw)
		if err != nil {
			t.Errorf("mpatch error")
		}

		patchImageTag, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ImageTag", mockDockerClientInterface.ImageTag)
		if err != nil {
			t.Errorf("mpatch error")
		}

		mockDockerClientInterface.EXPECT().
			ImageInspectWithRaw(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(types.ImageInspect{ID: "test_id"}, nil, nil)

		mockDockerClientInterface.EXPECT().
			ImageTag(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil)
		return []*mpatch.Patch{patchImageInspectWithRaw, patchImageTag}
	}

	imageInspectWithRawErrorFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		cli := &client.Client{}
		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)

		patchImageInspectWithRaw, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ImageInspectWithRaw", mockDockerClientInterface.ImageInspectWithRaw)
		if err != nil {
			t.Errorf("mpatch error")
		}

		mockDockerClientInterface.EXPECT().
			ImageInspectWithRaw(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(types.ImageInspect{}, nil, testError)
		return []*mpatch.Patch{patchImageInspectWithRaw}
	}

	imageTagErrorFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		cli := &client.Client{}
		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)

		patchImageInspectWithRaw, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ImageInspectWithRaw", mockDockerClientInterface.ImageInspectWithRaw)
		if err != nil {
			t.Errorf("mpatch error")
		}

		patchImageTag, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ImageTag", mockDockerClientInterface.ImageTag)
		if err != nil {
			t.Errorf("mpatch error")
		}

		mockDockerClientInterface.EXPECT().
			ImageInspectWithRaw(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(types.ImageInspect{ID: "test_id"}, nil, nil)

		mockDockerClientInterface.EXPECT().
			ImageTag(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(testError)
		return []*mpatch.Patch{patchImageInspectWithRaw, patchImageTag}
	}

	cases := []struct {
		imageTag       string
		newTag         string
		wantErr        error
		funcBeforeTest func(*testing.T, *gomock.Controller) []*mpatch.Patch
		funcAfterTest  func()
	}{
		{
			imageTag:       "test_image_tag",
			newTag:         "test_new_tag",
			wantErr:        nil,
			funcBeforeTest: normalFunc,
		},
		{
			imageTag:       "test_image_tag",
			newTag:         "test_new_tag",
			wantErr:        testError,
			funcBeforeTest: getDockerClientErrFunc,
		},
		{
			imageTag:       "test_image_tag",
			newTag:         "test_new_tag",
			wantErr:        testError,
			funcBeforeTest: imageInspectWithRawErrorFunc,
		},
		{
			imageTag:       "test_image_tag",
			newTag:         "test_new_tag",
			wantErr:        testError,
			funcBeforeTest: imageTagErrorFunc,
		},
	}
	for n, testCase := range cases {
		t.Logf("TestTagImage case %d start", n)
		func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			if testCase.funcBeforeTest != nil {
				pList := testCase.funcBeforeTest(t, ctrl)
				defer unpatchAll(t, pList)
			}

			err := TagImage(testCase.imageTag, testCase.newTag)

			if testCase.funcAfterTest != nil {
				testCase.funcAfterTest()
			}

			if !errors.Is(err, testCase.wantErr) &&
				(err == nil || !strings.Contains(err.Error(), testCase.wantErr.Error())) {
				t.Errorf("Unexpected error: %v", err)
			}

		}()
		t.Logf("TestTagImage case %d end", n)
	}

	t.Log("Done")
}

func TestGetCertPoolCfgWithCustomCa(t *testing.T) {
	wrong_tarball := "wrong_tarball"
	tlscfg, err := GetCertPoolCfgWithCustomCa(wrong_tarball)
	if err != nil {
		t.Errorf("Unexpect error returned. %v", err)
	} else if tlscfg != nil {
		t.Errorf("Unexpect tls config returned. %v", tlscfg)
	}

	patchSystemCertPool, err := mpatch.PatchMethod(x509.SystemCertPool, func() (*x509.CertPool, error) {
		return nil, nil
	})
	if err != nil {
		t.Errorf("mpatch error")
	}
	defer unpatch(t, patchSystemCertPool)

	tarballFileName := "tmp/TestGetCertPoolCfgWithCustomCa.test.data"
	patchReadFile(t, []byte("test"), nil)
	patchFileExists, err := mpatch.PatchMethod(eputils.FileExists, func(_ string) bool {
		return true
	})
	if err != nil {
		t.Errorf("mpatch error")
	}

	defer unpatch(t, patchFileExists)
	tlscfg, err = GetCertPoolCfgWithCustomCa(tarballFileName)
	if err != nil {
		t.Errorf("Unexpect error returned. %v", err)
	} else if tlscfg == nil {
		t.Errorf("Get empty tls config.")
	}

	patchReadFile(t, nil, testError)
	tlscfg, err = GetCertPoolCfgWithCustomCa(tarballFileName)
	if err == nil || !strings.Contains(err.Error(), testError.Error()) {
		t.Errorf("Unexpect error returned. %v", err)
	} else if tlscfg != nil {
		t.Errorf("Unexpect tls config returned. %v", tlscfg)
	}

	t.Log("Done")
}

func TestGetImageNewTag(t *testing.T) {
	registryurl := "testregistry"
	cases := []struct {
		input, expectedoutput string
	}{
		{
			input:          "testtag:latest",
			expectedoutput: registryurl + "/" + "testtag:latest",
		},
		{
			input:          "testtag:1.1",
			expectedoutput: registryurl + "/" + "testtag:1.1",
		},
		{
			input:          "docker.io/testtag:1.1",
			expectedoutput: registryurl + "/" + "docker.io/testtag:1.1",
		},
		{
			input:          "testregistry:5555/testtag:1.1",
			expectedoutput: registryurl + "/" + "testtag:1.1",
		},
		{
			input:          "testtag:1.1@SHA256",
			expectedoutput: registryurl + "/" + "testtag:1.1",
		},
	}
	for _, c := range cases {
		input := c.input
		expectedoutput := c.expectedoutput

		testoutput := GetImageNewTag(input, registryurl)
		if testoutput != expectedoutput {
			t.Errorf("GetImageNewTag(%s, %s) expected %s but got %s",
				input, registryurl, expectedoutput, testoutput)
		}
	}

	t.Log("Done")
}

func TestErrCase(t *testing.T) {
	wrong_tarball := "wrong_tarball"
	wrong_imgref := "wrong_imgref:not_exist"
	wrong_auth, _ := GetAuthConf("wrong_address", "wrong_port", "wrong_user", "wrong_password")

	if err := ImagePull(wrong_imgref, wrong_auth); err == nil {
		t.Errorf("ImagePull() not return error with wrong input")
	}
	if err := ImagePull(wrong_imgref, nil); err == nil {
		t.Errorf("ImagePull() not return error with wrong input")
	}
	if err := ImagePush(wrong_imgref, wrong_auth); err == nil {
		t.Logf("Expected wrong_imgref and wrong_auth.")
	}
	if err := ImagePush(wrong_imgref, nil); err == nil {
		t.Logf("Expected wrong_imgref.")
	}

	wrong_dockerfile := "wrong_dockerfile"
	wrong_tag := "wrong_tag"
	wrong_registryurl := "wrong_registryurl"
	_, cf, _, ok := runtime.Caller(0)
	if !ok {
		t.Errorf("Failed to get current test file.")
	}
	cwd := filepath.Join(filepath.Dir(cf))
	wrong_tarball2 := filepath.Join(cwd, "test-containers.yml")

	if err := ImageBuild(wrong_tarball, wrong_dockerfile, wrong_tag); err == nil {
		t.Errorf("ImageBuild() not return error with wrong input")
	}
	if err := ImageBuild(wrong_tarball2, wrong_dockerfile, wrong_tag); err == nil {
		t.Errorf("ImageBuild() not return error with wrong input")
	}
	if err := ImageLoad(wrong_tarball); err == nil {
		t.Errorf("ImageLoad() not return error with wrong input")
	}
	if err := ImageLoad(wrong_tarball2); err == nil {
		t.Logf("Expected wrong_tarball2.")
	}
	if newtag, err := TagImageToLocal(wrong_tag, wrong_registryurl); err == nil {
		t.Errorf("TagImageToLocal() not return error with wrong input, but return with %s", newtag)
	}
	if _, err := TagImageToLocal("hello-world:latest", ""); err == nil {
		t.Logf("Expected null tag.")
	}

	//auth.go
	if _, err := GetAuthConf("", "", "", ""); err == nil {
		t.Errorf("GetAuthConf() not return error with wrong input")
	}

	t.Log("Done")
}

func TestGetHostImages(t *testing.T) {
	normalFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		cli := &client.Client{}
		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)

		patchImageList, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ImageList", mockDockerClientInterface.ImageList)
		if err != nil {
			t.Errorf("mpatch error")
		}
		mockDockerClientInterface.EXPECT().
			ImageList(gomock.Any(), gomock.Any(), gomock.Any()).
			Return([]types.ImageSummary{
				{RepoTags: []string{"helloworld:latest", "helloworld:2.0"}},
				{RepoTags: []string{"test:1.0", "test:2.0"}},
			}, nil)
		return []*mpatch.Patch{patchImageList}
	}

	ImageListErrFunc := func(t *testing.T, ctrl *gomock.Controller) []*mpatch.Patch {
		cli := &client.Client{}
		mockDockerClientInterface := clientmock.NewMockDockerClientInterface(ctrl)

		patchImageList, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "ImageList", mockDockerClientInterface.ImageList)
		if err != nil {
			t.Errorf("mpatch error")
		}

		mockDockerClientInterface.EXPECT().
			ImageList(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, testError)
		return []*mpatch.Patch{patchImageList}
	}
	cases := []struct {
		name           string
		wantErr        error
		funcBeforeTest func(*testing.T, *gomock.Controller) []*mpatch.Patch
		funcAfterTest  func()
	}{
		{
			name:           "test_normal",
			wantErr:        nil,
			funcBeforeTest: normalFunc,
		},
		{
			name:           "test_get_docker_client_err",
			wantErr:        testError,
			funcBeforeTest: getDockerClientErrFunc,
		},
		{
			name:           "test_images_list_err",
			wantErr:        testError,
			funcBeforeTest: ImageListErrFunc,
		},
	}
	for _, testCase := range cases {
		t.Logf("TestGetHostImages case %s start", testCase.name)
		func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			if testCase.funcBeforeTest != nil {
				pList := testCase.funcBeforeTest(t, ctrl)
				defer unpatchAll(t, pList)
			}

			_, err := GetHostImages()

			if testCase.funcAfterTest != nil {
				testCase.funcAfterTest()
			}

			if !errors.Is(err, testCase.wantErr) &&
				(err == nil || !strings.Contains(err.Error(), testCase.wantErr.Error())) {
				t.Errorf("Unexpected error: %v", err)
			}

		}()
		t.Logf("TestTagImage case %s end", testCase.name)
	}

	t.Log("Done")
}
