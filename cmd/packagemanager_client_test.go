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

// Package cmd implements Scorecard commandline.
package cmd

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_GetURI_calls_client_get_with_input(t *testing.T) {
	t.Parallel()
	type args struct {
		input_url string
	}
	tests := []struct {
		name          string
		args          args
		want_uri      string
		want_response string
	}{
		{
			name: "GetURI_input_is_the_same_as_get_uri",

			args: args{
				input_url: "test",
			},
			want_uri:      "/test",
			want_response: "test",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.want_uri {
					t.Errorf("Expected to request '%s', got: %s", tt.want_uri, r.URL.Path)
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.want_response))
			}))
			defer server.Close()
			client := packageManager{}
			got, err := client.GetURI(server.URL + "/" + tt.args.input_url)
			if err != nil {
				log.Fatalln(err)
			}
			body, err := io.ReadAll(got.Body)
			if err != nil {
				log.Fatalln(err)
			}
			if string(body) != tt.want_response {
				t.Errorf("GetURI() = %v, want %v", got, tt.want_response)
			}
		})
	}
}

func Test_Get_calls_client_get_with_input(t *testing.T) {
	t.Parallel()
	type args struct {
		input_url    string
		package_name string
	}
	tests := []struct {
		name          string
		args          args
		want_uri      string
		want_response string
	}{
		{
			name: "Get_input_adds_package_name_for_get_uri",

			args: args{
				input_url:    "test-%s-test",
				package_name: "test_package",
			},
			want_uri:      "/test-test_package-test",
			want_response: "test",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.want_uri {
					t.Errorf("Expected to request '%s', got: %s", tt.want_uri, r.URL.Path)
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.want_response))
			}))
			defer server.Close()
			client := packageManager{}
			got, err := client.Get(server.URL+"/"+tt.args.input_url, tt.args.package_name)
			if err != nil {
				log.Fatalln(err)
			}
			body, err := io.ReadAll(got.Body)
			if err != nil {
				log.Fatalln(err)
			}
			if string(body) != tt.want_response {
				t.Errorf("GetURI() = %v, want %v", got, tt.want_response)
			}
		})
	}
}
