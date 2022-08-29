/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

//nolint: dupl
package docker

import (
	api "ep/pkg/api/plugins"
	"ep/pkg/eputils"
	clientmock "ep/pkg/eputils/docker/mock"
	"errors"
	"os/user"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/mount"
	"github.com/golang/mock/gomock"
	mpatch "github.com/undefinedlabs/go-mpatch"
)

var (
	err_docker = errors.New("docker error")
	errUsrCur  = errors.New("user.Current error.")
	errDckrCrt = errors.New("DockerCreate error.")
)

func unpatch(t *testing.T, m *mpatch.Patch) {
	err := m.Unpatch()
	if err != nil {
		t.Fatal(err)
	}
}

type mockInputOutput struct {
	*api.ContainersItems0
	containerId    string
	networkMode    string
	networkNames   []string
	userInfo       string
	needimagepull  bool
	binds          []string
	mounts         []mount.Mount
	volumes        map[string]struct{}
	ports          []string
	env            []string
	err            error
	restart        string
	readOnlyRootfs bool
}

// Test Entry
func TestDockerCreate(t *testing.T) {
	falseValue := false
	userCurrentPatchErrFunc := func(expectedErr error) *mpatch.Patch {
		p, err := mpatch.PatchMethod(user.Current, func() (*user.User, error) {
			return nil, expectedErr
		})
		if err != nil {
			t.Fatal(err)
		}
		return p
	}
	userCurrentPatchOKFunc := func(expectedErr error) *mpatch.Patch {
		p, err := mpatch.PatchMethod(user.Current, func() (*user.User, error) {
			return &user.User{Uid: "1000", Gid: "1000"}, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		return p
	}
	normalFunc := func(ctrl *gomock.Controller, testCase mockInputOutput) *mpatch.Patch {
		mockWrapperContainer := clientmock.NewMockDockerClientWrapperContainer(ctrl)
		p, err := mpatch.PatchMethod(CreateContainer, mockWrapperContainer.CreateContainer)
		if err != nil {
			t.Fatal(err)
		}
		mockWrapperContainer.EXPECT().CreateContainer(
			//image, name, hostname,
			testCase.Image, testCase.Name, testCase.HostName,
			//netmode, netname,    userinfo,
			testCase.networkMode, testCase.networkNames, testCase.userInfo,
			//priviledged, imagepull, background, readOnlyRootfs,
			false, testCase.needimagepull, testCase.RunInBackground, testCase.readOnlyRootfs,
			//cmd,   args,
			testCase.Command, testCase.Args,
			//binds,    mounts,          volumes,
			testCase.binds, testCase.mounts, testCase.volumes,
			//ports,        env,          caps,             securityOpt,
			testCase.ports, testCase.env, testCase.CapAdd, testCase.SecurityOpt,
			//restart
			testCase.restart,
		).Return(testCase.containerId, testCase.err)
		return p
	}
	cases := []struct {
		item                 mockInputOutput
		funcBeforeTest       func(ctrl *gomock.Controller, testCase mockInputOutput) *mpatch.Patch
		userCurrentPatchFunc func(err error) *mpatch.Patch
	}{
		{
			item: mockInputOutput{
				ContainersItems0: &api.ContainersItems0{
					Name:            "test_dockerrun",
					Image:           "testimage:latest",
					HostName:        "hostname",
					UserInContainer: "auto",
					Command:         []string{"testcmds"},
					Args:            []string{"testargs"},
					CapAdd:          []string{},
					Networks:        []string{"test_network"},
				},
				readOnlyRootfs: true,
				networkMode:    "bridge",
				userInfo:       "",
				needimagepull:  true,
				networkNames:   []string{"test_network"},
				mounts:         []mount.Mount{},
				volumes:        map[string]struct{}{},
				binds:          []string{},
				ports:          []string{},
				env:            []string{},
				containerId:    "id",
				err:            nil,
			},
			funcBeforeTest: normalFunc,
		},
		{
			item: mockInputOutput{
				ContainersItems0: &api.ContainersItems0{
					Name:            "test_dockerrun",
					Image:           "testimage:latest",
					HostName:        "hostname",
					UserInContainer: "auto",
					Command:         []string{"testcmds"},
					Args:            []string{"testargs"},
					CapAdd:          []string{},
					Networks:        []string{"test_network"},
					Ports: []*api.ContainersItems0PortsItems0{
						{
							Protocol: "tcp",
							HostIP:   "0.0.0.0",
						},
					},
				},
				containerId: "",
				err:         eputils.GetError("errIP"),
			},
		},
		{
			item: mockInputOutput{
				ContainersItems0: &api.ContainersItems0{
					Name:            "test_dockerrun",
					Image:           "testimage:latest",
					HostName:        "hostname",
					UserInContainer: "auto",
					Command:         []string{"testcmds"},
					Args:            []string{"testargs"},
					CapAdd:          []string{},
					ImagePullPolicy: "Never",
					VolumeMounts: []*api.ContainersItems0VolumeMountsItems0{
						{
							HostPath:  "/test/Source",
							MountPath: "/test/Target",
							ReadOnly:  true,
						},
					},
					BindMounts: []*api.ContainersItems0BindMountsItems0{
						{
							HostPath:  "test/HostPath1",
							MountPath: "test/MountPath1",
							ReadOnly:  true,
						},
						{
							HostPath:  "test/HostPath2",
							MountPath: "test/MountPath2",
							ReadOnly:  false,
						},
					},
					Networks:    []string{"test_network"},
					HostNetwork: true,
					Ports: []*api.ContainersItems0PortsItems0{
						{
							Protocol: "tcp",
							HostIP:   "localhost",
							HostPort: 8080,
						},
						{
							Protocol:      "udp",
							HostIP:        "127.0.0.2",
							HostPort:      8118,
							ContainerPort: 8118,
						},
						{
							Protocol: "udp",
							HostIP:   "",
						},
					},
					Env: []*api.ContainersItems0EnvItems0{
						{
							Name:  "test_name",
							Value: "test_value",
						},
					},
				},
				networkMode:   "host",
				userInfo:      "",
				needimagepull: false,
				networkNames:  []string{},
				mounts: []mount.Mount{
					{
						Type:     mount.TypeVolume,
						Source:   "/test/Source",
						Target:   "/test/Target",
						ReadOnly: true,
					},
				},
				volumes:        map[string]struct{}{},
				binds:          []string{"test/HostPath1:test/MountPath1:ro", "test/HostPath2:test/MountPath2"},
				ports:          []string{"127.0.0.1:8080:0/tcp", "127.0.0.2:8118:8118/udp", "127.0.0.1:0:0/udp"},
				env:            []string{"test_name=test_value"},
				containerId:    "id",
				readOnlyRootfs: true,
				err:            nil,
			},
			funcBeforeTest: normalFunc,
		},
		{
			item: mockInputOutput{
				ContainersItems0: &api.ContainersItems0{
					Name:        "test_dockerrun",
					Image:       "testimage:latest",
					HostName:    "hostname",
					Command:     []string{"testcmds"},
					Args:        []string{"testargs"},
					CapAdd:      []string{},
					Networks:    []string{"test_network"},
					HostNetwork: true,
				},
				containerId: "",
				err:         errUsrCur,
			},
			userCurrentPatchFunc: userCurrentPatchErrFunc,
		},
		{
			item: mockInputOutput{
				ContainersItems0: &api.ContainersItems0{
					Name:            "test_dockerrun",
					Image:           "testimage:latest",
					HostName:        "hostname",
					UserInContainer: "",
					Command:         []string{"testcmds"},
					Args:            []string{"testargs"},
					CapAdd:          []string{},
					Networks:        []string{"test_network"},
					Tmpfs:           []string{"tmp_fs"},
					ReadOnlyRootfs:  &falseValue,
				},
				networkMode:   "bridge",
				userInfo:      "1000:1000",
				needimagepull: true,
				networkNames:  []string{"test_network"},
				mounts: []mount.Mount{
					{
						Type:   mount.TypeTmpfs,
						Source: "",
						Target: "tmp_fs",
					},
				},
				volumes:        map[string]struct{}{},
				binds:          []string{},
				ports:          []string{},
				env:            []string{},
				containerId:    "",
				readOnlyRootfs: false,
				err:            errDckrCrt,
			},
			userCurrentPatchFunc: userCurrentPatchOKFunc,
			funcBeforeTest:       normalFunc,
		},
	}

	for n, testCase := range cases {
		t.Logf("TestDockerCreate case %d start", n)
		func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			var plist []*mpatch.Patch

			if testCase.userCurrentPatchFunc != nil {
				p := testCase.userCurrentPatchFunc(testCase.item.err)
				if p != nil {
					plist = append(plist, p)
				}
			}

			if testCase.funcBeforeTest != nil {
				p := testCase.funcBeforeTest(ctrl, testCase.item)
				if p != nil {
					plist = append(plist, p)
				}
			}

			containerId, err := DockerCreate(testCase.item.ContainersItems0)
			if !errors.Is(testCase.item.err, err) {
				if err == nil || !strings.Contains(err.Error(), testCase.item.err.Error()) {
					t.Errorf("Expect error not returned. %v", err)
				}
			}
			if containerId != testCase.item.containerId {
				t.Errorf("Unexpect return value: %v", containerId)
			}

			for _, p := range plist {
				unpatch(t, p)
			}
		}()
		t.Logf("TestDockerCreate case %d end", n)
	}

	t.Log("Done")
}

func TestDockerRun(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockWrapperContainer := clientmock.NewMockDockerClientWrapperContainer(ctrl)
	p1, err := mpatch.PatchMethod(GetContainerByName, mockWrapperContainer.GetContainerByName)
	if err != nil {
		t.Fatal(err)
	}
	defer unpatch(t, p1)

	p2, err := mpatch.PatchMethod(StartContainer, mockWrapperContainer.StartContainer)
	if err != nil {
		t.Fatal(err)
	}
	defer unpatch(t, p2)

	mockDockerRunner := clientmock.NewMockDockerInterface(ctrl)
	p3, err := mpatch.PatchMethod(DockerRemove, mockDockerRunner.DockerRemove)
	if err != nil {
		t.Fatal(err)
	}
	defer unpatch(t, p3)

	p4, err := mpatch.PatchMethod(DockerCreate, mockDockerRunner.DockerCreate)
	if err != nil {
		t.Fatal(err)
	}
	defer unpatch(t, p4)

	mockWrapperContainer.EXPECT().GetContainerByName(gomock.Any()).Return(nil, err_docker)
	if err := DockerRun(&api.ContainersItems0{Name: "test_dockerrun", Force: true}); !errors.Is(err, err_docker) {
		t.Error("Expect error not returned.")
	}

	mockWrapperContainer.EXPECT().GetContainerByName(gomock.Any()).Return(&types.Container{}, nil)
	mockDockerRunner.EXPECT().DockerRemove(gomock.Any()).Return(err_docker)
	if err := DockerRun(&api.ContainersItems0{Name: "test_dockerrun", Force: true}); !errors.Is(err, err_docker) {
		t.Error("Expect error not returned.")
	}

	mockDockerRunner.EXPECT().DockerCreate(gomock.Any()).Return("", err_docker)
	if err := DockerRun(&api.ContainersItems0{Name: "test_dockerrun", Force: false}); !errors.Is(err, err_docker) {
		t.Error("Expect error not returned.")
	}

	mockDockerRunner.EXPECT().DockerCreate(gomock.Any()).Return("id", nil)
	mockWrapperContainer.EXPECT().StartContainer(gomock.Any(), gomock.Any(), gomock.Any()).Return(err_docker)
	if err := DockerRun(&api.ContainersItems0{Name: "test_dockerrun", Force: false}); !errors.Is(err, err_docker) {
		t.Error("Expect error not returned.")
	}

	mockDockerRunner.EXPECT().DockerCreate(gomock.Any()).Return("id", nil)
	mockWrapperContainer.EXPECT().StartContainer(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	if err := DockerRun(&api.ContainersItems0{Name: "test_dockerrun", Force: false}); err != nil {
		t.Error("Expect success but failed with err.")
	}

	t.Log("Done")
}

func TestDockerStart(t *testing.T) {
	cases := []struct {
		name          string
		containername string
		expectError   error
	}{
		{
			name:          "valid_container",
			containername: "valid",
			expectError:   nil,
		},
		{
			name:          "invalid_container",
			containername: "invalid",
			expectError:   err_docker,
		},
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockWrapperContainer := clientmock.NewMockDockerClientWrapperContainer(ctrl)
	p, err := mpatch.PatchMethod(StartContainer, mockWrapperContainer.StartContainer)
	if err != nil {
		t.Fatal(err)
	}
	defer unpatch(t, p)

	mockWrapperContainer.EXPECT().StartContainer(gomock.Any(), "valid", gomock.Any()).AnyTimes().Return(nil)
	mockWrapperContainer.EXPECT().StartContainer(gomock.Any(), "invalid", gomock.Any()).AnyTimes().Return(err_docker)

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			container := &api.ContainersItems0{
				Name: tc.containername,
			}

			if err := DockerStart(container); err != nil {
				if errors.Is(err, tc.expectError) {
					t.Log("Output expected.")
				} else {
					t.Error("Output unexpected")
				}
			}
		})
	}

	t.Log("Done")
}

func TestDockerStop(t *testing.T) {
	cases := []struct {
		name          string
		containername string
		expectError   error
	}{
		{
			name:          "valid_container",
			containername: "valid",
			expectError:   nil,
		},
		{
			name:          "invalid_container",
			containername: "invalid",
			expectError:   err_docker,
		},
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockWrapperContainer := clientmock.NewMockDockerClientWrapperContainer(ctrl)
	p, err := mpatch.PatchMethod(StopContainer, mockWrapperContainer.StopContainer)
	if err != nil {
		t.Fatal(err)
	}
	defer unpatch(t, p)

	mockWrapperContainer.EXPECT().StopContainer("valid").AnyTimes().Return(nil)
	mockWrapperContainer.EXPECT().StopContainer("invalid").AnyTimes().Return(err_docker)

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			container := &api.ContainersItems0{
				Name: tc.containername,
			}

			if err := DockerStop(container); err != nil {
				if errors.Is(err, tc.expectError) {
					t.Log("Output expected.")
				} else {
					t.Error("Output unexpected")
				}
			}
		})
	}

	t.Log("Done")
}

func TestDockerRemove(t *testing.T) {
	cases := []struct {
		name          string
		containername string
		expectError   error
	}{
		{
			name:          "valid_container",
			containername: "valid",
			expectError:   nil,
		},
		{
			name:          "invalid_container",
			containername: "invalid",
			expectError:   err_docker,
		},
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockWrapperContainer := clientmock.NewMockDockerClientWrapperContainer(ctrl)
	p, err := mpatch.PatchMethod(RemoveContainer, mockWrapperContainer.RemoveContainer)
	if err != nil {
		t.Fatal(err)
	}
	defer unpatch(t, p)

	mockWrapperContainer.EXPECT().RemoveContainer("valid").AnyTimes().Return(nil)
	mockWrapperContainer.EXPECT().RemoveContainer("invalid").AnyTimes().Return(err_docker)

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			container := &api.ContainersItems0{
				Name: tc.containername,
			}

			if err := DockerRemove(container); err != nil {
				if errors.Is(err, tc.expectError) {
					t.Log("Output expected.")
				} else {
					t.Error("Output unexpected")
				}
			}
		})
	}

	t.Log("Done")
}
