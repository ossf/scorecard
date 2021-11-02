// Copyright 2015 Google LLC
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

package bigquery

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"cloud.google.com/go/internal"
	"cloud.google.com/go/internal/detect"
	"cloud.google.com/go/internal/version"
	gax "github.com/googleapis/gax-go/v2"
	bq "google.golang.org/api/bigquery/v2"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

const (
	// Scope is the Oauth2 scope for the service.
	// For relevant BigQuery scopes, see:
	// https://developers.google.com/identity/protocols/googlescopes#bigqueryv2
	Scope           = "https://www.googleapis.com/auth/bigquery"
	userAgentPrefix = "gcloud-golang-bigquery"
)

var xGoogHeader = fmt.Sprintf("gl-go/%s gccl/%s", version.Go(), version.Repo)

func setClientHeader(headers http.Header) {
	headers.Set("x-goog-api-client", xGoogHeader)
}

// Client may be used to perform BigQuery operations.
type Client struct {
	// Location, if set, will be used as the default location for all subsequent
	// dataset creation and job operations. A location specified directly in one of
	// those operations will override this value.
	Location string

	projectID string
	bqs       *bq.Service
}

// DetectProjectID is a sentinel value that instructs NewClient to detect the
// project ID. It is given in place of the projectID argument. NewClient will
// use the project ID from the given credentials or the default credentials
// (https://developers.google.com/accounts/docs/application-default-credentials)
// if no credentials were provided. When providing credentials, not all
// options will allow NewClient to extract the project ID. Specifically a JWT
// does not have the project ID encoded.
const DetectProjectID = "*detect-project-id*"

// NewClient constructs a new Client which can perform BigQuery operations.
// Operations performed via the client are billed to the specified GCP project.
//
// If the project ID is set to DetectProjectID, NewClient will attempt to detect
// the project ID from credentials.
func NewClient(ctx context.Context, projectID string, opts ...option.ClientOption) (*Client, error) {
	o := []option.ClientOption{
		option.WithScopes(Scope),
		option.WithUserAgent(fmt.Sprintf("%s/%s", userAgentPrefix, version.Repo)),
	}
	o = append(o, opts...)
	bqs, err := bq.NewService(ctx, o...)
	if err != nil {
		return nil, fmt.Errorf("bigquery: constructing client: %v", err)
	}

	// Handle project autodetection.
	projectID, err = detect.ProjectID(ctx, projectID, "", opts...)
	if err != nil {
		return nil, err
	}

	c := &Client{
		projectID: projectID,
		bqs:       bqs,
	}
	return c, nil
}

// Project returns the project ID or number for this instance of the client, which may have
// either been explicitly specified or autodetected.
func (c *Client) Project() string {
	return c.projectID
}

// Close closes any resources held by the client.
// Close should be called when the client is no longer needed.
// It need not be called at program exit.
func (c *Client) Close() error {
	return nil
}

// Calls the Jobs.Insert RPC and returns a Job.
func (c *Client) insertJob(ctx context.Context, job *bq.Job, media io.Reader) (*Job, error) {
	call := c.bqs.Jobs.Insert(c.projectID, job).Context(ctx)
	setClientHeader(call.Header())
	if media != nil {
		call.Media(media)
	}
	var res *bq.Job
	var err error
	invoke := func() error {
		res, err = call.Do()
		return err
	}
	// A job with a client-generated ID can be retried; the presence of the
	// ID makes the insert operation idempotent.
	// We don't retry if there is media, because it is an io.Reader. We'd
	// have to read the contents and keep it in memory, and that could be expensive.
	// TODO(jba): Look into retrying if media != nil.
	if job.JobReference != nil && media == nil {
		err = runWithRetry(ctx, invoke)
	} else {
		err = invoke()
	}
	if err != nil {
		return nil, err
	}
	return bqToJob(res, c)
}

// runQuery invokes the optimized query path.
// Due to differences in options it supports, it cannot be used for all existing
// jobs.insert requests that are query jobs.
func (c *Client) runQuery(ctx context.Context, queryRequest *bq.QueryRequest) (*bq.QueryResponse, error) {
	call := c.bqs.Jobs.Query(c.projectID, queryRequest)
	setClientHeader(call.Header())

	var res *bq.QueryResponse
	var err error
	invoke := func() error {
		res, err = call.Do()
		return err
	}

	// We control request ID, so we can always runWithRetry.
	err = runWithRetry(ctx, invoke)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Convert a number of milliseconds since the Unix epoch to a time.Time.
// Treat an input of zero specially: convert it to the zero time,
// rather than the start of the epoch.
func unixMillisToTime(m int64) time.Time {
	if m == 0 {
		return time.Time{}
	}
	return time.Unix(0, m*1e6)
}

// runWithRetry calls the function until it returns nil or a non-retryable error, or
// the context is done.
// See the similar function in ../storage/invoke.go. The main difference is the
// reason for retrying.
func runWithRetry(ctx context.Context, call func() error) error {
	// These parameters match the suggestions in https://cloud.google.com/bigquery/sla.
	backoff := gax.Backoff{
		Initial:    1 * time.Second,
		Max:        32 * time.Second,
		Multiplier: 2,
	}
	return internal.Retry(ctx, backoff, func() (stop bool, err error) {
		err = call()
		if err == nil {
			return true, nil
		}
		return !retryableError(err), err
	})
}

// This is the correct definition of retryable according to the BigQuery team. It
// also considers 502 ("Bad Gateway") and 503 ("Service Unavailable") errors
// retryable; these are returned by systems between the client and the BigQuery
// service.
func retryableError(err error) bool {
	if err == nil {
		return false
	}
	if err == io.ErrUnexpectedEOF {
		return true
	}
	// Special case due to http2: https://github.com/googleapis/google-cloud-go/issues/1793
	// Due to Go's default being higher for streams-per-connection than is accepted by the
	// BQ backend, it's possible to get streams refused immediately after a connection is
	// started but before we receive SETTINGS frame from the backend.  This generally only
	// happens when we try to enqueue > 100 requests onto a newly initiated connection.
	if err.Error() == "http2: stream closed" {
		return true
	}

	switch e := err.(type) {
	case *googleapi.Error:
		// We received a structured error from backend.
		var reason string
		if len(e.Errors) > 0 {
			reason = e.Errors[0].Reason
		}
		if e.Code == http.StatusServiceUnavailable || e.Code == http.StatusBadGateway || reason == "backendError" || reason == "rateLimitExceeded" {
			return true
		}
	case *url.Error:
		retryable := []string{"connection refused", "connection reset"}
		for _, s := range retryable {
			if strings.Contains(e.Error(), s) {
				return true
			}
		}
	case interface{ Temporary() bool }:
		if e.Temporary() {
			return true
		}
	}
	// Unwrap is only supported in go1.13.x+
	if e, ok := err.(interface{ Unwrap() error }); ok {
		return retryableError(e.Unwrap())
	}
	return false
}
