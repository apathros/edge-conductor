/*
* Copyright (c) 2021 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

//nolint: dupl
package capiutils

import (
	"bytes"
	pluginapi "ep/pkg/api/plugins"
	eputils "ep/pkg/eputils"
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/undefinedlabs/go-mpatch"
)

var (
	errTest = errors.New("test_error")
)

var (
	fakeByoh       = "capi-byoh"
	fakeMetal3     = "capi-metal3"
	fakeCONFIGBYOH = "byoh"
	fakeRuntime    = "default"
)

func unpatch(t *testing.T, m *mpatch.Patch) {
	err := m.Unpatch()
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetCapiTemplate(t *testing.T) {
	func_ConvertSchemaStruct_err := func(ctrl *gomock.Controller) []*mpatch.Patch {
		p, err := mpatch.PatchMethod(eputils.ConvertSchemaStruct, func(from interface{}, to interface{}) error { return errTest })
		if err != nil {
			t.Fatal(err)
		}
		return []*mpatch.Patch{p}
	}
	type args struct {
		epparams *pluginapi.EpParams
		setting  pluginapi.CapiSetting
		cp       *CapiTemplate
	}
	tests := []struct {
		name           string
		args           args
		wantErr        error
		funcBeforeTest func(*gomock.Controller) []*mpatch.Patch
	}{
		{
			name: "ConvertSchemaStruct_err",
			args: args{
				&pluginapi.EpParams{},
				pluginapi.CapiSetting{},
				&CapiTemplate{
					pluginapi.EpParams{},
					pluginapi.CapiSetting{},
				},
			},
			wantErr:        errTest,
			funcBeforeTest: func_ConvertSchemaStruct_err,
		},
		{
			// TODO: Add test cases.
			name: "Get Capi Template",
			args: args{
				&pluginapi.EpParams{
					Cmdline:   "",
					Kitconfig: &pluginapi.Kitconfig{},
					Extensions: []*pluginapi.EpParamsExtensionsItems0{
						{
							Name: "capi-metal3",
							Extension: &pluginapi.Extension{
								Extension: []*pluginapi.ExtensionItems0{
									{
										Name: "Infra-provider",
										Config: []*pluginapi.ExtensionItems0ConfigItems0{
											{
												Name:  "Management-cluster-kubeconfig",
												Value: "",
											},
											{
												Name:  "default",
												Value: "default",
											},
										},
									},
								},
							},
						},
					},
					Registrycert: &pluginapi.Certificate{},
				},
				pluginapi.CapiSetting{},
				&CapiTemplate{
					pluginapi.EpParams{},
					pluginapi.CapiSetting{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			if tt.funcBeforeTest != nil {
				plist := tt.funcBeforeTest(ctrl)
				for _, p := range plist {
					defer unpatch(t, p)
				}
			}

			if err := GetCapiTemplate(tt.args.epparams, tt.args.setting, tt.args.cp); !errors.Is(err, tt.wantErr) {
				t.Errorf("GetCapiTemplate() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				t.Log("Error expected.")
			}
		})
	}
}

func Test_generatorByohBinariesItem(t *testing.T) {
	type args struct {
		binariesItem *pluginapi.ClustermanifestCapiClusterProvidersItems0BinariesItems0
	}
	tests := []struct {
		name string
		args args
		want *pluginapi.CapiSettingBinariesItems0
	}{
		{
			name: "return CapiSettingBinaries",
			args: args{
				&pluginapi.ClustermanifestCapiClusterProvidersItems0BinariesItems0{
					Name:    "Test",
					Version: "v0,1",
					URL:     "httpmytest.com",
				},
			},
			want: &pluginapi.CapiSettingBinariesItems0{
				Name:    "Test",
				Version: "v0,1",
				URL:     "httpmytest.com",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generatorByohBinariesItem(tt.args.binariesItem); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generatorByohBinariesItem() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCapiClusterProviderConfig(t *testing.T) {
	type args struct {
		manifest          []*pluginapi.ClustermanifestCapiClusterProvidersItems0
		capi_cluster_name string
	}
	tests := []struct {
		name    string
		args    args
		want    *pluginapi.ClustermanifestCapiClusterProvidersItems0
		wantErr error
	}{
		// TODO: Add test cases.
		{
			name: "Manifest lost",
			args: args{
				manifest:          []*pluginapi.ClustermanifestCapiClusterProvidersItems0{},
				capi_cluster_name: "byoh",
			},
			want:    nil,
			wantErr: errCAPIProviderConfig,
		},
		{
			name: "Get_clusterProviderItem_ok",
			args: args{
				manifest: []*pluginapi.ClustermanifestCapiClusterProvidersItems0{
					{
						Name: &fakeByoh,
					},
				},
				capi_cluster_name: "capi-byoh",
			},
			want: &pluginapi.ClustermanifestCapiClusterProvidersItems0{
				Name: &fakeByoh,
			},
			wantErr: nil,
		},
		{
			name: "Get_clusterProviderItem_err",
			args: args{
				manifest: []*pluginapi.ClustermanifestCapiClusterProvidersItems0{
					{
						Name: &fakeMetal3,
					},
				},
				capi_cluster_name: "capi-byoh",
			},
			want:    nil,
			wantErr: errCAPIProviderConfig,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetCapiClusterProviderConfig(tt.args.manifest, tt.args.capi_cluster_name)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("GetCapiClusterProviderConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCapiClusterProviderConfig() = %v, want %v", got, tt.want)
			}
			t.Logf("{%s} Expected error catched, done.", tt.name)
		})
	}
}

func TestGetManifestConfigNameByCapiInfraProvider(t *testing.T) {
	tests := []struct {
		name string
		args CapiInfraProvider
		want string
	}{
		// TODO: Add test cases.
		{
			name: "metal3",
			args: "capi-metal3",
			want: "metal3",
		},
		{
			name: "byoh",
			args: "capi-byoh",
			want: "byoh",
		},
		{
			name: "defalut",
			args: "defalut",
			want: "",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := GetManifestConfigNameByCapiInfraProvider(tc.args); got != tc.want {
				t.Errorf("GetManifestConfigNameByCapiInfraProvider() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestGetManagementClusterKubeconfig(t *testing.T) {
	var cases = []struct {
		name           string
		ecparams       *pluginapi.EpParams
		wantConfigPath string
	}{
		// TODO: Add test cases.
		{
			name: "Get Management Cluster Kubeconfig",
			ecparams: &pluginapi.EpParams{
				Cmdline:   "",
				Kitconfig: &pluginapi.Kitconfig{},
				Extensions: []*pluginapi.EpParamsExtensionsItems0{
					{
						Name: "capi-metal3",
						Extension: &pluginapi.Extension{
							Extension: []*pluginapi.ExtensionItems0{
								{
									Name: "Infra-provider",
									Config: []*pluginapi.ExtensionItems0ConfigItems0{
										{
											Name:  "Management-cluster-kubeconfig",
											Value: "",
										},
										{
											Name:  "default",
											Value: "default",
										},
									},
								},
							},
						},
					},
				},
				Registrycert: &pluginapi.Certificate{},
			},
			wantConfigPath: "m_kubeconfig",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if gotConfigPath := GetManagementClusterKubeconfig(tc.ecparams); gotConfigPath != tc.wantConfigPath {
				t.Errorf("GetManagementClusterKubeconfig() = %v, want %v", gotConfigPath, tc.wantConfigPath)
			}
		})
	}
}

func TestGetInfraProvider(t *testing.T) {
	type args struct {
		inputKitconfig *pluginapi.Kitconfig
	}
	tests := []struct {
		name         string
		args         args
		wantProvider CapiInfraProvider
		wantErr      bool
	}{
		// TODO: Add test cases.
		{
			name: "Get InfraProvider_nil",
			args: args{
				&pluginapi.Kitconfig{},
			},
			wantProvider: "",
			wantErr:      true,
		},
		{
			name: "Get InfraProvider_ok",
			args: args{
				&pluginapi.Kitconfig{
					Parameters: &pluginapi.KitconfigParameters{
						Extensions: []string{
							"test",
						},
					},
				},
			},
			wantProvider: "",
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotProvider, err := GetInfraProvider(tt.args.inputKitconfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetInfraProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotProvider != tt.wantProvider {
				t.Errorf("GetInfraProvider() = %v, want %v", gotProvider, tt.wantProvider)
			}
		})
	}
}

var (
	inputClustermanifest = &pluginapi.Clustermanifest{
		CapiClusterProviders: []*pluginapi.ClustermanifestCapiClusterProvidersItems0{
			{
				Name:        &fakeCONFIGBYOH,
				Runtime:     &fakeRuntime,
				Providers:   pluginapi.Provider{},
				CertManager: &pluginapi.ClustermanifestCapiClusterProvidersItems0CertManager{},
				Binaries: []*pluginapi.ClustermanifestCapiClusterProvidersItems0BinariesItems0{
					{
						Name:    "defalut",
						Version: "0.3",
						URL:     "10.10.10.10",
					},
				},
				Images: []string{"helloworld", "busybox"},
			},
		},
	}
	inputCapiClusterConfig = &pluginapi.CapiClusterConfig{
		WorkloadCluster: &pluginapi.CapiClusterConfigWorkloadCluster{
			Name:      "defalut",
			Namespace: "defalutNamespace",
		},
	}
	inputCapiSetting = &pluginapi.CapiSetting{
		InfraProvider: &pluginapi.CapiSettingInfraProvider{
			WorkloadClusterControlplaneEndpoint: "",
			WorkloadClusterName:                 "",
			WorkloadClusterNamespace:            "",
		},
	}
	inputCapiSetting_Ironic = &pluginapi.CapiSetting{
		InfraProvider: &pluginapi.CapiSettingInfraProvider{
			WorkloadClusterControlplaneEndpoint: "",
			WorkloadClusterName:                 "",
			WorkloadClusterNamespace:            "",
		},
		IronicConfig: &pluginapi.CapiSettingIronicConfig{
			IronicProvisionNic: "",
			IronicProvisionIP:  "",
			IronicDhcpRange:    "",
			IronicHTTPPort:     "",
			IronicOsImage:      "",
		},
	}
)

func TestGetCapiSetting(t *testing.T) {
	type args struct {
		epparams        *pluginapi.EpParams
		clusterManifest *pluginapi.Clustermanifest
		clusterConfig   *pluginapi.CapiClusterConfig
		setting         *pluginapi.CapiSetting
	}
	tests := []struct {
		name               string
		args               args
		genExpectedSetting func(*pluginapi.CapiSetting) *pluginapi.CapiSetting
	}{
		{
			// 			// TODO: Add test cases.epparams.Kitconfig.Parameters.Customconfig.Registry
			name: "Get Capi-byoh Setting_CONFIG_WORKLOAD_CLUSTER_RUNTIME",
			args: args{
				&pluginapi.EpParams{
					Cmdline: "",
					Kitconfig: &pluginapi.Kitconfig{
						Parameters: &pluginapi.KitconfigParameters{
							Customconfig: &pluginapi.Customconfig{
								Registry: &pluginapi.CustomconfigRegistry{},
							},
						},
					},
					Extensions: []*pluginapi.EpParamsExtensionsItems0{
						{
							Name: "capi-byoh",
							Extension: &pluginapi.Extension{
								Extension: []*pluginapi.ExtensionItems0{
									{
										Name: "Infra-provider",
										Config: []*pluginapi.ExtensionItems0ConfigItems0{
											{
												Name:  "Workload-cluster-runtime",
												Value: "",
											},
										},
									},
								},
							},
						},
					},
					Registrycert: &pluginapi.Certificate{},
				},
				inputClustermanifest,
				inputCapiClusterConfig,
				inputCapiSetting,
			},
			genExpectedSetting: nil,
		},
		{
			name: "Get Capi-byoh Setting_CONFIG_MANAGEMENT_CLUSTER_KUBECONFIG",
			args: args{
				&pluginapi.EpParams{
					Cmdline: "",
					Kitconfig: &pluginapi.Kitconfig{
						Parameters: &pluginapi.KitconfigParameters{
							Customconfig: &pluginapi.Customconfig{
								Registry: &pluginapi.CustomconfigRegistry{},
							},
						},
					},
					Extensions: []*pluginapi.EpParamsExtensionsItems0{
						{
							Name: "capi-byoh",
							Extension: &pluginapi.Extension{
								Extension: []*pluginapi.ExtensionItems0{
									{
										Name: "Infra-provider",
										Config: []*pluginapi.ExtensionItems0ConfigItems0{
											{
												Name:  "Management-cluster-kubeconfig",
												Value: "",
											},
										},
									},
								},
							},
						},
					},
					Registrycert: &pluginapi.Certificate{},
				},
				inputClustermanifest,
				inputCapiClusterConfig,
				inputCapiSetting,
			},
			genExpectedSetting: nil,
		},
		{
			name: "Get Capi-byoh Setting",
			args: args{
				&pluginapi.EpParams{
					Cmdline: "",
					Kitconfig: &pluginapi.Kitconfig{
						Parameters: &pluginapi.KitconfigParameters{
							Customconfig: &pluginapi.Customconfig{
								Registry: &pluginapi.CustomconfigRegistry{},
							},
						},
					},
					Extensions: []*pluginapi.EpParamsExtensionsItems0{
						{
							Name: "capi-byoh",
							Extension: &pluginapi.Extension{
								Extension: []*pluginapi.ExtensionItems0{
									{
										Name: "Infra-provider",
										Config: []*pluginapi.ExtensionItems0ConfigItems0{
											{
												Name:  "Workload-cluster-network",
												Value: "",
											},
										},
									},
								},
							},
						},
					},
					Registrycert: &pluginapi.Certificate{},
				},
				inputClustermanifest,
				inputCapiClusterConfig,
				inputCapiSetting,
			},
			genExpectedSetting: nil,
		},
		{
			name: "Get Capi-byoh Setting_CONFIG_WORKLOAD_CLUSTER_NETWORK_GATEWAY",
			args: args{
				&pluginapi.EpParams{
					Cmdline: "",
					Kitconfig: &pluginapi.Kitconfig{
						Parameters: &pluginapi.KitconfigParameters{
							Customconfig: &pluginapi.Customconfig{
								Registry: &pluginapi.CustomconfigRegistry{},
							},
						},
					},
					Extensions: []*pluginapi.EpParamsExtensionsItems0{
						{
							Name: "capi-byoh",
							Extension: &pluginapi.Extension{
								Extension: []*pluginapi.ExtensionItems0{
									{
										Name: "Infra-provider",
										Config: []*pluginapi.ExtensionItems0ConfigItems0{
											{
												Name:  "Workload-cluster-network-gateway",
												Value: "",
											},
										},
									},
								},
							},
						},
					},
					Registrycert: &pluginapi.Certificate{},
				},
				inputClustermanifest,
				inputCapiClusterConfig,
				inputCapiSetting,
			},
			genExpectedSetting: nil,
		},
		{
			name: "Get Capi-byoh Setting_CONFIG_WORKLOAD_CLUSTER_CONTROLPLANE_ENDPOINT",
			args: args{
				&pluginapi.EpParams{
					Cmdline: "",
					Kitconfig: &pluginapi.Kitconfig{
						Parameters: &pluginapi.KitconfigParameters{
							Customconfig: &pluginapi.Customconfig{
								Registry: &pluginapi.CustomconfigRegistry{},
							},
						},
					},
					Extensions: []*pluginapi.EpParamsExtensionsItems0{
						{
							Name: "capi-byoh",
							Extension: &pluginapi.Extension{
								Extension: []*pluginapi.ExtensionItems0{
									{
										Name: "Infra-provider",
										Config: []*pluginapi.ExtensionItems0ConfigItems0{
											{
												Name:  "Workload-cluster-controlplane-endpoint",
												Value: "",
											},
										},
									},
								},
							},
						},
					},
					Registrycert: &pluginapi.Certificate{},
				},
				inputClustermanifest,
				inputCapiClusterConfig,
				inputCapiSetting,
			},
			genExpectedSetting: nil,
		},
		{
			name: "Get Capi-byoh Setting_CONFIG_WORKLOAD_CLUSTER_NODE_ADDRESS_START",
			args: args{
				&pluginapi.EpParams{
					Cmdline: "",
					Kitconfig: &pluginapi.Kitconfig{
						Parameters: &pluginapi.KitconfigParameters{
							Customconfig: &pluginapi.Customconfig{
								Registry: &pluginapi.CustomconfigRegistry{},
							},
						},
					},
					Extensions: []*pluginapi.EpParamsExtensionsItems0{
						{
							Name: "capi-byoh",
							Extension: &pluginapi.Extension{
								Extension: []*pluginapi.ExtensionItems0{
									{
										Name: "Infra-provider",
										Config: []*pluginapi.ExtensionItems0ConfigItems0{
											{
												Name:  "Workload-cluster-node-address-start",
												Value: "",
											},
										},
									},
								},
							},
						},
					},
					Registrycert: &pluginapi.Certificate{},
				},
				inputClustermanifest,
				inputCapiClusterConfig,
				inputCapiSetting,
			},
			genExpectedSetting: nil,
		},
		{
			name: "Get Capi-byoh Setting_CONFIG_WORKLOAD_CLUSTER_NODE_ADDRESS_END",
			args: args{
				&pluginapi.EpParams{
					Cmdline: "",
					Kitconfig: &pluginapi.Kitconfig{
						Parameters: &pluginapi.KitconfigParameters{
							Customconfig: &pluginapi.Customconfig{
								Registry: &pluginapi.CustomconfigRegistry{},
							},
						},
					},
					Extensions: []*pluginapi.EpParamsExtensionsItems0{
						{
							Name: "capi-byoh",
							Extension: &pluginapi.Extension{
								Extension: []*pluginapi.ExtensionItems0{
									{
										Name: "Infra-provider",
										Config: []*pluginapi.ExtensionItems0ConfigItems0{
											{
												Name:  "Workload-cluster-node-address-end",
												Value: "",
											},
										},
									},
								},
							},
						},
					},
					Registrycert: &pluginapi.Certificate{},
				},
				inputClustermanifest,
				inputCapiClusterConfig,
				inputCapiSetting,
			},
			genExpectedSetting: nil,
		},
		{
			name: "Get Capi-byoh Setting_CONFIG_WORKLOAD_CLUSTER_NODE_ADDRESS_PREFIX",
			args: args{
				&pluginapi.EpParams{
					Cmdline: "",
					Kitconfig: &pluginapi.Kitconfig{
						Parameters: &pluginapi.KitconfigParameters{
							Customconfig: &pluginapi.Customconfig{
								Registry: &pluginapi.CustomconfigRegistry{},
							},
						},
					},
					Extensions: []*pluginapi.EpParamsExtensionsItems0{
						{
							Name: "capi-byoh",
							Extension: &pluginapi.Extension{
								Extension: []*pluginapi.ExtensionItems0{
									{
										Name: "Infra-provider",
										Config: []*pluginapi.ExtensionItems0ConfigItems0{
											{
												Name:  "Workload-cluster-node-address-prefix",
												Value: "",
											},
										},
									},
								},
							},
						},
					},
					Registrycert: &pluginapi.Certificate{},
				},
				inputClustermanifest,
				inputCapiClusterConfig,
				inputCapiSetting,
			},
			genExpectedSetting: nil,
		},
		{
			name: "Get Capi-byoh Setting_CONFIG_WORKLOAD_CLUSTER_NODE_USERNAME",
			args: args{
				&pluginapi.EpParams{
					Cmdline: "",
					Kitconfig: &pluginapi.Kitconfig{
						Parameters: &pluginapi.KitconfigParameters{
							Customconfig: &pluginapi.Customconfig{
								Registry: &pluginapi.CustomconfigRegistry{},
							},
						},
					},
					Extensions: []*pluginapi.EpParamsExtensionsItems0{
						{
							Name: "capi-byoh",
							Extension: &pluginapi.Extension{
								Extension: []*pluginapi.ExtensionItems0{
									{
										Name: "Infra-provider",
										Config: []*pluginapi.ExtensionItems0ConfigItems0{
											{
												Name:  "Workload-cluster-node-username",
												Value: "",
											},
										},
									},
								},
							},
						},
					},
					Registrycert: &pluginapi.Certificate{},
				},
				inputClustermanifest,
				inputCapiClusterConfig,
				inputCapiSetting,
			},
			genExpectedSetting: nil,
		},
		{
			name: "Get Capi-byoh Setting_CONFIG_WORKLOAD_CLUSTER_NIC_NAME",
			args: args{
				&pluginapi.EpParams{
					Cmdline: "",
					Kitconfig: &pluginapi.Kitconfig{
						Parameters: &pluginapi.KitconfigParameters{
							Customconfig: &pluginapi.Customconfig{
								Registry: &pluginapi.CustomconfigRegistry{},
							},
						},
					},
					Extensions: []*pluginapi.EpParamsExtensionsItems0{
						{
							Name: "capi-byoh",
							Extension: &pluginapi.Extension{
								Extension: []*pluginapi.ExtensionItems0{
									{
										Name: "Infra-provider",
										Config: []*pluginapi.ExtensionItems0ConfigItems0{
											{
												Name:  "Workload-cluster-nic-name",
												Value: "",
											},
										},
									},
								},
							},
						},
					},
					Registrycert: &pluginapi.Certificate{},
				},
				inputClustermanifest,
				inputCapiClusterConfig,
				inputCapiSetting,
			},
			genExpectedSetting: nil,
		},
		{
			name: "Get Capi-byoh Setting_CONFIG_AUTHORIZED_SSH_PUBLIC_KEY",
			args: args{
				&pluginapi.EpParams{
					Cmdline: "",
					Kitconfig: &pluginapi.Kitconfig{
						Parameters: &pluginapi.KitconfigParameters{
							Customconfig: &pluginapi.Customconfig{
								Registry: &pluginapi.CustomconfigRegistry{},
							},
						},
					},
					Extensions: []*pluginapi.EpParamsExtensionsItems0{
						{
							Name: "capi-byoh",
							Extension: &pluginapi.Extension{
								Extension: []*pluginapi.ExtensionItems0{
									{
										Name: "Infra-provider",
										Config: []*pluginapi.ExtensionItems0ConfigItems0{
											{
												Name:  "Authorized-ssh-public-key",
												Value: "",
											},
										},
									},
								},
							},
						},
					},
					Registrycert: &pluginapi.Certificate{},
				},
				inputClustermanifest,
				inputCapiClusterConfig,
				inputCapiSetting,
			},
			genExpectedSetting: nil,
		},
		{
			name: "Get Capi-byoh Setting12_CONFIG_IRONIC_PROVISION_NIC",
			args: args{
				&pluginapi.EpParams{
					Cmdline: "",
					Kitconfig: &pluginapi.Kitconfig{
						Parameters: &pluginapi.KitconfigParameters{
							Customconfig: &pluginapi.Customconfig{
								Registry: &pluginapi.CustomconfigRegistry{},
							},
						},
					},
					Extensions: []*pluginapi.EpParamsExtensionsItems0{
						{
							Name: "capi-byoh",
							Extension: &pluginapi.Extension{
								Extension: []*pluginapi.ExtensionItems0{
									{
										Name: "Infra-provider",
										Config: []*pluginapi.ExtensionItems0ConfigItems0{
											{
												Name:  "Ironic-provision-nic",
												Value: "",
											},
										},
									},
								},
							},
						},
					},
					Registrycert: &pluginapi.Certificate{},
				},
				inputClustermanifest,
				inputCapiClusterConfig,
				inputCapiSetting_Ironic,
			},
			genExpectedSetting: nil,
		},
		{
			name: "Get Capi-byoh Setting_CONFIG_IRONIC_PROVISION_IP",
			args: args{
				&pluginapi.EpParams{
					Cmdline: "",
					Kitconfig: &pluginapi.Kitconfig{
						Parameters: &pluginapi.KitconfigParameters{
							Customconfig: &pluginapi.Customconfig{
								Registry: &pluginapi.CustomconfigRegistry{},
							},
						},
					},
					Extensions: []*pluginapi.EpParamsExtensionsItems0{
						{
							Name: "capi-byoh",
							Extension: &pluginapi.Extension{
								Extension: []*pluginapi.ExtensionItems0{
									{
										Name: "Infra-provider",
										Config: []*pluginapi.ExtensionItems0ConfigItems0{
											{
												Name:  "Ironic-provision-ip",
												Value: "",
											},
										},
									},
								},
							},
						},
					},
					Registrycert: &pluginapi.Certificate{},
				},
				inputClustermanifest,
				inputCapiClusterConfig,
				inputCapiSetting_Ironic,
			},
			genExpectedSetting: nil,
		},
		{
			name: "Get Capi-byoh Setting_CONFIG_IRONIC_DHCP",
			args: args{
				&pluginapi.EpParams{
					Cmdline: "",
					Kitconfig: &pluginapi.Kitconfig{
						Parameters: &pluginapi.KitconfigParameters{
							Customconfig: &pluginapi.Customconfig{
								Registry: &pluginapi.CustomconfigRegistry{},
							},
						},
					},
					Extensions: []*pluginapi.EpParamsExtensionsItems0{
						{
							Name: "capi-byoh",
							Extension: &pluginapi.Extension{
								Extension: []*pluginapi.ExtensionItems0{
									{
										Name: "Infra-provider",
										Config: []*pluginapi.ExtensionItems0ConfigItems0{
											{
												Name:  "Ironic-dhcp-range",
												Value: "",
											},
										},
									},
								},
							},
						},
					},
					Registrycert: &pluginapi.Certificate{},
				},
				inputClustermanifest,
				inputCapiClusterConfig,
				inputCapiSetting_Ironic,
			},
			genExpectedSetting: nil,
		},
		{
			name: "Get Capi-byoh Setting_CONFIG_IRONIC_HTTP_PORT",
			args: args{
				&pluginapi.EpParams{
					Cmdline: "",
					Kitconfig: &pluginapi.Kitconfig{
						Parameters: &pluginapi.KitconfigParameters{
							Customconfig: &pluginapi.Customconfig{
								Registry: &pluginapi.CustomconfigRegistry{},
							},
						},
					},
					Extensions: []*pluginapi.EpParamsExtensionsItems0{
						{
							Name: "capi-byoh",
							Extension: &pluginapi.Extension{
								Extension: []*pluginapi.ExtensionItems0{
									{
										Name: "Infra-provider",
										Config: []*pluginapi.ExtensionItems0ConfigItems0{
											{
												Name:  "Ironic-http-port",
												Value: "",
											},
										},
									},
								},
							},
						},
					},
					Registrycert: &pluginapi.Certificate{},
				},
				inputClustermanifest,
				inputCapiClusterConfig,
				inputCapiSetting_Ironic,
			},
			genExpectedSetting: nil,
		},
		{
			name: "Get Capi-byoh Setting_CONFIG_IRONIC_OS_IMAGE",
			args: args{
				&pluginapi.EpParams{
					Cmdline: "",
					Kitconfig: &pluginapi.Kitconfig{
						Parameters: &pluginapi.KitconfigParameters{
							Customconfig: &pluginapi.Customconfig{
								Registry: &pluginapi.CustomconfigRegistry{},
							},
						},
					},
					Extensions: []*pluginapi.EpParamsExtensionsItems0{
						{
							Name: "capi-byoh",
							Extension: &pluginapi.Extension{
								Extension: []*pluginapi.ExtensionItems0{
									{
										Name: "Infra-provider",
										Config: []*pluginapi.ExtensionItems0ConfigItems0{
											{
												Name:  "Ironic-os-image",
												Value: "",
											},
										},
									},
								},
							},
						},
					},
					Registrycert: &pluginapi.Certificate{},
				},
				inputClustermanifest,
				inputCapiClusterConfig,
				inputCapiSetting_Ironic,
			},
			genExpectedSetting: nil,
		},
		{
			name: "Get Capi-byoh Setting_WorkloadClusterNodeNum success",
			args: args{
				&pluginapi.EpParams{
					Kitconfig: &pluginapi.Kitconfig{
						Parameters: &pluginapi.KitconfigParameters{
							Nodes: []*pluginapi.Node{
								{
									Role: []string{"worker"},
								},
								{
									Role: []string{"worker"},
								},
								{
									Role: []string{"controlplane", "etcd"},
								},
							},
						},
					},
					Extensions: []*pluginapi.EpParamsExtensionsItems0{
						{
							Name:      "capi-byoh",
							Extension: &pluginapi.Extension{},
						},
					},
				},
				inputClustermanifest,
				inputCapiClusterConfig,
				inputCapiSetting,
			},
			genExpectedSetting: func(input *pluginapi.CapiSetting) *pluginapi.CapiSetting {
				input.InfraProvider.WorkloadClusterControlPlaneNum = 1
				input.InfraProvider.WorkloadClusterWorkerNodeNum = 2
				return input
			},
		},
		{
			name: "Get Capi-byoh Setting_WorkloadClusterNodeNum invalid",
			args: args{
				&pluginapi.EpParams{
					Kitconfig: &pluginapi.Kitconfig{
						Parameters: &pluginapi.KitconfigParameters{
							Nodes: []*pluginapi.Node{
								{
									Role: []string{"worker"},
								},
								{
									Role: []string{"worker", "controlplane"},
								},
								{
									Role: []string{"controlplane", "etcd"},
								},
							},
						},
					},
					Extensions: []*pluginapi.EpParamsExtensionsItems0{
						{
							Name:      "capi-byoh",
							Extension: &pluginapi.Extension{},
						},
					},
				},
				inputClustermanifest,
				inputCapiClusterConfig,
				inputCapiSetting,
			},
			genExpectedSetting: func(input *pluginapi.CapiSetting) *pluginapi.CapiSetting {
				input.InfraProvider.WorkloadClusterControlPlaneNum = 0
				input.InfraProvider.WorkloadClusterWorkerNodeNum = 0
				return input
			},
		},
		{
			name: "Get Capi-byoh Setting_WorkloadClusterNodeNum success",
			args: args{
				&pluginapi.EpParams{
					Kitconfig: &pluginapi.Kitconfig{
						Parameters: &pluginapi.KitconfigParameters{
							Nodes: []*pluginapi.Node{
								{
									Role: []string{"worker"},
								},
								{
									Role: []string{"worker"},
								},
								{
									Role: []string{"controlplane", "etcd"},
								},
							},
						},
					},
					Extensions: []*pluginapi.EpParamsExtensionsItems0{
						{
							Name:      "capi-byoh",
							Extension: &pluginapi.Extension{},
						},
					},
				},
				inputClustermanifest,
				inputCapiClusterConfig,
				inputCapiSetting,
			},
			genExpectedSetting: func(input *pluginapi.CapiSetting) *pluginapi.CapiSetting {
				input.InfraProvider.WorkloadClusterControlPlaneNum = 1
				input.InfraProvider.WorkloadClusterWorkerNodeNum = 2
				return input
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := GetCapiSetting(tt.args.epparams, tt.args.clusterManifest, tt.args.clusterConfig, tt.args.setting)
			if err == nil && tt.genExpectedSetting != nil {
				expectedSetting := tt.genExpectedSetting(tt.args.setting)
				binaryOri, errOri := tt.args.setting.MarshalBinary()
				binaryCom, errCom := expectedSetting.MarshalBinary()

				if errOri != nil || errCom != nil {
					t.Logf("MarshalBinary error: errOri: %v, errCom: %v", errOri, errCom)
				}

				if res := bytes.Compare(binaryOri, binaryCom); res != 0 {
					t.Error("Failed to get expected output.")
				}
			}
		})
	}
}

func TestTmplFileRendering(t *testing.T) {
	func_DownloadFile_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchDownloadFile, err := mpatch.PatchMethod(eputils.DownloadFile, func(filepath, fileurl string) error {
			return errTest
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}

		return []*mpatch.Patch{pathchDownloadFile}
	}
	func_ReadFile_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchDownloadFile, err := mpatch.PatchMethod(eputils.DownloadFile, func(filepath, fileurl string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchReadFile, err := mpatch.PatchMethod(ioutil.ReadFile, func(filename string) ([]byte, error) {
			return nil, errTest
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		return []*mpatch.Patch{pathchDownloadFile, pathchReadFile}
	}
	func_StringTemplateConvertWithParams_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchDownloadFile, err := mpatch.PatchMethod(eputils.DownloadFile, func(filepath, fileurl string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchReadFile, err := mpatch.PatchMethod(ioutil.ReadFile, func(filename string) ([]byte, error) {
			return nil, nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchStringTemplateConvertWithParams, err := mpatch.PatchMethod(eputils.StringTemplateConvertWithParams, func(str string, tempParams interface{}) (string, error) {
			return "", errTest
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}

		return []*mpatch.Patch{pathchDownloadFile, pathchReadFile, pathchStringTemplateConvertWithParams}
	}
	func_WriteStringToFile_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchDownloadFile, err := mpatch.PatchMethod(eputils.DownloadFile, func(filepath, fileurl string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchReadFile, err := mpatch.PatchMethod(ioutil.ReadFile, func(filename string) ([]byte, error) {
			return nil, nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchStringTemplateConvertWithParams, err := mpatch.PatchMethod(eputils.StringTemplateConvertWithParams, func(str string, tempParams interface{}) (string, error) {
			return "", nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchWriteStringToFile, err := mpatch.PatchMethod(eputils.WriteStringToFile, func(content string, filename string) error {
			return errTest
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		return []*mpatch.Patch{pathchDownloadFile, pathchReadFile, pathchStringTemplateConvertWithParams, pathchWriteStringToFile}
	}
	func_everything_ok := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchDownloadFile, err := mpatch.PatchMethod(eputils.DownloadFile, func(filepath, fileurl string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchReadFile, err := mpatch.PatchMethod(ioutil.ReadFile, func(filename string) ([]byte, error) {
			return nil, nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchStringTemplateConvertWithParams, err := mpatch.PatchMethod(eputils.StringTemplateConvertWithParams, func(str string, tempParams interface{}) (string, error) {
			return "", nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchWriteStringToFile, err := mpatch.PatchMethod(eputils.WriteStringToFile, func(content string, filename string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		return []*mpatch.Patch{pathchDownloadFile, pathchReadFile, pathchStringTemplateConvertWithParams, pathchWriteStringToFile}
	}
	type args struct {
		tmpl       *CapiTemplate
		workFolder string
		url        string
		dstFile    string
	}
	cases := []struct {
		name               string
		args               args
		expectErrorContent error
		funcBeforeTest     func(*gomock.Controller) []*mpatch.Patch
	}{
		{
			name: "Tmpl File Rendering_DownloadFile_err",
			args: args{
				tmpl:       &CapiTemplate{},
				workFolder: "/test",
				url:        "10.10.10.10",
				dstFile:    "",
			},
			expectErrorContent: errTest,
			funcBeforeTest:     func_DownloadFile_fail,
		},
		{
			name: "Tmpl File Rendering_ReadFile_err",
			args: args{
				tmpl:       &CapiTemplate{},
				workFolder: "/test",
				url:        "10.10.10.10",
				dstFile:    "",
			},
			expectErrorContent: errTest,
			funcBeforeTest:     func_ReadFile_fail,
		},
		{
			name: "Tmpl File Rendering_StringTemplateConvertWithParams_err",
			args: args{
				tmpl:       &CapiTemplate{},
				workFolder: "/test",
				url:        "10.10.10.10",
				dstFile:    "",
			},
			expectErrorContent: errTest,
			funcBeforeTest:     func_StringTemplateConvertWithParams_fail,
		},
		{
			name: "Tmpl File Rendering_WriteStringToFile_err",
			args: args{
				tmpl:       &CapiTemplate{},
				workFolder: "/test",
				url:        "10.10.10.10",
				dstFile:    "",
			},
			expectErrorContent: errTest,
			funcBeforeTest:     func_WriteStringToFile_fail,
		},
		{
			name: "Tmpl File Rendering_everything_ok",
			args: args{
				tmpl:       &CapiTemplate{},
				workFolder: "/test",
				url:        "10.10.10.10",
				dstFile:    "",
			},
			expectErrorContent: errTest,
			funcBeforeTest:     func_everything_ok,
		},
	}
	for _, tc := range cases {

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		var plist []*mpatch.Patch
		if tc.funcBeforeTest != nil {
			plist = tc.funcBeforeTest(ctrl)
		}

		t.Run(tc.name, func(t *testing.T) {
			// Run test cases in parallel if necessary.
			if result := TmplFileRendering(tc.args.tmpl, tc.args.workFolder, tc.args.url, tc.args.dstFile); result != nil {
				if fmt.Sprint(result) == fmt.Sprint(tc.expectErrorContent) {
					t.Log(tc.name, "Done")
				} else {
					t.Errorf("%s error %s", tc.name, result)
				}
			}
		})

		for _, p := range plist {
			unpatch(t, p)
		}
	}
}

func TestCheckCapiSetting(t *testing.T) {
	tests := []struct {
		name           string
		input          *pluginapi.CapiSetting
		expectError    bool
		expectErrorMsg string
	}{
		{
			name:           "Miss infra provider info",
			input:          &pluginapi.CapiSetting{},
			expectError:    true,
			expectErrorMsg: errCAPISetting.Error(),
		},
		{
			name: "Invaid node number",
			input: &pluginapi.CapiSetting{
				Provider: fakeByoh,
				InfraProvider: &pluginapi.CapiSettingInfraProvider{
					WorkloadClusterControlPlaneNum: 0,
					WorkloadClusterWorkerNodeNum:   0,
				},
			},
			expectError:    true,
			expectErrorMsg: errNumberNodes.Error(),
		},
		{
			name: "Pass checking",
			input: &pluginapi.CapiSetting{
				Provider: fakeByoh,
				InfraProvider: &pluginapi.CapiSettingInfraProvider{
					WorkloadClusterControlPlaneNum: 1,
					WorkloadClusterWorkerNodeNum:   3,
				},
			},
			expectError:    false,
			expectErrorMsg: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			if result := CheckCapiSetting(tc.input); result != nil {
				if tc.expectError {
					if fmt.Sprint(result) == tc.expectErrorMsg {
						t.Logf("Expected error: {%s} catched, done.", tc.expectErrorMsg)
						return
					} else if tc.expectErrorMsg == "" {
						return
					} else {
						t.Fatal("Unexpected error occurred.")
					}

				}
				t.Error(result)
			}
		})
	}
}
