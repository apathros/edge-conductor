/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

//nolint: dupl
package service

import (
	cmapi "ep/pkg/api/certmgr"
	epplugins "ep/pkg/api/plugins"
	certmgr "ep/pkg/certmgr"
	eputils "ep/pkg/eputils"
	kubeutils "ep/pkg/eputils/kubeutils"
	fakekubeutils "ep/pkg/eputils/test/fakekubeutils"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	mpatch "github.com/undefinedlabs/go-mpatch"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/release"
)

var (
	kubeerr    = errors.New("kubernetes error")
	errRelease = errors.New("release: not found")
	testerrmsg = "test_error"
	testerr    = errors.New(testerrmsg)
)

func unpatch(t *testing.T, m *mpatch.Patch) {
	err := m.Unpatch()
	if err != nil {
		t.Fatal(err)
	}
}

func checkError(t *testing.T, err, expectErr error) {
	if expectErr != nil {
		if err == nil {
			t.Error("Expected error but no error found.")
		} else {
			if fmt.Sprint(err) == fmt.Sprint(expectErr) {
				t.Log("Error expected.")
			} else {
				t.Error("Expect:", expectErr, "; But found:", err)
			}
		}
	} else {
		if err != nil {
			t.Error("Unexpected Error:", err)
		}
	}
}

func TestYamlDeployer(t *testing.T) {
	p, err := mpatch.PatchMethod(eputils.RunCMD, func(*exec.Cmd) (string, error) { return "", nil })
	if err != nil {
		t.Fatal(err)
	}
	defer unpatch(t, p)
	wait := &YamlWait{Timeout: 100}
	yamlDeployer := NewYamlDeployer("test", "ns", "testdata/kind-nginx-ingress.yml", wait)
	if "test" != yamlDeployer.GetName() {
		t.Error("Unexpected name found:", yamlDeployer.GetName())
	}
	if err := yamlDeployer.YamlInstall("kubeconfig"); err != nil {
		t.Error("Unexpected error found.")
	}
	if err := yamlDeployer.YamlUninstall("kubeconfig"); err != nil {
		t.Error("Unexpected error found.")
	}
	yamlDeployerNowait := NewYamlDeployer("test", "ns", "yamlName")
	if "test" != yamlDeployerNowait.GetName() {
		t.Error("Unexpected name found:", yamlDeployerNowait.GetName())
	}
	if err := yamlDeployerNowait.YamlInstall("kubeconfig"); err != nil {
		t.Error("Unexpected error found.")
	}
	if err := yamlDeployerNowait.YamlUninstall("kubeconfig"); err != nil {
		t.Error("Unexpected error found.")
	}
}

func patch_helmStatus(t *testing.T, mockfunc func(*action.Status, string) (*release.Release, error)) *mpatch.Patch {
	cli := action.NewStatus(new(action.Configuration))
	p, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "Run", mockfunc)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func patch_helmInstall(t *testing.T, mockfunc func(*action.Install, *chart.Chart, map[string]interface{}) (*release.Release, error)) *mpatch.Patch {
	cli := action.NewInstall(new(action.Configuration))
	p, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "Run", mockfunc)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func patch_helmUpgrade(t *testing.T, mockfunc func(*action.Upgrade, string, *chart.Chart, map[string]interface{}) (*release.Release, error)) *mpatch.Patch {
	cli := action.NewUpgrade(new(action.Configuration))
	p, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "Run", mockfunc)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func patch_chartLoad(t *testing.T) []*mpatch.Patch {
	testdep := &chart.Dependency{}
	testchart := &chart.Chart{Metadata: &chart.Metadata{Dependencies: []*chart.Dependency{testdep}}}

	p1, err := mpatch.PatchMethod(loader.Load, func(string) (*chart.Chart, error) { return testchart, nil })
	if err != nil {
		t.Fatal(err)
	}
	p2, err := mpatch.PatchMethod(loadChartRemote, func(string) (*chart.Chart, error) { return testchart, nil })
	if err != nil {
		t.Fatal(err)
	}
	p3, err := mpatch.PatchMethod(action.CheckDependencies, func(*chart.Chart, []*chart.Dependency) error { return kubeerr })
	if err != nil {
		t.Fatal(err)
	}
	dlMgr := &downloader.Manager{}
	p4, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(dlMgr), "Update", func(*downloader.Manager) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	return []*mpatch.Patch{p1, p2, p3, p4}
}

func patch_helmUninstall(t *testing.T, mockfunc func(*action.Uninstall, string) (*release.UninstallReleaseResponse, error)) *mpatch.Patch {
	cli := action.NewUninstall(new(action.Configuration))
	p, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli), "Run", mockfunc)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestHelmDeployer(t *testing.T) {
	helmDeployer := NewHelmDeployer("test", "ns", "charts", "testdata/override.yml")

	// General functions to run before test
	func_err_initHelm := func() []*mpatch.Patch {
		p1, err := mpatch.PatchMethod(initHelm, func(string, string) error { return kubeerr })
		if err != nil {
			t.Fatal(err)
		}
		return []*mpatch.Patch{p1}
	}

	func_err_readValues := func() []*mpatch.Patch {
		p1, err := mpatch.PatchMethod(readValues, func(string) (map[string]interface{}, error) { return nil, kubeerr })
		if err != nil {
			t.Fatal(err)
		}
		return []*mpatch.Patch{p1}
	}

	// Unit Test: helmDeployer.GetName()
	if "test" != helmDeployer.GetName() {
		t.Error("Unexpected name found:", helmDeployer.GetName())
	}

	// Unit Test: helmDeployer.HelmStatus()

	// mock functions for helmStatus
	func_status_run_ok := func(*action.Status, string) (*release.Release, error) {
		return &release.Release{
			Info:    &release.Info{Status: "Test"},
			Version: 1,
		}, nil
	}
	func_status_run_err := func(*action.Status, string) (*release.Release, error) {
		return nil, kubeerr
	}
	func_status_run_err_notdeployed := func(*action.Status, string) (*release.Release, error) {
		return nil, errRelease
	}
	// cases
	cases_status := []struct {
		name           string
		expectStatus   string
		expectRev      int
		funcMock       func(*action.Status, string) (*release.Release, error)
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			name:         "err_test_wrong_release_status",
			expectStatus: HELM_STATUS_UNKNOWN,
			funcMock:     func_status_run_err,
		},
		{
			name:           "err_test_status_inithelm",
			expectStatus:   HELM_STATUS_UNKNOWN,
			expectRev:      0,
			funcMock:       func_status_run_ok,
			funcBeforeTest: func_err_initHelm,
		},
		{
			name:         "test_release_status_notdeployed",
			expectStatus: HELM_STATUS_NOT_DEPLOYED,
			funcMock:     func_status_run_err_notdeployed,
		},
		{
			name:         "test_release_status",
			expectStatus: "Test",
			expectRev:    1,
			funcMock:     func_status_run_ok,
		},
	}
	for _, tc := range cases_status {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				plist := tc.funcBeforeTest()
				for _, p := range plist {
					defer unpatch(t, p)
				}
			}
			if tc.funcMock != nil {
				p := patch_helmStatus(t, tc.funcMock)
				defer unpatch(t, p)
			}

			status, rev := helmDeployer.HelmStatus("kubeconfig")

			if status != tc.expectStatus {
				t.Error("Expected", tc.expectStatus, "but got", status)
			}
			if rev != tc.expectRev {
				t.Error("Expected", tc.expectRev, "but got", rev)
			}
			t.Log("Done")

		})
	}

	// Unit Test: helmDeployer.HelmInstall()

	// mock functions for helmInstall
	func_install_run_ok := func(*action.Install, *chart.Chart, map[string]interface{}) (*release.Release, error) {
		return &release.Release{
			Info:    &release.Info{Status: "Test"},
			Version: 1,
		}, nil
	}
	func_install_run_err := func(*action.Install, *chart.Chart, map[string]interface{}) (*release.Release, error) {
		return nil, kubeerr
	}
	// cases
	cases_install := []struct {
		name           string
		expectErr      error
		funcMock       func(*action.Install, *chart.Chart, map[string]interface{}) (*release.Release, error)
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			name:      "err_test_install_failed",
			expectErr: kubeerr,
			funcMock:  func_install_run_err,
		},
		{
			name:           "err_test_install_inithelm",
			expectErr:      kubeerr,
			funcMock:       func_install_run_ok,
			funcBeforeTest: func_err_initHelm,
		},
		{
			name:           "err_test_install_readvalues",
			expectErr:      kubeerr,
			funcMock:       func_install_run_ok,
			funcBeforeTest: func_err_readValues,
		},
		{
			name:      "test_install_ok",
			expectErr: nil,
			funcMock:  func_install_run_ok,
		},
	}
	for _, tc := range cases_install {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				plist := tc.funcBeforeTest()
				for _, p := range plist {
					defer unpatch(t, p)
				}
			}
			if tc.funcMock != nil {
				p := patch_helmInstall(t, tc.funcMock)
				defer unpatch(t, p)
			}
			plist := patch_chartLoad(t)
			for _, p := range plist {
				defer unpatch(t, p)
			}

			err := helmDeployer.HelmInstall("kubeconfig", WithWaitAndTimeout(false, 300))
			checkError(t, err, tc.expectErr)
			t.Log("Done")
		})

		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				plist := tc.funcBeforeTest()
				for _, p := range plist {
					defer unpatch(t, p)
				}
			}
			if tc.funcMock != nil {
				p := patch_helmInstall(t, tc.funcMock)
				defer unpatch(t, p)
			}
			plist := patch_chartLoad(t)
			for _, p := range plist {
				defer unpatch(t, p)
			}

			err := helmDeployer.HelmInstall("kubeconfig")
			checkError(t, err, tc.expectErr)
			t.Log("Done")
		})
	}

	// Unit Test: helmDeployer.HelmUpgrade()

	// mock functions for helmUpgrade
	func_upgrade_run_ok := func(*action.Upgrade, string, *chart.Chart, map[string]interface{}) (*release.Release, error) {
		return &release.Release{
			Info:    &release.Info{Status: "Test"},
			Version: 1,
		}, nil
	}
	func_upgrade_run_err := func(*action.Upgrade, string, *chart.Chart, map[string]interface{}) (*release.Release, error) {
		return nil, kubeerr
	}
	// cases
	cases_upgrade := []struct {
		name           string
		expectErr      error
		funcMock       func(*action.Upgrade, string, *chart.Chart, map[string]interface{}) (*release.Release, error)
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			name:      "err_test_upgrade_failed",
			expectErr: kubeerr,
			funcMock:  func_upgrade_run_err,
		},
		{
			name:           "err_test_upgrade_inithelm",
			expectErr:      kubeerr,
			funcMock:       func_upgrade_run_ok,
			funcBeforeTest: func_err_initHelm,
		},
		{
			name:           "err_test_upgrade_readvalues",
			expectErr:      kubeerr,
			funcMock:       func_upgrade_run_ok,
			funcBeforeTest: func_err_readValues,
		},
		{
			name:      "test_upgrade_ok",
			expectErr: nil,
			funcMock:  func_upgrade_run_ok,
		},
	}
	for _, tc := range cases_upgrade {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				plist := tc.funcBeforeTest()
				for _, p := range plist {
					defer unpatch(t, p)
				}
			}
			if tc.funcMock != nil {
				p := patch_helmUpgrade(t, tc.funcMock)
				defer unpatch(t, p)
			}
			plist := patch_chartLoad(t)
			for _, p := range plist {
				defer unpatch(t, p)
			}

			err := helmDeployer.HelmUpgrade("kubeconfig")
			checkError(t, err, tc.expectErr)
			t.Log("Done")
		})
	}

	// Unit Test: helmDeployer.HelmUninstall()

	// mock functions for helmUninstall
	func_uninstall_run_ok := func(*action.Uninstall, string) (*release.UninstallReleaseResponse, error) {
		return &release.UninstallReleaseResponse{
			Info: "ok",
		}, nil
	}
	func_uninstall_run_err := func(*action.Uninstall, string) (*release.UninstallReleaseResponse, error) {
		return nil, kubeerr
	}
	// cases
	cases_uninstall := []struct {
		name           string
		expectErr      error
		funcMock       func(*action.Uninstall, string) (*release.UninstallReleaseResponse, error)
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			name:      "err_test_uninstall_failed",
			expectErr: kubeerr,
			funcMock:  func_uninstall_run_err,
		},
		{
			name:           "err_test_uninstall_inithelm",
			expectErr:      kubeerr,
			funcMock:       func_uninstall_run_ok,
			funcBeforeTest: func_err_initHelm,
		},
		{
			name:      "test_uninstall_ok",
			expectErr: nil,
			funcMock:  func_uninstall_run_ok,
		},
	}
	for _, tc := range cases_uninstall {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				plist := tc.funcBeforeTest()
				for _, p := range plist {
					defer unpatch(t, p)
				}
			}
			if tc.funcMock != nil {
				p := patch_helmUninstall(t, tc.funcMock)
				defer unpatch(t, p)
			}

			err := helmDeployer.HelmUninstall("kubeconfig")
			checkError(t, err, tc.expectErr)
			t.Log("Done")
		})
	}
}

func TestHelmDeployerUtils(t *testing.T) {
	var err error
	var p *mpatch.Patch
	resp := &http.Response{Body: io.NopCloser(strings.NewReader("test reader"))}

	// Unit Test: loadChartRemote()
	p, err = mpatch.PatchMethod(http.Get, func(string) (*http.Response, error) { return nil, kubeerr })
	if err != nil {
		t.Fatal(err)
	}
	_, err = loadChartRemote("http://wrongurl")
	checkError(t, err, kubeerr)
	unpatch(t, p)

	p, err = mpatch.PatchMethod(http.Get, func(string) (*http.Response, error) { return resp, nil })
	if err != nil {
		t.Fatal(err)
	}
	guard, err := mpatch.PatchMethod(loader.LoadArchive, func(io.Reader) (*chart.Chart, error) { return &chart.Chart{}, nil })
	if err != nil {
		t.Fatal(err)
	}
	_, err = loadChartRemote("http://wrongurl")
	if err != nil {
		t.Error("Unexpected error.")
	}
	unpatch(t, guard)

	guard, err = mpatch.PatchMethod(loader.LoadArchive, func(io.Reader) (*chart.Chart, error) { return nil, kubeerr })
	if err != nil {
		t.Fatal(err)
	}
	_, err = loadChartRemote("http://wrongurl")
	checkError(t, err, kubeerr)
	unpatch(t, guard)

	unpatch(t, p)

	// Unit Test: loadChartLocal()
	p, err = mpatch.PatchMethod(loader.Load, func(string) (*chart.Chart, error) { return &chart.Chart{}, nil })
	if err != nil {
		t.Fatal(err)
	}
	_, err = loadChartLocal("testdata/override.yml")
	if err != nil {
		t.Error("Unexpected error.")
	}
	unpatch(t, p)

	p, err = mpatch.PatchMethod(loader.Load, func(string) (*chart.Chart, error) { return &chart.Chart{}, kubeerr })
	if err != nil {
		t.Fatal(err)
	}
	_, err = loadChartLocal("testdata/override.yml")
	checkError(t, err, kubeerr)
	unpatch(t, p)

	// Unit Test: loadChartRemote()
	p, err = mpatch.PatchMethod(http.Get, func(string) (*http.Response, error) { return resp, nil })
	if err != nil {
		t.Fatal(err)
	}
	_, err = readFile("http://testurl")
	if err != nil {
		t.Error("Unexpected error.")
	}
	unpatch(t, p)

	p, err = mpatch.PatchMethod(http.Get, func(string) (*http.Response, error) { return nil, kubeerr })
	if err != nil {
		t.Fatal(err)
	}
	_, err = readFile("http://testurl")
	checkError(t, err, kubeerr)
	unpatch(t, p)
}

func TestInitHelmError(t *testing.T) {
	p1, err := mpatch.PatchMethod(filepath.Abs, func(string) (string, error) { return "", testerr })
	if err != nil {
		t.Fatal(err)
	}
	defer unpatch(t, p1)

	if err := initHelm("testconfig", "testspace"); errors.Is(err, testerr) {
		t.Log("Error Expected")
	} else {
		t.Error("Expect", testerr, "but found", err)
	}
}

func TestGenSvcTLSCertFromTLSExtension(t *testing.T) {
	func_GetCertBundleByName_err := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchGetCertBundleByName, err := mpatch.PatchMethod(certmgr.GetCertBundleByName, func(cname string, ctype string) (*cmapi.Certificate, certmgr.CertType, error) {
			return nil, -1, testerr
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}

		return []*mpatch.Patch{pathchGetCertBundleByName}
	}
	func_GenCertAndConfig_err := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchGetCertBundleByName, err := mpatch.PatchMethod(certmgr.GetCertBundleByName, func(cname string, ctype string) (*cmapi.Certificate, certmgr.CertType, error) {
			return &cmapi.Certificate{
				Ca: &cmapi.CertificateCa{
					Cert: "defalut",
					Csr:  "defalut",
					Key:  "defalut",
				},
			}, 1, nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchGenCertAndConfig, err := mpatch.PatchMethod(certmgr.GenCertAndConfig, func(certbundle cmapi.Certificate, hosts string) error {
			return testerr
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		return []*mpatch.Patch{pathchGetCertBundleByName, pathchGenCertAndConfig}
	}
	func_runnormal_ok := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchGetCertBundleByName, err := mpatch.PatchMethod(certmgr.GetCertBundleByName, func(cname string, ctype string) (*cmapi.Certificate, certmgr.CertType, error) {
			return &cmapi.Certificate{
				Ca: &cmapi.CertificateCa{
					Cert: "defalut",
					Csr:  "defalut",
					Key:  "defalut",
				},
			}, 1, nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchGenCertAndConfig, err := mpatch.PatchMethod(certmgr.GenCertAndConfig, func(certbundle cmapi.Certificate, hosts string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		return []*mpatch.Patch{pathchGetCertBundleByName, pathchGenCertAndConfig}
	}
	// cases
	type args struct {
		exts []*epplugins.EpParamsExtensionsItems0
		svc  string
	}
	cases := []struct {
		name               string
		args               args
		expectErrorContent error
		funcBeforeTest     func(*gomock.Controller) []*mpatch.Patch
	}{
		{
			name: "ExtensionNotFound_err",
			args: args{
				exts: []*epplugins.EpParamsExtensionsItems0{
					{
						Name:      "service-tls",
						Extension: nil,
					},
				},
				svc: "test_svc",
			},
			expectErrorContent: eputils.GetError("errExtNotFound"),
		},
		{
			name: "ConfigNotFound_err",
			args: args{
				exts: []*epplugins.EpParamsExtensionsItems0{
					{
						Name: "service-tls",
						Extension: &epplugins.Extension{
							Extension: []*epplugins.ExtensionItems0{
								{
									Config: nil,
								},
							},
						},
					},
				},
				svc: "test_svc",
			},
			expectErrorContent: eputils.GetError("errExtCfgNotFound"),
		},
		{
			name: "GetCertBundleByName_err",
			args: args{
				exts: []*epplugins.EpParamsExtensionsItems0{
					{
						Name: "service-tls",
						Extension: &epplugins.Extension{
							Extension: []*epplugins.ExtensionItems0{
								{
									Config: []*epplugins.ExtensionItems0ConfigItems0{
										{
											Name:  "service-name",
											Value: "test_svc",
										},
									},
								},
							},
						},
					},
				},
				svc: "test_svc",
			},
			expectErrorContent: testerr,
			funcBeforeTest:     func_GetCertBundleByName_err,
		},
		{
			name: "GenCertAndConfig_err",
			args: args{
				exts: []*epplugins.EpParamsExtensionsItems0{
					{
						Name: "service-tls",
						Extension: &epplugins.Extension{
							Extension: []*epplugins.ExtensionItems0{
								{
									Name: "test_svc",
									Config: []*epplugins.ExtensionItems0ConfigItems0{
										{
											Name:  "service-name",
											Value: "test_svc",
										},
										{
											Name:  "csr-filename",
											Value: "test_svc-csr.json",
										},
									},
								},
							},
						},
					},
				},
				svc: "test_svc",
			},
			expectErrorContent: testerr,
			funcBeforeTest:     func_GenCertAndConfig_err,
		},
		{
			name: "runnormal_ok",
			args: args{
				exts: []*epplugins.EpParamsExtensionsItems0{
					{
						Name: "service-tls",
						Extension: &epplugins.Extension{
							Extension: []*epplugins.ExtensionItems0{
								{
									Name: "test_svc",
									Config: []*epplugins.ExtensionItems0ConfigItems0{
										{
											Name:  "service-name",
											Value: "test_svc",
										},
										{
											Name:  "client-csr-filename",
											Value: "test_cli-csr.json",
										},
										{
											Name:  "csr-filename",
											Value: "test_svc-csr.json",
										},
									},
								},
							},
						},
					},
				},
				svc: "test_svc",
			},
			funcBeforeTest: func_runnormal_ok,
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
			if result := GenSvcTLSCertFromTLSExtension(tc.args.exts, tc.args.svc); result != nil {
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

func TestGenSvcSecretFromTLSExtension(t *testing.T) {
	func_GetCertBundleByName_err := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchGetCertBundleByName, err := mpatch.PatchMethod(certmgr.GetCertBundleByName, func(cname string, ctype string) (*cmapi.Certificate, certmgr.CertType, error) {
			return nil, -1, testerr
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}

		return []*mpatch.Patch{pathchGetCertBundleByName}
	}
	func_ca_secret_name_ok := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchGetCertBundleByName, err := mpatch.PatchMethod(certmgr.GetCertBundleByName, func(cname string, ctype string) (*cmapi.Certificate, certmgr.CertType, error) {
			return &cmapi.Certificate{
				Ca: &cmapi.CertificateCa{
					Cert: "defalut",
					Csr:  "defalut",
					Key:  "defalut",
				},
			}, 1, nil
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
		fakeSecret := fakekubeutils.FakeSecret{}
		pathchNewSecret, err := mpatch.PatchMethod(kubeutils.NewSecret, func(namespace string, name string, feildManager string, kubeconfig string) (kubeutils.SecretWrapper, error) {
			return &fakeSecret, nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchfakeNew, err := mpatch.PatchMethod(fakeSecret.New, func() error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchRenewStringData, err := mpatch.PatchMethod(fakeSecret.RenewStringData, func(key string, data string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		return []*mpatch.Patch{pathchGetCertBundleByName, pathchReadFile, pathchNewSecret, pathchfakeNew, pathchRenewStringData}
	}
	type args struct {
		exts    []*epplugins.EpParamsExtensionsItems0
		svc     string
		ns      string
		kubecfg string
	}
	cases := []struct {
		name               string
		args               args
		expectErrorContent error
		funcBeforeTest     func(*gomock.Controller) []*mpatch.Patch
	}{
		{
			name: "ExtensionNotFound_err",
			args: args{
				exts: []*epplugins.EpParamsExtensionsItems0{
					{
						Name:      "service-tls",
						Extension: nil,
					},
				},
				svc: "test_svc",
			},
			expectErrorContent: eputils.GetError("errExtNotFound"),
		},
		{
			name: "ConfigNotFound_err",
			args: args{
				exts: []*epplugins.EpParamsExtensionsItems0{
					{
						Name: "service-tls",
						Extension: &epplugins.Extension{
							Extension: []*epplugins.ExtensionItems0{
								{
									Config: nil,
								},
							},
						},
					},
				},
				svc: "test_svc",
			},
			expectErrorContent: eputils.GetError("errExtCfgNotFound"),
		},
		{
			name: "GetCertBundleByName_err",
			args: args{
				exts: []*epplugins.EpParamsExtensionsItems0{
					{
						Name: "service-tls",
						Extension: &epplugins.Extension{
							Extension: []*epplugins.ExtensionItems0{
								{
									Config: []*epplugins.ExtensionItems0ConfigItems0{
										{
											Name:  "service-name",
											Value: "test_svc",
										},
									},
								},
							},
						},
					},
				},
				svc: "test_svc",
			},
			expectErrorContent: testerr,
			funcBeforeTest:     func_GetCertBundleByName_err,
		},
		{
			name: "ca-secret-name_ok",
			args: args{
				exts: []*epplugins.EpParamsExtensionsItems0{
					{
						Name: "service-tls",
						Extension: &epplugins.Extension{
							Extension: []*epplugins.ExtensionItems0{
								{
									Name: "test_svc",
									Config: []*epplugins.ExtensionItems0ConfigItems0{
										{
											Name:  "service-name",
											Value: "test_svc",
										},
										{
											Name:  "ca-secret-name",
											Value: "test_svc-csr.json",
										},
									},
								},
							},
						},
					},
				},
				svc: "test_svc",
			},
			expectErrorContent: testerr,
			funcBeforeTest:     func_ca_secret_name_ok,
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
			if result := GenSvcSecretFromTLSExtension(tc.args.exts, tc.args.svc, tc.args.ns, tc.args.kubecfg); result != nil {
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
