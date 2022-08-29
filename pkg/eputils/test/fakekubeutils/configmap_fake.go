/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package fakekubeutils

type FakeConfigMap struct {
}

func (c *FakeConfigMap) New() error {
	return nil
}
func (c *FakeConfigMap) Get() error {
	return nil
}
func (c *FakeConfigMap) RenewData(key string, data string) error {
	return nil
}
func (c *FakeConfigMap) RenewBinaryData(key string, data []byte) error {
	return nil
}
func (c *FakeConfigMap) RemoveData(key string) error {
	return nil
}
func (c *FakeConfigMap) RemoveBinaryData(key string) error {
	return nil
}
func (c *FakeConfigMap) Update() error {
	return nil
}
func (c *FakeConfigMap) GetData() map[string]string {
	return nil
}
func (c *FakeConfigMap) GetBinaryData() map[string][]byte {
	return nil
}
