/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package nodeutils

import (
	"errors"
	pluginapi "github.com/intel/edge-conductor/pkg/api/plugins"
	"github.com/undefinedlabs/go-mpatch"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

var (
	testError = errors.New("testing")
)

func TestFindNodeInClusterByIP(t *testing.T) {
	cases := []struct {
		name           string
		nodelist       *corev1.NodeList
		ip             string
		expectedOutput bool
	}{
		{
			name:           "nodelist is nil",
			nodelist:       nil,
			expectedOutput: false,
		},
		{
			name: "find successful",
			nodelist: &corev1.NodeList{
				Items: []corev1.Node{
					{
						Status: corev1.NodeStatus{
							Addresses: []corev1.NodeAddress{
								{Address: "192.168.1.1"},
							},
						},
					},
				},
			},
			ip:             "192.168.1.1",
			expectedOutput: true,
		},
		{
			name:           "not find",
			nodelist:       &corev1.NodeList{},
			expectedOutput: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := FindNodeInClusterByIP(tc.nodelist, tc.ip)
			if result != tc.expectedOutput {
				t.Errorf("Unexpected output: %v, should be %v", result, tc.expectedOutput)
			} else {
				return
			}
		})
	}
}

//nolint: dupl
func TestGetClusterVersion(t *testing.T) {
	cases := []struct {
		name           string
		nodelist       *corev1.NodeList
		expectedOutput string
	}{
		{
			name:           "nodelist is nil",
			nodelist:       nil,
			expectedOutput: "",
		},
		{
			name: "get version",
			nodelist: &corev1.NodeList{
				Items: []corev1.Node{
					{
						Status: corev1.NodeStatus{
							NodeInfo: corev1.NodeSystemInfo{
								KubeletVersion: "1.1",
							},
						},
					},
				},
			},
			expectedOutput: "1.1",
		},
		{
			name:           "no items",
			nodelist:       &corev1.NodeList{},
			expectedOutput: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := GetClusterVersion(tc.nodelist)
			if result != tc.expectedOutput {
				t.Errorf("Unexpected output: %v, should be %v", result, tc.expectedOutput)
			} else {
				return
			}
		})
	}
}

//nolint: dupl
func TestGetCRI(t *testing.T) {
	cases := []struct {
		name           string
		nodelist       *corev1.NodeList
		expectedOutput string
	}{
		{
			name:           "nodelist is nil",
			nodelist:       nil,
			expectedOutput: "",
		},
		{
			name: "get version",
			nodelist: &corev1.NodeList{
				Items: []corev1.Node{
					{
						Status: corev1.NodeStatus{
							NodeInfo: corev1.NodeSystemInfo{
								ContainerRuntimeVersion: "1.9",
							},
						},
					},
				},
			},
			expectedOutput: "1.9",
		},
		{
			name:           "no items",
			nodelist:       &corev1.NodeList{},
			expectedOutput: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := GetCRI(tc.nodelist)
			if result != tc.expectedOutput {
				t.Errorf("Unexpected output: %v, should be %v", result, tc.expectedOutput)
			} else {
				return
			}
		})
	}

}

func TestGetKubeConfigContent(t *testing.T) {
	patch_read_failed := func() []*mpatch.Patch {
		patch1, _ := mpatch.PatchMethod(ioutil.ReadFile,
			func(string) ([]byte, error) {
				return nil, testError
			})
		return []*mpatch.Patch{patch1}
	}
	patch_read_successful := func() []*mpatch.Patch {
		patch1, _ := mpatch.PatchMethod(ioutil.ReadFile,
			func(string) ([]byte, error) {
				return []byte(`hello`), nil
			})
		return []*mpatch.Patch{patch1}
	}
	cases := []struct {
		name           string
		kubeConfigFile string
		expectedOutput *pluginapi.Filecontent
		expextedErr    bool
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			name:           "failed to read file",
			kubeConfigFile: "",
			expectedOutput: nil,
			expextedErr:    true,
			funcBeforeTest: patch_read_failed,
		},
		{
			name:           "read successful",
			kubeConfigFile: "",
			expectedOutput: &pluginapi.Filecontent{
				Content: "hello",
			},
			expextedErr:    false,
			funcBeforeTest: patch_read_successful,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				plist := tc.funcBeforeTest()
				for _, p := range plist {
					defer func(t *testing.T, m *mpatch.Patch) {
						err := m.Unpatch()
						if err != nil {
							t.Fatal(err)
						}
					}(t, p)
				}
			}
			result, err := GetKubeConfigContent(tc.kubeConfigFile)
			if err != nil {
				if tc.expextedErr {
					t.Log("Error expected.")
				} else {
					t.Error(result)
				}
			} else {
				if tc.expextedErr {
					t.Error(result)
				} else {
					t.Log("Error expected.")
				}
			}
			if result != nil {
				if result.Content != tc.expectedOutput.Content {
					t.Errorf("Unexpected output %v  %v\n", result.Content, tc.expectedOutput.Content)
				}
			} else if result != tc.expectedOutput {
				t.Errorf("Unexpected output\n")
			}
		})
	}
}
