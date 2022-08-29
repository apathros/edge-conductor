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

	"github.com/stretchr/testify/require"
	"github.com/undefinedlabs/go-mpatch"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	corev1 "k8s.io/client-go/applyconfigurations/core/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
)

var (
	errRestConf   = errors.New("restconfigfromkubeconfig.error")
	errNewForConf = errors.New("newforconfig.error")
)

func TestSecret_NewSecret(t *testing.T) {
	cases := []struct {
		name            string
		in_namespace    string
		in_name         string
		in_fieldmanager string
		in_kubeconfig   string
		expectError     error
		beforetest      func()
		teardown        func()
	}{
		{
			name:            "restconfigfromkubeconfig fail",
			in_namespace:    "default namespace",
			in_name:         "default name",
			in_fieldmanager: "default field name",
			in_kubeconfig:   "default kubeconfig",
			expectError:     errRestConf,
			beforetest: func() {
				patchrestconfigfromkubeconfigfail(t)
			},
		},
		{
			name:            "corev1client.newforconfig fail",
			in_namespace:    "default namespace",
			in_name:         "default name",
			in_fieldmanager: "default fieldmanager",
			in_kubeconfig:   "default kubeconfig",
			expectError:     errNewForConf,
			beforetest: func() {
				patchrestconfigfromkubeconfigok(t)
				patchcorev1clientnewforconfigfail(t)
			},
		},
		{
			name:            "NewSecret ok",
			in_namespace:    "default namespace",
			in_name:         "default name",
			in_fieldmanager: "default fieldmanager",
			in_kubeconfig:   "default kubeconfig",
			expectError:     nil,
			beforetest: func() {
				patchrestconfigfromkubeconfigok(t)
				patchcorev1clientnewforconfigok(t)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			for _, tc := range cases {
				if tc.beforetest != nil {
					tc.beforetest()
				}

				_, err := NewSecret(tc.in_namespace, tc.in_name, tc.in_fieldmanager, tc.in_kubeconfig)

				if !isExpectedError(err, tc.expectError) {
					t.Errorf("Unexpected error: %v", err)
				}

				// Teardown
				if tc.teardown != nil {
					tc.teardown()
				}

			}

		})
	}
}

func TestSecret_New(t *testing.T) {
	cases := []struct {
		name        string
		expectError error
		beforetest  func()
		teardown    func()
	}{
		{
			name:        "New secret",
			expectError: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			for _, tc := range cases {
				if tc.beforetest != nil {
					tc.beforetest()
				}

				s := &Secret{}
				err := s.New()

				if !isExpectedError(err, tc.expectError) {
					t.Errorf("Unexpected error: %v", err)
				}

				// Teardown
				if tc.teardown != nil {
					tc.teardown()
				}

			}

		})
	}
}

func patchclientok(t *testing.T, GetOK bool, CreateOK bool, UpdateOK bool) *mpatch.Patch {
	var patch *mpatch.Patch
	var patchErr error

	s, _ := NewSecret("default namespace", "default name", "default fieldmanager", kubeconfig_good_filepath)
	errNew := s.New()
	require.NoError(t, errNew, "s.New Error:")

	patch, patchErr = mpatch.PatchInstanceMethodByName(reflect.TypeOf(s.(*Secret).Client), "Secrets", func(*corev1client.CoreV1Client, string) corev1client.SecretInterface {
		return SecretMock{GetOK: GetOK, CreateOK: CreateOK, UpdateOK: UpdateOK}
	})

	if patchErr != nil {
		t.Errorf("patch error %v: ", patchErr)
	}

	return patch
}

func TestSecret_Get(t *testing.T) {
	var patchHandler *mpatch.Patch
	cases := []struct {
		name        string
		expectError error
		beforetest  func()
		teardown    func()
	}{
		{
			name:        "Secrets.Get fail",
			expectError: errGet,
			beforetest: func() {
				patchHandler = patchclientok(t, false, true, true)
			},
			teardown: func() {
				unpatch(t, patchHandler)
			},
		},
		{
			name:        "Secret.Get ok",
			expectError: nil,
			beforetest: func() {
				patchHandler = patchclientok(t, true, false, false)
			},
			teardown: func() {
				unpatch(t, patchHandler)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			for _, tc := range cases {
				if tc.beforetest != nil {
					tc.beforetest()
				}

				s, _ := NewSecret("default namespace", "default name", "default fieldmanager", kubeconfig_good_filepath)
				errNew := s.New()
				require.NoError(t, errNew, "s.New Error:")

				err := s.Get()

				if !isExpectedError(err, tc.expectError) {
					t.Errorf("Unexpected error: %v", err)
				}

				// Teardown
				if tc.teardown != nil {
					tc.teardown()
				}

			}

		})
	}
}

func TestSecret_RenewData(t *testing.T) {
	var patchHandler *mpatch.Patch
	cases := []struct {
		name         string
		in_key       string
		in_data      []byte
		expectError  error
		data_exists  bool
		data_is_null bool
		beforetest   func()
		teardown     func()
	}{
		{
			name:         "data exists",
			in_key:       "ek1",
			in_data:      []byte("ed1"),
			expectError:  nil,
			data_exists:  true,
			data_is_null: false,
			beforetest: func() {
				patchHandler = patchclientok(t, true, true, true)
			},
			teardown: func() {
				unpatch(t, patchHandler)
			},
		},
		{
			name:         "data does not exist",
			in_key:       "ek1",
			in_data:      []byte("ed1"),
			expectError:  nil,
			data_exists:  false,
			data_is_null: false,
			beforetest: func() {
				patchHandler = patchclientok(t, true, true, true)
			},
			teardown: func() {
				unpatch(t, patchHandler)
			},
		},
		{
			name:         "Data object is nil",
			in_key:       "ek1",
			in_data:      []byte("ed1"),
			expectError:  nil,
			data_exists:  false,
			data_is_null: true,
			beforetest: func() {
				patchHandler = patchclientok(t, true, true, true)

			},
			teardown: func() {
				unpatch(t, patchHandler)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			for _, tc := range cases {
				if tc.beforetest != nil {
					tc.beforetest()
				}

				s, _ := NewSecret("default namespace", "default name", "default fieldmanager", kubeconfig_good_filepath)
				errNew := s.New()
				require.NoError(t, errNew, "s.New Error:")

				if tc.data_exists {

					s.(*Secret).SecretObj.Data["ek1"] = []byte("ed1")
				}

				if tc.data_is_null {
					s.(*Secret).SecretObj.Data = nil
				}

				err := s.RenewData(tc.in_key, tc.in_data)

				if !isExpectedError(err, tc.expectError) {
					t.Errorf("Unexpected error: %v", err)
				}

				// Teardown
				if tc.teardown != nil {
					tc.teardown()
				}

			}

		})
	}
}

func TestSecret_RenewStringData(t *testing.T) {
	var patchHandler *mpatch.Patch
	cases := []struct {
		name         string
		in_key       string
		in_data      string
		expectError  error
		data_exists  bool
		data_is_null bool
		beforetest   func()
		teardown     func()
	}{
		{
			name:         "data exists",
			in_key:       "ek1",
			in_data:      "ed1",
			expectError:  nil,
			data_exists:  true,
			data_is_null: false,
			beforetest: func() {
				patchHandler = patchclientok(t, true, true, true)
			},
			teardown: func() {
				unpatch(t, patchHandler)
			},
		},
		{
			name:         "data does not exist",
			in_key:       "ek1",
			in_data:      "ed1",
			expectError:  nil,
			data_exists:  false,
			data_is_null: false,
			beforetest: func() {
				patchHandler = patchclientok(t, true, true, true)
			},
			teardown: func() {
				unpatch(t, patchHandler)
			},
		},
		{
			name:         "Data object is nil",
			in_key:       "ek1",
			in_data:      "ed1",
			expectError:  nil,
			data_exists:  false,
			data_is_null: true,
			beforetest: func() {
				patchHandler = patchclientok(t, true, true, true)

			},
			teardown: func() {
				unpatch(t, patchHandler)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			for _, tc := range cases {
				if tc.beforetest != nil {
					tc.beforetest()
				}

				s, _ := NewSecret("default namespace", "default name", "default fieldmanager", kubeconfig_good_filepath)
				errNew := s.New()
				require.NoError(t, errNew, "s.New Error:")

				if tc.data_exists {

					s.(*Secret).SecretObj.StringData["ek1"] = "ed1"
				}

				if tc.data_is_null {
					s.(*Secret).SecretObj.StringData = nil
				}

				err := s.RenewStringData(tc.in_key, tc.in_data)

				if !isExpectedError(err, tc.expectError) {
					t.Errorf("Unexpected error: %v", err)
				}

				// Teardown
				if tc.teardown != nil {
					tc.teardown()
				}

			}

		})
	}
}

func TestSecret_Update(t *testing.T) {
	var patchHandler *mpatch.Patch
	cases := []struct {
		name        string
		expectError error
		beforetest  func()
		teardown    func()
	}{
		{
			name:        "Get fail",
			expectError: errGet,
			beforetest: func() {
				patchHandler = patchclientok(t, false, true, true)
			},
			teardown: func() {
				unpatch(t, patchHandler)
			},
		},
		{
			name:        "Get fail, Create fail",
			expectError: errCreate,
			beforetest: func() {
				patchHandler = patchclientok(t, false, false, true)
			},
			teardown: func() {
				unpatch(t, patchHandler)
			},
		},
		{
			name:        "Get ok, Update fail",
			expectError: errUpdate,
			beforetest: func() {
				patchHandler = patchclientok(t, true, false, false)
			},
			teardown: func() {
				unpatch(t, patchHandler)
			},
		},
		{
			name:        "Get ok, Update ok",
			expectError: nil,
			beforetest: func() {
				patchHandler = patchclientok(t, true, true, true)
			},
			teardown: func() {
				unpatch(t, patchHandler)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			for _, tc := range cases {
				if tc.beforetest != nil {
					tc.beforetest()
				}

				// Test code
				s, _ := NewSecret("default namespace", "default name", "default fieldmanager", kubeconfig_good_filepath)
				errNew := s.New()
				require.NoError(t, errNew, "s.New Error:")

				err := s.Update()

				if !isExpectedError(err, tc.expectError) {
					t.Errorf("Unexpected error: %v", err)
				}

				// Teardown
				if tc.teardown != nil {
					tc.teardown()
				}

			}

		})
	}
}

func TestSecret_getdata(t *testing.T) {
	var patchHandler *mpatch.Patch
	cases := []struct {
		name        string
		in_key      string
		in_data     string
		expectError error
		beforetest  func()
		teardown    func()
	}{
		{
			name:        "Get data ok, get string data ok",
			in_key:      "ek1",
			in_data:     "ed1",
			expectError: nil,
			beforetest: func() {
				patchHandler = patchclientok(t, true, true, true)
			},
			teardown: func() {
				unpatch(t, patchHandler)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			for _, tc := range cases {
				if tc.beforetest != nil {
					tc.beforetest()
				}

				// Test code
				s, _ := NewSecret("default namespace", "default name", "default fieldmanager", kubeconfig_good_filepath)
				errNew := s.New()
				require.NoError(t, errNew, "s.New Error:")

				_ = s.GetStringData()
				_ = s.GetData()

				// Teardown
				if tc.teardown != nil {
					tc.teardown()
				}
			}

		})
	}
}

//region struct mock
type SecretMock struct {
	GetOK    bool
	CreateOK bool
	UpdateOK bool
}

func (s SecretMock) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.Secret, error) {
	if s.GetOK {

		return nil, nil
	} else {
		return nil, errGet
	}
}

func (s SecretMock) Create(ctx context.Context, configMap *v1.Secret, opts metav1.CreateOptions) (*v1.Secret, error) {
	if s.CreateOK {

		return nil, nil
	} else {
		return nil, errCreate
	}
}

func (s SecretMock) Update(ctx context.Context, configMap *v1.Secret, opts metav1.UpdateOptions) (*v1.Secret, error) {
	if s.UpdateOK {

		return nil, nil
	} else {
		return nil, errUpdate
	}
}

func (s SecretMock) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return nil
}
func (s SecretMock) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return nil
}
func (s SecretMock) List(ctx context.Context, opts metav1.ListOptions) (*v1.SecretList, error) {
	return nil, nil
}
func (s SecretMock) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return nil, nil
}
func (s SecretMock) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.Secret, err error) {
	return nil, nil
}
func (s SecretMock) Apply(ctx context.Context, configMap *corev1.SecretApplyConfiguration, opts metav1.ApplyOptions) (result *v1.Secret, err error) {
	return nil, nil
}

//endregion
