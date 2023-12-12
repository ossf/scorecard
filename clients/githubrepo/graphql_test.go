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
	"context"
	"net/http"
	"testing"

	"github.com/shurcooL/githubv4"
)

type badGatewayRoundTripper struct {
	requestCounter *int
}

func (b badGatewayRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	*b.requestCounter += 1
	return &http.Response{
		StatusCode: http.StatusBadGateway,
		Status:     "502 Bad Gateway",
	}, nil
}

func Test_getCommits_retry(t *testing.T) {
	t.Parallel()
	var nRequests int
	rt := badGatewayRoundTripper{requestCounter: &nRequests}
	handler := graphqlHandler{
		client: githubv4.NewClient(&http.Client{
			Transport: rt,
		}),
	}
	handler.init(context.Background(), &repoURL{}, 1)
	_, err := handler.getCommits()
	if err == nil {
		t.Error("expected error")
	}
	want := retryLimit + 1 // 1 represents the initial request
	if *rt.requestCounter != want {
		t.Errorf("wanted %d retries, got %d", want, *rt.requestCounter)
	}
}
