// Copyright 2023 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package tokens

import "testing"

// mockTokenAccessor implements TokenAccessor for testing.
type mockTokenAccessor struct {
	tokens  []string
	counter uint64
}

// Next implements TokenAccessor.Next.
func (m *mockTokenAccessor) Next() (uint64, string) {
	if len(m.tokens) == 0 {
		return 0, ""
	}
	token := m.tokens[0]
	m.tokens = m.tokens[1:]
	m.counter++
	return m.counter, token
}

// Release implements TokenAccessor.Release.
func (m *mockTokenAccessor) Release(id uint64) {
	// No-op for mock.
}

// NewMockTokenAccessor creates a new mockTokenAccessor.
func newMockTokenAccessor(tokens []string) *mockTokenAccessor {
	return &mockTokenAccessor{
		tokens: tokens,
	}
}

func TestTokenOverRPC_Next(t *testing.T) {
	mockClient := newMockTokenAccessor([]string{"token1", "token2", "token3"})
	rpc := NewTokenOverRPC(mockClient)
	token := &Token{}
	x := struct{}{}
	err := rpc.Next(x, token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.Value != "token1" {
		t.Fatalf("unexpected token: %s", token.Value)
	}
	err = rpc.Next(x, token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.Value != "token2" {
		t.Fatalf("unexpected token: %s", token.Value)
	}
}

func TestTokenOverRPC_Release(t *testing.T) {
	mockClient := newMockTokenAccessor([]string{"token1", "token2", "token3"})
	rpc := NewTokenOverRPC(mockClient)

	var reply struct{}
	err := rpc.Release(2, &reply)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}
