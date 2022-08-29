/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
//nolint: dupl
package capihostprovision

import (
	pluginapi "ep/pkg/api/plugins"
	"ep/pkg/eputils"
	"ep/pkg/eputils/capiutils"
	"ep/pkg/eputils/kubeutils"
	"ep/pkg/executor"
	"errors"
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/undefinedlabs/go-mpatch"
	appsv1 "k8s.io/api/apps/v1"
	appsv1client "k8s.io/client-go/kubernetes/typed/apps/v1"
)

var (
	errTest = errors.New("test_error")

	errByohNotReady = errors.New("E001.329: BYOH host not ready")
)

func unpatch(t *testing.T, m *mpatch.Patch) {
	err := m.Unpatch()
	if err != nil {
		t.Fatal(err)
	}
}

type fakeDeployment struct {
	Namespace    string
	Name         string
	FieldManager string

	Client        appsv1client.AppsV1Interface
	DeploymentObj *appsv1.Deployment
}

func (s fakeDeployment) Get() error {
	return nil
}

func (s fakeDeployment) New() error {
	return nil
}

func (s fakeDeployment) GetStatus() appsv1.DeploymentStatus {
	return appsv1.DeploymentStatus{}
}

func Test_waitByohCtlMgrDeploymentReady(t *testing.T) {
	func_Deployment_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchNewDeployment, err := mpatch.PatchMethod(kubeutils.NewDeployment, func(namespace string, name string, feildManager string, kubeconfig string) (kubeutils.DeploymentWrapper, error) {
			return nil, errTest
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}

		return []*mpatch.Patch{pathchNewDeployment}
	}
	func_byohCtlMgrDeploymentGet_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchNewDeployment, err := mpatch.PatchMethod(kubeutils.NewDeployment, func(namespace string, name string, feildManager string, kubeconfig string) (kubeutils.DeploymentWrapper, error) {
			return &fakeDeployment{
				Namespace:     "",
				Name:          "",
				FieldManager:  "",
				Client:        nil,
				DeploymentObj: nil,
			}, nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchfakeDeploymentGet, err := mpatch.PatchMethod(fakeDeployment.Get, func(fakeDeployment) error {
			return errTest
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}

		return []*mpatch.Patch{pathchNewDeployment, pathchfakeDeploymentGet}
	}
	func_byohCtlMgrDeploymentGetStatus_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchNewDeployment, err := mpatch.PatchMethod(kubeutils.NewDeployment, func(namespace string, name string, feildManager string, kubeconfig string) (kubeutils.DeploymentWrapper, error) {
			return &fakeDeployment{
				Namespace:     "",
				Name:          "",
				FieldManager:  "",
				Client:        nil,
				DeploymentObj: nil,
			}, nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchfakeDeploymentGet, err := mpatch.PatchMethod(fakeDeployment.Get, func(fakeDeployment) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		return []*mpatch.Patch{pathchNewDeployment, pathchfakeDeploymentGet}
	}

	type args struct {
		management_kubeconfig string
	}
	tests := []struct {
		name               string
		args               args
		expectErrorContent error
		funcBeforeTest     func(*gomock.Controller) []*mpatch.Patch
	}{
		{
			name: "NewDeployment_err",
			args: args{
				management_kubeconfig: "",
			},
			expectErrorContent: errTest,
			funcBeforeTest:     func_Deployment_fail,
		},
		{
			name: "byohCtlMgrDeploymentGet_err",
			args: args{
				management_kubeconfig: "",
			},
			expectErrorContent: errTest,
			funcBeforeTest:     func_byohCtlMgrDeploymentGet_fail,
		},
		{
			name: "byohCtlMgrDeployment_ok",
			args: args{
				management_kubeconfig: "",
			},
			funcBeforeTest: func_byohCtlMgrDeploymentGetStatus_fail,
		},
	}
	for _, tc := range tests {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		var plist []*mpatch.Patch
		if tc.funcBeforeTest != nil {
			plist = tc.funcBeforeTest(ctrl)
		}

		t.Run(tc.name, func(t *testing.T) {
			// Run test cases in parallel if necessary.
			if result := DeploymentReady(tc.args.management_kubeconfig, "byoh-system", "byoh-controller-manager"); result != nil {
				if fmt.Sprint(result) == fmt.Sprint(tc.expectErrorContent) {
					t.Log(tc.name, "Done")
				} else {
					t.Errorf("%s error %s", tc.name, result)
				}
			} else {
				t.Log(tc.name, "Done")
			}
		})

		for _, p := range plist {
			unpatch(t, p)
		}
	}
}

func Test_checkByoHosts(t *testing.T) {
	func_RUNCMD_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchRunCMD, err := mpatch.PatchMethod(eputils.RunCMD, func(cmd *exec.Cmd) (string, error) {
			return "", errTest
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}

		return []*mpatch.Patch{pathchRunCMD}
	}
	func_NodeNotReady_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchRunCMD, err := mpatch.PatchMethod(eputils.RunCMD, func(cmd *exec.Cmd) (string, error) {
			return "first\n", nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		patchSleep, err := mpatch.PatchMethod(time.Sleep, func(d time.Duration) {})
		if err != nil {
			t.Fatal(err)
		}
		return []*mpatch.Patch{pathchRunCMD, patchSleep}
	}
	func_CheckByoHostsReady_ok := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchRunCMD, err := mpatch.PatchMethod(eputils.RunCMD, func(cmd *exec.Cmd) (string, error) {
			return "first\nsecond\nthird\n", nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}

		return []*mpatch.Patch{pathchRunCMD}
	}

	type args struct {
		ep_params             *pluginapi.EpParams
		workFolder            string
		management_kubeconfig string
		clusterConfig         *pluginapi.CapiClusterConfig
		tmpl                  *capiutils.CapiTemplate
	}

	tests := []struct {
		name               string
		args               args
		expectErrorContent error
		funcBeforeTest     func(*gomock.Controller) []*mpatch.Patch
	}{
		{
			name: "RunCMD_err",
			args: args{
				ep_params: &pluginapi.EpParams{
					Workspace: "default",
				},
				workFolder:            "",
				management_kubeconfig: "",
				clusterConfig: &pluginapi.CapiClusterConfig{
					WorkloadCluster: &pluginapi.CapiClusterConfigWorkloadCluster{
						Namespace: "default",
					},
				},
				tmpl: nil,
			},
			expectErrorContent: errTest,
			funcBeforeTest:     func_RUNCMD_fail,
		},
		{
			name: "NodeNotReady_err",
			args: args{
				ep_params: &pluginapi.EpParams{
					Workspace: "default",
				},
				workFolder:            "",
				management_kubeconfig: "",
				clusterConfig: &pluginapi.CapiClusterConfig{
					WorkloadCluster: &pluginapi.CapiClusterConfigWorkloadCluster{
						Namespace: "default",
					},
				},
				tmpl: nil,
			},
			expectErrorContent: errByohNotReady,
			funcBeforeTest:     func_NodeNotReady_fail,
		},
		{
			name: "CheckByoHostsReady_ok",
			args: args{
				ep_params: &pluginapi.EpParams{
					Workspace: "default",
				},
				workFolder:            "",
				management_kubeconfig: "",
				clusterConfig: &pluginapi.CapiClusterConfig{
					WorkloadCluster: &pluginapi.CapiClusterConfigWorkloadCluster{
						Namespace: "default",
					},
				},
				tmpl: nil,
			},
			funcBeforeTest: func_CheckByoHostsReady_ok,
		},
	}
	for _, tc := range tests {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		var plist []*mpatch.Patch
		if tc.funcBeforeTest != nil {
			plist = tc.funcBeforeTest(ctrl)
		}
		t.Run(tc.name, func(t *testing.T) {
			// Run test cases in parallel if necessary.
			if result := checkByoHosts(tc.args.ep_params, tc.args.workFolder, tc.args.management_kubeconfig, tc.args.clusterConfig, tc.args.tmpl); result != nil {
				if fmt.Sprint(result) == fmt.Sprint(tc.expectErrorContent) {
					t.Log(tc.name, "Done")
				} else {
					t.Errorf("%s error %s", tc.name, result)
				}
			} else {
				t.Log(tc.name, "Done")
			}
		})

		for _, p := range plist {
			unpatch(t, p)
		}
	}
}

func Test_byohHostProvision(t *testing.T) {
	/*
		func_crioReleaseDownload_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
			pathchcrioReleaseDownload, err := mpatch.PatchMethod(crioReleaseDownload, func(ep_params *pluginapi.EpParams, workFolder string, capiSetting *pluginapi.CapiSetting) error {
				return errTest
			})
			if err != nil {
				t.Errorf("patch error: %v", err)
			}

			return []*mpatch.Patch{pathchcrioReleaseDownload}
		}
	*/
	func_waitByohCtlMgrDeploymentReady_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		/*
			pathchcrioReleaseDownload, err := mpatch.PatchMethod(crioReleaseDownload, func(ep_params *pluginapi.EpParams, workFolder string, capiSetting *pluginapi.CapiSetting) error {
				return nil
			})
			if err != nil {
				t.Errorf("patch error: %v", err)
			}
		*/
		pathchwaitByohCtlMgrDeploymentReady, err := mpatch.PatchMethod(DeploymentReady, func(management_kubeconfig, namespace, deploymentName string) error {
			return errTest
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}

		//return []*mpatch.Patch{pathchcrioReleaseDownload, pathchwaitByohCtlMgrDeploymentReady}
		return []*mpatch.Patch{pathchwaitByohCtlMgrDeploymentReady}
	}
	func_executorRun_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		/*
			pathchcrioReleaseDownload, err := mpatch.PatchMethod(crioReleaseDownload, func(ep_params *pluginapi.EpParams, workFolder string, capiSetting *pluginapi.CapiSetting) error {
				return nil
			})
			if err != nil {
				t.Errorf("patch error: %v", err)
			}
		*/
		pathchwaitByohCtlMgrDeploymentReady, err := mpatch.PatchMethod(DeploymentReady, func(management_kubeconfig, namespace, deploymentName string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchexecutorRun, err := mpatch.PatchMethod(executor.Run, func(specFile string, epparams *pluginapi.EpParams, value interface{}) error {
			return errTest
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}

		//return []*mpatch.Patch{pathchcrioReleaseDownload, pathchwaitByohCtlMgrDeploymentReady, pathchexecutorRun}
		return []*mpatch.Patch{pathchwaitByohCtlMgrDeploymentReady, pathchexecutorRun}
	}
	func_checkByoHosts_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		/*
			pathchcrioReleaseDownload, err := mpatch.PatchMethod(crioReleaseDownload, func(ep_params *pluginapi.EpParams, workFolder string, capiSetting *pluginapi.CapiSetting) error {
				return nil
			})
			if err != nil {
				t.Errorf("patch error: %v", err)
			}
		*/
		pathchwaitByohCtlMgrDeploymentReady, err := mpatch.PatchMethod(DeploymentReady, func(management_kubeconfig, namespace, deploymentName string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchexecutorRun, err := mpatch.PatchMethod(executor.Run, func(specFile string, epparams *pluginapi.EpParams, value interface{}) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathcheckByoHosts, err := mpatch.PatchMethod(checkByoHosts, func(ep_params *pluginapi.EpParams, workFolder string, management_kubeconfig string, clusterConfig *pluginapi.CapiClusterConfig, tmpl *capiutils.CapiTemplate) error {
			return errTest
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		//return []*mpatch.Patch{pathchcrioReleaseDownload, pathchwaitByohCtlMgrDeploymentReady, pathchexecutorRun, pathcheckByoHosts}
		return []*mpatch.Patch{pathchwaitByohCtlMgrDeploymentReady, pathchexecutorRun, pathcheckByoHosts}
	}
	func_byohHostProvision_ok := func(ctrl *gomock.Controller) []*mpatch.Patch {
		/*
			pathchcrioReleaseDownload, err := mpatch.PatchMethod(crioReleaseDownload, func(ep_params *pluginapi.EpParams, workFolder string, capiSetting *pluginapi.CapiSetting) error {
				return nil
			})
			if err != nil {
				t.Errorf("patch error: %v", err)
			}
		*/
		pathchwaitByohCtlMgrDeploymentReady, err := mpatch.PatchMethod(DeploymentReady, func(management_kubeconfig, namespace, deploymentName string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchexecutorRun, err := mpatch.PatchMethod(executor.Run, func(specFile string, epparams *pluginapi.EpParams, value interface{}) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathcheckByoHosts, err := mpatch.PatchMethod(checkByoHosts, func(ep_params *pluginapi.EpParams, workFolder string, management_kubeconfig string, clusterConfig *pluginapi.CapiClusterConfig, tmpl *capiutils.CapiTemplate) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		//return []*mpatch.Patch{pathchcrioReleaseDownload, pathchwaitByohCtlMgrDeploymentReady, pathchexecutorRun, pathcheckByoHosts}
		return []*mpatch.Patch{pathchwaitByohCtlMgrDeploymentReady, pathchexecutorRun, pathcheckByoHosts}
	}
	type args struct {
		ep_params             *pluginapi.EpParams
		workFolder            string
		management_kubeconfig string
		clusterConfig         *pluginapi.CapiClusterConfig
		tmpl                  *capiutils.CapiTemplate
	}
	tests := []struct {
		name               string
		args               args
		expectErrorContent error
		funcBeforeTest     func(*gomock.Controller) []*mpatch.Patch
	}{
		/*
			{
				name: "crioReleaseDownload_err",
				args: args{
					ep_params: &pluginapi.EpParams{
						Workspace:  "default",
						Runtimedir: "default",
						Kitconfig: &pluginapi.Kitconfig{
							Parameters: &pluginapi.KitconfigParameters{
								GlobalSettings: &pluginapi.KitconfigParametersGlobalSettings{
									ProviderIP:   "10.10.10.10",
									RegistryPort: "8080",
								},
								Customconfig: &pluginapi.Customconfig{
									Registry: &pluginapi.CustomconfigRegistry{
										User:     "testName",
										Password: "testPwd",
									},
								},
							},
						},
					},
					workFolder:            "",
					management_kubeconfig: "",
					clusterConfig: &pluginapi.CapiClusterConfig{
						WorkloadCluster: &pluginapi.CapiClusterConfigWorkloadCluster{
							Namespace: "default",
						},
					},
					tmpl: &capiutils.CapiTemplate{
						CapiSetting: pluginapi.CapiSetting{
							CRI: &pluginapi.CapiSettingCRI{
								Name: "defalut",
							},
						},
					},
				},
				expectErrorContent: errTest,
				funcBeforeTest:     func_crioReleaseDownload_fail,
			},
		*/
		{
			name: "waitByohCtlMgrDeploymentReady_err",
			args: args{
				ep_params: &pluginapi.EpParams{
					Workspace:  "default",
					Runtimedir: "default",
					Kitconfig: &pluginapi.Kitconfig{
						Parameters: &pluginapi.KitconfigParameters{
							GlobalSettings: &pluginapi.KitconfigParametersGlobalSettings{
								ProviderIP:   "10.10.10.10",
								RegistryPort: "8080",
							},
							Customconfig: &pluginapi.Customconfig{
								Registry: &pluginapi.CustomconfigRegistry{
									User:     "testName",
									Password: "testPwd",
								},
							},
						},
					},
				},
				workFolder:            "",
				management_kubeconfig: "",
				clusterConfig: &pluginapi.CapiClusterConfig{
					WorkloadCluster: &pluginapi.CapiClusterConfigWorkloadCluster{
						Namespace: "default",
					},
				},
				tmpl: &capiutils.CapiTemplate{
					CapiSetting: pluginapi.CapiSetting{
						CRI: &pluginapi.CapiSettingCRI{
							Name: "defalut",
						},
					},
				},
			},
			expectErrorContent: errTest,
			funcBeforeTest:     func_waitByohCtlMgrDeploymentReady_fail,
		},
		{
			name: "executorRun_err",
			args: args{
				ep_params: &pluginapi.EpParams{
					Workspace: "default",
				},
				workFolder:            "",
				management_kubeconfig: "",
				clusterConfig: &pluginapi.CapiClusterConfig{
					WorkloadCluster: &pluginapi.CapiClusterConfigWorkloadCluster{
						Namespace: "default",
					},
					ByohAgent: &pluginapi.CapiClusterConfigByohAgent{
						InitScript: "defalut",
					},
				},
				tmpl: &capiutils.CapiTemplate{
					CapiSetting: pluginapi.CapiSetting{},
				},
			},
			expectErrorContent: errTest,
			funcBeforeTest:     func_executorRun_fail,
		},
		{
			name: "checkByoHosts_err",
			args: args{
				ep_params: &pluginapi.EpParams{
					Workspace: "default",
				},
				workFolder:            "",
				management_kubeconfig: "",
				clusterConfig: &pluginapi.CapiClusterConfig{
					WorkloadCluster: &pluginapi.CapiClusterConfigWorkloadCluster{
						Namespace: "default",
					},
					ByohAgent: &pluginapi.CapiClusterConfigByohAgent{
						InitScript: "defalut",
					},
				},
				tmpl: &capiutils.CapiTemplate{
					CapiSetting: pluginapi.CapiSetting{},
				},
			},
			expectErrorContent: errTest,
			funcBeforeTest:     func_checkByoHosts_fail,
		},
		{
			name: "checkByoHosts_ok",
			args: args{
				ep_params: &pluginapi.EpParams{
					Workspace: "default",
				},
				workFolder:            "",
				management_kubeconfig: "",
				clusterConfig: &pluginapi.CapiClusterConfig{
					WorkloadCluster: &pluginapi.CapiClusterConfigWorkloadCluster{
						Namespace: "default",
					},
					ByohAgent: &pluginapi.CapiClusterConfigByohAgent{
						InitScript: "defalut",
					},
				},
				tmpl: &capiutils.CapiTemplate{
					CapiSetting: pluginapi.CapiSetting{},
				},
			},
			funcBeforeTest: func_byohHostProvision_ok,
		},
	}
	for _, tc := range tests {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		var plist []*mpatch.Patch
		if tc.funcBeforeTest != nil {
			plist = tc.funcBeforeTest(ctrl)
		}
		t.Run(tc.name, func(t *testing.T) {
			// Run test cases in parallel if necessary.
			if result := byohHostProvision(tc.args.ep_params, tc.args.workFolder, tc.args.management_kubeconfig, tc.args.clusterConfig, tc.args.tmpl); result != nil {
				if fmt.Sprint(result) == fmt.Sprint(tc.expectErrorContent) {
					t.Log(tc.name, "Done")
				} else {
					t.Errorf("%s error %s", tc.name, result)
				}
			} else {
				t.Log(tc.name, "Done")
			}
		})

		for _, p := range plist {
			unpatch(t, p)
		}
	}
}

/*
func Test_crioReleaseDownload(t *testing.T) {
	func_FileExists_err := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchFileExists, err := mpatch.PatchMethod(eputils.FileExists, func(filename string) bool {
			return false
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}

		return []*mpatch.Patch{pathchFileExists}
	}
	func_RunCMD_err := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchFileExists, err := mpatch.PatchMethod(eputils.FileExists, func(filename string) bool {
			return true
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchRunCMD, err := mpatch.PatchMethod(eputils.RunCMD, func(cmd *exec.Cmd) (string, error) {
			return "", errTest
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		return []*mpatch.Patch{pathchFileExists, pathchRunCMD}
	}
	func_Runnormal_ok := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchFileExists, err := mpatch.PatchMethod(eputils.FileExists, func(filename string) bool {
			return true
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchRunCMD, err := mpatch.PatchMethod(eputils.RunCMD, func(cmd *exec.Cmd) (string, error) {
			return "success", nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		return []*mpatch.Patch{pathchFileExists, pathchRunCMD}
	}
	type args struct {
		ep_params   *pluginapi.EpParams
		workFolder  string
		capiSetting *pluginapi.CapiSetting
	}
	tests := []struct {
		name               string
		args               args
		expectErrorContent error
		funcBeforeTest     func(*gomock.Controller) []*mpatch.Patch
	}{
		{
			name: "CRIName_err",
			args: args{
				ep_params:  &pluginapi.EpParams{},
				workFolder: "",
				capiSetting: &pluginapi.CapiSetting{
					CRI: &pluginapi.CapiSettingCRI{
						Name: "test_err",
					},
				},
			},
			expectErrorContent: nil,
			funcBeforeTest:     func_FileExists_err,
		},
		{
			name: "FileExists_err",
			args: args{
				ep_params: &pluginapi.EpParams{
					Runtimedir: "default",
					Kitconfig: &pluginapi.Kitconfig{
						Parameters: &pluginapi.KitconfigParameters{
							GlobalSettings: &pluginapi.KitconfigParametersGlobalSettings{
								ProviderIP:   "10.10.10.10",
								RegistryPort: "8080",
							},
							Customconfig: &pluginapi.Customconfig{
								Registry: &pluginapi.CustomconfigRegistry{
									User:     "testName",
									Password: "testPwd",
								},
							},
						},
					},
				},
				workFolder: "",
				capiSetting: &pluginapi.CapiSetting{
					CRI: &pluginapi.CapiSettingCRI{
						Name: "crio",
					},
				},
			},
			expectErrorContent: errImgPkg,
			funcBeforeTest:     func_FileExists_err,
		},
		{
			name: "RunCMD_err",
			args: args{
				ep_params: &pluginapi.EpParams{
					Runtimedir: "default",
					Kitconfig: &pluginapi.Kitconfig{
						Parameters: &pluginapi.KitconfigParameters{
							GlobalSettings: &pluginapi.KitconfigParametersGlobalSettings{
								ProviderIP:   "10.10.10.10",
								RegistryPort: "8080",
							},
							Customconfig: &pluginapi.Customconfig{
								Registry: &pluginapi.CustomconfigRegistry{
									User:     "testName",
									Password: "testPwd",
								},
							},
						},
					},
				},
				workFolder: "",
				capiSetting: &pluginapi.CapiSetting{
					CRI: &pluginapi.CapiSettingCRI{
						Name: "crio",
					},
				},
			},
			expectErrorContent: errTest,
			funcBeforeTest:     func_RunCMD_err,
		},
		{
			name: "Runnormal_ok",
			args: args{
				ep_params: &pluginapi.EpParams{
					Runtimedir: "default",
					Kitconfig: &pluginapi.Kitconfig{
						Parameters: &pluginapi.KitconfigParameters{
							GlobalSettings: &pluginapi.KitconfigParametersGlobalSettings{
								ProviderIP:   "10.10.10.10",
								RegistryPort: "8080",
							},
							Customconfig: &pluginapi.Customconfig{
								Registry: &pluginapi.CustomconfigRegistry{
									User:     "testName",
									Password: "testPwd",
								},
							},
						},
					},
				},
				workFolder: "",
				capiSetting: &pluginapi.CapiSetting{
					CRI: &pluginapi.CapiSettingCRI{
						Name: "crio",
					},
				},
			},
			funcBeforeTest: func_Runnormal_ok,
		},
	}
	for _, tc := range tests {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		var plist []*mpatch.Patch
		if tc.funcBeforeTest != nil {
			plist = tc.funcBeforeTest(ctrl)
		}
		t.Run(tc.name, func(t *testing.T) {
			if result := crioReleaseDownload(tc.args.ep_params, tc.args.workFolder, tc.args.capiSetting); result != nil {
				if fmt.Sprint(result) == fmt.Sprint(tc.expectErrorContent) {
					t.Log(tc.name, "Done")
				} else {
					t.Errorf("%s error %s", tc.name, result)
				}
			} else {
				t.Log(tc.name, "Done")
			}
		})
		for _, p := range plist {
			unpatch(t, p)
		}
	}
}
*/
