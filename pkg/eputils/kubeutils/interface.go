/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

//go:generate mockgen -destination=./mock/cliapi_mock.go -package=mock -copyright_file=../../../api/schemas/license-header.txt ep/pkg/eputils/kubeutils KubeClientWrapper
//go:generate mockgen -destination=./mock/configmap_mock.go -package=mock -copyright_file=../../../api/schemas/license-header.txt ep/pkg/eputils/kubeutils ConfigMapWrapper
//go:generate mockgen -destination=./mock/secret_mock.go -package=mock -copyright_file=../../../api/schemas/license-header.txt ep/pkg/eputils/kubeutils SecretWrapper

package kubeutils

import (
	pluginapi "ep/pkg/api/plugins"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

type KubeClientWrapper interface {
	ClientFromKubeConfig(kubeconfig string) (*kubernetes.Clientset, error)
	ClientFromEPKubeConfig(input_kubeconfig *pluginapi.Filecontent) (*kubernetes.Clientset, error)
	RestClientFromEPKubeConfig(input_kubeconfig *pluginapi.Filecontent) (*restclient.Config, error)
	RestConfigFromKubeConfig(kubeconfig string) (*restclient.Config, error)
	NewRestClient(kubeconfigcontent []byte) (*restclient.Config, error)
	NewClient(kubeconfigcontent []byte) (*kubernetes.Clientset, error)

	CreateNamespace(kubeconfig, namespace string) error
	GetNodeList(input_kubeconfig *pluginapi.Filecontent, labelSelectorStr string) (*corev1.NodeList, error)

	NewConfigMap(namespace, name, feildManager, kubeconfig string) (ConfigMapWrapper, error)
	NewSecret(namespace, name, feildManager, kubeconfig string) (SecretWrapper, error)

	NewDeployment(namespace, name, feildManager, kubeconfig string) (DeploymentWrapper, error)
}

type ConfigMapWrapper interface {
	New() error
	Get() error
	RenewData(key string, data string) error
	RenewBinaryData(key string, data []byte) error
	RemoveData(key string) error
	RemoveBinaryData(key string) error
	Update() error
	GetData() map[string]string
	GetBinaryData() map[string][]byte
}

type SecretWrapper interface {
	New() error
	Get() error
	RenewData(key string, data []byte) error
	RenewStringData(key string, data string) error
	Update() error
	GetData() map[string][]byte
	GetStringData() map[string]string
}

type DeploymentWrapper interface {
	New() error
	Get() error
	GetStatus() appsv1.DeploymentStatus
}
