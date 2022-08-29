/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

//nolint: dupl
package kubeutils

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/undefinedlabs/go-mpatch"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	corev1 "k8s.io/client-go/applyconfigurations/core/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
)

var (
	errGet          = errors.New("get.error")
	errRestConfig   = errors.New("restconfigfromkubeconfig.error")
	errNewForConfig = errors.New("newforconfig.error")
	errCreate       = errors.New("create.error")
	errUpdate       = errors.New("update.error")
)

func TestGetData(t *testing.T) {
	var patchHandler *mpatch.Patch
	cases := []struct {
		name        string
		expectError error
		beforeTest  func()
		teardown    func()
	}{
		{
			name:        "Get Success",
			expectError: nil,
			beforeTest: func() {
				patchHandler = patchconfigmapok(t, true, false, false)
			},
			teardown: func() {
				unpatch(t, patchHandler)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, tc := range cases {
				if tc.beforeTest != nil {
					tc.beforeTest()
				}

				c, err := NewConfigMap("mock namespace", "mock name", "mock fieldmanager", kubeconfig_good_filepath)
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				err = c.New()
				c.GetData()
				c.GetBinaryData()

				if !isExpectedError(err, tc.expectError) {
					t.Errorf("Unexpected error: %v", err)
				}

				if tc.teardown != nil {
					tc.teardown()
				}

			}
		})
	}

}

func TestGet(t *testing.T) {
	var patchHandler *mpatch.Patch
	cases := []struct {
		name        string
		input       string
		expectError error
		beforeTest  func()
		teardown    func()
	}{
		{
			name:        "Get Failed",
			input:       "",
			expectError: errGet,
			beforeTest: func() {
				patchHandler = patchconfigmapok(t, false, false, false)
			},
			teardown: func() {
				unpatch(t, patchHandler)
			},
		},
		{
			name:        "Get Success",
			input:       "",
			expectError: nil,
			beforeTest: func() {
				patchHandler = patchconfigmapok(t, true, false, false)
			},
			teardown: func() {
				unpatch(t, patchHandler)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, tc := range cases {
				if tc.beforeTest != nil {
					tc.beforeTest()
				}

				c, _ := NewConfigMap("mock namespace", "mock name", "mock fieldmanager", kubeconfig_good_filepath)
				err := c.Get()

				if !isExpectedError(err, tc.expectError) {
					t.Errorf("Unexpected error: %v", err)
				}

				if tc.teardown != nil {
					tc.teardown()
				}

			}
		})
	}
}

func TestNewConfigMap(t *testing.T) {
	cases := []struct {
		name        string
		input       []string
		expectError error
		beforeTest  func()
		teardown    func()
	}{
		{
			name: "RestConfigFromKubeConfig fail",
			input: []string{
				"default namespace",
				"default name",
				"default fieldmanager",
				"default kubeconfig",
			},
			expectError: errRestConfig,
			beforeTest: func() {
				patchrestconfigfromkubeconfigfail(t)
			},
		},
		{
			name: "corev1client.NewForConfig fail",
			input: []string{
				"default namespace", "default name", "default fieldmanager", "default kubeconfig",
			},
			expectError: errNewForConfig,
			beforeTest: func() {
				patchrestconfigfromkubeconfigok(t)
				patchcorev1clientnewforconfigfail(t)
			},
		},
		{
			name: "NewConfigMap ok",
			input: []string{
				"default namespace", "default name", "default fieldmanager", "default kubeconfig",
			},
			expectError: nil,
			beforeTest: func() {
				patchrestconfigfromkubeconfigok(t)
				patchcorev1clientnewforconfigok(t)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, tc := range cases {
				if tc.beforeTest != nil {
					tc.beforeTest()
				}

				_, err := NewConfigMap(tc.input[0], tc.input[1], tc.input[2], tc.input[3])

				if !isExpectedError(err, tc.expectError) {
					t.Errorf("Unexpected error: %v", err)
				}

			}
		})
	}
}

func TestRenewData(t *testing.T) {
	cases := []struct {
		name        string
		input       []string
		expectError error
		data_exists bool
		beforeTest  func()
	}{
		{
			name: "key exists",
			input: []string{
				"mock key", "mock data",
			},
			expectError: nil,
			beforeTest: func() {
				patchupdateok(t)
			},
		},
		{
			name: "key does not exist",
			input: []string{
				"ek1", "ed1",
			},
			data_exists: true,
			expectError: nil,
			beforeTest: func() {
				patchupdateok(t)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, tc := range cases {
				if tc.beforeTest != nil {
					tc.beforeTest()
				}

				c, _ := NewConfigMap("mock namespace", "mock name", "mock fieldmanager", kubeconfig_good_filepath)
				_ = c.New()
				if tc.data_exists == true {
					c.(*ConfigMap).ConfigMapObj.Data["ek1"] = "ed1"
				}

				err := c.RenewData(tc.input[0], tc.input[1])

				if !isExpectedError(err, tc.expectError) {
					t.Errorf("Unexpected error: %v", err)
				}

			}
		})
	}
}

func TestRenewBinaryData(t *testing.T) {
	cases := []struct {
		name        string
		input_key   string
		input_data  []byte
		data_exists bool
		expectError error
		beforeTest  func()
	}{
		{
			name:        "key does not exist",
			input_key:   "mock key",
			input_data:  []byte("mock data"),
			data_exists: false,
			expectError: nil,
			beforeTest: func() {
				patchupdateok(t)
			},
		},
		{
			name:        "key exists",
			input_key:   "ek1",
			input_data:  []byte("ed1"),
			data_exists: true,
			expectError: nil,
			beforeTest: func() {
				patchupdateok(t)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, tc := range cases {
				if tc.beforeTest != nil {
					tc.beforeTest()
				}

				c, _ := NewConfigMap("mock namespace", "mock name", "mock fieldmanager", kubeconfig_good_filepath)
				_ = c.New()
				if tc.data_exists == true {
					c.(*ConfigMap).ConfigMapObj.BinaryData["ek1"] = []byte("ed1")
				}

				err := c.RenewBinaryData(tc.input_key, tc.input_data)

				if !isExpectedError(err, tc.expectError) {
					t.Errorf("Unexpected error: %v", err)
				}

			}
		})
	}
}

func TestConfigMap_RemoveData(t *testing.T) {
	cases := []struct {
		name        string
		input_key   string
		data_exists bool
		expectError error
		beforeTest  func()
	}{
		{
			name:        "key exists",
			input_key:   "ek1",
			expectError: nil,
			data_exists: true,
			beforeTest: func() {
				patchupdateok(t)
			},
		},
		{
			name:        "key does not exist",
			input_key:   "mock key",
			data_exists: true,
			expectError: nil,
			beforeTest: func() {
				patchupdateok(t)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, tc := range cases {
				if tc.beforeTest != nil {
					tc.beforeTest()
				}

				c, _ := NewConfigMap("mock namespace", "mock name", "mock fieldmanager", kubeconfig_good_filepath)
				_ = c.New()
				if tc.data_exists == true {
					c.(*ConfigMap).ConfigMapObj.Data["ek1"] = "ed1"
				}

				err := c.(*ConfigMap).RemoveData(tc.input_key)

				if !isExpectedError(err, tc.expectError) {
					t.Errorf("Unexpected error: %v", err)
				}

			}
		})
	}
}

func TestNew(t *testing.T) {

	cases := []struct {
		name           string
		input          string
		expectError    error
		expectErrorMsg string
		beforeTest     func()
	}{
		{
			name:           "Create new configmap",
			input:          "",
			expectError:    nil,
			expectErrorMsg: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, tc := range cases {
				if tc.beforeTest != nil {
					tc.beforeTest()
				}

				c, _ := NewConfigMap("mock namespace", "mock name", "mock fieldmanager", kubeconfig_good_filepath)
				err := c.New()

				if !isExpectedError(err, tc.expectError) {
					t.Errorf("Unexpected error: %v", err)
				}

			}
		})
	}
}

func TestConfigMap_RemoveBinaryData(t *testing.T) {
	cases := []struct {
		name           string
		input_key      string
		data_exists    bool
		expectError    error
		expectErrorMsg string
		beforeTest     func()
	}{
		{
			name:        "key exists",
			input_key:   "ek1",
			expectError: nil,
			data_exists: true,
			beforeTest: func() {
				patchupdateok(t)
			},
		},
		{
			name:        "key does not exist",
			input_key:   "mock key",
			data_exists: true,
			expectError: nil,
			beforeTest: func() {
				patchupdateok(t)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, tc := range cases {
				if tc.beforeTest != nil {
					tc.beforeTest()
				}

				c, _ := NewConfigMap("mock namespace", "mock name", "mock fieldmanager", kubeconfig_good_filepath)
				_ = c.New()

				if tc.data_exists == true {
					c.(*ConfigMap).ConfigMapObj.BinaryData["ek1"] = []byte("ed1")
				}
				err := c.(*ConfigMap).RemoveBinaryData(tc.input_key)

				if !isExpectedError(err, tc.expectError) {
					t.Errorf("Unexpected error: %v", err)
				}

			}
		})
	}
}

func TestConfigMap_Update(t *testing.T) {
	var patchHandler *mpatch.Patch
	cases := []struct {
		name        string
		expectError error
		beforeTest  func()
		teardown    func()
	}{
		{
			name:        "ConfigMaps.Get fail, Update fail",
			expectError: errCreate,
			beforeTest: func() {
				patchHandler = patchconfigmapok(t, false, false, false)
			},
			teardown: func() {
				unpatch(t, patchHandler)
			},
		},
		{
			name:        "ConfigMaps.Get ok, Update fail",
			expectError: errUpdate,
			beforeTest: func() {
				patchHandler = patchconfigmapok(t, true, false, false)
			},
			teardown: func() {
				unpatch(t, patchHandler)
			},
		},
		{
			name:        "ConfigMaps.Get ok, Update ok",
			expectError: nil,
			beforeTest: func() {
				patchHandler = patchconfigmapok(t, true, false, true)
			},
			teardown: func() {
				unpatch(t, patchHandler)
			},
		},
		{
			name:        "ConfigMaps.get false, Create ok",
			expectError: errGet,
			beforeTest: func() {
				patchHandler = patchconfigmapok(t, false, true, false)
			},
			teardown: func() {
				unpatch(t, patchHandler)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.beforeTest != nil {
				tc.beforeTest()
			}

			c, _ := NewConfigMap("mock namespace", "mock name", "mock fieldmanager", kubeconfig_good_filepath)
			_ = c.New()

			err := c.Update()

			if !isExpectedError(err, tc.expectError) {
				t.Errorf("Unexpected error: %v", err)
			}

			if tc.teardown != nil {
				tc.teardown()
			}
		})
	}
}

//region patches

func patchrestconfigfromkubeconfigok(t *testing.T) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(RestConfigFromKubeConfig, func(kubeconfig string) (*restclient.Config, error) {
		unpatch(t, patch)
		return &restclient.Config{}, nil
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

}

func patchrestconfigfromkubeconfigfail(t *testing.T) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(RestConfigFromKubeConfig, func(kubeconfig string) (*restclient.Config, error) {
		unpatch(t, patch)
		return nil, errRestConfig
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

}

func patchcorev1clientnewforconfigfail(t *testing.T) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(corev1client.NewForConfig, func(c *restclient.Config) (*corev1client.CoreV1Client, error) {
		unpatch(t, patch)
		return nil, errNewForConfig
	})

	if patchErr != nil {
		t.Errorf("patch error %v", patchErr)
	}

}

func patchcorev1clientnewforconfigok(t *testing.T) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(corev1client.NewForConfig, func(c *restclient.Config) (*corev1client.CoreV1Client, error) {
		unpatch(t, patch)
		return nil, nil
	})

	if patchErr != nil {
		t.Errorf("patch error %v", patchErr)
	}

}

func patchconfigmapok(t *testing.T, GetOK bool, CreateOK bool, UpdateOK bool) *mpatch.Patch {
	var patch *mpatch.Patch
	var patchErr error
	c, _ := NewConfigMap("mock namespace", "mock name", "mock fieldmanager", kubeconfig_good_filepath)

	patch, patchErr = mpatch.PatchInstanceMethodByName(reflect.TypeOf(c.(*ConfigMap).Client), "ConfigMaps", func(*corev1client.CoreV1Client, string) corev1client.ConfigMapInterface {
		return ConfigMapPatch{GetOK: GetOK, CreateOK: CreateOK, UpdateOK: UpdateOK}
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

	return patch
}

func patchupdateok(t *testing.T) {
	c, _ := NewConfigMap("mock namespace", "mock name", "mock fieldmanager", kubeconfig_good_filepath)
	_ = c.New()

	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchInstanceMethodByName(reflect.TypeOf(c), "Update", func(c *ConfigMap) error {
		unpatch(t, patch)
		return nil
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

//endregion

//region patch structs
type ConfigMapPatch struct {
	GetOK    bool
	CreateOK bool
	UpdateOK bool
}

func (s ConfigMapPatch) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.ConfigMap, error) {
	if s.GetOK {
		return nil, nil
	} else {
		return nil, errGet
	}
}

func (s ConfigMapPatch) Create(ctx context.Context, configMap *v1.ConfigMap, opts metav1.CreateOptions) (*v1.ConfigMap, error) {
	if s.CreateOK {
		return nil, nil
	} else {
		return nil, errCreate
	}
}

func (s ConfigMapPatch) Update(ctx context.Context, configMap *v1.ConfigMap, opts metav1.UpdateOptions) (*v1.ConfigMap, error) {
	if s.UpdateOK {

		return nil, nil
	} else {
		return nil, errUpdate
	}
}

func (s ConfigMapPatch) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return nil
}
func (s ConfigMapPatch) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return nil
}
func (s ConfigMapPatch) List(ctx context.Context, opts metav1.ListOptions) (*v1.ConfigMapList, error) {
	return nil, nil
}
func (s ConfigMapPatch) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return nil, nil
}
func (s ConfigMapPatch) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.ConfigMap, err error) {
	return nil, nil
}
func (s ConfigMapPatch) Apply(ctx context.Context, configMap *corev1.ConfigMapApplyConfiguration, opts metav1.ApplyOptions) (result *v1.ConfigMap, err error) {
	return nil, nil
}

//endregion
