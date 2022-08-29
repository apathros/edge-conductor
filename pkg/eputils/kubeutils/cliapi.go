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
	"io/ioutil"

	"ep/pkg/eputils"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"
)

func ClientFromKubeConfig(kubeconfig string) (*kubernetes.Clientset, error) {
	configcontent, err := ioutil.ReadFile(kubeconfig)
	if err != nil {
		log.Errorf("Failed to read file %s: %s", kubeconfig, err)
		return nil, err
	}

	return NewClient(configcontent)
}

func ClientFromEPKubeConfig(input_kubeconfig *pluginapi.Filecontent) (*kubernetes.Clientset, error) {
	if client, err := NewClient([]byte(input_kubeconfig.Content)); err != nil {
		return nil, err
	} else {
		return client, nil
	}
}

func RestClientFromEPKubeConfig(input_kubeconfig *pluginapi.Filecontent) (*restclient.Config, error) {
	if client, err := NewRestClient([]byte(input_kubeconfig.Content)); err != nil {
		return nil, err
	} else {
		return client, nil
	}
}

func RestConfigFromKubeConfig(kubeconfig string) (*restclient.Config, error) {
	configcontent, err := ioutil.ReadFile(kubeconfig)
	if err != nil {
		log.Errorf("Failed to read file %s: %s", kubeconfig, err)
		return nil, err
	}

	return NewRestClient(configcontent)
}

func NewRestClient(kubeconfigcontent []byte) (*restclient.Config, error) {
	kubeconfig, err := yaml.YAMLToJSON(kubeconfigcontent)
	if err != nil {
		log.Errorf("Failed to read kubeconfig  %s", err)
		return nil, err
	}
	restconf, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		log.Errorf("Failed to get restconf %s", err)
		return nil, err
	}
	return restconf, nil
}

func NewClient(kubeconfigcontent []byte) (*kubernetes.Clientset, error) {
	restconf, err := NewRestClient(kubeconfigcontent)
	if err != nil {
		log.Errorf("Failed to get restconf %s", err)
		return nil, err
	}
	client, err := kubernetes.NewForConfig(restconf)
	if err != nil {
		log.Errorf("Failed to gen new k8s client. %s", err)
		return nil, err
	}
	return client, nil
}

func CreateNamespace(kubeconfig, namespace string) error {
	client, err := ClientFromKubeConfig(kubeconfig)
	if err != nil {
		return err
	}

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	if _, err := client.CoreV1().Namespaces().Get(
		context.Background(), namespace, metav1.GetOptions{}); err != nil {
		_, err = client.CoreV1().Namespaces().Create(
			context.Background(), ns, metav1.CreateOptions{})
		if err != nil {
			log.Errorln("Failed to create namespace:", namespace, err)
			return err
		} else {
			log.Infoln("Namespace", namespace, "created.")
		}
	} else {
		log.Infoln("Namespace", namespace, "already exists.")
	}

	return nil
}

func GetNodeList(input_kubeconfig *pluginapi.Filecontent, labelSelectorStr string) (*corev1.NodeList, error) {
	client, err := ClientFromEPKubeConfig(input_kubeconfig)
	if err != nil {
		log.Errorf("Failed to Get ClientSet. %s", err)
		return nil, err
	}
	node_client := client.CoreV1().Nodes()
	nodelist, err := node_client.List(context.TODO(), metav1.ListOptions{LabelSelector: labelSelectorStr})
	if err != nil {
		log.Errorf("Failed to Get K8S nodes. %s", err)
		return nil, err
	}
	if len(nodelist.Items) > 0 {
		return nodelist, nil
	} else {
		return nil, eputils.GetError("errNok8sNode")
	}
}
