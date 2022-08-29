/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package kubeutils

import (
	"context"
	pluginapi "ep/pkg/api/plugins"
	"errors"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"

	"ep/pkg/eputils"
	"github.com/undefinedlabs/go-mpatch"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	corev1ac "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	restclient "k8s.io/client-go/rest"
	"sigs.k8s.io/yaml"
)

var (
	kubeconfig_bad_filepath  = "testdata/testkubeconfigbad"
	kubeconfig_good_filepath = "testdata/testkubeconfigcontent"
)

var (
	errEmpty           = errors.New("")
	errKubeNewForConf  = errors.New("kubernetes.NewForConfig.error")
	errYML2JSON        = errors.New("yamltojsonfail")
	errNewClient       = errors.New("newclient.error")
	errInvalidConf     = errors.New("invalid configuration: no configuration has been provided, try setting KUBERNETES_MASTER environment variable")
	errClntFromKubeCnf = errors.New("clientfromkubeconfig.error")
	errReadfile        = errors.New("ioutils.readfile.error")
	errClntFrompKube   = errors.New("clientfromepkubeconfig.error")
	errList            = errors.New("list.error")
)

func unpatch(t *testing.T, m *mpatch.Patch) {
	err := m.Unpatch()
	if err != nil {
		t.Fatal(err)
	}
}

func readkubeconfig_tobyte(t *testing.T, filepath string) []byte {

	if b, err := ioutil.ReadFile(filepath); err != nil {
		t.Error("[Fatal]: Cannot read file from filepath`" + filepath + "`.")
		return nil
	} else {
		return b
	}

}

func readkubeconfig_tofilecontent(t *testing.T, filepath string) *pluginapi.Filecontent {
	if b, err := ioutil.ReadFile(filepath); err != nil {
		t.Error("[Fatal]: Cannot read testdata from `" + kubeconfig_good_filepath + "`.")
		return nil
	} else {
		return &pluginapi.Filecontent{
			Content: string(b),
		}
	}

}

func isExpectedError(returnErr error, wantError error) bool {
	if !errors.Is(returnErr, wantError) &&
		(returnErr == nil || wantError == nil || !strings.Contains(returnErr.Error(), wantError.Error())) {
		return false
	}
	return true
}

func TestNewClient(t *testing.T) {
	cases := []struct {
		name           string
		expectError    error
		input          []byte
		funcBeforeTest func()
	}{
		{
			name:        "NewRestClient fail",
			input:       readkubeconfig_tobyte(t, kubeconfig_good_filepath),
			expectError: errEmpty,
			funcBeforeTest: func() {
				patchnewrestclientfail(t)
			},
		},
		{
			name:           "NewClient ok",
			expectError:    nil,
			input:          readkubeconfig_tobyte(t, kubeconfig_good_filepath),
			funcBeforeTest: nil,
		},
		{
			name:        "NewClient fail",
			expectError: errKubeNewForConf,
			input:       readkubeconfig_tobyte(t, kubeconfig_good_filepath),
			funcBeforeTest: func() {
				patchnewforconfigfail(t)
			},
		},
	}

	for _, tc := range cases {
		if tc.funcBeforeTest != nil {
			tc.funcBeforeTest()
		}

		_, err := NewClient(tc.input)

		if !isExpectedError(err, tc.expectError) {
			t.Errorf("Unexpected error: %v", err)
		}

	}
}

func TestNewRestClient(t *testing.T) {
	cases := []struct {
		name           string
		expectError    error
		input          []byte
		funcBeforeTest func()
	}{
		{
			name:           "NewRestClient ok",
			expectError:    nil,
			input:          readkubeconfig_tobyte(t, kubeconfig_good_filepath),
			funcBeforeTest: nil,
		},
		{
			name:        "kubeconfigcontent bad",
			expectError: errYML2JSON,
			input:       readkubeconfig_tobyte(t, kubeconfig_bad_filepath),
			funcBeforeTest: func() {
				patchyamltojson(t)
			},
		},
	}

	for _, tc := range cases {

		if tc.funcBeforeTest != nil {
			tc.funcBeforeTest()
		}

		_, err := NewRestClient(tc.input)

		if !isExpectedError(err, tc.expectError) {
			t.Errorf("Unexpected error: %v", err)
		}

	}
}

func TestCreateNewClientSetFromFileContent(t *testing.T) {
	cases := []struct {
		name           string
		expectError    error
		input          *pluginapi.Filecontent
		funcBeforeTest func()
	}{
		{
			name:        "ClientFromEPKubeConfig ok",
			expectError: nil,
			input:       readkubeconfig_tofilecontent(t, kubeconfig_good_filepath),
			funcBeforeTest: func() {
				patchnewclientok(t)
			},
		},
		{
			name:        "ClientFromEPKubeConfig fail",
			expectError: errNewClient,
			input:       readkubeconfig_tofilecontent(t, kubeconfig_bad_filepath),
			funcBeforeTest: func() {
				patchnewclientfail(t)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Run funcBeforeTest
			if tc.funcBeforeTest != nil {
				tc.funcBeforeTest()
			}

			_, err := ClientFromEPKubeConfig(tc.input)

			if !isExpectedError(err, tc.expectError) {
				t.Errorf("Unexpected error: %v", err)
			}

		})
	}

}

func TestCreateNewRESTClientSetFromFileContentSuccess(t *testing.T) {
	cases := []struct {
		name           string
		expectError    error
		input          *pluginapi.Filecontent
		funcBeforeTest func()
	}{
		{
			name:           "NewRESTClientSetFromFileContent ok",
			expectError:    nil,
			input:          readkubeconfig_tofilecontent(t, kubeconfig_good_filepath),
			funcBeforeTest: nil,
		},
		{
			name:           "RestClientFromEPKubeConfig fail",
			expectError:    errInvalidConf,
			input:          readkubeconfig_tofilecontent(t, kubeconfig_bad_filepath),
			funcBeforeTest: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			_, err := RestClientFromEPKubeConfig(tc.input)

			if !isExpectedError(err, tc.expectError) {
				t.Errorf("Unexpected error: %v", err)
			}

		})
	}
}

func TestCreateNewClient(t *testing.T) {
	cases := []struct {
		name           string
		expectError    error
		input          []byte
		funcBeforeTest func()
	}{
		{
			name:           "CreateNewClient fail",
			expectError:    nil,
			input:          readkubeconfig_tobyte(t, kubeconfig_good_filepath),
			funcBeforeTest: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			_, err := NewClient(tc.input)

			if !isExpectedError(err, tc.expectError) {
				t.Errorf("Unexpected error: %v", err)
			}

		})
	}
}

func TestCreateNamespace(t *testing.T) {
	var patchholder *mpatch.Patch
	cases := []struct {
		name           string
		expectError    error
		input          string
		funcBeforeTest func()
		tearDownFunc   func()
	}{
		{
			name:        "ClientFromKubeConfig fail",
			expectError: errClntFromKubeCnf,
			input:       kubeconfig_good_filepath,
			funcBeforeTest: func() {
				patchclientfromkubeconfigfail(t)
			},
		},
		{
			name:        "CreateNamespace fail (Namespaces.Get fail, Namespaces.Create fail)",
			expectError: errEmpty,
			input:       kubeconfig_bad_filepath,
			funcBeforeTest: func() {
				patchclientfromkubeconfigok(t)
				patchholder = patchgetcorev1ok(t, false, false, false, false)
			},
			tearDownFunc: func() {
				unpatch(t, patchholder)
			},
		},
		{
			name:        "CreateNamespace ok (Namespaces.Get fail, Namespaces.Create ok)",
			expectError: nil,
			input:       kubeconfig_bad_filepath,
			funcBeforeTest: func() {
				patchclientfromkubeconfigok(t)
				patchholder = patchgetcorev1ok(t, false, true, false, false)
			},
			tearDownFunc: func() {
				unpatch(t, patchholder)
			},
		},
		{
			name:        "CreateNamespace ok",
			expectError: nil,
			input:       kubeconfig_good_filepath,
			funcBeforeTest: func() {
				patchclientfromkubeconfigok(t)
				patchgetcorev1ok(t, true, true, false, false)
			},
			tearDownFunc: func() {
				unpatch(t, patchholder)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			if tc.funcBeforeTest != nil {
				tc.funcBeforeTest()
			}

			err := CreateNamespace(tc.input, "default")

			if !isExpectedError(err, tc.expectError) {
				t.Errorf("Unexpected error: %v", err)
			}

			// Teardown
			if tc.tearDownFunc != nil {
				tc.tearDownFunc()
			}
		})
	}
}

func TestCreateNewRestClientSet(t *testing.T) {
	cases := []struct {
		name           string
		expectError    error
		input          string
		funcBeforeTest func()
	}{
		{
			name:           "CreateNewRestClientSetFromKubeConfig ok",
			expectError:    nil,
			input:          kubeconfig_good_filepath,
			funcBeforeTest: nil,
		},
		{
			name:        "CreateNewRestConfigFromKubeConfig fail",
			expectError: errReadfile,
			input:       kubeconfig_bad_filepath,
			funcBeforeTest: func() {
				patchreadfilefail(t)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			// Run funcBeforeTest
			if tc.funcBeforeTest != nil {
				tc.funcBeforeTest()
			}

			_, err := RestConfigFromKubeConfig(tc.input)

			if !isExpectedError(err, tc.expectError) {
				t.Errorf("Unexpected error: %v", err)
			}

		})
	}
}

func TestCreateNewClientSet(t *testing.T) {
	cases := []struct {
		name           string
		expectError    error
		input          string
		expectedOutput string
		funcBeforeTest func()
	}{
		{
			name:           "CreateNewClientSetFromKubeConfigFile ok",
			expectError:    nil,
			input:          kubeconfig_good_filepath,
			expectedOutput: "",
		},
		{
			name:        "ClientFromKubeConfig fail",
			expectError: errReadfile,
			input:       kubeconfig_bad_filepath,
			funcBeforeTest: func() {
				patchreadfilefail(t)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				tc.funcBeforeTest()
			}

			_, err := ClientFromKubeConfig(tc.input)

			if !isExpectedError(err, tc.expectError) {
				t.Errorf("Unexpected error: %v", err)
			}

		})
	}
}

func TestGetNodeList(t *testing.T) {
	var patchHandler *mpatch.Patch
	cases := []struct {
		name                   string
		expectError            error
		input_kubeconfig       *pluginapi.Filecontent
		input_labelselectorstr string

		funcBeforeTest func()
		tearDownFunc   func()
	}{
		{
			name:                   "ClientFromEPKubeConfig fail",
			expectError:            errClntFrompKube,
			input_kubeconfig:       readkubeconfig_tofilecontent(t, kubeconfig_bad_filepath),
			input_labelselectorstr: "default",
			funcBeforeTest: func() {
				patchclientfromepkubeconfigfail(t)
			},
		},
		{
			name:                   "no items",
			expectError:            eputils.GetError("errNok8sNode"),
			input_kubeconfig:       readkubeconfig_tofilecontent(t, kubeconfig_good_filepath),
			input_labelselectorstr: "default",
			funcBeforeTest: func() {
				patchclientfromepkubeconfigok(t)
				patchHandler = patchgetcorev1ok(t, false, false, true, false)
			},
			tearDownFunc: func() {
				unpatch(t, patchHandler)
			},
		},
		{
			name:                   "list node clients fail",
			expectError:            errList,
			input_kubeconfig:       readkubeconfig_tofilecontent(t, kubeconfig_good_filepath),
			input_labelselectorstr: "default",
			funcBeforeTest: func() {
				patchclientfromepkubeconfigok(t)
				patchHandler = patchgetcorev1ok(t, false, false, false, false)
			},
			tearDownFunc: func() {
				unpatch(t, patchHandler)
			},
		},
		{
			name:                   "nodelist.Item > 0",
			expectError:            nil,
			input_kubeconfig:       readkubeconfig_tofilecontent(t, kubeconfig_good_filepath),
			input_labelselectorstr: "default",
			funcBeforeTest: func() {
				patchclientfromepkubeconfigok(t)
				patchHandler = patchgetcorev1ok(t, false, false, true, true)
			},
			tearDownFunc: func() {
				unpatch(t, patchHandler)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Run pretest function
			if tc.funcBeforeTest != nil {
				tc.funcBeforeTest()
			}

			// Test
			_, err := GetNodeList(tc.input_kubeconfig, tc.input_labelselectorstr)

			if !isExpectedError(err, tc.expectError) {
				t.Errorf("Unexpected error: %v", err)
			}

			// Teardown
			if tc.tearDownFunc != nil {
				tc.tearDownFunc()
			}

		})
	}
}

//region patches

func patchclientfromepkubeconfigok(t *testing.T) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(ClientFromEPKubeConfig, func(input_kubeconfig *pluginapi.Filecontent) (*kubernetes.Clientset, error) {
		unpatch(t, patch)
		return &kubernetes.Clientset{}, nil
	})

	if patchErr != nil {
		t.Errorf("patch error %v", patchErr)
	}
}

func patchclientfromepkubeconfigfail(t *testing.T) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(ClientFromEPKubeConfig, func(input_kubeconfig *pluginapi.Filecontent) (*kubernetes.Clientset, error) {
		unpatch(t, patch)
		return nil, errClntFrompKube
	})

	if patchErr != nil {
		t.Errorf("patch error %v", patchErr)
	}

}

func patchreadfilefail(t *testing.T) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(ioutil.ReadFile, func(filename string) ([]byte, error) {
		unpatch(t, patch)
		return nil, errReadfile
	})
	if patchErr != nil {
		t.Errorf("patch error %v", patchErr)
	}
}

func patchyamltojson(t *testing.T) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(yaml.YAMLToJSON, func(y []byte) ([]byte, error) {
		unpatch(t, patch)
		return nil, errYML2JSON
	})

	if patchErr != nil {
		t.Errorf("patch error %v", patchErr)
	}
}

func patchnewrestclientfail(t *testing.T) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(NewRestClient, func(kubeconfigcontent []byte) (*restclient.Config, error) {
		unpatch(t, patch)
		return nil, errEmpty
	})

	if patchErr != nil {
		t.Errorf("patch error %v", patchErr)
	}
}

func patchnewforconfigfail(t *testing.T) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(kubernetes.NewForConfig, func(c *rest.Config) (*kubernetes.Clientset, error) {
		unpatch(t, patch)
		return nil, errKubeNewForConf
	})
	if patchErr != nil {
		t.Errorf("patch error %v", patchErr)
	}
}

func patchnewclientok(t *testing.T) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(NewClient, func(kubeconfigcontent []byte) (*kubernetes.Clientset, error) {
		unpatch(t, patch)
		return nil, nil
	})
	if patchErr != nil {
		t.Errorf("patch error %v", patchErr)
	}
}

func patchnewclientfail(t *testing.T) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(NewClient, func(kubeconfigcontent []byte) (*kubernetes.Clientset, error) {
		unpatch(t, patch)
		return nil, errNewClient
	})
	if patchErr != nil {
		t.Errorf("patch error %v", patchErr)
	}
}

func patchclientfromkubeconfigfail(t *testing.T) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(ClientFromKubeConfig, func(kubeconfig string) (*kubernetes.Clientset, error) {
		unpatch(t, patch)
		return nil, errClntFromKubeCnf
	})
	if patchErr != nil {
		t.Errorf("patch error %v", patchErr)
	}
}

func patchclientfromkubeconfigok(t *testing.T) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(ClientFromKubeConfig, func(kubeconfig string) (*kubernetes.Clientset, error) {
		unpatch(t, patch)
		return &kubernetes.Clientset{}, nil
	})
	if patchErr != nil {
		t.Errorf("patch error %v", patchErr)
	}
}

func patchgetcorev1ok(t *testing.T, GetOK bool, CreateOK bool, ListOK bool, ItemsOK bool) *mpatch.Patch {
	var patch *mpatch.Patch
	var patchErr error

	patch, patchErr = mpatch.PatchInstanceMethodByName(reflect.TypeOf(&kubernetes.Clientset{}), "CoreV1", func(c *kubernetes.Clientset) corev1.CoreV1Interface {
		return &CoreV1ClientMock{GetOK: GetOK, CreateOK: CreateOK, ListOK: ListOK, ItemsOK: ItemsOK}
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

	return patch
}

//endregion

//region struct mocks

type NamespacesMock struct {
	CreateOK bool
	GetOK    bool
}

func (n NamespacesMock) Create(ctx context.Context, namespace *v1.Namespace, opts metav1.CreateOptions) (*v1.Namespace, error) {
	if n.CreateOK {
		return nil, nil
	} else {
		return nil, errEmpty
	}
}

func (n NamespacesMock) Update(ctx context.Context, namespace *v1.Namespace, opts metav1.UpdateOptions) (*v1.Namespace, error) {
	return nil, nil
}
func (n NamespacesMock) UpdateStatus(ctx context.Context, namespace *v1.Namespace, opts metav1.UpdateOptions) (*v1.Namespace, error) {
	return nil, nil
}
func (n NamespacesMock) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return nil
}
func (n NamespacesMock) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.Namespace, error) {
	if n.GetOK {
		return nil, nil
	} else {
		return nil, errEmpty
	}
}
func (n NamespacesMock) List(ctx context.Context, opts metav1.ListOptions) (*v1.NamespaceList, error) {
	return nil, nil
}
func (n NamespacesMock) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return nil, nil
}
func (n NamespacesMock) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.Namespace, err error) {
	return nil, nil
}
func (n NamespacesMock) Apply(ctx context.Context, namespace *corev1ac.NamespaceApplyConfiguration, opts metav1.ApplyOptions) (result *v1.Namespace, err error) {
	return nil, nil
}
func (n NamespacesMock) ApplyStatus(ctx context.Context, namespace *corev1ac.NamespaceApplyConfiguration, opts metav1.ApplyOptions) (result *v1.Namespace, err error) {
	return nil, nil
}
func (n NamespacesMock) Finalize(ctx context.Context, item *v1.Namespace, opts metav1.UpdateOptions) (*v1.Namespace, error) {
	return nil, nil
}

type NodeMock struct {
	ListOK  bool
	ItemsOK bool
}

func (n NodeMock) Create(ctx context.Context, node *v1.Node, opts metav1.CreateOptions) (*v1.Node, error) {
	return nil, nil
}
func (n NodeMock) Update(ctx context.Context, node *v1.Node, opts metav1.UpdateOptions) (*v1.Node, error) {
	return nil, nil
}
func (n NodeMock) UpdateStatus(ctx context.Context, node *v1.Node, opts metav1.UpdateOptions) (*v1.Node, error) {
	return nil, nil
}
func (n NodeMock) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return nil
}
func (n NodeMock) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return nil
}
func (n NodeMock) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.Node, error) {
	return nil, nil
}

func (n NodeMock) List(ctx context.Context, opts metav1.ListOptions) (*v1.NodeList, error) {
	if n.ListOK {
		if n.ItemsOK {
			return &v1.NodeList{Items: []v1.Node{
				{},
				{},
			}}, nil
		} else {

			return &v1.NodeList{Items: []v1.Node{}}, nil
		}

	} else {
		return nil, errList
	}
}
func (n NodeMock) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return nil, nil
}

func (n NodeMock) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.Node, err error) {
	return nil, nil
}
func (n NodeMock) Apply(ctx context.Context, node *corev1ac.NodeApplyConfiguration, opts metav1.ApplyOptions) (result *v1.Node, err error) {
	return nil, nil
}
func (n NodeMock) ApplyStatus(ctx context.Context, node *corev1ac.NodeApplyConfiguration, opts metav1.ApplyOptions) (result *v1.Node, err error) {
	return nil, nil
}
func (n NodeMock) PatchStatus(ctx context.Context, nodeName string, data []byte) (*v1.Node, error) {
	return nil, nil
}

type CoreV1ClientMock struct {
	CreateOK bool
	GetOK    bool
	ListOK   bool
	ItemsOK  bool
}

func (c CoreV1ClientMock) RESTClient() rest.Interface {
	return nil
}

func (c CoreV1ClientMock) ComponentStatuses() corev1.ComponentStatusInterface {
	return nil
}

func (c CoreV1ClientMock) ConfigMaps(namespace string) corev1.ConfigMapInterface {
	return nil
}

func (c CoreV1ClientMock) Endpoints(namespace string) corev1.EndpointsInterface {
	return nil
}

func (c CoreV1ClientMock) Events(namespace string) corev1.EventInterface {
	return nil
}

func (c CoreV1ClientMock) LimitRanges(namespace string) corev1.LimitRangeInterface {
	return nil
}

func (c CoreV1ClientMock) Namespaces() corev1.NamespaceInterface {
	return &NamespacesMock{GetOK: c.GetOK, CreateOK: c.CreateOK}
}

func (c CoreV1ClientMock) Nodes() corev1.NodeInterface {
	return &NodeMock{ListOK: c.ListOK, ItemsOK: c.ItemsOK}
}

func (c CoreV1ClientMock) PersistentVolumes() corev1.PersistentVolumeInterface {
	return nil
}

func (c CoreV1ClientMock) PersistentVolumeClaims(namespace string) corev1.PersistentVolumeClaimInterface {
	return nil
}

func (c CoreV1ClientMock) Pods(namespace string) corev1.PodInterface {
	return nil
}

func (c CoreV1ClientMock) PodTemplates(namespace string) corev1.PodTemplateInterface {
	return nil
}

func (c CoreV1ClientMock) ReplicationControllers(namespace string) corev1.ReplicationControllerInterface {
	return nil
}

func (c CoreV1ClientMock) ResourceQuotas(namespace string) corev1.ResourceQuotaInterface {
	return nil
}

func (c CoreV1ClientMock) Secrets(namespace string) corev1.SecretInterface {
	return nil
}

func (c CoreV1ClientMock) Services(namespace string) corev1.ServiceInterface {
	return nil
}

func (c CoreV1ClientMock) ServiceAccounts(namespace string) corev1.ServiceAccountInterface {
	return nil
}

//endregion
