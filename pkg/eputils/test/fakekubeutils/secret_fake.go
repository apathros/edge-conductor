/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package fakekubeutils

type FakeSecret struct {
	Name         string
	Namespace    string
	FieldManager string
	Kubeconfig   string
}

func (s *FakeSecret) New() error {
	return nil
}

func (s *FakeSecret) Get() error {
	return nil
}

func (s *FakeSecret) RenewData(key string, data []byte) error {
	return nil
}

func (s *FakeSecret) RenewStringData(key string, data string) error {
	return nil
}

func (s *FakeSecret) Update() error {
	return nil
}

func (s *FakeSecret) GetData() map[string][]byte {
	return map[string][]byte{"": []byte("")}
}

func (s *FakeSecret) GetStringData() map[string]string {
	return map[string]string{"": ""}
}
