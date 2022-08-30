/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Auto generated, do not modify.

package test_files

import (
	"github.com/intel/edge-conductor/pkg/api/plugins"
	"testing"
)

func TestGen(t *testing.T) {

	ep_params := plugins.EpParams{
		Cmdline: "??",
		Extensions: []*plugins.EpParamsExtensionsItems0{
			{
				Extension: &plugins.Extension{
					Extension: []*plugins.ExtensionItems0{
						{
							Config: []*plugins.ExtensionItems0ConfigItems0{
								{
									Name:  "",
									Value: "",
								},
							},
							Name: "",
						},
					},
				},
				Name: "",
			},
		},
		Kitconfig: &plugins.Kitconfig{
			Cluster: &plugins.KitconfigCluster{
				Config:             "",
				ExportConfigFolder: "",
				Manifests: []string{
					"",
					"",
				},
				Provider: "???",
			},
			Components: &plugins.KitconfigComponents{
				Manifests: []string{
					"",
				},
				Selector: []*plugins.KitconfigComponentsSelectorItems0{
					{},
				},
			},
			OS: &plugins.KitconfigOS{
				Config:    "?",
				Manifests: []string{""},
				Provider:  "",
			},

			Parameters: &plugins.KitconfigParameters{
				Customconfig: &plugins.Customconfig{
					Cluster: &plugins.Cluster{
						ManagementCluster: &plugins.ClusterManagementCluster{
							Capath: "",
							Endpoint: &plugins.ClusterManagementClusterEndpoint{
								ApiserverPort: "",
								IP:            "",
								Password:      "",
								Token:         "",
								User:          "",
							},
						},
						Namespace: "",
						WorkCluster: &plugins.ClusterWorkCluster{
							Apiserver:  "",
							Cacerthash: "",
							Controller: &plugins.Node{
								BmcEndpoint: "",
								BmcPassword: "",
								BmcProtocol: "",
								BmcUser:     "",
								Critype:     "",
								IP:          "",
								Mac:         "",
								Name:        "",
								Role: []string{
									"",
									"",
								},
								SSHKey:     "",
								SSHKeyPath: "",
								SSHPasswd:  "",
								SSHPort:    0,
								User:       "",
							},
							Discoverytoken: "",
							Name:           "",
						},
					},
					Ironic: &plugins.CustomconfigIronic{
						Dhcprange:               "",
						Httpport:                "",
						Ironicinspectorpassword: "",
						Ironicinspectoruser:     "",
						Ironicpassword:          "",
						Ironicuser:              "",
						Kubeconfigpath:          "",
						Mariadbpassword:         "",
						Provisioninginterface:   "",
						Provisioningip:          "",
					},
					Registry: &plugins.CustomconfigRegistry{
						Capath:      "",
						Externalurl: "",
						Password:    "",
						User:        "",
					},
					Resources: []*plugins.CustomconfigResourcesItems0{},
				},
				DefaultSSHKeyPath: "",
				Extensions: []string{
					"capi-byoh",
				},
				GlobalSettings: &plugins.KitconfigParametersGlobalSettings{
					DNSServer: []string{
						"",
						"",
					},
					HTTPProxy:    "",
					HTTPSProxy:   "",
					NoProxy:      "",
					NtpServer:    "",
					ProviderIP:   "",
					RegistryPort: "",
					WorkflowPort: "",
				},
				Nodes: []*plugins.Node{
					{
						BmcEndpoint: "",
						BmcPassword: "",
						BmcProtocol: "",
						BmcUser:     "",
						Critype:     "",
						IP:          "",
						Mac:         "",
						Name:        "",
						Role: []string{
							"controlplane",
							"etcd",
							"worker",
						},
						SSHKey:     "",
						SSHKeyPath: "",
						SSHPasswd:  "",
						SSHPort:    0,
						User:       "",
					},
				},
			},
			Use: []string{
				"",
			},
		},
		Kitconfigpath: "testkitconfig",
		Kubeconfig:    "testkubeconfig",
		Registrycert: &plugins.Certificate{
			Ca: &plugins.CertificateCa{
				Cert: "",
				Csr:  "",
				Key:  "",
			},
			Client: &plugins.CertificateClient{
				Cert: "",
				Csr:  "",
				Key:  "",
			},
			Name: "",
			Server: &plugins.CertificateServer{
				Cert: "",
				Csr:  "",
				Key:  "",
			},
		},
		Runtimebin:  "",
		Runtimedata: "",
		Runtimedir:  "",
		User:        "",
		Workspace:   "",
	}

	r, _ := ep_params.MarshalBinary()
	_ = r

	n := "byoh"
	cluster_manifest := plugins.Clustermanifest{
		CapiClusterProviders: []*plugins.ClustermanifestCapiClusterProvidersItems0{
			{
				Binaries: []*plugins.ClustermanifestCapiClusterProvidersItems0BinariesItems0{
					{
						Name:    "",
						Sha256:  "",
						URL:     "",
						Version: "",
					},
				},
				CertManager: &plugins.ClustermanifestCapiClusterProvidersItems0CertManager{
					URL:     "",
					Version: "",
				},
				Images: []string{
					"", "",
				},
				Name: &n,
				Providers: plugins.Provider{
					{
						Name: "",
						Parameters: &plugins.ProviderItems0Parameters{
							Metadata:      "",
							ProviderLabel: "",
							Version:       "",
						},
						ProviderType: "",
						URL:          "",
					},
				},
				Runtime: new(string),
			},
		},
		ClusterProviders: []*plugins.ClustermanifestClusterProvidersItems0{
			{
				Binaries: []*plugins.ClustermanifestClusterProvidersItems0BinariesItems0{
					{
						Name:   "",
						Sha256: "",
						URL:    "",
					},
				},
				Images: []*plugins.ClustermanifestClusterProvidersItems0ImagesItems0{{
					Name:    "",
					RepoTag: "",
				}},
				Name:            "",
				Registrystorage: "",
				Resources: []*plugins.ClustermanifestClusterProvidersItems0ResourcesItems0{{
					Name:  "",
					Value: "",
				}},
				Version: "",
			},
		},
	}

	r, _ =
		cluster_manifest.MarshalBinary()
	_ = r
}
