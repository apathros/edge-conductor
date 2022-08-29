/*
* Copyright (c) 2021 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

//nolint: dupl
package capihostprovision

import (
	cmapi "ep/pkg/api/certmgr"
	pluginapi "ep/pkg/api/plugins"
	certmgr "ep/pkg/certmgr"
	eputils "ep/pkg/eputils"
	capiutils "ep/pkg/eputils/capiutils"
	kubeutils "ep/pkg/eputils/kubeutils"
	serviceutil "ep/pkg/eputils/service"
	"ep/pkg/eputils/test/fakeserviceutils"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/undefinedlabs/go-mpatch"
)

var (
	errMetalTest = errors.New("test_error")
)

func multiplePatchGenCertAndConfig(t *testing.T, err error, nextPatch func()) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(certmgr.GenCertAndConfig, func(certbundle cmapi.Certificate, hosts string) error {
		unpatch(t, patch)
		if nextPatch != nil {
			nextPatch()
		}
		return err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func Test_ironic_tls_setup(t *testing.T) {
	func_GenCertAndConfig_fail_1 := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchGenCertAndConfig, err := mpatch.PatchMethod(certmgr.GenCertAndConfig, func(certbundle cmapi.Certificate, hosts string) error {
			return errMetalTest
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}

		return []*mpatch.Patch{pathchGenCertAndConfig}
	}

	func_GenCertAndConfig_fail_2 := func(ctrl *gomock.Controller) []*mpatch.Patch {
		multiplePatchGenCertAndConfig(t, nil, func() {
			multiplePatchGenCertAndConfig(t, errMetalTest, nil)
		})
		return nil
	}
	func_GenCertAndConfig_fail_3 := func(ctrl *gomock.Controller) []*mpatch.Patch {
		multiplePatchGenCertAndConfig(t, nil, func() {
			multiplePatchGenCertAndConfig(t, nil, func() {
				multiplePatchGenCertAndConfig(t, errMetalTest, nil)
			})
		})
		return nil
	}
	func_ChmodMARIADBKEYFIL_err := func(ctrl *gomock.Controller) []*mpatch.Patch {
		multiplePatchGenCertAndConfig(t, nil, func() {
			multiplePatchGenCertAndConfig(t, nil, func() {
				multiplePatchGenCertAndConfig(t, nil, nil)
			})
		})
		pathchFileExists, err := mpatch.PatchMethod(eputils.FileExists, func(filename string) bool {
			return true
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchChmod, err := mpatch.PatchMethod(os.Chmod, func(name string, mode fs.FileMode) error {
			return errMetalTest
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}

		return []*mpatch.Patch{pathchFileExists, pathchChmod}
	}
	func_runNormal_ok := func(ctrl *gomock.Controller) []*mpatch.Patch {
		multiplePatchGenCertAndConfig(t, nil, func() {
			multiplePatchGenCertAndConfig(t, nil, func() {
				multiplePatchGenCertAndConfig(t, nil, nil)
			})
		})
		pathchFileExists, err := mpatch.PatchMethod(eputils.FileExists, func(filename string) bool {
			return true
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchChmod, err := mpatch.PatchMethod(os.Chmod, func(name string, mode fs.FileMode) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}

		return []*mpatch.Patch{pathchFileExists, pathchChmod}
	}
	type args struct {
		ep_params      *pluginapi.EpParams
		ironic_host_ip string
	}
	tests := []struct {
		name               string
		args               args
		expectErrorContent error
		funcBeforeTest     func(*gomock.Controller) []*mpatch.Patch
	}{
		{
			name: "GenCertAndConfig_err_1",
			args: args{
				ep_params: &pluginapi.EpParams{
					Workspace: "default",
				},
				ironic_host_ip: "",
			},
			expectErrorContent: errMetalTest,
			funcBeforeTest:     func_GenCertAndConfig_fail_1,
		},
		{
			name: "GenCertAndConfig_err_2",
			args: args{
				ep_params: &pluginapi.EpParams{
					Workspace: "default",
				},
				ironic_host_ip: "",
			},
			expectErrorContent: errMetalTest,
			funcBeforeTest:     func_GenCertAndConfig_fail_2,
		},
		{
			name: "GenCertAndConfig_err_3",
			args: args{
				ep_params: &pluginapi.EpParams{
					Workspace: "default",
				},
				ironic_host_ip: "",
			},
			expectErrorContent: errMetalTest,
			funcBeforeTest:     func_GenCertAndConfig_fail_3,
		},
		{
			name: "Chmod_MARIADBKEYFILE_err",
			args: args{
				ep_params: &pluginapi.EpParams{
					Workspace: "default",
				},
				ironic_host_ip: "",
			},
			expectErrorContent: errMetalTest,
			funcBeforeTest:     func_ChmodMARIADBKEYFIL_err,
		},
		{
			name: "runNormal_ok",
			args: args{
				ep_params: &pluginapi.EpParams{
					Workspace: "default",
				},
				ironic_host_ip: "",
			},
			expectErrorContent: errMetalTest,
			funcBeforeTest:     func_runNormal_ok,
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
			if result := ironic_tls_setup(tc.args.ep_params, tc.args.ironic_host_ip); result != nil {
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

//launchIronicContainers(ep_params *pluginapi.EpParams, workFolder string, clusterConfig *pluginapi.CapiClusterConfig, tmpl *capiutils.CapiTemplate) error {
func Test_launchIronicContainers(t *testing.T) {
	func_TmplFileRendering_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchTmplFileRendering, err := mpatch.PatchMethod(capiutils.TmplFileRendering, func(tmpl *capiutils.CapiTemplate, workFolder, url, dstFile string) error {
			return errMetalTest
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}

		return []*mpatch.Patch{pathchTmplFileRendering}
	}
	func_LoadSchemaStructFromYamlFile_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchTmplFileRendering, err := mpatch.PatchMethod(capiutils.TmplFileRendering, func(tmpl *capiutils.CapiTemplate, workFolder, url, dstFile string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchLoadSchemaStructFromYamlFile, err := mpatch.PatchMethod(eputils.LoadSchemaStructFromYamlFile, func(v eputils.SchemaStruct, file string) error {
			return errMetalTest
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		return []*mpatch.Patch{pathchTmplFileRendering, pathchLoadSchemaStructFromYamlFile}
	}

	func_runnormal_ok := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchTmplFileRendering, err := mpatch.PatchMethod(capiutils.TmplFileRendering, func(tmpl *capiutils.CapiTemplate, workFolder, url, dstFile string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchLoadSchemaStructFromYamlFile, err := mpatch.PatchMethod(eputils.LoadSchemaStructFromYamlFile, func(v eputils.SchemaStruct, file string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		return []*mpatch.Patch{pathchTmplFileRendering, pathchLoadSchemaStructFromYamlFile}
	}
	type args struct {
		ep_params     *pluginapi.EpParams
		workFolder    string
		clusterConfig *pluginapi.CapiClusterConfig
		tmpl          *capiutils.CapiTemplate
	}
	tests := []struct {
		name               string
		args               args
		expectErrorContent error
		funcBeforeTest     func(*gomock.Controller) []*mpatch.Patch
	}{
		// TODO: Add test cases.
		{
			name: "TmplFileRendering_err",
			args: args{
				ep_params:  &pluginapi.EpParams{},
				workFolder: "",
				clusterConfig: &pluginapi.CapiClusterConfig{
					BaremetelOperator: &pluginapi.CapiClusterConfigBaremetelOperator{
						IronicContainers: "default",
					},
				},
				tmpl: nil,
			},
			expectErrorContent: errMetalTest,
			funcBeforeTest:     func_TmplFileRendering_fail,
		},
		{
			name: "LoadSchemaStructFromYamlFile_err",
			args: args{
				ep_params:  &pluginapi.EpParams{},
				workFolder: "",
				clusterConfig: &pluginapi.CapiClusterConfig{
					BaremetelOperator: &pluginapi.CapiClusterConfigBaremetelOperator{
						IronicContainers: "default",
					},
				},
				tmpl: nil,
			},
			expectErrorContent: errMetalTest,
			funcBeforeTest:     func_LoadSchemaStructFromYamlFile_fail,
		},
		{
			name: "DockerRun_err",
			args: args{
				ep_params:  &pluginapi.EpParams{},
				workFolder: "",
				clusterConfig: &pluginapi.CapiClusterConfig{
					BaremetelOperator: &pluginapi.CapiClusterConfigBaremetelOperator{
						IronicContainers: "default",
					},
				},
				tmpl: nil,
			},
			expectErrorContent: errMetalTest,
			funcBeforeTest:     func_runnormal_ok,
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
			if result := launchIronicContainers(tc.args.ep_params, tc.args.workFolder, tc.args.clusterConfig, tc.args.tmpl); result != nil {
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

func Test_launchBmo(t *testing.T) {
	func_TmplFileRendering_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchTmplFileRendering, err := mpatch.PatchMethod(capiutils.TmplFileRendering, func(tmpl *capiutils.CapiTemplate, workFolder, url, dstFile string) error {
			return errMetalTest
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}

		return []*mpatch.Patch{pathchTmplFileRendering}
	}

	func_YamlInstall_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchTmplFileRendering, err := mpatch.PatchMethod(capiutils.TmplFileRendering, func(tmpl *capiutils.CapiTemplate, workFolder, url, dstFile string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		fakeYamlDeployer := fakeserviceutils.FakeYamlDeployer{}
		pathchNewYamlDeployer, err := mpatch.PatchMethod(serviceutil.NewYamlDeployer, func(name string, namespace string, yamlfile string, x ...interface{}) serviceutil.YamlDeployerWrapper {
			return &fakeYamlDeployer
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchYamlInstall, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(&fakeYamlDeployer), "YamlInstall", func(*fakeserviceutils.FakeYamlDeployer, string) error {
			return errMetalTest
		})
		if err != nil {
			t.Fatal(err)
		}
		return []*mpatch.Patch{pathchTmplFileRendering, pathchNewYamlDeployer, pathchYamlInstall}
	}
	func_runNormal_ok := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchTmplFileRendering, err := mpatch.PatchMethod(capiutils.TmplFileRendering, func(tmpl *capiutils.CapiTemplate, workFolder, url, dstFile string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		fakeYamlDeployer := fakeserviceutils.FakeYamlDeployer{}
		pathchNewYamlDeployer, err := mpatch.PatchMethod(serviceutil.NewYamlDeployer, func(name string, namespace string, yamlfile string, x ...interface{}) serviceutil.YamlDeployerWrapper {
			return &fakeYamlDeployer
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchYamlInstall, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(&fakeYamlDeployer), "YamlInstall", func(*fakeserviceutils.FakeYamlDeployer, string) error {
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		return []*mpatch.Patch{pathchTmplFileRendering, pathchNewYamlDeployer, pathchYamlInstall}
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
		// TODO: Add test cases. clusterConfig.BaremetelOperator.URL,
		{
			name: "TmplFileRendering_err",
			args: args{
				ep_params:             &pluginapi.EpParams{},
				workFolder:            "",
				management_kubeconfig: "",
				clusterConfig: &pluginapi.CapiClusterConfig{
					BaremetelOperator: &pluginapi.CapiClusterConfigBaremetelOperator{
						URL: "10.10.10.10",
					},
				},
				tmpl: nil,
			},
			expectErrorContent: errMetalTest,
			funcBeforeTest:     func_TmplFileRendering_fail,
		},
		{
			name: "YamlInstall_err",
			args: args{
				ep_params:             &pluginapi.EpParams{},
				workFolder:            "",
				management_kubeconfig: "",
				clusterConfig: &pluginapi.CapiClusterConfig{
					BaremetelOperator: &pluginapi.CapiClusterConfigBaremetelOperator{
						URL: "10.10.10.10",
					},
					WorkloadCluster: &pluginapi.CapiClusterConfigWorkloadCluster{
						Namespace: "defalut",
					},
				},
				tmpl: nil,
			},
			expectErrorContent: errMetalTest,
			funcBeforeTest:     func_YamlInstall_fail,
		},
		{
			name: "runNormal_ok",
			args: args{
				ep_params:             &pluginapi.EpParams{},
				workFolder:            "",
				management_kubeconfig: "",
				clusterConfig: &pluginapi.CapiClusterConfig{
					BaremetelOperator: &pluginapi.CapiClusterConfigBaremetelOperator{
						URL: "10.10.10.10",
					},
					WorkloadCluster: &pluginapi.CapiClusterConfigWorkloadCluster{
						Namespace: "defalut",
					},
				},
				tmpl: nil,
			},
			expectErrorContent: errMetalTest,
			funcBeforeTest:     func_runNormal_ok,
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
			if result := launchBmo(tc.args.ep_params, tc.args.workFolder, tc.args.management_kubeconfig, tc.args.clusterConfig, tc.args.tmpl); result != nil {
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

func Test_waitBmoDeploymentReady(t *testing.T) {
	func_Deployment_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchNewDeployment, err := mpatch.PatchMethod(kubeutils.NewDeployment, func(namespace string, name string, feildManager string, kubeconfig string) (kubeutils.DeploymentWrapper, error) {
			return nil, errMetalTest
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}

		return []*mpatch.Patch{pathchNewDeployment}
	}
	func_bmoDeploymentGet_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
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
			return errMetalTest
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		return []*mpatch.Patch{pathchNewDeployment, pathchfakeDeploymentGet}
	}
	func_bmoDeploymentGetStatus_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
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
		// TODO: Add test cases.
		{
			name: "NewDeployment_err",
			args: args{
				management_kubeconfig: "",
			},
			expectErrorContent: errMetalTest,
			funcBeforeTest:     func_Deployment_fail,
		},
		{
			name: "bmoDeploymentGet_err",
			args: args{
				management_kubeconfig: "",
			},
			expectErrorContent: errMetalTest,
			funcBeforeTest:     func_bmoDeploymentGet_fail,
		},
		{
			name: "byohCtlMgrDeployment_ok",
			args: args{
				management_kubeconfig: "",
			},
			funcBeforeTest: func_bmoDeploymentGetStatus_fail,
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
			if result := DeploymentReady(tc.args.management_kubeconfig, "baremetal-operator-system", "baremetal-operator-controller-manager"); result != nil {
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

func Test_checkBmHosts(t *testing.T) {
	func_RUNCMD_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchRunCMD, err := mpatch.PatchMethod(eputils.RunCMD, func(cmd *exec.Cmd) (string, error) {
			return "", errMetalTest
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}

		return []*mpatch.Patch{pathchRunCMD}
	}
	func_NodeNotReady_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchRunCMD, err := mpatch.PatchMethod(eputils.RunCMD, func(cmd *exec.Cmd) (string, error) {
			return "a", nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		patchSleep, err := mpatch.PatchMethod(time.Sleep, func(d time.Duration) {})
		if err != nil {
			t.Fatal(err)
		}
		//`.*\savaliable\s.*`
		return []*mpatch.Patch{pathchRunCMD, patchSleep}
	}
	func_CheckBmHostsReady_ok := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchRunCMD, err := mpatch.PatchMethod(eputils.RunCMD, func(cmd *exec.Cmd) (string, error) {
			return "node-1 available metal3-workers-7sfxj true 17h", nil
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
			expectErrorContent: errMetalTest,
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
			name: "CheckBmHostsReady_ok",
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
			funcBeforeTest: func_CheckBmHostsReady_ok,
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
			if result := checkBmHosts(tc.args.ep_params, tc.args.workFolder, tc.args.management_kubeconfig, tc.args.clusterConfig, tc.args.tmpl); result != nil {
				if fmt.Sprint(result) == fmt.Sprint(tc.expectErrorContent) {

					t.Log(tc.name, "Done")
				} else {
					t1 := fmt.Sprint(result)
					_ = t1
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

func Test_metal3HostProvision(t *testing.T) {
	func_ironic_tls_setup_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchironic_tls_setup, err := mpatch.PatchMethod(ironic_tls_setup, func(ep_params *pluginapi.EpParams, ironic_host_ip string) error {
			return errMetalTest
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}

		return []*mpatch.Patch{pathchironic_tls_setup}
	}
	func_launchBmo_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchironic_tls_setup, err := mpatch.PatchMethod(ironic_tls_setup, func(ep_params *pluginapi.EpParams, ironic_host_ip string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchlaunchBmo, err := mpatch.PatchMethod(launchBmo, func(ep_params *pluginapi.EpParams, workFolder string, management_kubeconfig string, clusterConfig *pluginapi.CapiClusterConfig, tmpl *capiutils.CapiTemplate) error {
			return errMetalTest
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		return []*mpatch.Patch{pathchironic_tls_setup, pathchlaunchBmo}
	}
	func_waitBmoDeploymentReady_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchironic_tls_setup, err := mpatch.PatchMethod(ironic_tls_setup, func(ep_params *pluginapi.EpParams, ironic_host_ip string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchlaunchBmo, err := mpatch.PatchMethod(launchBmo, func(ep_params *pluginapi.EpParams, workFolder string, management_kubeconfig string, clusterConfig *pluginapi.CapiClusterConfig, tmpl *capiutils.CapiTemplate) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchwaitBmoDeploymentReady, err := mpatch.PatchMethod(DeploymentReady, func(management_kubeconfig, namespace, deploymentName string) error {
			return errMetalTest
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		return []*mpatch.Patch{pathchironic_tls_setup, pathchlaunchBmo, pathchwaitBmoDeploymentReady}
	}
	func_launchIronicContainers_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchironic_tls_setup, err := mpatch.PatchMethod(ironic_tls_setup, func(ep_params *pluginapi.EpParams, ironic_host_ip string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchlaunchBmo, err := mpatch.PatchMethod(launchBmo, func(ep_params *pluginapi.EpParams, workFolder string, management_kubeconfig string, clusterConfig *pluginapi.CapiClusterConfig, tmpl *capiutils.CapiTemplate) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchwaitBmoDeploymentReady, err := mpatch.PatchMethod(DeploymentReady, func(management_kubeconfig, namespace, deploymentName string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchlaunchIronicContainers, err := mpatch.PatchMethod(launchIronicContainers, func(ep_params *pluginapi.EpParams, workFolder string, clusterConfig *pluginapi.CapiClusterConfig, tmpl *capiutils.CapiTemplate) error {
			return errMetalTest
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		return []*mpatch.Patch{pathchironic_tls_setup, pathchlaunchBmo, pathchwaitBmoDeploymentReady, pathchlaunchIronicContainers}
	}
	func_makeBmHosts_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchironic_tls_setup, err := mpatch.PatchMethod(ironic_tls_setup, func(ep_params *pluginapi.EpParams, ironic_host_ip string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchlaunchBmo, err := mpatch.PatchMethod(launchBmo, func(ep_params *pluginapi.EpParams, workFolder string, management_kubeconfig string, clusterConfig *pluginapi.CapiClusterConfig, tmpl *capiutils.CapiTemplate) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchwaitBmoDeploymentReady, err := mpatch.PatchMethod(DeploymentReady, func(management_kubeconfig, namespace, deploymentName string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchlaunchIronicContainers, err := mpatch.PatchMethod(launchIronicContainers, func(ep_params *pluginapi.EpParams, workFolder string, clusterConfig *pluginapi.CapiClusterConfig, tmpl *capiutils.CapiTemplate) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchmakeBmHosts, err := mpatch.PatchMethod(makeBmHosts, func(ep_params *pluginapi.EpParams, workFolder string, management_kubeconfig string, clusterConfig *pluginapi.CapiClusterConfig, tmpl *capiutils.CapiTemplate) error {
			return errMetalTest
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		return []*mpatch.Patch{pathchironic_tls_setup, pathchlaunchBmo, pathchwaitBmoDeploymentReady, pathchlaunchIronicContainers, pathchmakeBmHosts}
	}
	func_checkBmHosts_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchironic_tls_setup, err := mpatch.PatchMethod(ironic_tls_setup, func(ep_params *pluginapi.EpParams, ironic_host_ip string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchlaunchBmo, err := mpatch.PatchMethod(launchBmo, func(ep_params *pluginapi.EpParams, workFolder string, management_kubeconfig string, clusterConfig *pluginapi.CapiClusterConfig, tmpl *capiutils.CapiTemplate) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchwaitBmoDeploymentReady, err := mpatch.PatchMethod(DeploymentReady, func(management_kubeconfig, namespace, deploymentName string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchlaunchIronicContainers, err := mpatch.PatchMethod(launchIronicContainers, func(ep_params *pluginapi.EpParams, workFolder string, clusterConfig *pluginapi.CapiClusterConfig, tmpl *capiutils.CapiTemplate) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchmakeBmHosts, err := mpatch.PatchMethod(makeBmHosts, func(ep_params *pluginapi.EpParams, workFolder string, management_kubeconfig string, clusterConfig *pluginapi.CapiClusterConfig, tmpl *capiutils.CapiTemplate) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchcheckBmHosts, err := mpatch.PatchMethod(checkBmHosts, func(ep_params *pluginapi.EpParams, workFolder string, management_kubeconfig string, clusterConfig *pluginapi.CapiClusterConfig, tmpl *capiutils.CapiTemplate) error {
			return errMetalTest
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		return []*mpatch.Patch{pathchironic_tls_setup, pathchlaunchBmo, pathchwaitBmoDeploymentReady, pathchlaunchIronicContainers, pathchmakeBmHosts, pathchcheckBmHosts}
	}
	func_runNormal_ok := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchironic_tls_setup, err := mpatch.PatchMethod(ironic_tls_setup, func(ep_params *pluginapi.EpParams, ironic_host_ip string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchlaunchBmo, err := mpatch.PatchMethod(launchBmo, func(ep_params *pluginapi.EpParams, workFolder string, management_kubeconfig string, clusterConfig *pluginapi.CapiClusterConfig, tmpl *capiutils.CapiTemplate) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchwaitBmoDeploymentReady, err := mpatch.PatchMethod(DeploymentReady, func(management_kubeconfig, namespace, deploymentName string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchlaunchIronicContainers, err := mpatch.PatchMethod(launchIronicContainers, func(ep_params *pluginapi.EpParams, workFolder string, clusterConfig *pluginapi.CapiClusterConfig, tmpl *capiutils.CapiTemplate) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchmakeBmHosts, err := mpatch.PatchMethod(makeBmHosts, func(ep_params *pluginapi.EpParams, workFolder string, management_kubeconfig string, clusterConfig *pluginapi.CapiClusterConfig, tmpl *capiutils.CapiTemplate) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchcheckBmHosts, err := mpatch.PatchMethod(checkBmHosts, func(ep_params *pluginapi.EpParams, workFolder string, management_kubeconfig string, clusterConfig *pluginapi.CapiClusterConfig, tmpl *capiutils.CapiTemplate) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		return []*mpatch.Patch{pathchironic_tls_setup, pathchlaunchBmo, pathchwaitBmoDeploymentReady, pathchlaunchIronicContainers, pathchmakeBmHosts, pathchcheckBmHosts}
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
		// TODO: Add test cases. launchIronicContainers(ep_params, workFolder, clusterConfig, tmpl)
		{
			name: "ironic_tls_setup_err",
			args: args{
				ep_params:             &pluginapi.EpParams{},
				workFolder:            "",
				management_kubeconfig: "",
				clusterConfig:         &pluginapi.CapiClusterConfig{},
				tmpl: &capiutils.CapiTemplate{
					CapiSetting: pluginapi.CapiSetting{
						IronicConfig: &pluginapi.CapiSettingIronicConfig{
							IronicProvisionIP: "10.10.10.10",
						},
					},
				},
			},
			expectErrorContent: errMetalTest,
			funcBeforeTest:     func_ironic_tls_setup_fail,
		},
		{
			name: "launchBmo_err",
			args: args{
				ep_params:             &pluginapi.EpParams{},
				workFolder:            "",
				management_kubeconfig: "",
				clusterConfig:         &pluginapi.CapiClusterConfig{},
				tmpl: &capiutils.CapiTemplate{
					CapiSetting: pluginapi.CapiSetting{
						IronicConfig: &pluginapi.CapiSettingIronicConfig{
							IronicProvisionIP: "10.10.10.10",
						},
					},
				},
			},
			expectErrorContent: errMetalTest,
			funcBeforeTest:     func_launchBmo_fail,
		},
		{
			name: "waitBmoDeploymentReady_err",
			args: args{
				ep_params:             &pluginapi.EpParams{},
				workFolder:            "",
				management_kubeconfig: "",
				clusterConfig:         &pluginapi.CapiClusterConfig{},
				tmpl: &capiutils.CapiTemplate{
					CapiSetting: pluginapi.CapiSetting{
						IronicConfig: &pluginapi.CapiSettingIronicConfig{
							IronicProvisionIP: "10.10.10.10",
						},
					},
				},
			},
			expectErrorContent: errMetalTest,
			funcBeforeTest:     func_waitBmoDeploymentReady_fail,
		},
		{
			name: "launchIronicContainers_err",
			args: args{
				ep_params:             &pluginapi.EpParams{},
				workFolder:            "",
				management_kubeconfig: "",
				clusterConfig:         &pluginapi.CapiClusterConfig{},
				tmpl: &capiutils.CapiTemplate{
					CapiSetting: pluginapi.CapiSetting{
						IronicConfig: &pluginapi.CapiSettingIronicConfig{
							IronicProvisionIP: "10.10.10.10",
						},
					},
				},
			},
			expectErrorContent: errMetalTest,
			funcBeforeTest:     func_launchIronicContainers_fail,
		},
		{
			name: "makeBmHosts_err",
			args: args{
				ep_params:             &pluginapi.EpParams{},
				workFolder:            "",
				management_kubeconfig: "",
				clusterConfig:         &pluginapi.CapiClusterConfig{},
				tmpl: &capiutils.CapiTemplate{
					CapiSetting: pluginapi.CapiSetting{
						IronicConfig: &pluginapi.CapiSettingIronicConfig{
							IronicProvisionIP: "10.10.10.10",
						},
					},
				},
			},
			expectErrorContent: errMetalTest,
			funcBeforeTest:     func_makeBmHosts_fail,
		},
		{
			name: "checkBmHosts_err",
			args: args{
				ep_params:             &pluginapi.EpParams{},
				workFolder:            "",
				management_kubeconfig: "",
				clusterConfig:         &pluginapi.CapiClusterConfig{},
				tmpl: &capiutils.CapiTemplate{
					CapiSetting: pluginapi.CapiSetting{
						IronicConfig: &pluginapi.CapiSettingIronicConfig{
							IronicProvisionIP: "10.10.10.10",
						},
					},
				},
			},
			expectErrorContent: errMetalTest,
			funcBeforeTest:     func_checkBmHosts_fail,
		},
		{
			name: "runNormal_ok",
			args: args{
				ep_params:             &pluginapi.EpParams{},
				workFolder:            "",
				management_kubeconfig: "",
				clusterConfig:         &pluginapi.CapiClusterConfig{},
				tmpl: &capiutils.CapiTemplate{
					CapiSetting: pluginapi.CapiSetting{
						IronicConfig: &pluginapi.CapiSettingIronicConfig{
							IronicProvisionIP: "10.10.10.10",
						},
					},
				},
			},
			expectErrorContent: errMetalTest,
			funcBeforeTest:     func_runNormal_ok,
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
			if result := metal3HostProvision(tc.args.ep_params, tc.args.workFolder, tc.args.management_kubeconfig, tc.args.clusterConfig, tc.args.tmpl); result != nil {
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

func Test_makeBmHosts(t *testing.T) {
	func_TmplFileRendering_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchTmplFileRendering, err := mpatch.PatchMethod(capiutils.TmplFileRendering, func(tmpl *capiutils.CapiTemplate, workFolder, url, dstFile string) error {
			return errMetalTest
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}

		return []*mpatch.Patch{pathchTmplFileRendering}
	}
	func_YamlInstall_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchTmplFileRendering, err := mpatch.PatchMethod(capiutils.TmplFileRendering, func(tmpl *capiutils.CapiTemplate, workFolder, url, dstFile string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		fakeYamlDeployer := fakeserviceutils.FakeYamlDeployer{}
		pathchNewYamlDeployer, err := mpatch.PatchMethod(serviceutil.NewYamlDeployer, func(name string, namespace string, yamlfile string, x ...interface{}) serviceutil.YamlDeployerWrapper {
			return &fakeYamlDeployer
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchYamlInstall, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(&fakeYamlDeployer), "YamlInstall", func(*fakeserviceutils.FakeYamlDeployer, string) error {
			return errMetalTest
		})
		if err != nil {
			t.Fatal(err)
		}
		return []*mpatch.Patch{pathchTmplFileRendering, pathchNewYamlDeployer, pathchYamlInstall}
	}
	func_runNormal_ok := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchTmplFileRendering, err := mpatch.PatchMethod(capiutils.TmplFileRendering, func(tmpl *capiutils.CapiTemplate, workFolder, url, dstFile string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		fakeYamlDeployer := fakeserviceutils.FakeYamlDeployer{}
		pathchNewYamlDeployer, err := mpatch.PatchMethod(serviceutil.NewYamlDeployer, func(name string, namespace string, yamlfile string, x ...interface{}) serviceutil.YamlDeployerWrapper {
			return &fakeYamlDeployer
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchYamlInstall, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(&fakeYamlDeployer), "YamlInstall", func(*fakeserviceutils.FakeYamlDeployer, string) error {
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		return []*mpatch.Patch{pathchTmplFileRendering, pathchNewYamlDeployer, pathchYamlInstall}
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
		// TODO: Add test cases. clusterConfig.BaremetelOperator.URL,
		{
			name: "TmplFileRendering_err",
			args: args{
				ep_params:             &pluginapi.EpParams{},
				workFolder:            "",
				management_kubeconfig: "",
				clusterConfig: &pluginapi.CapiClusterConfig{
					BaremetelOperator: &pluginapi.CapiClusterConfigBaremetelOperator{
						URL: "10.10.10.10",
					},
				},
				tmpl: nil,
			},
			expectErrorContent: errMetalTest,
			funcBeforeTest:     func_TmplFileRendering_fail,
		},
		{
			name: "YamlInstall_err",
			args: args{
				ep_params:             &pluginapi.EpParams{},
				workFolder:            "",
				management_kubeconfig: "",
				clusterConfig: &pluginapi.CapiClusterConfig{
					BaremetelOperator: &pluginapi.CapiClusterConfigBaremetelOperator{
						URL: "10.10.10.10",
					},
					WorkloadCluster: &pluginapi.CapiClusterConfigWorkloadCluster{
						Namespace: "defalut",
					},
				},
				tmpl: nil,
			},
			expectErrorContent: errMetalTest,
			funcBeforeTest:     func_YamlInstall_fail,
		},
		{
			name: "runNormal_ok",
			args: args{
				ep_params:             &pluginapi.EpParams{},
				workFolder:            "",
				management_kubeconfig: "",
				clusterConfig: &pluginapi.CapiClusterConfig{
					BaremetelOperator: &pluginapi.CapiClusterConfigBaremetelOperator{
						URL: "10.10.10.10",
					},
					WorkloadCluster: &pluginapi.CapiClusterConfigWorkloadCluster{
						Namespace: "defalut",
					},
				},
				tmpl: nil,
			},
			expectErrorContent: errMetalTest,
			funcBeforeTest:     func_runNormal_ok,
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
			if result := makeBmHosts(tc.args.ep_params, tc.args.workFolder, tc.args.management_kubeconfig, tc.args.clusterConfig, tc.args.tmpl); result != nil {
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
