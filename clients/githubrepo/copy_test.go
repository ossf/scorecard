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
package githubrepo

import (
	"reflect"
	"testing"
	"time"

	"github.com/ossf/scorecard/v4/clients"
)

func TestCopyBoolPtr(t *testing.T) {
	t.Parallel()
	tests := []struct { //nolint:govet
		name string
		src  *bool
		dest **bool
		want *bool
	}{
		{
			name: "nil_src",
			src:  nil,
			dest: new(*bool),
			want: nil,
		},
		{
			name: "non_nil_src_true",
			src:  BoolPtr(true),
			dest: new(*bool),
			want: BoolPtr(true),
		},
		{
			name: "non_nil_src_false",
			src:  BoolPtr(false),
			dest: new(*bool),
			want: BoolPtr(false),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			copyBoolPtr(tt.src, tt.dest)
			if (tt.want == nil && *tt.dest != nil) || (tt.want != nil && *tt.dest == nil) || (tt.want != nil && *tt.dest != nil && **tt.dest != *tt.want) { //nolint:lll
				t.Errorf("copyBoolPtr() got = %v, want %v", *tt.dest, tt.want)
			}
		})
	}
}

// BoolPtr is a utility function to get a pointer to a boolean value.
func BoolPtr(value bool) *bool {
	return &value
}

func TestCopyInt32Ptr(t *testing.T) {
	t.Parallel()
	tests := []struct { //nolint:govet
		name string
		src  *int32
		dest **int32
		want *int32
	}{
		{
			name: "nil_src",
			src:  nil,
			dest: new(*int32),
			want: nil,
		},
		{
			name: "non_nil_src_positive",
			src:  Int32Ptr(123),
			dest: new(*int32),
			want: Int32Ptr(123),
		},
		{
			name: "non_nil_src_negative",
			src:  Int32Ptr(-456),
			dest: new(*int32),
			want: Int32Ptr(-456),
		},
		{
			name: "non_nil_src_zero",
			src:  Int32Ptr(0),
			dest: new(*int32),
			want: Int32Ptr(0),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			copyInt32Ptr(tt.src, tt.dest)
			if (tt.want == nil && *tt.dest != nil) || (tt.want != nil && *tt.dest == nil) || (tt.want != nil && *tt.dest != nil && **tt.dest != *tt.want) { //nolint:lll
				t.Errorf("copyInt32Ptr() got = %v, want %v", *tt.dest, tt.want)
			}
		})
	}
}

// Int32Ptr is a utility function to get a pointer to an int32 value.
func Int32Ptr(value int32) *int32 {
	return &value
}

func TestCopyStringPtr(t *testing.T) {
	t.Parallel()
	tests := []struct { //nolint:govet
		name string
		src  *string
		dest **string
		want *string
	}{
		{
			name: "nil_src",
			src:  nil,
			dest: new(*string),
			want: nil,
		},
		{
			name: "non_nil_src_empty_string",
			src:  StringPtr(""),
			dest: new(*string),
			want: StringPtr(""),
		},
		{
			name: "non_nil_src_hello",
			src:  StringPtr("hello"),
			dest: new(*string),
			want: StringPtr("hello"),
		},
		{
			name: "non_nil_src_world",
			src:  StringPtr("world"),
			dest: new(*string),
			want: StringPtr("world"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			copyStringPtr(tt.src, tt.dest)
			if (tt.want == nil && *tt.dest != nil) || (tt.want != nil && *tt.dest == nil) || (tt.want != nil && *tt.dest != nil && **tt.dest != *tt.want) { //nolint:lll
				t.Errorf("copyStringPtr() got = %v, want %v", *tt.dest, tt.want)
			}
		})
	}
}

// StringPtr is a utility function to get a pointer to a string value.
func StringPtr(value string) *string {
	return &value
}

func TestCopyTimePtr(t *testing.T) {
	t.Parallel()
	// Define some example time values for testing
	time1 := time.Date(2023, 4, 16, 12, 0, 0, 0, time.UTC)
	time2 := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct { //nolint:govet
		name string
		src  *time.Time
		dest **time.Time
		want *time.Time
	}{
		{
			name: "nil_src",
			src:  nil,
			dest: new(*time.Time),
			want: nil,
		},
		{
			name: "non_nil_src_time1",
			src:  &time1,
			dest: new(*time.Time),
			want: &time1,
		},
		{
			name: "non_nil_src_time2",
			src:  &time2,
			dest: new(*time.Time),
			want: &time2,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			copyTimePtr(tt.src, tt.dest)
			if (tt.want == nil && *tt.dest != nil) || (tt.want != nil && *tt.dest == nil) || (tt.want != nil && *tt.dest != nil && !(*tt.dest).Equal(*tt.want)) { //nolint:lll
				t.Errorf("copyTimePtr() got = %v, want %v", *tt.dest, tt.want)
			}
		})
	}
}

func TestCopyStringSlice(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		src  []string
		dest []string
		want []string
	}{
		{
			name: "empty_src",
			src:  []string{},
			dest: nil,
			want: []string{},
		},
		{
			name: "single_element_src",
			src:  []string{"hello"},
			dest: nil,
			want: []string{"hello"},
		},
		{
			name: "multiple_elements_src",
			src:  []string{"hello", "world", "foo", "bar"},
			dest: nil,
			want: []string{"hello", "world", "foo", "bar"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			copyStringSlice(tt.src, &tt.dest)
			if !reflect.DeepEqual(tt.dest, tt.want) {
				t.Errorf("copyStringSlice() got = %v, want %v", tt.dest, tt.want)
			}
		})
	}
}

func TestCopyRepoAssociationPtr(t *testing.T) {
	t.Parallel()
	var repoAssoc1 clients.RepoAssociation

	var repoAssoc2 clients.RepoAssociation

	tests := []struct { //nolint:govet
		name string
		src  *clients.RepoAssociation
		dest **clients.RepoAssociation
		want *clients.RepoAssociation
	}{
		{
			name: "nil_src",
			src:  nil,
			dest: new(*clients.RepoAssociation),
			want: nil,
		},
		{
			name: "non_nil_src_repoAssoc1",
			src:  &repoAssoc1,
			dest: new(*clients.RepoAssociation),
			want: &repoAssoc1,
		},
		{
			name: "non_nil_src_repoAssoc2",
			src:  &repoAssoc2,
			dest: new(*clients.RepoAssociation),
			want: &repoAssoc2,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			copyRepoAssociationPtr(tt.src, tt.dest)
			if (tt.want == nil && *tt.dest != nil) || (tt.want != nil && *tt.dest == nil) || (tt.want != nil && *tt.dest != nil && !reflect.DeepEqual(**tt.dest, *tt.want)) { //nolint:lll
				t.Errorf("copyRepoAssociationPtr() got = %v, want %v", *tt.dest, tt.want)
			}
		})
	}
}
