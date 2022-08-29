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

type Secret struct {
	Namespace    string
	Name         string
	FieldManager string

	Client    corev1client.CoreV1Interface
	SecretObj *corev1.Secret
}

func NewSecret(namespace, name, feildManager, kubeconfig string) (SecretWrapper, error) {
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

	return &Secret{
		Namespace:    namespace,
		Name:         name,
		FieldManager: feildManager,
		Client:       client,
		SecretObj:    nil,
	}, nil
}

func (s *Secret) New() error {
	// Create Secret
	s.SecretObj = &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Name,
			Namespace: s.Namespace,
		},
		Type: corev1.SecretTypeOpaque,
	}
	s.SecretObj.Name = s.Name
	s.SecretObj.Data = map[string][]byte{}
	s.SecretObj.StringData = map[string]string{}

	return nil
}

func (s *Secret) Get() error {
	if secret, err := s.Client.Secrets(s.Namespace).Get(
		context.Background(), s.Name, metav1.GetOptions{}); err != nil {
		// Secret not found on the cluster
		return err
	} else {
		// Found Secret
		s.SecretObj = secret
	}
	return nil
}

func (s *Secret) RenewData(key string, data []byte) error {
	if s.SecretObj.Data == nil {
		s.SecretObj.Data = map[string][]byte{}
	}

	if _, exists := s.SecretObj.Data[key]; exists {
		log.Infoln(s.Name, "Data[", key, "] will be updated.")
	} else {
		log.Infoln(s.Name, "Data[", key, "] will be added.")
	}
	s.SecretObj.Data[key] = data
	return s.Update()
}

func (s *Secret) RenewStringData(key string, data string) error {
	if s.SecretObj.StringData == nil {
		s.SecretObj.StringData = map[string]string{}
	}

	if _, exists := s.SecretObj.StringData[key]; exists {
		log.Infoln(s.Name, "StringData[", key, "] will be updated.")
	} else {
		log.Infoln(s.Name, "StringData[", key, "] will be added.")
	}
	s.SecretObj.StringData[key] = data
	return s.Update()
}

//nolint: dupl   //TODO: Need fix this error in v0.6
func (s *Secret) Update() error {
	if _, err := s.Client.Secrets(s.Namespace).Get(
		context.Background(), s.Name, metav1.GetOptions{}); err != nil {
		// Secret not found on the cluster
		log.Infoln("Create Secret", s.Name, "on behalf of", s.FieldManager)
		_, err = s.Client.Secrets(s.Namespace).Create(
			context.Background(), s.SecretObj,
			metav1.CreateOptions{FieldManager: s.FieldManager})
		if err != nil {
			log.Errorln("Failed to create Secret on target cluster:", err)
			return err
		}
	} else {
		// Found Secret
		log.Infoln("Update Secret", s.Name, "on behalf of", s.FieldManager)
		_, err = s.Client.Secrets(s.Namespace).Update(
			context.Background(), s.SecretObj,
			metav1.UpdateOptions{FieldManager: s.FieldManager})
		if err != nil {
			log.Errorln("Failed to update Secret on target cluster:", err)
			return err
		}
	}
	return s.Get()
}

func (s *Secret) GetData() map[string][]byte {
	return s.SecretObj.Data
}

func (s *Secret) GetStringData() map[string]string {
	return s.SecretObj.StringData
}
