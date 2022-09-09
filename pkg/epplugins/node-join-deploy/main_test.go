/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Template auto-generated once, maintained by plugin owner.

//nolint: dupl
package nodejoindeploy

import (
	"errors"
	pluginapi "github.com/intel/edge-conductor/pkg/api/plugins"
	eputils "github.com/intel/edge-conductor/pkg/eputils"
	kubeutils "github.com/intel/edge-conductor/pkg/eputils/kubeutils"
	nodeutils "github.com/intel/edge-conductor/pkg/eputils/nodeutils"
	"github.com/undefinedlabs/go-mpatch"
	"golang.org/x/crypto/ssh"
	"io"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	kubeadmapiv1 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
	kubeadmcmd "k8s.io/kubernetes/cmd/kubeadm/app/cmd"
	cmdutil "k8s.io/kubernetes/cmd/kubeadm/app/cmd/util"
	"testing"
)

var (
	patch_genconfig_failed_loops = 0
	testError                    = errors.New("testing")
)

func TestGetNodeJoinCMD(t *testing.T) {
	patch_client_failed := func() []*mpatch.Patch {
		patch1, _ := mpatch.PatchMethod(kubeutils.ClientFromEPKubeConfig,
			func(*pluginapi.Filecontent) (*kubernetes.Clientset, error) {
				return nil, testError
			})
		return []*mpatch.Patch{patch1}
	}
	patch_createtoken_failed := func() []*mpatch.Patch {
		patch1, _ := mpatch.PatchMethod(kubeutils.ClientFromEPKubeConfig,
			func(*pluginapi.Filecontent) (*kubernetes.Clientset, error) {
				return &kubernetes.Clientset{}, nil
			})
		patch2, _ := mpatch.PatchMethod(cmdutil.DefaultInitConfiguration,
			func() *kubeadmapiv1.InitConfiguration {
				return &kubeadmapiv1.InitConfiguration{}
			})
		patch3, _ := mpatch.PatchMethod(kubeadmcmd.RunCreateToken,
			func(io.Writer, kubernetes.Interface, string, *kubeadmapiv1.InitConfiguration, bool, string, string) error {
				return testError
			})

		return []*mpatch.Patch{patch1, patch2, patch3}
	}

	patch_successful := func() []*mpatch.Patch {
		patch1, _ := mpatch.PatchMethod(kubeutils.ClientFromEPKubeConfig,
			func(*pluginapi.Filecontent) (*kubernetes.Clientset, error) {
				return &kubernetes.Clientset{}, nil
			})
		patch2, _ := mpatch.PatchMethod(cmdutil.DefaultInitConfiguration,
			func() *kubeadmapiv1.InitConfiguration {
				return &kubeadmapiv1.InitConfiguration{}
			})
		patch3, _ := mpatch.PatchMethod(kubeadmcmd.RunCreateToken,
			func(io.Writer, kubernetes.Interface, string, *kubeadmapiv1.InitConfiguration, bool, string, string) error {
				return nil
			})

		return []*mpatch.Patch{patch1, patch2, patch3}
	}

	cases := []struct {
		name           string
		param1         *pluginapi.Filecontent
		param2         string
		expectedOutput string
		expectedErr    bool
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			name:           "Get ClientSet Failed",
			param1:         &pluginapi.Filecontent{},
			param2:         "",
			expectedOutput: "",
			expectedErr:    true,
			funcBeforeTest: patch_client_failed,
		},
		{
			name:           "RunCreateToken",
			param1:         &pluginapi.Filecontent{},
			param2:         "",
			expectedOutput: "",
			expectedErr:    true,
			funcBeforeTest: patch_createtoken_failed,
		},
		{
			name:           "successful",
			param1:         &pluginapi.Filecontent{},
			param2:         "",
			expectedOutput: "",
			expectedErr:    false,
			funcBeforeTest: patch_successful,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				plist := tc.funcBeforeTest()
				for _, p := range plist {
					t.Logf("patch:%v\n", p)
					defer func(t *testing.T, m *mpatch.Patch) {
						err := m.Unpatch()
						if err != nil {
							t.Fatal(err)
						}
					}(t, p)
				}
			}

			result, err := GetNodeJoinCMD(tc.param1, tc.param2)
			if err != nil {
				if tc.expectedErr {
					t.Log("Error expected.")
					return
				} else {
					t.Logf("Failed to run GetNodeJoinCMD when input is %v, %v.", tc.param1, tc.param2)
					t.Error(result)
				}
			}

			if result != tc.expectedOutput {
				t.Logf("The expected value is not match")
			}
		})
	}
}

func TestPluginMain(t *testing.T) {
	patch_kubeconfig_failed := func() []*mpatch.Patch {
		patch1, _ := mpatch.PatchMethod(nodeutils.GetKubeConfigContent,
			func(string) (*pluginapi.Filecontent, error) {
				return nil, testError
			})
		return []*mpatch.Patch{patch1}
	}
	patch_getnode_failed := func() []*mpatch.Patch {
		patch1, _ := mpatch.PatchMethod(nodeutils.GetKubeConfigContent,
			func(string) (*pluginapi.Filecontent, error) {
				return &pluginapi.Filecontent{Content: ""}, nil
			})
		patch2, _ := mpatch.PatchMethod(kubeutils.GetNodeList,
			func(*pluginapi.Filecontent, string) (*corev1.NodeList, error) {
				return nil, testError
			})
		return []*mpatch.Patch{patch1, patch2}
	}

	patch_genconfig_failed := func() []*mpatch.Patch {
		patch1, _ := mpatch.PatchMethod(nodeutils.GetKubeConfigContent,
			func(string) (*pluginapi.Filecontent, error) {
				return &pluginapi.Filecontent{Content: ""}, nil
			})
		patch2, _ := mpatch.PatchMethod(kubeutils.GetNodeList,
			func(*pluginapi.Filecontent, string) (*corev1.NodeList, error) {
				return &corev1.NodeList{}, nil
			})
		patch3, _ := mpatch.PatchMethod(nodeutils.GetCRI,
			func(*corev1.NodeList) string {
				return ""
			})
		patch4, _ := mpatch.PatchMethod(nodeutils.FindNodeInClusterByIP,
			func(*corev1.NodeList, string) bool {
				patch_genconfig_failed_loops++
				if patch_genconfig_failed_loops == 1 {
					return true
				} else {
					return false
				}
			})
		patch5, _ := mpatch.PatchMethod(GetNodeJoinCMD,
			func(*pluginapi.Filecontent, string) (string, error) {
				return "", testError
			})
		patch6, _ := mpatch.PatchMethod(eputils.GenSSHConfig,
			func(*pluginapi.Node) (*ssh.ClientConfig, error) {
				return nil, testError
			})
		return []*mpatch.Patch{patch1, patch2, patch3, patch4, patch5, patch6}
	}
	patch_enable_container_failed := func() []*mpatch.Patch {
		patch1, _ := mpatch.PatchMethod(nodeutils.GetKubeConfigContent,
			func(string) (*pluginapi.Filecontent, error) {
				return &pluginapi.Filecontent{Content: ""}, nil
			})
		patch2, _ := mpatch.PatchMethod(kubeutils.GetNodeList,
			func(*pluginapi.Filecontent, string) (*corev1.NodeList, error) {
				return &corev1.NodeList{}, nil
			})
		patch3, _ := mpatch.PatchMethod(nodeutils.GetCRI,
			func(*corev1.NodeList) string {
				return "containerd"
			})
		patch4, _ := mpatch.PatchMethod(nodeutils.FindNodeInClusterByIP,
			func(*corev1.NodeList, string) bool {
				return false
			})
		patch5, _ := mpatch.PatchMethod(GetNodeJoinCMD,
			func(*pluginapi.Filecontent, string) (string, error) {
				return "", testError
			})
		patch6, _ := mpatch.PatchMethod(eputils.GenSSHConfig,
			func(*pluginapi.Node) (*ssh.ClientConfig, error) {
				return &ssh.ClientConfig{}, nil
			})
		patch7, _ := mpatch.PatchMethod(eputils.RunRemoteCMD,
			func(string, *ssh.ClientConfig, string) error {
				return testError
			})
		return []*mpatch.Patch{patch1, patch2, patch3, patch4, patch5, patch6, patch7}
	}
	patch_successful := func() []*mpatch.Patch {
		patch1, _ := mpatch.PatchMethod(nodeutils.GetKubeConfigContent,
			func(string) (*pluginapi.Filecontent, error) {
				return &pluginapi.Filecontent{Content: ""}, nil
			})
		patch2, _ := mpatch.PatchMethod(kubeutils.GetNodeList,
			func(*pluginapi.Filecontent, string) (*corev1.NodeList, error) {
				return &corev1.NodeList{}, nil
			})
		patch3, _ := mpatch.PatchMethod(nodeutils.GetCRI,
			func(*corev1.NodeList) string {
				return "testing"
			})
		patch4, _ := mpatch.PatchMethod(nodeutils.FindNodeInClusterByIP,
			func(*corev1.NodeList, string) bool {
				return false
			})
		patch5, _ := mpatch.PatchMethod(GetNodeJoinCMD,
			func(*pluginapi.Filecontent, string) (string, error) {
				return "", testError
			})
		patch6, _ := mpatch.PatchMethod(eputils.GenSSHConfig,
			func(*pluginapi.Node) (*ssh.ClientConfig, error) {
				return &ssh.ClientConfig{}, nil
			})
		patch7, _ := mpatch.PatchMethod(eputils.RunRemoteCMD,
			func(string, *ssh.ClientConfig, string) error {
				return nil
			})
		return []*mpatch.Patch{patch1, patch2, patch3, patch4, patch5, patch6, patch7}
	}

	cases := []struct {
		name                  string
		input, expectedOutput map[string][]byte
		expectError           bool
		funcBeforeTest        func() []*mpatch.Patch
	}{
		// TODO: Add the values to complete your test cases.
		// Add the values for input and expectedoutput with particular struct marshal data in json format.
		// They will be used to generate "SchemaMapData" as inputs and expected outputs of plugins under test.
		// if the inputs in the Plugin Input List is not required in your test case, keep the value as nil.
		{
			name: "get kube config content failed",
			input: map[string][]byte{
				"ep-params": []byte(`{"kubeconfig": ""}`),
			},
			funcBeforeTest: patch_kubeconfig_failed,
			expectError:    true,
		},
		{
			name: "get node list failed",
			input: map[string][]byte{
				"ep-params": []byte(`{"kubeconfig": ""}`),
			},
			funcBeforeTest: patch_getnode_failed,
			expectError:    true,
		},
		{
			name: "gen config failed",
			input: map[string][]byte{
				"ep-params": []byte(`{"kubeconfig": "",
						      "kitconfig": { "Parameters": { "nodes": [{"ip":"127.0.0.1"}, {"ip":"127.0.0.1","sshport":22}]
						     }}}`),
			},
			funcBeforeTest: patch_genconfig_failed,
			expectError:    true,
		},
		{
			name: "enable container failed",
			input: map[string][]byte{
				"ep-params": []byte(`{"kubeconfig": "",
						      "kitconfig": { "Parameters": { "nodes": [{"ip":"127.0.0.1", "sshport":22}]
						     }}}`),
			},
			funcBeforeTest: patch_enable_container_failed,
			expectError:    true,
		},
		{
			name: "successful",
			input: map[string][]byte{
				"ep-params": []byte(`{"kubeconfig": "",
						      "kitconfig": { "Parameters": { "nodes": [{"ip":"127.0.0.1", "sshport":22}]
						     }}}`),
			},
			funcBeforeTest: patch_successful,
			expectError:    true,
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
