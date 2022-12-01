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

package monitoring

import (
	"log"
	"time"

	"go.opencensus.io/stats/view"
)

type printerExporter struct{}

func (pe *printerExporter) ExportView(viewData *view.Data) {
	log.Printf("Printing viewData for view: %s\n", viewData.View.Name)
	log.Printf("StartTime: %s, EndTime: %s", viewData.Start.Format(time.UnixDate), viewData.End.Format(time.UnixDate))
	for i, row := range viewData.Rows {
		log.Printf("Row: %d\n%v\n", i, row)
	}
}

func (pe *printerExporter) StartMetricsExporter() error {
	return nil
}

func (pe *printerExporter) StopMetricsExporter() {}

func (pe *printerExporter) Flush() {}
