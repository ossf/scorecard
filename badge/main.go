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
// WITHOUT WARRANTIES OR CONDITIONS OF ANY healthz, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	e "errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/narqo/go-badge"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type dailyscore struct {
	Results []struct {
		Repo   string `json:"Repo"`
		Date   string `json:"Date"`
		Checks []struct {
			Checkname  string   `json:"CheckName"`
			Pass       bool     `json:"Pass"`
			Confidence int      `json:"Confidence"`
			Details    []string `json:"Details"`
		} `json:"Checks"`
		Metadata []interface{} `json:"MetaData"`
	} `json:"results"`
}

var (
	logf        func(s string, f ...interface{})
	ErrNotFound error = errors.New("item not found in the scorecard results")
)

// healthz is a health check function.
func healthz(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// liveness and readiness probe.
		w.WriteHeader(http.StatusOK)
		return

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// calculateScore calcualtes the score based on scorecard results that were run by dailyscore.
func calculateScore(host, owner, repo string) (int, error) {
	const checks int = 13
	const hundred int = 100
	// the dailyscore results are stored in GCS bucket.
	//nolint
	resp, err := http.Get("https://storage.googleapis.com/ossf-scorecards/latest.json")
	scores := &dailyscore{}
	if err != nil {
		return 0, errors.Wrap(err, "unable to http request")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, errors.Wrap(err, "unable to read http response")
	}
	defer resp.Body.Close()
	if err := json.Unmarshal(body, &scores); err != nil {
		return 0, errors.Wrap(err, "unable to unmarshal the json response")
	}

	for _, item := range scores.Results {
		if item.Repo == fmt.Sprintf("%s/%s/%s", host, owner, repo) {
			// simple check for now to get the result. The score calculation will be based on tiers
			// which still needs to be worked on.
			pass := 0
			for _, c := range item.Checks {
				if c.Pass {
					pass++
				}
			}
			return (hundred * pass) / checks, nil
		}
	}
	return 0, errors.Wrap(ErrNotFound, "item not found")
}

func getBadge(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		m := mux.Vars(r)
		w.Header().Set("Content-Disposition", "attachment; filename=badge.svg")
		w.Header().Set("Content-Type", r.Header.Get("image/svg+xml"))
		result, err := calculateScore(m["host"], m["owner"], m["repo"])
		if err != nil && e.Is(err, ErrNotFound) {
			logf("item not found for badge %s", err.Error())
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			logf("error in handling request for badge %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = badge.Render("scorecard", strconv.Itoa(result)+"%", badge.ColorBlue, w)
		if err != nil {
			logf("error in handling request for badge %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
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
	logf = sugar.Infof

	r := mux.NewRouter()
	r.HandleFunc("/{host}/{owner}/{repo}", getBadge)
	r.HandleFunc("/healthz/", healthz)

	sugar.Info("Starting server for testing ...\n")

	srv := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:8000",
	}

	sugar.Fatal(srv.ListenAndServe())
}
