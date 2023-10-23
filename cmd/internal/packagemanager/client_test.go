// Copyright 2020 OpenSSF Scorecard Authors
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

// Package packagemanager implements a packagemanager client
package packagemanager

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_GetURI_calls_client_get_with_input(t *testing.T) {
	t.Parallel()
	type args struct {
		inputURL string
	}
	tests := []struct {
		name         string
		args         args
		wantURL      string
		wantResponse string
	}{
		{
			name: "GetURI_input_is_the_same_as_get_uri",

			args: args{
				inputURL: "test",
			},
			wantURL:      "/test",
			wantResponse: "test",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.wantURL {
					t.Errorf("Expected to request '%s', got: %s", tt.wantURL, r.URL.Path)
				}
				// nolint
				w.WriteHeader(http.StatusOK)
				// nolint
				w.Write([]byte(tt.wantResponse))
			}))
			defer server.Close()
			client := PackageManagerClient{}
			got, err := client.GetURI(server.URL + "/" + tt.args.inputURL)
			if err != nil {
				t.Errorf("Test_GetURI_calls_client_get_with_input() error in Get= %v", err)
				return
			}
			defer got.Body.Close()
			body, err := io.ReadAll(got.Body)
			if err != nil {
				t.Errorf("Test_GetURI_calls_client_get_with_input() error in ReadAll= %v", err)
				return
			}
			if string(body) != tt.wantResponse {
				t.Errorf("GetURI() = %v, want %v", got, tt.wantResponse)
			}
		})
	}
}

func Test_Get_calls_client_get_with_input(t *testing.T) {
	t.Parallel()
	type args struct {
		inputURL    string
		packageName string
	}
	tests := []struct {
		name         string
		args         args
		wantURL      string
		wantResponse string
	}{
		{
			name: "Get_input_adds_package_name_for_get_uri",

			args: args{
				inputURL:    "test-%s-test",
				packageName: "test_package",
			},
			wantURL:      "/test-test_package-test",
			wantResponse: "test",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.wantURL {
					t.Errorf("Expected to request '%s', got: %s", tt.wantURL, r.URL.Path)
				}
				// nolint
				w.WriteHeader(http.StatusOK)
				// nolint
				w.Write([]byte(tt.wantResponse))
			}))
			defer server.Close()
			client := PackageManagerClient{}
			got, err := client.Get(server.URL+"/"+tt.args.inputURL, tt.args.packageName)
			if err != nil {
				t.Errorf("Test_Get_calls_client_get_with_input() error in Get = %v", err)
				return
			}
			defer got.Body.Close()
			body, err := io.ReadAll(got.Body)
			if err != nil {
				t.Errorf("Test_Get_calls_client_get_with_input() error in ReadAll = %v", err)
				return
			}
			if string(body) != tt.wantResponse {
				t.Errorf("GetURI() = %v, want %v", got, tt.wantResponse)
			}
		})
	}
}
