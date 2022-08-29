/*
* Copyright (c) 2021 Intel Corporation.
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
	v1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"

	appsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
	applyconfigurationsautoscalingv1 "k8s.io/client-go/applyconfigurations/autoscaling/v1"
	appsv1client "k8s.io/client-go/kubernetes/typed/apps/v1"
	restclient "k8s.io/client-go/rest"
)

var (
	errRestCfg = errors.New("RestConfigFromKubeConfig.error")
	errNewFor  = errors.New("newforconfig.error")
	//errGet     = errors.New("get.error")
)

func patchRestConfigFromKubeConfigfail(t *testing.T) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(RestConfigFromKubeConfig, func(kubeconfig string) (*restclient.Config, error) {
		unpatch(t, patch)
		return nil, errRestCfg
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

}
func patchRestConfigFromKubeConfigok(t *testing.T) {
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
func patchappsv1clientnewforconfigfail(t *testing.T) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(appsv1client.NewForConfig, func(c *restclient.Config) (*appsv1client.AppsV1Client, error) {
		unpatch(t, patch)
		return nil, errNewFor
	})

	if patchErr != nil {
		t.Errorf("patch error %v", patchErr)
	}

}
func patchappsv1clientnewforconfigok(t *testing.T) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(appsv1client.NewForConfig, func(c *restclient.Config) (*appsv1client.AppsV1Client, error) {
		unpatch(t, patch)
		return nil, nil
	})

	if patchErr != nil {
		t.Errorf("patch error %v", patchErr)
	}

}

func TestNewDeployment(t *testing.T) {
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
			expectError: errRestCfg,
			beforeTest: func() {
				patchRestConfigFromKubeConfigfail(t)
			},
		},
		{
			name: "appsv1client.NewForConfig fail",
			input: []string{
				"default namespace",
				"default name",
				"default fieldmanager",
				"default kubeconfig",
			},
			expectError: errNewFor,
			beforeTest: func() {
				patchRestConfigFromKubeConfigok(t)
				patchappsv1clientnewforconfigfail(t)
			},
		},
		{
			name: "NewConfigMap ok",
			input: []string{
				"default namespace",
				"default name",
				"default fieldmanager",
				"default kubeconfig",
			},
			expectError: nil,
			beforeTest: func() {
				patchRestConfigFromKubeConfigok(t)
				patchappsv1clientnewforconfigok(t)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, tc := range cases {
				if tc.beforeTest != nil {
					tc.beforeTest()
				}

				_, err := NewDeployment(tc.input[0], tc.input[1], tc.input[2], tc.input[3])

				if !isExpectedError(err, tc.expectError) {
					t.Errorf("Unexpected error: %v", err)
				}

			}
		})
	}
}

func patchDeploymentok(t *testing.T, GetOK bool, CreateOK bool, UpdateOK bool) *mpatch.Patch {
	var patch *mpatch.Patch
	var patchErr error
	d, _ := NewDeployment("mock namespace", "mock name", "mock fieldmanager", kubeconfig_good_filepath)

	patch, patchErr = mpatch.PatchInstanceMethodByName(reflect.TypeOf(d.(*Deployment).Client), "Deployments", func(*appsv1client.AppsV1Client, string) appsv1client.DeploymentInterface {
		return DeploymentPatch{GetOK: GetOK, CreateOK: CreateOK, UpdateOK: UpdateOK}
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

	return patch
}

type DeploymentPatch struct {
	GetOK    bool
	CreateOK bool
	UpdateOK bool
}

func (s DeploymentPatch) Apply(ctx context.Context, deployment *appsv1.DeploymentApplyConfiguration, opts metav1.ApplyOptions) (result *v1.Deployment, err error) {
	return nil, nil
}
func (s DeploymentPatch) Create(ctx context.Context, deployment *v1.Deployment, opts metav1.CreateOptions) (*v1.Deployment, error) {
	return nil, nil
}
func (s DeploymentPatch) Update(ctx context.Context, deployment *v1.Deployment, opts metav1.UpdateOptions) (*v1.Deployment, error) {
	return nil, nil
}
func (s DeploymentPatch) UpdateStatus(ctx context.Context, deployment *v1.Deployment, opts metav1.UpdateOptions) (*v1.Deployment, error) {
	return nil, nil
}
func (s DeploymentPatch) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return nil
}
func (s DeploymentPatch) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return nil
}
func (s DeploymentPatch) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.Deployment, error) {
	if s.GetOK {
		return nil, nil
	} else {
		return nil, errGet
	}
}
func (s DeploymentPatch) List(ctx context.Context, opts metav1.ListOptions) (*v1.DeploymentList, error) {
	return nil, nil
}
func (s DeploymentPatch) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return nil, nil
}
func (s DeploymentPatch) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.Deployment, err error) {
	return nil, nil
}
func (s DeploymentPatch) ApplyStatus(ctx context.Context, deployment *appsv1.DeploymentApplyConfiguration, opts metav1.ApplyOptions) (result *v1.Deployment, err error) {
	return nil, nil
}
func (s DeploymentPatch) GetScale(ctx context.Context, deploymentName string, options metav1.GetOptions) (*autoscalingv1.Scale, error) {
	return nil, nil
}
func (s DeploymentPatch) UpdateScale(ctx context.Context, deploymentName string, scale *autoscalingv1.Scale, opts metav1.UpdateOptions) (*autoscalingv1.Scale, error) {
	return nil, nil
}
func (s DeploymentPatch) ApplyScale(ctx context.Context, deploymentName string, scale *applyconfigurationsautoscalingv1.ScaleApplyConfiguration, opts metav1.ApplyOptions) (*autoscalingv1.Scale, error) {
	return nil, nil
}
func TestDeployment_Get(t *testing.T) {
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
				patchHandler = patchDeploymentok(t, false, false, false)
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
				patchHandler = patchDeploymentok(t, true, false, false)
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

				d, _ := NewDeployment("mock namespace", "mock name", "mock fieldmanager", kubeconfig_good_filepath)
				err := d.Get()

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

func TestDeployment_GetStatus(t *testing.T) {
	var patchHandler *mpatch.Patch
	cases := []struct {
		name        string
		expectedRet int64
		beforeTest  func()
		teardown    func()
	}{
		{
			name:        "Deployment Get Status",
			expectedRet: 10,
			beforeTest: func() {
				patchHandler = patchDeploymentok(t, true, false, false)
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
				d, err := NewDeployment("mock namespace1", "mock name1", "mock fieldmanager1", kubeconfig_good_filepath)
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				err = d.New()
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				result := d.GetStatus()

				if tc.expectedRet != result.ObservedGeneration {
					t.Errorf("Test case %s: fail", tc.name)
				}

				if tc.teardown != nil {
					tc.teardown()
				}

			}
		})
	}
}

func TestDeployment_New(t *testing.T) {
	cases := []struct {
		name           string
		input          string
		expectError    error
		expectErrorMsg string
		beforeTest     func()
	}{
		{
			name:           "Create new Deployment",
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

				d, err := NewDeployment("mock namespace1", "mock name1", "mock fieldmanager1", kubeconfig_good_filepath)
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				err = d.New()
				if !isExpectedError(err, tc.expectError) {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
