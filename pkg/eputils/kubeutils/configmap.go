/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package kubeutils

import (
	"context"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
)

type ConfigMap struct {
	Namespace    string
	Name         string
	FieldManager string

	Client       corev1client.CoreV1Interface
	ConfigMapObj *corev1.ConfigMap
}

func NewConfigMap(namespace, name, feildManager, kubeconfig string) (ConfigMapWrapper, error) {
	// Create client
	restConfig, err := RestConfigFromKubeConfig(kubeconfig)
	if err != nil {
		log.Errorln("Failed to create restclient config:", err)
		return nil, err
	}
	client, err := corev1client.NewForConfig(restConfig)
	if err != nil {
		log.Errorln("Failed to create restclient:", err)
		return nil, err
	}

	return &ConfigMap{
		Namespace:    namespace,
		Name:         name,
		FieldManager: feildManager,
		Client:       client,
		ConfigMapObj: nil,
	}, nil
}

func (c *ConfigMap) New() error {
	// Create ConfigMap
	c.ConfigMapObj = &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.Name,
			Namespace: c.Namespace,
		},
	}
	c.ConfigMapObj.Name = c.Name
	c.ConfigMapObj.Data = map[string]string{}
	c.ConfigMapObj.BinaryData = map[string][]byte{}

	return nil
}

func (c *ConfigMap) Get() error {
	if cm, err := c.Client.ConfigMaps(c.Namespace).Get(
		context.Background(), c.Name, metav1.GetOptions{}); err != nil {
		// ConfigMap not found on the cluster
		return err
	} else {
		// Found ConfigMap
		c.ConfigMapObj = cm
	}
	return nil
}

func (c *ConfigMap) RenewData(key string, data string) error {
	if _, exists := c.ConfigMapObj.Data[key]; exists {
		log.Infoln(c.Name, "Data[", key, "] will be updated.")
	} else {
		log.Infoln(c.Name, "Data[", key, "] will be added.")
	}
	c.ConfigMapObj.Data[key] = data
	return c.Update()
}

func (c *ConfigMap) RenewBinaryData(key string, data []byte) error {
	if _, exists := c.ConfigMapObj.BinaryData[key]; exists {
		log.Infoln(c.Name, "BinaryData[", key, "] will be updated.")
	} else {
		log.Infoln(c.Name, "BinaryData[", key, "] will be added.")
	}
	c.ConfigMapObj.BinaryData[key] = data
	return c.Update()
}

func (c *ConfigMap) RemoveData(key string) error {
	if _, exists := c.ConfigMapObj.Data[key]; exists {
		log.Infoln(c.Name, "Data[", key, "] will be removed.")
	}
	delete(c.ConfigMapObj.Data, key)
	return c.Update()
}

func (c *ConfigMap) RemoveBinaryData(key string) error {
	if _, exists := c.ConfigMapObj.BinaryData[key]; exists {
		log.Infoln(c.Name, "BinaryData[", key, "] will be removed.")
	}
	delete(c.ConfigMapObj.BinaryData, key)
	return c.Update()
}

//nolint: dupl   //TODO: Need fix this error in v0.6
func (c *ConfigMap) Update() error {
	if _, err := c.Client.ConfigMaps(c.Namespace).Get(
		context.Background(), c.Name, metav1.GetOptions{}); err != nil {
		// ConfigMap not found on the cluster
		log.Infoln("Create ConfigMap", c.Name, "on behalf of", c.FieldManager)
		_, err = c.Client.ConfigMaps(c.Namespace).Create(
			context.Background(), c.ConfigMapObj,
			metav1.CreateOptions{FieldManager: c.FieldManager})
		if err != nil {
			log.Errorln("Failed to create ConfigMap on target cluster:", err)
			return err
		}
	} else {
		// Found ConfigMap
		log.Infoln("Update ConfigMap", c.Name, "on behalf of", c.FieldManager)
		_, err = c.Client.ConfigMaps(c.Namespace).Update(
			context.Background(), c.ConfigMapObj,
			metav1.UpdateOptions{FieldManager: c.FieldManager})
		if err != nil {
			log.Errorln("Failed to update ConfigMap on target cluster:", err)
			return err
		}
	}
	return c.Get()
}

func (c *ConfigMap) GetData() map[string]string {
	return c.ConfigMapObj.Data
}

func (c *ConfigMap) GetBinaryData() map[string][]byte {
	return c.ConfigMapObj.BinaryData
}
