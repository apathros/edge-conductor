/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Template auto-generated once, maintained by plugin owner.

package serviceparser

import (
	"fmt"
	"os"
	"testing"

	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"

	eputils "github.com/intel/edge-conductor/pkg/eputils"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	mpatch "github.com/undefinedlabs/go-mpatch"
)

var emptyChartNameInput = "Components:\n  - name: testhelmrepo\n    helmrepo: https://prometheus-community.github.io/helm-charts\n    chartversion: 14.1.3\n    chartoverride: file://testoverrideurl\n    supported-clusters:\n      - default\n    type: helm\n"
var emptyRepoInput = "Components:\n  - name: testhelmrepo\n    chartname: prometheus\n    chartversion: 14.1.3\n    chartoverride: file://testoverrideurl\n    supported-clusters:\n      - default\n    type: helm\n"

func unpatch(t *testing.T, m *mpatch.Patch) {
	err := m.Unpatch()
	if err != nil {
		t.Fatal(err)
	}
}

func TestPluginMain(t *testing.T) {
	success_find_helm_repofunc := func(ctrl *gomock.Controller) []*mpatch.Patch {
		p, err := mpatch.PatchMethod(repo.FindChartInRepoURL, FindFakeChartInRepoURL)
		if err != nil {
			t.Fatal(err)
		}
		return []*mpatch.Patch{p}
	}
	cases := []struct {
		name                  string
		input, expectedOutput map[string][]byte
		expectError           bool
		expectErrorMsg        string
		funcBeforeTest        func(*gomock.Controller) []*mpatch.Patch
	}{
		{
			name: "success_yaml",
			input: map[string][]byte{
				"ep-params": []byte(`{"kitconfig":{
					"Cluster":{"Provider":"kind"},
					"Components":{"selector":[{"name": "testyaml"}], "manifests":["testdata/fake.yml"]}},
					"Runtimedir": "testdata"
				}`),
			},
			expectedOutput: map[string][]byte{
				"serviceconfig": []byte(`{
				"Components":[{"images":null,"name":"testyaml","resources":null,"supported-clusters":["default"],"type":"yaml","url":"file://testyamlurl"}]}`),
				"downloadfiles": []byte(`{"files":[{"url":"file://testyamlurl","urlreplacement":{"new":"yaml/testyaml","origin":"file://"}}]}`),
				"docker-images": []byte(`{"images":null}`),
			},
			expectError:    false,
			expectErrorMsg: "",
			funcBeforeTest: success_find_helm_repofunc,
		},
		{
			name: "success_helm",
			input: map[string][]byte{
				"ep-params": []byte(`{"kitconfig":{
					"Cluster":{"Provider":"kind"},
					"Components":{"selector":[{"name": "testhelm"}], "manifests":["testdata/fake.yml"]}},
					"Runtimedir": "testdata"
				}`),
			},
			expectedOutput: map[string][]byte{
				"serviceconfig": []byte(`{
				"Components":[{"images":null,"name":"testhelm","chartoverride":"file://testfakeoverride","resources":null,"supported-clusters":["default"],"type":"helm","url":"file://testhelmurl"}]}`),
				"downloadfiles": []byte(`{"files":[{"url":"file://testhelmurl","urlreplacement":{"new":"helm/testhelm","origin":"file://"}},
					{"url":"file://testfakeoverride","urlreplacement":{"new":"helm/testhelm","origin":"file://"}}]}`),
				"docker-images": []byte(`{"images":null}`),
			},
			expectError:    false,
			expectErrorMsg: "",
			funcBeforeTest: success_find_helm_repofunc,
		},
		{
			name: "success_helmrepo",
			input: map[string][]byte{
				"ep-params": []byte(`{"kitconfig":{
					"Cluster":{"Provider":"kind"},
					"Components":{"selector":[{"name": "testhelmrepo"}], "manifests":["testdata/fake.yml"]}},
					"Runtimedir": "testdata"
				}`),
			},
			expectedOutput: map[string][]byte{
				"serviceconfig": []byte(`{
				"Components":[{"chartname":"prometheus","chartversion":"14.1.3","helmrepo":"https://prometheus-community.github.io/helm-charts","images":null,"name":"testhelmrepo","chartoverride":"file://testoverrideurl","resources":null,"supported-clusters":["default"],"type":"helm","url":"https://github.com/prometheus-community/helm-charts/releases/download/prometheus-14.1.3/prometheus-14.1.3.tgz"}]}`),
				"downloadfiles": []byte(`{"files":[
					{"url":"https://github.com/prometheus-community/helm-charts/releases/download/prometheus-14.1.3/prometheus-14.1.3.tgz","urlreplacement":{"new":"helm/testhelmrepo","origin":"https://github.com/prometheus-community/helm-charts/releases/download/prometheus-14.1.3/"}},
					{"url":"file://testoverrideurl","urlreplacement":{"new":"helm/testhelmrepo","origin":"file://"}}]}`),
				"docker-images": []byte(`{"images":null}`),
			},
			expectError:    false,
			expectErrorMsg: "",
			funcBeforeTest: success_find_helm_repofunc,
		},
		{
			name: "success_repo",
			input: map[string][]byte{
				"ep-params": []byte(`{"kitconfig":{
					"Cluster":{"Provider":"kind"},
					"Components":{"selector":[{"name": "testrepo"}], "manifests":["testdata/fake.yml"]}},
					"Runtimedir": "testdata"
				}`),
			},
			expectedOutput: map[string][]byte{
				"serviceconfig": []byte(`{
				"Components":[{"images":null,"name":"testrepo","resources":null,"supported-clusters":["default"],"type":"repo","url":"file://testrepourl"}]}`),
				"downloadfiles": []byte(`{"files":null}`),
				"docker-images": []byte(`{"images":null}`),
			},
			expectError:    false,
			expectErrorMsg: "",
			funcBeforeTest: success_find_helm_repofunc,
		},
		{
			name: "service_not_supported",
			input: map[string][]byte{
				"ep-params": []byte(`{"kitconfig":{
					"Cluster":{"Provider":"kind"},
					"Components":{"selector":[{"name": "testrepo_rke"}], "manifests":["testdata/fake.yml"]}},
					"Runtimedir": "testdata"
				}`),
			},
			expectedOutput: map[string][]byte{
				"serviceconfig": []byte(`{
				"Components":null}`),
				"downloadfiles": []byte(`{"files":null}`),
				"docker-images": []byte(`{"images":null}`),
			},
			expectError:    false,
			expectErrorMsg: "",
			funcBeforeTest: success_find_helm_repofunc,
		},
		{
			name: "override_yaml",
			input: map[string][]byte{
				"ep-params": []byte(`{"kitconfig":{
					"Cluster":{"Provider":"kind"},
					"Components":{"selector":[{"name": "testrepo","__override_yaml__":"{'url':'file://testrepourl_override'}"}], "manifests":["testdata/fake.yml"]}},
					"Runtimedir": "testdata"
				}`),
			},
			expectedOutput: map[string][]byte{
				"serviceconfig": []byte(`{
				"Components":[{"images":null,"name":"testrepo","resources":null,"supported-clusters":["default"],"type":"repo","url":"file://testrepourl_override"}]}`),
				"downloadfiles": []byte(`{"files":null}`),
				"docker-images": []byte(`{"images":null}`),
			},
			expectError:    false,
			expectErrorMsg: "",
			funcBeforeTest: success_find_helm_repofunc,
		},
		{
			name: "success_withimage",
			input: map[string][]byte{
				"ep-params": []byte(`{"kitconfig":{
					"Cluster":{"Provider":"kind"},
					"Components":{"selector":[{"name": "testimage"}], "manifests":["testdata/fake.yml"]}},
					"Runtimedir": "testdata"
				}`),
			},
			expectedOutput: map[string][]byte{
				"serviceconfig": []byte(`{
				"Components":[{"images":["k8s.gcr.io/ingress-nginx/controller:v1.1.0","k8s.gcr.io/ingress-nginx/kube-webhook-certgen:v1.1.1"],"name":"testimage","resources":null,"supported-clusters":["default"],"type":"yaml","url":"file://testyamlurl"}]}`),
				"downloadfiles": []byte(`{"files":[{"url":"file://testyamlurl","urlreplacement":{"new":"yaml/testimage","origin":"file://"}}]}`),
				"docker-images": []byte(`{"images":[
					{"name":"testimage","url":"k8s.gcr.io/ingress-nginx/controller:v1.1.0"},
					{"name":"testimage","url":"k8s.gcr.io/ingress-nginx/kube-webhook-certgen:v1.1.1"}]
				}`),
			},
			expectError:    false,
			expectErrorMsg: "",
			funcBeforeTest: success_find_helm_repofunc,
		},
		{
			name: "err_empty_HelmChart_Name",
			input: map[string][]byte{
				"ep-params": []byte(`{"kitconfig":{"Cluster":{"Provider":"kind"},"Components":{"selector":[{"name": "testhelmrepo"}],"manifests":["test-chart-name.yml"]}}}`),
			},
			expectedOutput: nil,
			expectError:    true,
			expectErrorMsg: eputils.GetError("errHelmEmpty").Error(),
			funcBeforeTest: nil,
		},
		{
			name: "err_empty_HelmRepo",
			input: map[string][]byte{
				"ep-params": []byte(`{"kitconfig":{"Cluster":{"Provider":"kind"},"Components":{"selector":[{"name": "testhelmrepo"}],"manifests":["test-helm-repo.yml"]}}}`),
			},
			expectedOutput: nil,
			expectError:    true,
			expectErrorMsg: eputils.GetError("errHelmEmpty").Error(),
			funcBeforeTest: nil,
		},
	}

	errChart := eputils.WriteStringToFile(emptyChartNameInput, "test-chart-name.yml")
	require.NoError(t, errChart, "Write String To File Error:")

	defer os.RemoveAll("test-chart-name.yml")

	errHelm := eputils.WriteStringToFile(emptyRepoInput, "test-helm-repo.yml")
	require.NoError(t, errHelm, "Write String To File Error:")

	defer os.RemoveAll("test-helm-repo.yml")

	errMkdir := os.MkdirAll("testdata/data", 0755)
	require.NoError(t, errMkdir, "Create data folder")
	defer os.RemoveAll("testdata/data")

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			if tc.funcBeforeTest != nil {
				plist := tc.funcBeforeTest(ctrl)
				for _, p := range plist {
					defer unpatch(t, p)
				}
			}

			input := generateInput(tc.input)
			if input == nil {
				t.Fatalf("Failed to generateInput %s", tc.input)
			}
			testOutput := generateOutput(nil)

			err := PluginMain(input, &testOutput)

			if tc.expectError {
				if err == nil {
					t.Error("Expected error but no error found.")
				} else {
					if fmt.Sprint(err) == tc.expectErrorMsg {
						t.Log("Output expected.")
					} else {
						t.Error("Expect:", tc.expectErrorMsg, "; But found:", err)
					}
				}
			} else {
				if err != nil {
					t.Error("Unexpected Error:", err)
				} else {
					expectedOutput := generateOutput(tc.expectedOutput)
					if testOutput.EqualWith(expectedOutput) {
						t.Log("Done")
					} else {
						output, _ := testOutput.MarshalBinary()
						t.Errorf("Expect output %s but returned %s.", tc.expectedOutput, output)
					}
				}
			}
		})
	}
}

func FindFakeChartInRepoURL(string, string, string, string, string, string, getter.Providers) (string, error) {
	return "https://github.com/prometheus-community/helm-charts/releases/download/prometheus-14.1.3/prometheus-14.1.3.tgz", nil
}
