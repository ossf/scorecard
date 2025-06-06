// Copyright 2021 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

// Code generated by MockGen. DO NOT EDIT.
// Source: clients/repo.go

// Package mockrepo is a generated GoMock package.
package mockrepo

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockRepo is a mock of Repo interface.
type MockRepo struct {
	ctrl     *gomock.Controller
	recorder *MockRepoMockRecorder
}

// MockRepoMockRecorder is the mock recorder for MockRepo.
type MockRepoMockRecorder struct {
	mock *MockRepo
}

// NewMockRepo creates a new mock instance.
func NewMockRepo(ctrl *gomock.Controller) *MockRepo {
	mock := &MockRepo{ctrl: ctrl}
	mock.recorder = &MockRepoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRepo) EXPECT() *MockRepoMockRecorder {
	return m.recorder
}

// AppendMetadata mocks base method.
func (m *MockRepo) AppendMetadata(metadata ...string) {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range metadata {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "AppendMetadata", varargs...)
}

// AppendMetadata indicates an expected call of AppendMetadata.
func (mr *MockRepoMockRecorder) AppendMetadata(metadata ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AppendMetadata", reflect.TypeOf((*MockRepo)(nil).AppendMetadata), metadata...)
}

// Host mocks base method.
func (m *MockRepo) Host() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Host")
	ret0, _ := ret[0].(string)
	return ret0
}

// Host indicates an expected call of Host.
func (mr *MockRepoMockRecorder) Host() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Host", reflect.TypeOf((*MockRepo)(nil).Host))
}

// IsValid mocks base method.
func (m *MockRepo) IsValid() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsValid")
	ret0, _ := ret[0].(error)
	return ret0
}

// IsValid indicates an expected call of IsValid.
func (mr *MockRepoMockRecorder) IsValid() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsValid", reflect.TypeOf((*MockRepo)(nil).IsValid))
}

// Metadata mocks base method.
func (m *MockRepo) Metadata() []string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Metadata")
	ret0, _ := ret[0].([]string)
	return ret0
}

// Metadata indicates an expected call of Metadata.
func (mr *MockRepoMockRecorder) Metadata() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Metadata", reflect.TypeOf((*MockRepo)(nil).Metadata))
}

// Path mocks base method.
func (m *MockRepo) Path() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Path")
	ret0, _ := ret[0].(string)
	return ret0
}

// Path indicates an expected call of Path.
func (mr *MockRepoMockRecorder) Path() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Path", reflect.TypeOf((*MockRepo)(nil).Path))
}

// String mocks base method.
func (m *MockRepo) String() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "String")
	ret0, _ := ret[0].(string)
	return ret0
}

// String indicates an expected call of String.
func (mr *MockRepoMockRecorder) String() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "String", reflect.TypeOf((*MockRepo)(nil).String))
}

// URI mocks base method.
func (m *MockRepo) URI() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "URI")
	ret0, _ := ret[0].(string)
	return ret0
}

// URI indicates an expected call of URI.
func (mr *MockRepoMockRecorder) URI() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "URI", reflect.TypeOf((*MockRepo)(nil).URI))
}
