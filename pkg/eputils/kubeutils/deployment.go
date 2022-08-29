/*
* Copyright (c) 2021 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package kubeutils

import (
	"context"

	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1client "k8s.io/client-go/kubernetes/typed/apps/v1"
)

type Deployment struct {
	Namespace    string
	Name         string
	FieldManager string

	Client        appsv1client.AppsV1Interface
	DeploymentObj *appsv1.Deployment
}

func NewDeployment(namespace, name, feildManager, kubeconfig string) (DeploymentWrapper, error) {
	// Create client
	restConfig, err := RestConfigFromKubeConfig(kubeconfig)
	if err != nil {
		log.Errorln("Failed to create restclient config:", err)
		return nil, err
	}
	client, err := appsv1client.NewForConfig(restConfig)
	if err != nil {
		log.Errorln("Failed to create restclient:", err)
		return nil, err
	}

	return &Deployment{
		Namespace:     namespace,
		Name:          name,
		FieldManager:  feildManager,
		Client:        client,
		DeploymentObj: nil,
	}, nil
}

func (d *Deployment) Get() error {
	if deployment, err := d.Client.Deployments(d.Namespace).Get(
		context.Background(), d.Name, metav1.GetOptions{}); err != nil {
		// Deployment not found on the cluster
		return err
	} else {
		// Found Deployment
		d.DeploymentObj = deployment
	}
	return nil
}

func (d *Deployment) GetStatus() appsv1.DeploymentStatus {
	return d.DeploymentObj.Status
}

var replicas1 int32 = 1

func (d *Deployment) New() error {
	// Create Deployment
	d.DeploymentObj = &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.Name,
			Namespace: d.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas1,
		},
		Status: appsv1.DeploymentStatus{
			ObservedGeneration: 10,
		},
	}

	return nil
}
