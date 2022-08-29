/*
* Copyright (c) 2021 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Template auto-generated once, maintained by plugin owner.

package capiproviderlaunch

import (
	pluginapi "ep/pkg/api/plugins"
	eputils "ep/pkg/eputils"
	capiutils "ep/pkg/eputils/capiutils"
	repoutils "ep/pkg/eputils/repoutils"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/undefinedlabs/go-mpatch"
)

var (
	errGeneral = errors.New("error")
)

func Test_generateLocalPRPath(t *testing.T) {
	type args struct {
		file   *pluginapi.FilesItems0
		config *ClusterCtlConfig
	}
	tests := []struct {
		name       string
		args       args
		wantTarget string
	}{
		{
			name: "BootstrapProvider",
			args: args{
				file: &pluginapi.FilesItems0{
					URL:       "BootstrapProvider.Label",
					Mirrorurl: "/cluster-api/BootstrapProvider.Label",
				},
				config: &ClusterCtlConfig{
					CoreProvider: &ProviderConfig{
						Name:  "kubeadm",
						Label: "test",
					},
					ControlPlaneProvider: &ProviderConfig{
						Name:  "kubeadm",
						Label: "test",
					},
					InfrastructureProvider: &ProviderConfig{
						Name:  "kubeadm",
						Label: "test",
					},
					BootstrapProvider: &ProviderConfig{
						Name:    "kubeadm",
						Label:   "/cluster-api/BootstrapProvider.Label",
						Version: "v0.1",
					},
					Binaries: &BinariesConfig{
						Name: "kubeadm",
						Url:  "https://test/kubeadm.tar.gz",
					},
					RuntimeDir: "/tmp",
				},
			},
			wantTarget: "/tmp/cluster-api/BootstrapProvider.Label/v0.1/BootstrapProvider.Label",
		},
		{
			name: "ControlPlaneProvider",
			args: args{
				file: &pluginapi.FilesItems0{
					URL:       "ControlPlaneProvider.Label",
					Mirrorurl: "/cluster-api/ControlPlaneProvider.Label",
				},
				config: &ClusterCtlConfig{
					CoreProvider: &ProviderConfig{
						Name:  "kubeadm",
						Label: "test",
					},
					BootstrapProvider: &ProviderConfig{
						Name:  "kubeadm",
						Label: "test",
					},
					ControlPlaneProvider: &ProviderConfig{
						Name:    "kubeadm",
						Label:   "/cluster-api/ControlPlaneProvider.Label",
						Version: "v0.1",
					},
					InfrastructureProvider: &ProviderConfig{
						Name:  "kubeadm",
						Label: "test",
					},
					Binaries: &BinariesConfig{
						Name: "kubeadm",
						Url:  "https://test/kubeadm.tar.gz",
					},
					RuntimeDir: "/tmp",
				},
			},
			wantTarget: "/tmp/cluster-api/ControlPlaneProvider.Label/v0.1/ControlPlaneProvider.Label",
		},
		{
			name: "InfrastructureProvider",
			args: args{
				file: &pluginapi.FilesItems0{
					URL:       "InfrastructureProvider.Label",
					Mirrorurl: "/cluster-api/InfrastructureProvider.Label",
				},
				config: &ClusterCtlConfig{
					CoreProvider: &ProviderConfig{
						Name:  "kubeadm",
						Label: "test",
					},
					BootstrapProvider: &ProviderConfig{
						Name:  "kubeadm",
						Label: "test",
					},
					ControlPlaneProvider: &ProviderConfig{
						Name:  "kubeadm",
						Label: "test",
					},
					InfrastructureProvider: &ProviderConfig{
						Name:    "kubeadm",
						Label:   "/cluster-api/InfrastructureProvider.Label",
						Version: "v0.1",
					},
					Binaries: &BinariesConfig{
						Name: "kubeadm",
						Url:  "https://test/kubeadm.tar.gz",
					},
					RuntimeDir: "/tmp",
				},
			},
			wantTarget: "/tmp/cluster-api/InfrastructureProvider.Label/v0.1/InfrastructureProvider.Label",
		},
		{
			name: "cert-manager",
			args: args{
				file: &pluginapi.FilesItems0{
					URL:       "cert-manager",
					Mirrorurl: "cert-manager",
				},
				config: &ClusterCtlConfig{
					CoreProvider: &ProviderConfig{
						Name:  "kubeadm",
						Label: "test",
					},
					BootstrapProvider: &ProviderConfig{
						Name:  "kubeadm",
						Label: "test",
					},
					ControlPlaneProvider: &ProviderConfig{
						Name:  "kubeadm",
						Label: "test",
					},
					InfrastructureProvider: &ProviderConfig{
						Name:  "kubeadm",
						Label: "test",
					},
					CertManager: &CertManagerConfig{
						Version: "v0.1",
					},
					Binaries: &BinariesConfig{
						Name: "kubeadm",
						Url:  "https://test/kubeadm.tar.gz",
					},
					RuntimeDir: "/tmp",
				},
			},
			wantTarget: "/tmp/cert-manager/v0.1/cert-manager",
		},
		{
			name: "oras",
			args: args{
				file: &pluginapi.FilesItems0{
					URL:       "oras.tar.gz",
					Mirrorurl: "oras",
				},
				config: &ClusterCtlConfig{
					CoreProvider: &ProviderConfig{
						Name:  "kubeadm",
						Label: "test",
					},
					BootstrapProvider: &ProviderConfig{
						Name:  "kubeadm",
						Label: "test",
					},
					ControlPlaneProvider: &ProviderConfig{
						Name:  "kubeadm",
						Label: "test",
					},
					InfrastructureProvider: &ProviderConfig{
						Name:  "kubeadm",
						Label: "test",
					},
					CertManager: &CertManagerConfig{
						Version: "v0.1",
					},
					Binaries: &BinariesConfig{
						Name:    "oras",
						Version: "v0.1",
						Url:     "https://test/oras.tar.gz",
					},
					RuntimeDir: "/tmp",
				},
			},
			wantTarget: "/tmp/oras/oras/v0.1/oras.tar.gz",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotTarget := generateLocalPRPath(tt.args.file, tt.args.config); gotTarget != tt.wantTarget {
				t.Errorf("generateLocalPRPath() = %v, want %v", gotTarget, tt.wantTarget)
			}
		})
	}
}

func TestPluginMain(t *testing.T) {
	func_patch_os_file := func() []*mpatch.Patch {
		patch1, _ := mpatch.PatchMethod(os.Chmod, func(string, os.FileMode) error { return nil })
		patch2, _ := mpatch.PatchMethod(eputils.WriteStringToFile, func(string, string) error { return nil })
		patch3, _ := mpatch.PatchMethod(eputils.FileTemplateConvert, func(string, string) error { return nil })
		patch4, _ := mpatch.PatchMethod(eputils.RunCMD, func(*exec.Cmd) (string, error) { return "", nil })
		patch5, _ := mpatch.PatchMethod(eputils.CreateFolderIfNotExist, func(string) error { return nil })
		return []*mpatch.Patch{patch1, patch2, patch3, patch4, patch5}
	}
	func_patch_create_folder := func() []*mpatch.Patch {
		patch1, _ := mpatch.PatchMethod(eputils.CreateFolderIfNotExist, func(string) error { return nil })
		return []*mpatch.Patch{patch1}
	}
	func_CertManage_config_lost := func() []*mpatch.Patch {
		patch1, err := mpatch.PatchMethod(capiutils.GetCapiClusterProviderConfig, func(manifest []*pluginapi.ClustermanifestCapiClusterProvidersItems0, capi_cluster_name string) (*pluginapi.ClustermanifestCapiClusterProvidersItems0, error) {
			return &pluginapi.ClustermanifestCapiClusterProvidersItems0{CertManager: nil}, nil
		})
		if err != nil {
			t.Fatal(err)
		}

		return []*mpatch.Patch{patch1}
	}

	func_providerParameters_lost := func() []*mpatch.Patch {
		patchGetCapiClusterProviderConfig, err := mpatch.PatchMethod(capiutils.GetCapiClusterProviderConfig, func(manifest []*pluginapi.ClustermanifestCapiClusterProvidersItems0, capi_cluster_name string) (*pluginapi.ClustermanifestCapiClusterProvidersItems0, error) {
			return &pluginapi.ClustermanifestCapiClusterProvidersItems0{
				CertManager: &pluginapi.ClustermanifestCapiClusterProvidersItems0CertManager{
					Version: "v0.1",
				},
				Binaries: []*pluginapi.ClustermanifestCapiClusterProvidersItems0BinariesItems0{
					{
						Name: "testBinaries",
					},
				},
				Providers: []*pluginapi.ProviderItems0{
					{
						Name:       "testName",
						Parameters: nil,
					},
				},
			}, nil
		})
		if err != nil {
			t.Fatal(err)
		}

		return []*mpatch.Patch{patchGetCapiClusterProviderConfig}
	}
	func_failed_CopyFile := func() []*mpatch.Patch {
		patches := func_patch_os_file()
		patch, _ := mpatch.PatchMethod(eputils.CopyFile, func(dstName string, srcName string) (written int64, err error) {
			return 0, errGeneral
		})
		patches = append(patches, patch)
		return patches
	}

	func_failed_pull_clusterctl := func() []*mpatch.Patch {
		patches := func_patch_os_file()
		patchCopyFile, _ := mpatch.PatchMethod(eputils.CopyFile, func(dstName string, srcName string) (written int64, err error) {
			return 1, nil
		})
		patchPullFileFromRepo, _ := mpatch.PatchMethod(repoutils.PullFileFromRepo, func(file string, url string) error {
			if strings.Contains(file, "clusterctl") {
				return errGeneral
			} else {
				return nil
			}
		})
		patches = append(patches, patchCopyFile, patchPullFileFromRepo)
		return patches
	}

	cases := []struct {
		name           string
		input          map[string][]byte
		expectError    bool
		expectErrorMsg string
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			name: "Kitconfigs_lost",
			input: map[string][]byte{
				"ep-params": nil,
			},
			funcBeforeTest: func_patch_create_folder,
			expectError:    true,
			expectErrorMsg: eputils.GetError("errIncorrectParam").Error(),
		},
		{
			name: "Infra_provider_lost_in_Kit_config",
			input: map[string][]byte{
				"ep-params": []byte(`{"kitconfig": {"Cluster": {"provider": "clusterapi"}}}`),
				"cluster-manifest": []byte(`{
					"capi_cluster_providers":[
						{
							"name": "byoh",
						 	"images": ["test:test"]
						}
					]
				}`),
			},
			funcBeforeTest: func_patch_create_folder,
			expectError:    true,
			expectErrorMsg: eputils.GetError("errProvider").Error(),
		},
		{
			name: "CAPI_manifest_lost",
			input: map[string][]byte{
				"ep-params": []byte(`{"kitconfig": {"Cluster": {"provider": "clusterapi"}, "Parameters": {"extensions": ["capi-metal3"]}}}`),
			},
			funcBeforeTest: func_patch_create_folder,
			expectError:    true,
			expectErrorMsg: eputils.GetError("errCAPIManifest").Error(),
		},
		{
			name: "CertManage_config_lost",
			input: map[string][]byte{
				"ep-params": []byte(`{"kitconfig": {"Cluster": {"provider": "clusterapi"}, "Parameters": {"extensions": ["capi-metal3"]}}}`),
				"cluster-manifest": []byte(`{
					"capi_cluster_providers":[
						{
							"name": "byoh",
						 	"images": ["test:test"]
						}
					]
				}`),
			},
			funcBeforeTest: func_CertManage_config_lost,
			expectError:    true,
			expectErrorMsg: eputils.GetError("errCAPIProvider").Error(),
		},
		{
			name: "providerParameters_lost",
			input: map[string][]byte{
				"ep-params": []byte(`{"kitconfig": {"Cluster": {"provider": "clusterapi"}, "Parameters": {"extensions": ["capi-metal3"]}}}`),
				"cluster-manifest": []byte(`{
					"capi_cluster_providers":[
						{
							"name": "byoh",
						 	"images": ["test:test"]
						}
					]
				}`),
			},
			funcBeforeTest: func_providerParameters_lost,
			expectError:    true,
			expectErrorMsg: eputils.GetError("errCAPIProvider").Error(),
		},

		{
			name: "Provider lost",
			input: map[string][]byte{
				"ep-params": []byte(`{"kitconfig": {"Cluster": {"provider": "clusterapi"}, "Parameters": {"extensions": ["capi-metal3"]}}}`),
				"cluster-manifest": []byte(`{
					"capi_cluster_providers":[
						{
							"name": "byoh",
						 	"images": ["test:test"]
						}
					]
				}`),
			},
			funcBeforeTest: func_patch_create_folder,
			expectError:    true,
			expectErrorMsg: eputils.GetError("errCAPIProvider").Error(),
		},
		{
			name: "config in manefest lost",
			input: map[string][]byte{
				"ep-params": []byte(`{"kitconfig": {"Cluster": {"provider": "clusterapi"}, "Parameters": {"extensions": ["capi-metal3"]}}}`),
				"cluster-manifest": []byte(`{
					"capi_cluster_providers":[
						{
							"name": "byoh",
						 	"images": ["test:test"]
						}
					]
				}`),
			},
			funcBeforeTest: func_patch_create_folder,
			expectError:    true,
			expectErrorMsg: eputils.GetError("errCAPIProvider").Error(),
		},
		{
			name: "Provider parameter lost.",
			input: map[string][]byte{
				"ep-params": []byte(`{"kitconfig": {"Cluster": {"provider": "clusterapi"}, "Parameters": {"extensions": ["capi-metal3"]}}}`),
				"cluster-manifest": []byte(`{
					"capi_cluster_providers":[
						{
							"name": "byoh",
						 	"images": ["test:test"],
							 "providers": [{
								"provider_type": "CoreProvider",
								"name": ""
							}],
							"cert-manager": {"url": "bbb/cert-manager.yaml"}
						}
					]
				}`),
			},
			funcBeforeTest: func_patch_create_folder,
			expectError:    true,
			expectErrorMsg: eputils.GetError("errCAPIProvider").Error(),
		},
		{
			name: "Provider config lost.",
			input: map[string][]byte{
				"ep-params": []byte(`{"kitconfig": {"Cluster": {"provider": "clusterapi"}, "Parameters": {"extensions": ["capi-metal3"]}}}`),
				"cluster-manifest": []byte(`{
					"capi_cluster_providers":[
						{
							"name": "byoh",
						 	"images": ["test:test"],
							 "providers": [{
								"provider_type": "CoreProvider",
								"name": ""
							}],
							"cert-manager": {"url": "bbb/cert-manager.yaml"}
						}
					]
				}`),
			},
			funcBeforeTest: func_patch_create_folder,
			expectError:    true,
			expectErrorMsg: eputils.GetError("errCAPIProvider").Error(),
		},
		{
			name: "config of kitconfig missing",
			input: map[string][]byte{
				"ep-params": []byte(`{"kitconfig": {"Cluster": {"provider": "clusterapi"}, "Parameters": {"extensions": ["capi-metal3"]}}}`),
				"cluster-manifest": []byte(`{
					"capi_cluster_providers":[
						{
							"name": "metal3",
						 	"images": ["test:test"],
							"binaries": [
								{
									"name": "oras",
									"url": "https://test/oras.tar.gz"
								}
							],
							"providers": [
								{
									"provider_type": "CoreProvider",
									"name": "",
									"parameters" : {"version": "", "provider_label": ""}
								},
								{
									"provider_type": "BootstrapProvider",
									"name": "",
									"parameters" : {"version": "", "provider_label": ""}
								},
								{
									"provider_type": "ControlPlaneProvider",
									"name": "",
									"parameters" : {"version": "", "provider_label": ""}
								},
								{
									"provider_type": "InfrastructureProvider",
									"name": "",
									"parameters" : {"version": "", "provider_label": ""}
								}
							],
							"cert-manager": {"url": "bbb/cert-manager.yaml"}
						}
					],
					"cluster_providers":[{"name":"kind","images":[{"name":"img_node","repo_tag":""},{"name":"img_haproxy","repo_tag":""}],"binaries":[{"name":"kindtool","url":"","sha256":""}]}]}`),
				"files": []byte(`{"files":[{"url":"core-provider","mirrorurl": "/cluster-api/core-provider"}, {"mirrorurl": "capi/kind"}]}`),
			},
			funcBeforeTest: func_patch_os_file,
			expectError:    true,
			expectErrorMsg: eputils.GetError("errStartMgmtClster").Error(),
		},
		{
			name: "Failed to Copy clusterMgrKubeconfig",
			input: map[string][]byte{
				"ep-params": []byte(`{"kitconfig": {"Cluster": {"provider": "clusterapi"}, "Parameters": {"extensions": ["capi-metal3"], "global_settings": {"provider_ip": "", "registry_port": ""}}}, "extensions": [{"name": "capi-metal3", "extension": {"extension": [{"name": "Infra-provider", "config": [{"name": "Management-cluster-kubeconfig", "value": "test"}]}]}}]}`),
				"cluster-manifest": []byte(`{
					"capi_cluster_providers":[
						{
							"name": "metal3",
						 	"images": ["test:test"],
							"binaries": [
								{
									"name": "oras",
									"url": "https://test/oras.tar.gz"
								}
							],
							"providers": [
								{
									"provider_type": "CoreProvider",
									"name": "",
									"parameters" : {"version": "", "provider_label": ""}
								},
								{
									"provider_type": "BootstrapProvider",
									"name": "",
									"parameters" : {"version": "", "provider_label": ""}
								},
								{
									"provider_type": "ControlPlaneProvider",
									"name": "",
									"parameters" : {"version": "", "provider_label": ""}
								},
								{
									"provider_type": "InfrastructureProvider",
									"name": "",
									"parameters" : {"version": "", "provider_label": ""}
								}
							],
							"cert-manager": {"url": "bbb/cert-manager.yaml"}
						}
					],
					"cluster_providers":[{"name":"kind","images":[{"name":"img_node","repo_tag":""},{"name":"img_haproxy","repo_tag":""}],"binaries":[{"name":"kindtool","url":"","sha256":""}]}]}`),
				"files": []byte(`{"files":[{"url":"BootstrapProvider.Label","mirrorurl": "/cluster-api/BootstrapProvider.Label"}, {"mirrorurl": "capi/kind"}]}`),
			},
			funcBeforeTest: func_failed_CopyFile,
			expectError:    true,
			expectErrorMsg: eputils.GetError("errLaunchMgmtClster").Error(),
		},
		{
			name: "Failed to init clusterctl",
			input: map[string][]byte{
				"ep-params": []byte(`{"kitconfig": {"Cluster": {"provider": "clusterapi"}, "Parameters": {"extensions": ["capi-metal3"], "global_settings": {"provider_ip": "", "registry_port": ""}}}}`),
				"cluster-manifest": []byte(`{
					"capi_cluster_providers":[
						{
							"name": "metal3",
						 	"images": ["test:test"],
							"binaries": [
								{
									"name": "oras",
									"url": "https://test/oras.tar.gz"
								}
							],
							"providers": [
								{
									"provider_type": "CoreProvider",
									"name": "",
									"parameters" : {"version": "", "provider_label": ""}
								},
								{
									"provider_type": "BootstrapProvider",
									"name": "",
									"parameters" : {"version": "", "provider_label": ""}
								},
								{
									"provider_type": "ControlPlaneProvider",
									"name": "",
									"parameters" : {"version": "", "provider_label": ""}
								},
								{
									"provider_type": "InfrastructureProvider",
									"name": "",
									"parameters" : {"version": "", "provider_label": ""}
								}
							],
							"cert-manager": {"url": "bbb/cert-manager.yaml"}
						}
					],
					"cluster_providers":[{"name":"kind","images":[{"name":"img_node","repo_tag":""},{"name":"img_haproxy","repo_tag":""}],"binaries":[{"name":"kindtool","url":"","sha256":""}]}]}`),
				"files": []byte(`{"files":[{"url":"core-provider","mirrorurl": "/cluster-api/core-provider"}, {"mirrorurl": "capi/kind"}]}`),
			},
			funcBeforeTest: func_failed_pull_clusterctl,
			expectError:    true,
			expectErrorMsg: eputils.GetError("errInitClusterctl").Error(),
		},
		{
			name: "provider launch success",
			input: map[string][]byte{
				"ep-params": []byte(`{"kitconfig": {"Cluster": {"provider": "clusterapi"}, "Parameters": {"extensions": ["capi-metal3"], "global_settings": {"provider_ip": "", "registry_port": ""}}}, "extensions": [{"name": "capi-metal3", "extension": {"extension": [{"name": "Infra-provider", "config": [{"name": "Management-cluster-kubeconfig", "value": ""}]}]}}]}`),
				"cluster-manifest": []byte(`{
					"capi_cluster_providers":[
						{
							"name": "metal3",
						 	"images": ["test:test"],
							"binaries": [
								{
									"name": "oras",
									"url": "https://test/oras.tar.gz"
								}
							],
							"providers": [
								{
									"provider_type": "CoreProvider",
									"name": "",
									"parameters" : {"version": "", "provider_label": ""}
								},
								{
									"provider_type": "BootstrapProvider",
									"name": "",
									"parameters" : {"version": "", "provider_label": ""}
								},
								{
									"provider_type": "ControlPlaneProvider",
									"name": "",
									"parameters" : {"version": "", "provider_label": ""}
								},
								{
									"provider_type": "InfrastructureProvider",
									"name": "",
									"parameters" : {"version": "", "provider_label": ""}
								}
							],
							"cert-manager": {"url": "bbb/cert-manager.yaml"}
						}
					],
					"cluster_providers":[{"name":"kind","images":[{"name":"img_node","repo_tag":""},{"name":"img_haproxy","repo_tag":""}],"binaries":[{"name":"kindtool","url":"","sha256":""}]}]}`),
				"files": []byte(`{"files":[{"url":"core-provider","mirrorurl": "/cluster-api/core-provider"}, {"mirrorurl": "capi/kind"}, {"url":"kubeadm","mirrorurl": "/bootstrap-kubeadm/kubeadm"}, {"url":"kubeadm","mirrorurl": "/control-plane-kubeadm/kubeadm"}, {"url":"metal3","mirrorurl": "/infrastructure-metal3/metal3"}]}`),
			},
			funcBeforeTest: func_patch_os_file,
			expectError:    false,
			expectErrorMsg: "",
		},
	}

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
					if fmt.Sprint(result) == tc.expectErrorMsg {
						t.Logf("Expected error: {%s} catched, done.", tc.expectErrorMsg)
						return
					} else if tc.expectErrorMsg == "" {
						return
					} else {
						t.Logf("Expected error: {%s}.", tc.expectErrorMsg)
						t.Logf("active error msg: {%s}.", result)
						t.Fatal("Unexpected error occurred.")
					}

				}
				t.Logf("Failed to run PluginMain when input is %s.", tc.input)
				t.Error(result)
			}

			_ = testOutput
		})
	}
}

func Test_generateLocalProviderRepo(t *testing.T) {
	type args struct {
		files             *pluginapi.Files
		clusterctl_config *ClusterCtlConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test11",
			args: args{
				files: &pluginapi.Files{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := generateLocalProviderRepo(tt.args.files, tt.args.clusterctl_config); (err != nil) != tt.wantErr {
				t.Errorf("generateLocalProviderRepo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
