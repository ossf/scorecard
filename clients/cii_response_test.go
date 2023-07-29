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
package clients

import (
	"reflect"
	"testing"
)

func TestParseBadgeResponseFromJSON(t *testing.T) {
	t.Parallel()
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []BadgeResponse
		wantErr bool
	}{
		{
			name: "Test ParseBadgeResponseFromJSON",
			args: args{
				data: []byte(`[{"badge_level":"in_progress"}]`),
			},
			want: []BadgeResponse{
				{
					BadgeLevel: "in_progress",
				},
			},
		},
		{
			name: "Fail Test ParseBadgeResponseFromJSON",
			args: args{
				data: []byte(`foo`),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseBadgeResponseFromJSON(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBadgeResponseFromJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseBadgeResponseFromJSON() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBadgeResponse_AsJSON(t *testing.T) {
	type fields struct {
		BadgeLevel string
	}

	// Single test case
	tt := struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		name: "Test BadgeResponse_AsJSON",
		fields: fields{
			BadgeLevel: "in_progress",
		},
		want: []byte(`[{"badge_level":"in_progress"}]`),
	}

	t.Run(tt.name, func(t *testing.T) {
		resp := BadgeResponse{
			BadgeLevel: tt.fields.BadgeLevel,
		}
		got, err := resp.AsJSON()
		if (err != nil) != tt.wantErr {
			t.Errorf("AsJSON() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("AsJSON() got = %v, want %v", got, tt.want)
		}
	})
}

func TestBadgeResponse_getBadgeLevel(t *testing.T) {
	t.Parallel()
	type fields struct {
		BadgeLevel string
	}
	tests := []struct {
		name    string
		fields  fields
		want    BadgeLevel
		wantErr bool
	}{
		{
			name: "Test inProgress getBadgeLevel",
			fields: fields{
				BadgeLevel: "in_progress",
			},
			want: InProgress,
		},
		{
			name: "Fail Test getBadgeLevel",
			fields: fields{
				BadgeLevel: "foo",
			},
			wantErr: true,
		},
		{
			name: "Test passing getBadgeLevel",
			fields: fields{
				BadgeLevel: "passing",
			},
			want: Passing,
		},
		{
			name: "Test silver getBadgeLevel",
			fields: fields{
				BadgeLevel: "silver",
			},
			want: Silver,
		},
		{
			name: "Test gold getBadgeLevel",
			fields: fields{
				BadgeLevel: "gold",
			},
			want: Gold,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			resp := BadgeResponse{
				BadgeLevel: tt.fields.BadgeLevel,
			}
			got, err := resp.getBadgeLevel()
			if (err != nil) != tt.wantErr {
				t.Errorf("getBadgeLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getBadgeLevel() got = %v, want %v", got, tt.want)
			}
		})
	}
}
