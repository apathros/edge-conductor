/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Template auto-generated once, maintained by plugin owner.

//nolint: dupl
package nodejoinprepare

import (
	"errors"
	pluginapi "github.com/intel/edge-conductor/pkg/api/plugins"
	eputils "github.com/intel/edge-conductor/pkg/eputils"
	kubeutils "github.com/intel/edge-conductor/pkg/eputils/kubeutils"
	nodeutils "github.com/intel/edge-conductor/pkg/eputils/nodeutils"
	"github.com/intel/edge-conductor/pkg/executor"
	"github.com/undefinedlabs/go-mpatch"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

var (
	testError = errors.New("testing")
)

func TestPluginMain(t *testing.T) {
	patch_kubeconfig_failed := func() []*mpatch.Patch {
		patch1, _ := mpatch.PatchMethod(nodeutils.GetKubeConfigContent,
			func(string) (*pluginapi.Filecontent, error) {
				return nil, testError
			})
		return []*mpatch.Patch{patch1}
	}

	patch_download_failed := func() []*mpatch.Patch {
		patch1, _ := mpatch.PatchMethod(nodeutils.GetKubeConfigContent,
			func(string) (*pluginapi.Filecontent, error) {
				return &pluginapi.Filecontent{Content: ""}, nil
			})
		patch2, _ := mpatch.PatchMethod(eputils.DownloadFile,
			func(string, string) error {
				return testError
			})
		return []*mpatch.Patch{patch1, patch2}
	}

	patch_getnode_failed := func() []*mpatch.Patch {
		patch1, _ := mpatch.PatchMethod(nodeutils.GetKubeConfigContent,
			func(string) (*pluginapi.Filecontent, error) {
				return &pluginapi.Filecontent{Content: ""}, nil
			})
		patch2, _ := mpatch.PatchMethod(eputils.DownloadFile,
			func(string, string) error {
				return nil
			})
		patch3, _ := mpatch.PatchMethod(kubeutils.GetNodeList,
			func(*pluginapi.Filecontent, string) (*corev1.NodeList, error) {
				return nil, testError
			})
		return []*mpatch.Patch{patch1, patch2, patch3}
	}
	patch_preprovision_failed := func() []*mpatch.Patch {
		patch1, _ := mpatch.PatchMethod(nodeutils.GetKubeConfigContent,
			func(string) (*pluginapi.Filecontent, error) {
				return &pluginapi.Filecontent{Content: ""}, nil
			})
		patch2, _ := mpatch.PatchMethod(eputils.DownloadFile,
			func(string, string) error {
				return nil
			})
		patch3, _ := mpatch.PatchMethod(kubeutils.GetNodeList,
			func(*pluginapi.Filecontent, string) (*corev1.NodeList, error) {
				return &corev1.NodeList{}, nil
			})
		patch4, _ := mpatch.PatchMethod(nodeutils.GetClusterVersion,
			func(*corev1.NodeList) string {
				return ""
			})
		patch5, _ := mpatch.PatchMethod(nodeutils.GetCRI,
			func(*corev1.NodeList) string {
				return "a://b"
			})
		patch6, _ := mpatch.PatchMethod(executor.Run,
			func(string, *pluginapi.EpParams, interface{}) error {
				return testError
			})
		return []*mpatch.Patch{patch1, patch2, patch3, patch4, patch5, patch6}
	}
	patch_successful := func() []*mpatch.Patch {
		patch1, _ := mpatch.PatchMethod(nodeutils.GetKubeConfigContent,
			func(string) (*pluginapi.Filecontent, error) {
				return &pluginapi.Filecontent{Content: ""}, nil
			})
		patch2, _ := mpatch.PatchMethod(eputils.DownloadFile,
			func(string, string) error {
				return nil
			})
		patch3, _ := mpatch.PatchMethod(kubeutils.GetNodeList,
			func(*pluginapi.Filecontent, string) (*corev1.NodeList, error) {
				return &corev1.NodeList{}, nil
			})
		patch4, _ := mpatch.PatchMethod(nodeutils.GetClusterVersion,
			func(*corev1.NodeList) string {
				return ""
			})
		patch5, _ := mpatch.PatchMethod(nodeutils.GetCRI,
			func(*corev1.NodeList) string {
				return "a://b"
			})
		patch6, _ := mpatch.PatchMethod(executor.Run,
			func(string, *pluginapi.EpParams, interface{}) error {
				return nil
			})
		patch7, _ := mpatch.PatchMethod(nodeutils.FindNodeInClusterByIP,
			func(*corev1.NodeList, string) bool {
				return false
			})
		return []*mpatch.Patch{patch1, patch2, patch3, patch4, patch5, patch6, patch7}
	}

	cases := []struct {
		name                  string
		input, expectedOutput map[string][]byte
		expectError           bool
		funcBeforeTest        func() []*mpatch.Patch
	}{
		{
			name: "get kube config content failed",
			input: map[string][]byte{
				"ep-params": []byte(`{"kubeconfig": ""}`),
			},
			funcBeforeTest: patch_kubeconfig_failed,
			expectError:    true,
		},
		{
			name: "download oras tool failed",
			input: map[string][]byte{
				"ep-params": []byte(`{"kubeconfig": "", "runtimedir":"test", "extensions": [{"name": "capi-a"}] }`),
			},
			funcBeforeTest: patch_download_failed,
			expectError:    true,
		},
		{
			name: "get node list failed",
			input: map[string][]byte{
				"ep-params": []byte(`{"kubeconfig": "", "runtimedir":"test"}`),
			},
			funcBeforeTest: patch_getnode_failed,
			expectError:    true,
		},
		{
			name: "ByohAgent pre-provision failed",
			input: map[string][]byte{
				"ep-params": []byte(`{"kubeconfig": "",
						      "runtimedir":"test",
						      "kitconfig": { "Parameters": {
									    "customconfig": {"registry": {
												"user":"test",
												"password":"test123"
						     }}}}}`),
			},
			funcBeforeTest: patch_preprovision_failed,
			expectError:    true,
		},
		{
			name: "successful",
			input: map[string][]byte{
				"ep-params": []byte(`{"kubeconfig": "",
						      "runtimedir":"test",
						      "kitconfig": { "Parameters": {
									    "customconfig": {"registry": {
												"user":"test",
												"password":"test123"
									    }},
									    "nodes": [{"ip":""},{"ip":"127.0.0.1"}]
						       }}}`),
			},
			funcBeforeTest: patch_successful,
			expectError:    false,
		},
	}

	// Optional: add setup for the test series
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := generateInput(tc.input)
			if input == nil {
				t.Fatalf("Failed to generateInput %s", tc.input)
			}

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

			testOutput := generateOutput(nil)

			if result := PluginMain(input, &testOutput); result != nil {
				if tc.expectError {
					t.Log("Error expected.")
					return
				} else {
					t.Logf("Failed to run PluginMain when input is %s.", tc.input)
					t.Error(result)
				}
			}

			_ = testOutput
		})
	}

	// Optional: add teardown for the test series
}
