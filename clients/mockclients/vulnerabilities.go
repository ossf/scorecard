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
// Source: clients/vulnerabilities.go
//
// Generated by this command:
//
//	mockgen -source=clients/vulnerabilities.go -destination=clients/mockclients/vulnerabilities.go -package=mockrepo -copyright_file=clients/mockclients/license.txt
//

// Package mockrepo is a generated GoMock package.
package mockrepo

import (
	context "context"
	reflect "reflect"

	clients "github.com/ossf/scorecard/v5/clients"
	gomock "go.uber.org/mock/gomock"
)

// MockVulnerabilitiesClient is a mock of VulnerabilitiesClient interface.
type MockVulnerabilitiesClient struct {
	ctrl     *gomock.Controller
	recorder *MockVulnerabilitiesClientMockRecorder
	isgomock struct{}
}

// MockVulnerabilitiesClientMockRecorder is the mock recorder for MockVulnerabilitiesClient.
type MockVulnerabilitiesClientMockRecorder struct {
	mock *MockVulnerabilitiesClient
}

// NewMockVulnerabilitiesClient creates a new mock instance.
func NewMockVulnerabilitiesClient(ctrl *gomock.Controller) *MockVulnerabilitiesClient {
	mock := &MockVulnerabilitiesClient{ctrl: ctrl}
	mock.recorder = &MockVulnerabilitiesClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockVulnerabilitiesClient) EXPECT() *MockVulnerabilitiesClientMockRecorder {
	return m.recorder
}

// ListUnfixedVulnerabilities mocks base method.
func (m *MockVulnerabilitiesClient) ListUnfixedVulnerabilities(context context.Context, commit, localDir string) (clients.VulnerabilitiesResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListUnfixedVulnerabilities", context, commit, localDir)
	ret0, _ := ret[0].(clients.VulnerabilitiesResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListUnfixedVulnerabilities indicates an expected call of ListUnfixedVulnerabilities.
func (mr *MockVulnerabilitiesClientMockRecorder) ListUnfixedVulnerabilities(context, commit, localDir any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListUnfixedVulnerabilities", reflect.TypeOf((*MockVulnerabilitiesClient)(nil).ListUnfixedVulnerabilities), context, commit, localDir)
}
