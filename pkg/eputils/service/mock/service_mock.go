//
//   Copyright (c) 2022 Intel Corporation.
//
//   SPDX-License-Identifier: Apache-2.0
//
//
//

// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/intel/edge-conductor/pkg/eputils/service (interfaces: ServiceDeployer,HelmDeployerWrapper,YamlDeployerWrapper,ServiceTLSExtensionWrapper)

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	plugins "github.com/intel/edge-conductor/pkg/api/plugins"
	service "github.com/intel/edge-conductor/pkg/eputils/service"
)

// MockServiceDeployer is a mock of ServiceDeployer interface.
type MockServiceDeployer struct {
	ctrl     *gomock.Controller
	recorder *MockServiceDeployerMockRecorder
}

// MockServiceDeployerMockRecorder is the mock recorder for MockServiceDeployer.
type MockServiceDeployerMockRecorder struct {
	mock *MockServiceDeployer
}

// NewMockServiceDeployer creates a new mock instance.
func NewMockServiceDeployer(ctrl *gomock.Controller) *MockServiceDeployer {
	mock := &MockServiceDeployer{ctrl: ctrl}
	mock.recorder = &MockServiceDeployerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockServiceDeployer) EXPECT() *MockServiceDeployerMockRecorder {
	return m.recorder
}

// NewHelmDeployer mocks base method.
func (m *MockServiceDeployer) NewHelmDeployer(arg0, arg1, arg2, arg3 string) service.HelmDeployerWrapper {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewHelmDeployer", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(service.HelmDeployerWrapper)
	return ret0
}

// NewHelmDeployer indicates an expected call of NewHelmDeployer.
func (mr *MockServiceDeployerMockRecorder) NewHelmDeployer(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewHelmDeployer", reflect.TypeOf((*MockServiceDeployer)(nil).NewHelmDeployer), arg0, arg1, arg2, arg3)
}

// NewYamlDeployer mocks base method.
func (m *MockServiceDeployer) NewYamlDeployer(arg0, arg1, arg2 string, arg3 ...interface{}) service.YamlDeployerWrapper {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "NewYamlDeployer", varargs...)
	ret0, _ := ret[0].(service.YamlDeployerWrapper)
	return ret0
}

// NewYamlDeployer indicates an expected call of NewYamlDeployer.
func (mr *MockServiceDeployerMockRecorder) NewYamlDeployer(arg0, arg1, arg2 interface{}, arg3 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewYamlDeployer", reflect.TypeOf((*MockServiceDeployer)(nil).NewYamlDeployer), varargs...)
}

// MockHelmDeployerWrapper is a mock of HelmDeployerWrapper interface.
type MockHelmDeployerWrapper struct {
	ctrl     *gomock.Controller
	recorder *MockHelmDeployerWrapperMockRecorder
}

// MockHelmDeployerWrapperMockRecorder is the mock recorder for MockHelmDeployerWrapper.
type MockHelmDeployerWrapperMockRecorder struct {
	mock *MockHelmDeployerWrapper
}

// NewMockHelmDeployerWrapper creates a new mock instance.
func NewMockHelmDeployerWrapper(ctrl *gomock.Controller) *MockHelmDeployerWrapper {
	mock := &MockHelmDeployerWrapper{ctrl: ctrl}
	mock.recorder = &MockHelmDeployerWrapperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockHelmDeployerWrapper) EXPECT() *MockHelmDeployerWrapperMockRecorder {
	return m.recorder
}

// GetName mocks base method.
func (m *MockHelmDeployerWrapper) GetName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetName")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetName indicates an expected call of GetName.
func (mr *MockHelmDeployerWrapperMockRecorder) GetName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetName", reflect.TypeOf((*MockHelmDeployerWrapper)(nil).GetName))
}

// HelmInstall mocks base method.
func (m *MockHelmDeployerWrapper) HelmInstall(arg0 string, arg1 ...service.InstallOpt) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "HelmInstall", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// HelmInstall indicates an expected call of HelmInstall.
func (mr *MockHelmDeployerWrapperMockRecorder) HelmInstall(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HelmInstall", reflect.TypeOf((*MockHelmDeployerWrapper)(nil).HelmInstall), varargs...)
}

// HelmStatus mocks base method.
func (m *MockHelmDeployerWrapper) HelmStatus(arg0 string) (string, int) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HelmStatus", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(int)
	return ret0, ret1
}

// HelmStatus indicates an expected call of HelmStatus.
func (mr *MockHelmDeployerWrapperMockRecorder) HelmStatus(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HelmStatus", reflect.TypeOf((*MockHelmDeployerWrapper)(nil).HelmStatus), arg0)
}

// HelmUninstall mocks base method.
func (m *MockHelmDeployerWrapper) HelmUninstall(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HelmUninstall", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// HelmUninstall indicates an expected call of HelmUninstall.
func (mr *MockHelmDeployerWrapperMockRecorder) HelmUninstall(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HelmUninstall", reflect.TypeOf((*MockHelmDeployerWrapper)(nil).HelmUninstall), arg0)
}

// HelmUpgrade mocks base method.
func (m *MockHelmDeployerWrapper) HelmUpgrade(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HelmUpgrade", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// HelmUpgrade indicates an expected call of HelmUpgrade.
func (mr *MockHelmDeployerWrapperMockRecorder) HelmUpgrade(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HelmUpgrade", reflect.TypeOf((*MockHelmDeployerWrapper)(nil).HelmUpgrade), arg0)
}

// MockYamlDeployerWrapper is a mock of YamlDeployerWrapper interface.
type MockYamlDeployerWrapper struct {
	ctrl     *gomock.Controller
	recorder *MockYamlDeployerWrapperMockRecorder
}

// MockYamlDeployerWrapperMockRecorder is the mock recorder for MockYamlDeployerWrapper.
type MockYamlDeployerWrapperMockRecorder struct {
	mock *MockYamlDeployerWrapper
}

// NewMockYamlDeployerWrapper creates a new mock instance.
func NewMockYamlDeployerWrapper(ctrl *gomock.Controller) *MockYamlDeployerWrapper {
	mock := &MockYamlDeployerWrapper{ctrl: ctrl}
	mock.recorder = &MockYamlDeployerWrapperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockYamlDeployerWrapper) EXPECT() *MockYamlDeployerWrapperMockRecorder {
	return m.recorder
}

// GetName mocks base method.
func (m *MockYamlDeployerWrapper) GetName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetName")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetName indicates an expected call of GetName.
func (mr *MockYamlDeployerWrapperMockRecorder) GetName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetName", reflect.TypeOf((*MockYamlDeployerWrapper)(nil).GetName))
}

// YamlInstall mocks base method.
func (m *MockYamlDeployerWrapper) YamlInstall(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "YamlInstall", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// YamlInstall indicates an expected call of YamlInstall.
func (mr *MockYamlDeployerWrapperMockRecorder) YamlInstall(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "YamlInstall", reflect.TypeOf((*MockYamlDeployerWrapper)(nil).YamlInstall), arg0)
}

// YamlUninstall mocks base method.
func (m *MockYamlDeployerWrapper) YamlUninstall(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "YamlUninstall", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// YamlUninstall indicates an expected call of YamlUninstall.
func (mr *MockYamlDeployerWrapperMockRecorder) YamlUninstall(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "YamlUninstall", reflect.TypeOf((*MockYamlDeployerWrapper)(nil).YamlUninstall), arg0)
}

// MockServiceTLSExtensionWrapper is a mock of ServiceTLSExtensionWrapper interface.
type MockServiceTLSExtensionWrapper struct {
	ctrl     *gomock.Controller
	recorder *MockServiceTLSExtensionWrapperMockRecorder
}

// MockServiceTLSExtensionWrapperMockRecorder is the mock recorder for MockServiceTLSExtensionWrapper.
type MockServiceTLSExtensionWrapperMockRecorder struct {
	mock *MockServiceTLSExtensionWrapper
}

// NewMockServiceTLSExtensionWrapper creates a new mock instance.
func NewMockServiceTLSExtensionWrapper(ctrl *gomock.Controller) *MockServiceTLSExtensionWrapper {
	mock := &MockServiceTLSExtensionWrapper{ctrl: ctrl}
	mock.recorder = &MockServiceTLSExtensionWrapperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockServiceTLSExtensionWrapper) EXPECT() *MockServiceTLSExtensionWrapperMockRecorder {
	return m.recorder
}

// GenSvcSecretFromTLSExtension mocks base method.
func (m *MockServiceTLSExtensionWrapper) GenSvcSecretFromTLSExtension(arg0 []*plugins.EpParamsExtensionsItems0, arg1, arg2, arg3 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenSvcSecretFromTLSExtension", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// GenSvcSecretFromTLSExtension indicates an expected call of GenSvcSecretFromTLSExtension.
func (mr *MockServiceTLSExtensionWrapperMockRecorder) GenSvcSecretFromTLSExtension(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenSvcSecretFromTLSExtension", reflect.TypeOf((*MockServiceTLSExtensionWrapper)(nil).GenSvcSecretFromTLSExtension), arg0, arg1, arg2, arg3)
}

// GenSvcTLSCertFromTLSExtension mocks base method.
func (m *MockServiceTLSExtensionWrapper) GenSvcTLSCertFromTLSExtension(arg0 []*plugins.EpParamsExtensionsItems0, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenSvcTLSCertFromTLSExtension", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// GenSvcTLSCertFromTLSExtension indicates an expected call of GenSvcTLSCertFromTLSExtension.
func (mr *MockServiceTLSExtensionWrapperMockRecorder) GenSvcTLSCertFromTLSExtension(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenSvcTLSCertFromTLSExtension", reflect.TypeOf((*MockServiceTLSExtensionWrapper)(nil).GenSvcTLSCertFromTLSExtension), arg0, arg1)
}
