//
//   Copyright (c) 2022 Intel Corporation.
//
//   SPDX-License-Identifier: Apache-2.0
//
//
//

// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/intel/edge-conductor/pkg/eputils (interfaces: FileWrapper)

// Package mock is a generated GoMock package.
package mock

import (
	fs "io/fs"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	eputils "github.com/intel/edge-conductor/pkg/eputils"
)

// MockFileWrapper is a mock of FileWrapper interface.
type MockFileWrapper struct {
	ctrl     *gomock.Controller
	recorder *MockFileWrapperMockRecorder
}

// MockFileWrapperMockRecorder is the mock recorder for MockFileWrapper.
type MockFileWrapperMockRecorder struct {
	mock *MockFileWrapper
}

// NewMockFileWrapper creates a new mock instance.
func NewMockFileWrapper(ctrl *gomock.Controller) *MockFileWrapper {
	mock := &MockFileWrapper{ctrl: ctrl}
	mock.recorder = &MockFileWrapperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFileWrapper) EXPECT() *MockFileWrapperMockRecorder {
	return m.recorder
}

// CheckFileLink mocks base method.
func (m *MockFileWrapper) CheckFileLink(arg0 string) (eputils.Filelink, string) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckFileLink", arg0)
	ret0, _ := ret[0].(eputils.Filelink)
	ret1, _ := ret[1].(string)
	return ret0, ret1
}

// CheckFileLink indicates an expected call of CheckFileLink.
func (mr *MockFileWrapperMockRecorder) CheckFileLink(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckFileLink", reflect.TypeOf((*MockFileWrapper)(nil).CheckFileLink), arg0)
}

// CompressTar mocks base method.
func (m *MockFileWrapper) CompressTar(arg0, arg1 string, arg2 fs.FileMode) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CompressTar", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// CompressTar indicates an expected call of CompressTar.
func (mr *MockFileWrapperMockRecorder) CompressTar(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CompressTar", reflect.TypeOf((*MockFileWrapper)(nil).CompressTar), arg0, arg1, arg2)
}

// CopyFile mocks base method.
func (m *MockFileWrapper) CopyFile(arg0, arg1 string) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CopyFile", arg0, arg1)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CopyFile indicates an expected call of CopyFile.
func (mr *MockFileWrapperMockRecorder) CopyFile(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CopyFile", reflect.TypeOf((*MockFileWrapper)(nil).CopyFile), arg0, arg1)
}

// CreateFolderIfNotExist mocks base method.
func (m *MockFileWrapper) CreateFolderIfNotExist(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateFolderIfNotExist", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateFolderIfNotExist indicates an expected call of CreateFolderIfNotExist.
func (mr *MockFileWrapperMockRecorder) CreateFolderIfNotExist(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateFolderIfNotExist", reflect.TypeOf((*MockFileWrapper)(nil).CreateFolderIfNotExist), arg0)
}

// DownloadFile mocks base method.
func (m *MockFileWrapper) DownloadFile(arg0, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DownloadFile", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DownloadFile indicates an expected call of DownloadFile.
func (mr *MockFileWrapperMockRecorder) DownloadFile(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DownloadFile", reflect.TypeOf((*MockFileWrapper)(nil).DownloadFile), arg0, arg1)
}

// FileExists mocks base method.
func (m *MockFileWrapper) FileExists(arg0 string) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FileExists", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// FileExists indicates an expected call of FileExists.
func (mr *MockFileWrapperMockRecorder) FileExists(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FileExists", reflect.TypeOf((*MockFileWrapper)(nil).FileExists), arg0)
}

// GzipCompress mocks base method.
func (m *MockFileWrapper) GzipCompress(arg0, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GzipCompress", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// GzipCompress indicates an expected call of GzipCompress.
func (mr *MockFileWrapperMockRecorder) GzipCompress(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GzipCompress", reflect.TypeOf((*MockFileWrapper)(nil).GzipCompress), arg0, arg1)
}

// IsValidFile mocks base method.
func (m *MockFileWrapper) IsValidFile(arg0 string) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsValidFile", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsValidFile indicates an expected call of IsValidFile.
func (mr *MockFileWrapperMockRecorder) IsValidFile(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsValidFile", reflect.TypeOf((*MockFileWrapper)(nil).IsValidFile), arg0)
}

// LoadJsonFromFile mocks base method.
func (m *MockFileWrapper) LoadJsonFromFile(arg0 string, arg1 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LoadJsonFromFile", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// LoadJsonFromFile indicates an expected call of LoadJsonFromFile.
func (mr *MockFileWrapperMockRecorder) LoadJsonFromFile(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LoadJsonFromFile", reflect.TypeOf((*MockFileWrapper)(nil).LoadJsonFromFile), arg0, arg1)
}

// MakeDir mocks base method.
func (m *MockFileWrapper) MakeDir(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MakeDir", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// MakeDir indicates an expected call of MakeDir.
func (mr *MockFileWrapperMockRecorder) MakeDir(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MakeDir", reflect.TypeOf((*MockFileWrapper)(nil).MakeDir), arg0)
}

// RemoveFile mocks base method.
func (m *MockFileWrapper) RemoveFile(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveFile", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveFile indicates an expected call of RemoveFile.
func (mr *MockFileWrapperMockRecorder) RemoveFile(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveFile", reflect.TypeOf((*MockFileWrapper)(nil).RemoveFile), arg0)
}

// UncompressTgz mocks base method.
func (m *MockFileWrapper) UncompressTgz(arg0, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UncompressTgz", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UncompressTgz indicates an expected call of UncompressTgz.
func (mr *MockFileWrapperMockRecorder) UncompressTgz(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UncompressTgz", reflect.TypeOf((*MockFileWrapper)(nil).UncompressTgz), arg0, arg1)
}

// WriteStringToFile mocks base method.
func (m *MockFileWrapper) WriteStringToFile(arg0, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WriteStringToFile", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// WriteStringToFile indicates an expected call of WriteStringToFile.
func (mr *MockFileWrapperMockRecorder) WriteStringToFile(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteStringToFile", reflect.TypeOf((*MockFileWrapper)(nil).WriteStringToFile), arg0, arg1)
}
