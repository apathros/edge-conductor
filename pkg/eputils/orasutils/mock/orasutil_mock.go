//
//   Copyright (c) 2022 Intel Corporation.
//
//   SPDX-License-Identifier: Apache-2.0
//
//
//

// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/intel/edge-conductor/pkg/eputils/orasutils (interfaces: OrasInterface,OrasUtilInterface)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	content "github.com/containerd/containerd/content"
	remotes "github.com/containerd/containerd/remotes"
	oras "github.com/deislabs/oras/pkg/oras"
	gomock "github.com/golang/mock/gomock"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

// MockOrasInterface is a mock of OrasInterface interface.
type MockOrasInterface struct {
	ctrl     *gomock.Controller
	recorder *MockOrasInterfaceMockRecorder
}

// MockOrasInterfaceMockRecorder is the mock recorder for MockOrasInterface.
type MockOrasInterfaceMockRecorder struct {
	mock *MockOrasInterface
}

// NewMockOrasInterface creates a new mock instance.
func NewMockOrasInterface(ctrl *gomock.Controller) *MockOrasInterface {
	mock := &MockOrasInterface{ctrl: ctrl}
	mock.recorder = &MockOrasInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockOrasInterface) EXPECT() *MockOrasInterfaceMockRecorder {
	return m.recorder
}

// Pull mocks base method.
func (m *MockOrasInterface) Pull(arg0 context.Context, arg1 remotes.Resolver, arg2 string, arg3 content.Ingester, arg4 ...oras.PullOpt) (v1.Descriptor, []v1.Descriptor, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2, arg3}
	for _, a := range arg4 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Pull", varargs...)
	ret0, _ := ret[0].(v1.Descriptor)
	ret1, _ := ret[1].([]v1.Descriptor)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Pull indicates an expected call of Pull.
func (mr *MockOrasInterfaceMockRecorder) Pull(arg0, arg1, arg2, arg3 interface{}, arg4 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2, arg3}, arg4...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Pull", reflect.TypeOf((*MockOrasInterface)(nil).Pull), varargs...)
}

// Push mocks base method.
func (m *MockOrasInterface) Push(arg0 context.Context, arg1 remotes.Resolver, arg2 string, arg3 content.Provider, arg4 []v1.Descriptor, arg5 ...oras.PushOpt) (v1.Descriptor, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2, arg3, arg4}
	for _, a := range arg5 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Push", varargs...)
	ret0, _ := ret[0].(v1.Descriptor)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Push indicates an expected call of Push.
func (mr *MockOrasInterfaceMockRecorder) Push(arg0, arg1, arg2, arg3, arg4 interface{}, arg5 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2, arg3, arg4}, arg5...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Push", reflect.TypeOf((*MockOrasInterface)(nil).Push), varargs...)
}

// MockOrasUtilInterface is a mock of OrasUtilInterface interface.
type MockOrasUtilInterface struct {
	ctrl     *gomock.Controller
	recorder *MockOrasUtilInterfaceMockRecorder
}

// MockOrasUtilInterfaceMockRecorder is the mock recorder for MockOrasUtilInterface.
type MockOrasUtilInterfaceMockRecorder struct {
	mock *MockOrasUtilInterface
}

// NewMockOrasUtilInterface creates a new mock instance.
func NewMockOrasUtilInterface(ctrl *gomock.Controller) *MockOrasUtilInterface {
	mock := &MockOrasUtilInterface{ctrl: ctrl}
	mock.recorder = &MockOrasUtilInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockOrasUtilInterface) EXPECT() *MockOrasUtilInterfaceMockRecorder {
	return m.recorder
}

// OrasPullFile mocks base method.
func (m *MockOrasUtilInterface) OrasPullFile(arg0, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OrasPullFile", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// OrasPullFile indicates an expected call of OrasPullFile.
func (mr *MockOrasUtilInterfaceMockRecorder) OrasPullFile(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OrasPullFile", reflect.TypeOf((*MockOrasUtilInterface)(nil).OrasPullFile), arg0, arg1)
}

// OrasPushFile mocks base method.
func (m *MockOrasUtilInterface) OrasPushFile(arg0, arg1, arg2 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OrasPushFile", arg0, arg1, arg2)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// OrasPushFile indicates an expected call of OrasPushFile.
func (mr *MockOrasUtilInterfaceMockRecorder) OrasPushFile(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OrasPushFile", reflect.TypeOf((*MockOrasUtilInterface)(nil).OrasPushFile), arg0, arg1, arg2)
}
