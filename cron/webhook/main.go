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

// Package main implements a Webhook which runs a bash script when called.
// This is called on every successful completion of BQ transfer.
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/v1/google"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/ossf/scorecard/v2/cron/data"
)

const stableTag = "stable"

var images = []string{
	"gcr.io/openssf/scorecard",
	"gcr.io/openssf/scorecard-batch-controller",
	"gcr.io/openssf/scorecard-batch-worker",
	"gcr.io/openssf/scorecard-bq-transfer",
}

func scriptHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		jsonBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("unable to read request body: %v", err),
				http.StatusInternalServerError)
			return
		}
		var metadata data.ShardMetadata
		if err := protojson.Unmarshal(jsonBytes, &metadata); err != nil {
			http.Error(w, fmt.Sprintf("unable to parse request as ShardMetadata proto: %v", err),
				http.StatusBadRequest)
			return
		}
		authn, err := google.NewEnvAuthenticator()
		if err != nil {
			http.Error(w, fmt.Sprintf("error in NewEnvAuthenticator: %v", err), http.StatusInternalServerError)
			return
		}
		for _, image := range images {
			if err := crane.Tag(
				fmt.Sprintf("%s:%s", image, metadata.GetCommitSha()), stableTag,
				crane.WithAuth(authn)); err != nil {
				http.Error(w, fmt.Sprintf("crane.Tag: %v", err), http.StatusInternalServerError)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("successfully tagged images\n")); err != nil {
			log.Panic(fmt.Errorf("error during Write: %w", err))
		}
	default:
		http.Error(w, "only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}
}

func main() {
	http.HandleFunc("/", scriptHandler)
	fmt.Printf("Starting HTTP server on port 8080 ...\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
