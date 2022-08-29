/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package executor

import (
	"context"
	"ep/pkg/api/ep"
	pluginapi "ep/pkg/api/plugins"
	"ep/pkg/eputils"
	"errors"
	"io"
	"io/ioutil"
	"os/user"
	"reflect"
	"strings"
	"testing"

	"github.com/undefinedlabs/go-mpatch"
	"golang.org/x/crypto/ssh"
	"sigs.k8s.io/yaml"
)

var (
	errExEmpty = errors.New("")
)

func isExpectedError(returnErr error, wantError error) bool {
	if !errors.Is(returnErr, wantError) &&
		(returnErr == nil || wantError == nil || !strings.Contains(returnErr.Error(), wantError.Error())) {
		return false
	}
	return true
}

func patchUserCurrent(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(user.Current, func() (*user.User, error) {
		if fail {
			return nil, errExEmpty
		} else {
			return &user.User{}, nil
		}
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

/**
 * Test function StringOverrideWithNode
 **/
func TestStringOverrideWithNode(t *testing.T) {
	var cases = []struct {
		name string

		executor *Executor
		s        string
		ni       *nodeInfo

		expectError    error
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			name: "success",
			s:    `\{\{test\}\}`,
			executor: func() *Executor {
				return &Executor{
					Execspec: ep.Execspec{},
					tempParams: tempParameter{
						EpParams: pluginapi.EpParams{
							Kitconfig: &pluginapi.Kitconfig{
								Parameters: &pluginapi.KitconfigParameters{
									Nodes: []*pluginapi.Node{{
										IP: "192.168.1.1",
									}},
								},
							},
						},
					},
					nodesByIP: map[string]*nodeInfo{
						"192.168.1.1": {
							name:   "192.168.1.1",
							ip:     "192.168.1.1",
							client: &sshClient{},
						},
					},
					nodesByRole: map[string][]*nodeInfo{},
				}
			}(),
			ni: &nodeInfo{
				name: "192.168.1.1",
				ip:   "192.168.1.1",
				client: &sshClient{
					host:     "192.168.1.1",
					user:     "sysadmin",
					password: "sysadmin",
					key:      "ssh-rsa",
					port:     22,
					config:   &ssh.ClientConfig{},
					client:   &ssh.Client{},
				},
			},
			expectError: nil,
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchStringTemplateConvertWithParams(t, false)
				patch2 := patchUserCurrent(t, false)
				return []*mpatch.Patch{patch1, patch2}
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

			_, err := tc.executor.StringOverrideWithNode(tc.s, tc.ni)

			if !isExpectedError(err, tc.expectError) {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestNew(t *testing.T) {
	cases := []struct {
		name           string
		funcBeforeTest func()
	}{
		{
			name: "Generate a new instance of executor",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				tc.funcBeforeTest()
			}

			New()

		})
	}

}

// Test update node information
func TestNodeListUpdate(t *testing.T) {
	cases := []struct {
		name               string
		expectError        bool
		kitconfigparameter *pluginapi.KitconfigParameters
		funcBeforeTest     func() []*mpatch.Patch
	}{
		{
			name:        "Readfile failed",
			expectError: true,
			kitconfigparameter: &pluginapi.KitconfigParameters{
				Nodes: []*pluginapi.Node{
					{},
				},
			},
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchReadFile(t, false)
				return []*mpatch.Patch{patch1}
			},
		},
		{
			name:        "Node IP is empty, ignore",
			expectError: true,
			kitconfigparameter: &pluginapi.KitconfigParameters{
				Nodes: []*pluginapi.Node{{
					IP: "192.168.1.1",
				},
				},
			},
		},
		{
			name:        "Has Node IP",
			expectError: true,
			kitconfigparameter: &pluginapi.KitconfigParameters{
				Nodes: []*pluginapi.Node{{
					IP: "192.168.1.1",
				},
				},
			},
		},
		{
			name:        "Has User",
			expectError: true,
			kitconfigparameter: &pluginapi.KitconfigParameters{
				Nodes: []*pluginapi.Node{{
					IP:   "192.168.1.1",
					User: "sysadmin",
				},
				},
			},
		},
		{
			name:        "Has User/sshkey",
			expectError: true,
			kitconfigparameter: &pluginapi.KitconfigParameters{
				Nodes: []*pluginapi.Node{{
					IP:     "192.168.1.1",
					SSHKey: "xxxxxxxxxxxxxxxx",
					User:   "sysadmin",
				},
				},
			},
		},
		{
			name:        "Has User/sshkey/role",
			expectError: true,
			kitconfigparameter: &pluginapi.KitconfigParameters{
				Nodes: []*pluginapi.Node{{
					IP:     "192.168.1.1",
					Role:   []string{"controlplane", "etcd"},
					SSHKey: "xxxxxxxxxxxxxxxx",
					User:   "sysadmin",
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

			executor := Executor{
				Execspec:   ep.Execspec{},
				tempParams: tempParameter{},
				nodesByIP: map[string]*nodeInfo{
					"node1": {
						name:   "node1",
						ip:     "10.10.10.2",
						client: nil,
					},
				},
				nodesByRole: map[string][]*nodeInfo{
					"controleplane": {
						&nodeInfo{
							name:   "controlplane",
							ip:     "10.10.10.254",
							client: nil,
						},
					},
				},
			}

			_ = executor.NodeListUpdate(tc.kitconfigparameter)

		})
	}
}

func patchReadFile(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(ioutil.ReadFile, func(filename string) ([]byte, error) {
		if fail {
			return []byte(``), errExEmpty
		} else {
			return []byte(``), nil
		}
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func TestLoadSpecFromString(t *testing.T) {
	var cases = []struct {
		name           string
		specStr        string
		expectError    bool
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			name:        "UnmarshalBinary fail",
			expectError: true,
			specStr:     "",
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchStringTemplateConvertWithParams(t, false)
				patch2 := patchYAMLToJSON(t, false)
				patch3 := patchUnmarshalBinary(t, true)
				return []*mpatch.Patch{patch1, patch2, patch3}
			},
		},
		{
			name:        "YAMLToJSON fail",
			expectError: true,
			specStr:     "",
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchStringTemplateConvertWithParams(t, false)
				patch2 := patchYAMLToJSON(t, true)
				patch3 := patchUnmarshalBinary(t, false)
				return []*mpatch.Patch{patch1, patch2, patch3}
			},
		},
		{
			name:        "StringTemplateConvertWithParams fail",
			expectError: true,
			specStr:     "",
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchStringTemplateConvertWithParams(t, true)
				patch2 := patchYAMLToJSON(t, false)
				patch3 := patchUnmarshalBinary(t, false)
				return []*mpatch.Patch{patch1, patch2, patch3}
			},
		},
		{
			name:        "success",
			expectError: false,
			specStr:     "",
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchStringTemplateConvertWithParams(t, false)
				patch2 := patchYAMLToJSON(t, false)
				patch3 := patchUnmarshalBinary(t, false)
				return []*mpatch.Patch{patch1, patch2, patch3}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				pList := tc.funcBeforeTest()
				defer unpatchAll(t, pList)
			}

			err := New().LoadSpecFromString(tc.specStr)
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

func patchUnmarshalBinary(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchInstanceMethodByName(reflect.TypeOf(&ep.Execspec{}), "UnmarshalBinary", func(e *ep.Execspec, b []byte) error {
		if fail {
			return errExEmpty
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

func patchYAMLToJSON(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(yaml.YAMLToJSON, func(y []byte) ([]byte, error) {
		if fail {
			return []byte(``), errExEmpty
		} else {
			return []byte(``), nil
		}
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func patchStringTemplateConvertWithParams(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(eputils.StringTemplateConvertWithParams, func(str string, tempParams interface{}) (string, error) {
		if fail {
			return "", errExEmpty
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

func TestSetECParams(t *testing.T) {
	var cases = []struct {
		name           string
		epparams       *pluginapi.EpParams
		expectError    bool
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			name:        "Convert schema struct failed",
			expectError: true,
			epparams: &pluginapi.EpParams{
				Kitconfig: &pluginapi.Kitconfig{
					Parameters: &pluginapi.KitconfigParameters{},
				},
			},
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchConvertSchemaStruct(t, true)
				patch2 := patchNodeListUpdate(t, false)
				return []*mpatch.Patch{patch1, patch2}
			},
		},
		{
			name:        "default",
			expectError: false,
			epparams: &pluginapi.EpParams{
				Kitconfig: &pluginapi.Kitconfig{
					Parameters: &pluginapi.KitconfigParameters{},
				},
			},
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchConvertSchemaStruct(t, false)
				patch2 := patchNodeListUpdate(t, false)
				return []*mpatch.Patch{patch1, patch2}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				pList := tc.funcBeforeTest()
				defer unpatchAll(t, pList)
			}

			err := New().SetECParams(tc.epparams)
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

func patchNodeListUpdate(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchInstanceMethodByName(reflect.TypeOf(New()), "NodeListUpdate", func(e *Executor, params *pluginapi.KitconfigParameters) error {
		if fail {
			return errExEmpty
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

func patchConvertSchemaStruct(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(eputils.ConvertSchemaStruct, func(from interface{}, to interface{}) error {
		if fail {
			return errExEmpty
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

func TestLoadSpecFromFile(t *testing.T) {
	var cases = []struct {
		name           string
		specFile       string
		expectError    bool
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			name:        "LoadSpecFromString failed",
			specFile:    "",
			expectError: true,
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchReadFile(t, false)
				patch2 := patchCheckHashForContent(t, false)
				patch3 := patchLoadSpecFromString(t, true)
				return []*mpatch.Patch{patch1, patch2, patch3}
			},
		},
		{
			name:        "CheckHashForContent failed",
			specFile:    "",
			expectError: true,
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchReadFile(t, false)
				patch2 := patchCheckHashForContent(t, true)
				patch3 := patchLoadSpecFromString(t, false)
				return []*mpatch.Patch{patch1, patch2, patch3}
			},
		},
		{
			name:        "ReadFile failed",
			specFile:    "",
			expectError: true,
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchReadFile(t, true)
				patch2 := patchCheckHashForContent(t, false)
				patch3 := patchLoadSpecFromString(t, false)
				return []*mpatch.Patch{patch1, patch2, patch3}
			},
		},
		{
			name:        "success",
			specFile:    "",
			expectError: false,
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchReadFile(t, false)
				patch2 := patchCheckHashForContent(t, false)
				patch3 := patchLoadSpecFromString(t, false)
				return []*mpatch.Patch{patch1, patch2, patch3}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				pList := tc.funcBeforeTest()
				defer unpatchAll(t, pList)
			}

			err := New().LoadSpecFromFile(tc.specFile)
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

func patchLoadSpecFromString(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchInstanceMethodByName(reflect.TypeOf(New()), "LoadSpecFromString", func(e *Executor, specStr string) error {
		if fail {
			return errExEmpty
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

func patchCheckHashForContent(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(eputils.CheckHashForContent, func(content []byte, target string, workspace string) error {
		if fail {
			return errExEmpty
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

func TestExecutorRun(t *testing.T) {
	var cases = []struct {
		name           string
		ctx            context.Context
		expectError    bool
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			name:        "success",
			expectError: false,
			ctx:         nil,
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchBackground(t)
				patch2 := patchRunWithAttachIO(t, false)
				return []*mpatch.Patch{patch1, patch2}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				pList := tc.funcBeforeTest()
				defer unpatchAll(t, pList)
			}

			err := New().Run(tc.ctx)
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

func patchRunWithAttachIO(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchInstanceMethodByName(reflect.TypeOf(New()), "RunWithAttachIO", func(e *Executor, ctx context.Context, stdin io.Reader, stdout, stderr io.Writer) error {
		if fail {
			return errExEmpty
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

func patchBackground(t *testing.T) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(context.Background, func() context.Context {
		return nil
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func TestRunWithAttachIO(t *testing.T) {
	var cases = []struct {
		name  string
		ctx   context.Context
		stdin io.Reader
		stdout,
		stderr io.Writer
		expectError    bool
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			name:        "unknown command",
			expectError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				pList := tc.funcBeforeTest()
				defer unpatchAll(t, pList)
			}

			executor := &Executor{
				Execspec: ep.Execspec{

					Spec: &ep.ExecspecSpec{
						Steps: []*ep.ExecspecSpecStepsItems0{
							{
								Commands: []*ep.ExecspecSpecStepsItems0CommandsItems0{
									{
										Cmd: []string{"tail"},
									},
								},
								Nodes: &ep.ExecspecSpecStepsItems0Nodes{
									AnyOf:  []string{"controlplane"},
									AllOf:  []string{"worker"},
									NoneOf: []string{"day-0"},
								},
							},
							{
								Nodes: &ep.ExecspecSpecStepsItems0Nodes{},
							},
						},
					},
				},
				tempParams: tempParameter{},
				nodesByRole: map[string][]*nodeInfo{
					"controlplane": {
						&nodeInfo{
							name: "node1",
							ip:   "192.168.1.1",
						},
						&nodeInfo{
							name: "node2",
							ip:   "192.168.1.2",
						},
					},
				},
			}

			err := executor.RunWithAttachIO(tc.ctx, tc.stdin, tc.stdout, tc.stderr)
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
