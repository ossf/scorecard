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

// Package mockrepo is a generated GoMock package.
package mockrepo

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"

	clients "github.com/ossf/scorecard/v5/clients"
)

// MockCIIBestPracticesClient is a mock of CIIBestPracticesClient interface.
type MockCIIBestPracticesClient struct {
	ctrl     *gomock.Controller
	recorder *MockCIIBestPracticesClientMockRecorder
}

// MockCIIBestPracticesClientMockRecorder is the mock recorder for MockCIIBestPracticesClient.
type MockCIIBestPracticesClientMockRecorder struct {
	mock *MockCIIBestPracticesClient
}

// NewMockCIIBestPracticesClient creates a new mock instance.
func NewMockCIIBestPracticesClient(ctrl *gomock.Controller) *MockCIIBestPracticesClient {
	mock := &MockCIIBestPracticesClient{ctrl: ctrl}
	mock.recorder = &MockCIIBestPracticesClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCIIBestPracticesClient) EXPECT() *MockCIIBestPracticesClientMockRecorder {
	return m.recorder
}

// GetBadgeLevel mocks base method.
func (m *MockCIIBestPracticesClient) GetBadgeLevel(ctx context.Context, uri string) (clients.BadgeLevel, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBadgeLevel", ctx, uri)
	ret0, ok := ret[0].(clients.BadgeLevel)
	if !ok {
		ret0 = clients.Unknown
	}
	ret1, ok := ret[1].(error)
	if !ok {
		ret1 = nil
	}
	return ret0, ret1
}

// GetBadgeLevel indicates an expected call of GetBadgeLevel.
func (mr *MockCIIBestPracticesClientMockRecorder) GetBadgeLevel(ctx, uri any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(
		mr.mock,
		"GetBadgeLevel",
		reflect.TypeOf((*MockCIIBestPracticesClient)(nil).GetBadgeLevel),
		ctx,
		uri,
	)
}
