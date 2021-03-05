// Copyright 2020 Security Scorecard Authors
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

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ossf/scorecard/gitcache/pkg"
	"go.uber.org/zap"
)

type cache struct {
	URL string `json:"url"`
}

var (
	blob, tempDir string
	logf          func(s string, f ...interface{})
)

func handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		d := json.NewDecoder(r.Body)
		c := &cache{}
		if err := d.Decode(c); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		cache, err := pkg.NewCacheService(blob, tempDir, logf)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := cache.UpdateCache(c.URL); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "I can't do that.")
	}
}

func main() {
	logLevel := zap.LevelFlag("verbosity", zap.InfoLevel, "override the default log level")

	cfg := zap.NewProductionConfig()
	cfg.Level.SetLevel(*logLevel)
	logger, err := cfg.Build()
	if err != nil {
		log.Panic(err)
	}

	//nolint
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()
	// no need to lock this as it being written only within this method.
	blob = os.Getenv("BLOB_URL")
	if blob == "" {
		log.Panic("BLOB_URL env is not set.")
	}
	// tempDir is the storage space for archiving the repository.
	// not using the tempfs https://en.wikipedia.org/wiki/Tmpfs as it is in memory and some of the
	// repositories can be in large.
	tempDir = os.Getenv("TEMP_DIR")
	if tempDir == "" {
		log.Panic("TEMP_DIR env is not set.")
	}
	sugar.Info("BLOB_URL ", blob)
	// no need to lock this as it being written only within this method.
	logf = sugar.Infof

	http.HandleFunc("/", handler)
	sugar.Info("Starting server for testing HTTP POST...\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		sugar.Fatal(err)
	}
}
