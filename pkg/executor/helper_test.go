/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
//nolint: dupl
package executor

import (
	"context"
	"errors"
	"github.com/intel/edge-conductor/pkg/api/ep"
	pluginapi "github.com/intel/edge-conductor/pkg/api/plugins"
	"github.com/intel/edge-conductor/pkg/eputils/docker"
	repoutils "github.com/intel/edge-conductor/pkg/eputils/repoutils"
	"github.com/intel/edge-conductor/pkg/eputils/restfulcli"
	"io"
	"reflect"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/undefinedlabs/go-mpatch"
	"golang.org/x/crypto/ssh"
)

var (
	errEmpty = errors.New("")
)

var helper_executor = &Executor{
	Execspec: ep.Execspec{},
	tempParams: tempParameter{
		EpParams: pluginapi.EpParams{
			Kitconfig: &pluginapi.Kitconfig{
				Parameters: &pluginapi.KitconfigParameters{
					Customconfig: &pluginapi.Customconfig{
						Registry: &pluginapi.CustomconfigRegistry{
							Password: "123456",
							User:     "admin",
						},
					},
					GlobalSettings: &pluginapi.KitconfigParametersGlobalSettings{
						ProviderIP:   "192.168.1.1",
						RegistryPort: "9000",
					},
				},
			},
		},
	},
	nodesByRole: map[string][]*nodeInfo{
		"day-0": {&nodeInfo{
			ip: "192.168.1.1",
			client: &sshClient{
				client: &ssh.Client{},
			},
		}},
	},
}

var day0_nodes = map[string]*nodeInfo{
	"day-0": {
		name: "",
		ip:   "192.168.1.1",
		client: &sshClient{
			client: &ssh.Client{},
		},
	},
}

var day0_invalid_nodes = map[string]*nodeInfo{
	"day-0": {
		name: "",
		ip:   "192.168.1.100",
		client: &sshClient{
			client: &ssh.Client{},
		},
	},
}

/**
 * Test function helperCreateProjectOnHarbor
 **/
func TestHelperCreateProjectOnHarbor(t *testing.T) {
	var cases = []struct {
		name           string
		executor       *Executor
		expectError    error
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			name:        "success",
			executor:    helper_executor,
			expectError: nil,
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchMapImageURLCreateHarborProject(t, false)
				return []*mpatch.Patch{patch1}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				pList := tc.funcBeforeTest()
				defer unpatchAll(t, pList)
			}

			// Run test code
			err := tc.executor.helperCreateProjectOnHarbor(context.TODO(), nil, nil)

			if !isExpectedError(err, tc.expectError) {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func patchMapImageURLCreateHarborProject(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(restfulcli.MapImageURLCreateHarborProject, func(harborIP, harborPort, harborUser, harborPass string, image []string) ([]string, error) {
		if fail {
			return nil, errEmpty
		} else {
			return []string{"busybox"}, nil
		}
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func TestHelperPushImage(t *testing.T) {
	var cases = []struct {
		name           string
		executor       *Executor
		ctx            context.Context
		nodes          map[string]*nodeInfo
		cmd            []string
		from           *nodeInfo
		fromCmd        []string
		expectError    bool
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			"success",
			helper_executor,
			context.TODO(),
			map[string]*nodeInfo{
				"192.168.1.1": {
					ip: "192.168.1.1",
				},
			},
			[]string{"docker.com"},
			&nodeInfo{},
			[]string{"bash"},
			false,
			func() []*mpatch.Patch {
				patch1 := patchGetAuthConf(t, false)
				patch2 := patchTlsBasicAuth(t)
				patch3 := patchRegistryProjectExists(t, false)
				patch4 := patchRegistryCreateProject(t, false)
				patch5 := patchTagImageToLocal(t, false)
				patch6 := patchImagePush(t, false)
				patch7 := patchMapImageURLCreateHarborProject(t, false)
				return []*mpatch.Patch{patch1, patch2, patch3, patch4, patch5, patch6, patch7}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				pList := tc.funcBeforeTest()
				defer unpatchAll(t, pList)
			}

			err := tc.executor.helperPushImage(tc.ctx, tc.nodes, tc.cmd)
			if (err != nil && !tc.expectError) ||
				(err == nil && tc.expectError) {
				t.Logf("Test case %s failed.", tc.name)
				t.Error(err)
			} else {
				t.Log("Done")
			}
		})
	}
}

func patchImagePush(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(docker.ImagePush, func(imageRef string, authConf *types.AuthConfig) error {
		if fail {
			return errEmpty
		} else {
			return nil
		}

	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func patchPushFileToRepo(t *testing.T, fail bool) (string, *mpatch.Patch) {
	patch, patchErr := mpatch.PatchMethod(repoutils.PushFileToRepo, func(filename, subRef, rev string) (string, error) {
		if fail {
			return "FAIL", errEmpty
		} else {
			return "OK", nil
		}

	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return "FAIL", nil
	}
	return "OK", patch
}

func patchPullFileFromRepo(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(repoutils.PullFileFromRepo, func(filepath string, targeturl string) error {
		if fail {
			return errEmpty
		} else {
			return nil
		}

	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func patchTagImageToLocal(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(docker.TagImageToLocal, func(imageTag, registryURL string) (string, error) {
		if fail {
			return "", errEmpty
		} else {
			return "", nil
		}

	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func patchRegistryCreateProject(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(restfulcli.RegistryProjectExists, func(harborUrl, project, authStr, certFilePath string) (bool, error) {
		if fail {
			return false, errEmpty
		} else {
			return true, nil
		}

	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func patchRegistryProjectExists(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(restfulcli.RegistryCreateProject, func(harborUrl, project, authStr, certFilePath string) error {
		if fail {
			return errEmpty
		} else {
			return nil
		}

	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func patchTlsBasicAuth(t *testing.T) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(restfulcli.TlsBasicAuth, func(username, password string) string {
		return ""

	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch

}

func patchGetAuthConf(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(docker.GetAuthConf, func(server, port, user, password string) (*types.AuthConfig, error) {
		if fail {
			return nil, errEmpty
		} else {
			return &types.AuthConfig{}, nil
		}

	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch

}

func TestHelperPushFile(t *testing.T) {
	var cases = []struct {
		name           string
		executor       *Executor
		ctx            context.Context
		nodes          map[string]*nodeInfo
		cmd            []string
		expectError    bool
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			name:        "PushFile success",
			executor:    helper_executor,
			ctx:         context.TODO(),
			nodes:       day0_nodes,
			cmd:         []string{"/tmp/something", "subRef", "0.0.0"},
			expectError: false,
			funcBeforeTest: func() []*mpatch.Patch {
				_, patch1 := patchPushFileToRepo(t, false)
				return []*mpatch.Patch{patch1}
			},
		},

		{
			name:        "PushFile failure",
			executor:    helper_executor,
			ctx:         context.TODO(),
			nodes:       day0_nodes,
			cmd:         []string{"/tmp/something", "subRef", "0.0.0"},
			expectError: true,
		},

		{
			name:        "PushFile not day0",
			executor:    helper_executor,
			ctx:         context.TODO(),
			nodes:       day0_invalid_nodes,
			cmd:         []string{"/tmp/something", "subRef", "0.0.0"},
			expectError: true,
		},

		{
			name:        "PushFile invalid command",
			executor:    helper_executor,
			ctx:         context.TODO(),
			nodes:       day0_nodes,
			expectError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				pList := tc.funcBeforeTest()
				defer unpatchAll(t, pList)
			}

			err := tc.executor.helperPushFile(tc.ctx, tc.nodes, tc.cmd)
			if (err != nil && !tc.expectError) ||
				(err == nil && tc.expectError) {
				t.Logf("Test case %s failed.", tc.name)
				t.Error(err)
			} else {
				t.Log("Done")
			}
		})
	}
}

func TestHelperPullFile(t *testing.T) {
	var cases = []struct {
		name           string
		executor       *Executor
		ctx            context.Context
		nodes          map[string]*nodeInfo
		cmd            []string
		expectError    bool
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			name:        "PullFile success",
			executor:    helper_executor,
			ctx:         context.TODO(),
			nodes:       day0_nodes,
			cmd:         []string{"/tmp/something", "subRef", "0.0.0"},
			expectError: false,
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchPullFileFromRepo(t, false)
				return []*mpatch.Patch{patch1}
			},
		},

		{
			name:        "PullFile failure",
			executor:    helper_executor,
			ctx:         context.TODO(),
			nodes:       day0_nodes,
			cmd:         []string{"/tmp/something", "subRef", "0.0.0"},
			expectError: true,
		},

		{
			name:        "PullFile not day0",
			executor:    helper_executor,
			ctx:         context.TODO(),
			nodes:       day0_invalid_nodes,
			cmd:         []string{"/tmp/something", "subRef", "0.0.0"},
			expectError: true,
		},

		{
			name:        "PullFile invalid command",
			executor:    helper_executor,
			ctx:         context.TODO(),
			nodes:       day0_nodes,
			expectError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				pList := tc.funcBeforeTest()
				defer unpatchAll(t, pList)
			}

			err := tc.executor.helperPullFile(tc.ctx, tc.nodes, tc.cmd)
			if (err != nil && !tc.expectError) ||
				(err == nil && tc.expectError) {
				t.Logf("Test case %s failed.", tc.name)
				t.Error(err)
			} else {
				t.Log("Done")
			}
		})
	}
}

func TestHelperShell(t *testing.T) {
	var cases = []struct {
		name           string
		ctx            context.Context
		nodes          map[string]*nodeInfo
		cmd            []string
		expectError    bool
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			name:        "success",
			expectError: false,
			ctx:         context.TODO(),
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchConnect(t, false)
				patch3 := patchCmdWithAttachIO(t, false)
				patch4 := patchDisconnect(t, false)
				return []*mpatch.Patch{patch1, patch3, patch4}
			},
			cmd: []string{"bash"},
			nodes: map[string]*nodeInfo{
				"192.168.1.1": {
					ip: "192.168.1.1",
					client: &sshClient{
						client: &ssh.Client{},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				pList := tc.funcBeforeTest()
				defer unpatchAll(t, pList)
			}

			err := New().helperShell(tc.ctx, tc.nodes, tc.cmd)
			if (err != nil && !tc.expectError) ||
				(err == nil && tc.expectError) {
				t.Logf("Test case %s failed.", tc.name)
				t.Error(err)
			} else {
				t.Log("Done")
			}
		})
	}
}

func TestRunPipeTo(t *testing.T) {
	var cases = []struct {
		name           string
		ctx            context.Context
		nodes          map[string]*nodeInfo
		cmd            []string
		to             *nodeInfo
		toCmd          []string
		expectError    bool
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			name:        "success",
			expectError: false,
			ctx:         context.TODO(),
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchConnect(t, false)
				patch3 := patchCmdWithAttachIO(t, false)
				patch4 := patchDisconnect(t, false)
				return []*mpatch.Patch{patch1, patch3, patch4}
			},
			to: &nodeInfo{
				client: &sshClient{
					client: &ssh.Client{},
				},
			},
			nodes: map[string]*nodeInfo{
				"192.168.1.1": {
					name: "",
					ip:   "192.168.1.1",
					client: &sshClient{
						client: &ssh.Client{},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				pList := tc.funcBeforeTest()
				defer unpatchAll(t, pList)
			}

			err := New().runPipeTo(tc.ctx, tc.nodes, tc.cmd, tc.to, tc.toCmd)
			if (err != nil && !tc.expectError) ||
				(err == nil && tc.expectError) {
				t.Logf("Test case %s failed.", tc.name)
				t.Error(err)
			} else {
				t.Log("Done")
			}
		})
	}
}

func TestRunPipeFrom(t *testing.T) {
	var cases = []struct {
		name           string
		ctx            context.Context
		nodes          map[string]*nodeInfo
		cmd            []string
		from           *nodeInfo
		fromCmd        []string
		expectError    bool
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			name:        "success",
			expectError: false,
			ctx:         context.TODO(),
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchConnect(t, false)
				patch3 := patchCmdWithAttachIO(t, false)
				patch4 := patchDisconnect(t, false)
				return []*mpatch.Patch{patch1, patch3, patch4}
			},
			from: &nodeInfo{
				client: &sshClient{
					client: &ssh.Client{},
				},
			},
			nodes: map[string]*nodeInfo{
				"192.168.1.1": {
					ip: "192.168.1.1",
					client: &sshClient{
						client: &ssh.Client{},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				pList := tc.funcBeforeTest()
				defer unpatchAll(t, pList)
			}

			err := New().runPipeFrom(tc.ctx, tc.nodes, tc.cmd, tc.from, tc.fromCmd)
			if (err != nil && !tc.expectError) ||
				(err == nil && tc.expectError) {
				t.Logf("Test case %s failed.", tc.name)
				t.Error(err)
			} else {
				t.Log("Done")
			}
		})
	}
}

func patchDisconnect(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchInstanceMethodByName(reflect.TypeOf(&sshClient{}), "Disconnect", func(s *sshClient) error {
		if fail {
			return errEmpty
		} else {
			return nil
		}

	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func patchCmdWithAttachIO(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchInstanceMethodByName(reflect.TypeOf(&sshClient{}), "CmdWithAttachIO", func(s *sshClient, ctx context.Context, cmd []string, stdin io.Reader, stdout, stderr io.Writer, tty bool) error {
		if fail {
			return errEmpty
		} else {
			return nil
		}

	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func patchConnect(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchInstanceMethodByName(reflect.TypeOf(&sshClient{}), "Connect", func(s *sshClient) error {
		if fail {
			return errEmpty
		} else {
			return nil
		}

	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func TestHelperhelperCopyToDay0(t *testing.T) {
	var cases = []struct {
		name           string
		ctx            context.Context
		nodes          map[string]*nodeInfo
		cmd            []string
		expectError    bool
		executor       Executor
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			"connect failed",
			context.TODO(),
			day0_nodes,
			[]string{"/tmp/something", "/tmp/"},
			true,
			Executor{
				Execspec:   ep.Execspec{},
				tempParams: tempParameter{},
				nodesByIP:  nil,
				nodesByRole: map[string][]*nodeInfo{
					"day-0": {&nodeInfo{
						client: &sshClient{
							client: &ssh.Client{},
						}},
					},
				},
			},
			func() []*mpatch.Patch {
				patch1 := patchConnect(t, true)
				patch3 := patchCmdWithAttachIO(t, false)
				patch4 := patchDisconnect(t, false)
				return []*mpatch.Patch{patch1, patch3, patch4}
			},
		},
		{
			"command failed",
			context.TODO(),
			day0_nodes,
			[]string{"/tmp/something", "/tmp/"},
			true,
			Executor{
				Execspec:   ep.Execspec{},
				tempParams: tempParameter{},
				nodesByRole: map[string][]*nodeInfo{
					"day-0": {&nodeInfo{

						client: &sshClient{
							client: &ssh.Client{},
						}},
					},
				},
			},
			func() []*mpatch.Patch {
				patch1 := patchConnect(t, false)
				patch3 := patchCmdWithAttachIO(t, true)
				patch4 := patchDisconnect(t, false)
				return []*mpatch.Patch{patch1, patch3, patch4}
			},
		},
		{
			"disconnect failed",
			context.TODO(),
			day0_nodes,
			[]string{"/tmp/something", "/tmp/"},
			true,
			Executor{
				Execspec:   ep.Execspec{},
				tempParams: tempParameter{},
				nodesByIP:  nil,
				nodesByRole: map[string][]*nodeInfo{
					"day-0": {&nodeInfo{
						client: &sshClient{
							client: &ssh.Client{},
						}},
					},
				},
			},
			func() []*mpatch.Patch {
				patch1 := patchConnect(t, false)
				patch3 := patchCmdWithAttachIO(t, false)
				patch4 := patchDisconnect(t, true)
				return []*mpatch.Patch{patch1, patch3, patch4}
			},
		},
		{
			"success",
			context.TODO(),
			day0_nodes,
			[]string{"/tmp/something", "/tmp/"},
			false,
			Executor{
				Execspec:   ep.Execspec{},
				tempParams: tempParameter{},
				nodesByIP:  nil,
				nodesByRole: map[string][]*nodeInfo{
					"day-0": {&nodeInfo{
						client: &sshClient{
							client: &ssh.Client{},
						}},
					},
				},
			},
			func() []*mpatch.Patch {
				patch1 := patchConnect(t, false)
				patch3 := patchCmdWithAttachIO(t, false)
				patch4 := patchDisconnect(t, false)
				return []*mpatch.Patch{patch1, patch3, patch4}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				pList := tc.funcBeforeTest()
				defer unpatchAll(t, pList)
			}

			err := tc.executor.helperCopyToDay0(tc.ctx, tc.nodes, tc.cmd)
			if (err != nil && !tc.expectError) ||
				(err == nil && tc.expectError) {
				t.Logf("Test case %s failed.", tc.name)
				t.Error(err)
			} else {
				t.Log("Done")
			}
		})
	}
}

func TestHelperCopyFromDay0(t *testing.T) {
	var cases = []struct {
		name           string
		ctx            context.Context
		nodes          map[string]*nodeInfo
		cmd            []string
		expectError    bool
		executor       Executor
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			"success",
			context.TODO(),
			day0_nodes,
			[]string{"/tmp/something", "/tmp/"},
			false,
			Executor{
				Execspec:   ep.Execspec{},
				tempParams: tempParameter{},
				nodesByIP:  nil,
				nodesByRole: map[string][]*nodeInfo{
					"day-0": {&nodeInfo{
						client: &sshClient{
							client: &ssh.Client{},
						}},
					},
				},
			},
			func() []*mpatch.Patch {
				patch1 := patchConnect(t, false)
				patch3 := patchCmdWithAttachIO(t, false)
				patch4 := patchDisconnect(t, false)
				return []*mpatch.Patch{patch1, patch3, patch4}
			},
		},
		{
			"connect failed",
			context.TODO(),
			day0_nodes,
			[]string{"/tmp/something", "/tmp/"},
			true,
			Executor{
				Execspec:   ep.Execspec{},
				tempParams: tempParameter{},
				nodesByIP:  nil,
				nodesByRole: map[string][]*nodeInfo{
					"day-0": {&nodeInfo{
						client: &sshClient{
							client: &ssh.Client{},
						}},
					},
				},
			},
			func() []*mpatch.Patch {
				patch1 := patchConnect(t, true)
				patch3 := patchCmdWithAttachIO(t, false)
				patch4 := patchDisconnect(t, false)
				return []*mpatch.Patch{patch1, patch3, patch4}
			},
		},
		{
			"command failed",
			context.TODO(),
			day0_nodes,
			[]string{"/tmp/something", "/tmp/"},
			true,
			Executor{
				Execspec:   ep.Execspec{},
				tempParams: tempParameter{},
				nodesByIP:  nil,
				nodesByRole: map[string][]*nodeInfo{
					"day-0": {&nodeInfo{
						client: &sshClient{
							client: &ssh.Client{},
						}},
					},
				},
			},
			func() []*mpatch.Patch {
				patch1 := patchConnect(t, false)
				patch3 := patchCmdWithAttachIO(t, true)
				patch4 := patchDisconnect(t, false)
				return []*mpatch.Patch{patch1, patch3, patch4}
			},
		},
		{
			"disconnect failed",
			context.TODO(),
			day0_nodes,
			[]string{"/tmp/something", "/tmp/"},
			true,
			Executor{
				Execspec:   ep.Execspec{},
				tempParams: tempParameter{},
				nodesByIP:  nil,
				nodesByRole: map[string][]*nodeInfo{
					"day-0": {&nodeInfo{
						client: &sshClient{
							client: &ssh.Client{},
						}},
					},
				},
			},
			func() []*mpatch.Patch {
				patch1 := patchConnect(t, false)
				patch3 := patchCmdWithAttachIO(t, false)
				patch4 := patchDisconnect(t, true)
				return []*mpatch.Patch{patch1, patch3, patch4}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				pList := tc.funcBeforeTest()
				defer unpatchAll(t, pList)
			}

			err := tc.executor.helperCopyFromDay0(tc.ctx, tc.nodes, tc.cmd)
			if (err != nil && !tc.expectError) ||
				(err == nil && tc.expectError) {
				t.Logf("Test case %s failed.", tc.name)
				t.Error(err)
			} else {
				t.Log("Done")
			}
		})
	}
}
