// Copyright 2026 OpenSSF Scorecard Authors
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

package cdn

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

const token = "test-token"

func TestFastlyClient_Purge(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PURGE" {
			t.Errorf("expected method PURGE, got %s", r.Method)
		}
		if r.Header.Get("Fastly-Key") != token {
			t.Errorf("expected Fastly-Key header %s, got %s", token, r.Header.Get("Fastly-Key"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	client := NewFastlyClient(token, server.URL)
	if err := client.Purge(context.Background(), "/foo"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestFastlyClient_Purge_Error(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	client := NewFastlyClient(token, server.URL)
	if err := client.Purge(context.Background(), "/foo"); err == nil {
		t.Error("expected error, got nil")
	}
}
