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

// Package monitoring defines exporters to be used by opencensus package in the cron job.
package monitoring

import (
	"errors"
	"fmt"
	"time"

	mexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	"github.com/ossf/scorecard/v4/cron/config"
)

var errorUndefinedExporter = errors.New("unsupported exporterType")

type exporterType string

const (
	stackdriverTimeSeriesQuota              = 200
	stackdriverTimeoutMinutes               = 10
	stackDriver                exporterType = "stackdriver"
	printer                    exporterType = "printer"
)

// GetExporter defines a factory for returning opencensus Exporter.
// Ensure config.ReadConfig() is called at some point before this function.
func GetExporter() (metric.Exporter, error) {
	exporter, err := config.GetMetricExporter()
	if err != nil {
		return nil, fmt.Errorf("error during GetMetricExporter: %w", err)
	}

	fmt.Printf("Using %s exporter for metrics\n", exporter)

	switch exporterType(exporter) {
	case stackDriver:
		return newStackDriverExporter()
	case printer:
		return newStdoutExporter()
	default:
		return nil, fmt.Errorf("%w: %s", errorUndefinedExporter, exporter)
	}
}

func newStdoutExporter() (metric.Exporter, error) {
	exp, err := stdoutmetric.New()
	if err != nil {
		return nil, fmt.Errorf("couldn't get stdoutexporter: %w", err)
	}

	return exp, nil
}

func newStackDriverExporter() (metric.Exporter, error) {
	exp, err := mexporter.New()
	if err != nil {
		return nil, fmt.Errorf("couldn't get cloudmetrics exporter: %w", err)
	}

	return exp, nil
}

func NewMeterProvider(exp metric.Exporter) (*metric.MeterProvider, error) {
	res, err := resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName("ossf-scorecard"),
			semconv.ServiceVersion("0.1.0"),
		))
	if err != nil {
		return nil, fmt.Errorf("NewMeterProvider: %w", err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(exp,
			metric.WithInterval(1*time.Minute))),
	)
	return meterProvider, nil
}
