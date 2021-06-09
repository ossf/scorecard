// Copyright 2021 Security Scorecard Authors
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

package monitoring

import (
	"errors"
	"fmt"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"contrib.go.opencensus.io/exporter/stackdriver/monitoredresource/gcp"
	"go.opencensus.io/stats/view"

	"github.com/ossf/scorecard/cron/config"
)

var errorUndefinedExporter = errors.New("unsupported exporterType")

type exporterType string

const (
	stackdriverTimeSeriesQuota              = 200
	stackdriverTimeoutMinutes               = 10
	stackDriver                exporterType = "stackdriver"
	printer                    exporterType = "printer"
)

type Exporter interface {
	ExportView(viewData *view.Data)
	StartMetricsExporter() error
	StopMetricsExporter()
	Flush()
}

func GetExporter() (Exporter, error) {
	exporter, err := config.GetMetricExporter()
	if err != nil {
		return nil, fmt.Errorf("error during GetMetricExporter: %w", err)
	}
	switch exporterType(exporter) {
	case stackDriver:
		return newStackDriverExporter()
	case printer:
		return new(printerExporter), nil
	default:
		return nil, fmt.Errorf("%w: %s", errorUndefinedExporter, exporter)
	}
}

func newStackDriverExporter() (*stackdriver.Exporter, error) {
	projectID, err := config.GetProjectID()
	if err != nil {
		return nil, fmt.Errorf("error getting ProjectID: %w", err)
	}
	exporter, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID:         projectID,
		MetricPrefix:      "scorecard-cron",
		MonitoredResource: gcp.Autodetect(),
		Timeout:           stackdriverTimeoutMinutes * time.Minute,
		// Stackdriver specific quotas based on https://cloud.google.com/monitoring/quotas
		// `Time series included in a request`
		BundleCountThreshold: stackdriverTimeSeriesQuota,
	})
	if err != nil {
		return nil, fmt.Errorf("error during stackdriver.NewExporter: %w", err)
	}
	return exporter, nil
}
