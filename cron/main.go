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

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
	"time"

	"github.com/jszwec/csvutil"
)

type Repository struct {
	Repo     string `csv:"repo"`
	Metadata string `csv:"metadata,omitempty"`
}

func main() {
	fileName := fmt.Sprintf("%02d-%02d-%d.json", time.Now().Month(), time.Now().Day(), time.Now().Year())
	result, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer result.Close()
	projects, err := os.OpenFile(os.Args[1], os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer projects.Close()

	repos := []Repository{}
	data, err := ioutil.ReadAll(projects)
	if err != nil {
		panic(err)
	}
	err = csvutil.Unmarshal(data, &repos)
	if err != nil {
		panic(err)
	}
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].Repo < repos[j].Repo
	})

	//nolint
	const checks string = "--checks=Active,CI-Tests,CII-Best-Practices,Code-Review,Contributors,Frozen-Deps,Fuzzing,Packaging,Pull-Requests,SAST,Security-Policy,Signed-Releases,Signed-Tags"

	if _, err = result.Write([]byte(`{"results":[`)); err != nil {
		panic(err)
	}
	for i, r := range repos {
		fmt.Println(r.Repo)
		//nolint
		cmd := exec.Command("../scorecard", fmt.Sprintf("--repo=%s", r.Repo), fmt.Sprintf("--metadata=%s", r.Metadata), checks, "--show-details", "--format=json")
		// passing the external github token the cmd
		cmd.Env = append(cmd.Env, fmt.Sprintf("GITHUB_AUTH_TOKEN=%s", os.Getenv("GITHUB_AUTH_TOKEN")))
		cmd.Stderr = io.Writer(os.Stderr)

		// Execute the command
		data, err := cmd.Output()
		if err != nil {
			fmt.Printf("error:failed for the repo %s and the error is %s", r.Repo, err.Error())
			// continuing because this is just for that repo that failed.
			continue
		}

		_, err = result.WriteString(string(data))
		if err != nil {
			panic(err)
		}
		if i < len(repos)-1 {
			_, err = result.WriteString(",")
			if err != nil {
				panic(err)
			}
		}
	}
	if _, err := result.WriteString("\n]}\n"); err != nil {
		panic(err)
	}

	// copying the file to the GCS bucket
	if err := exec.Command("gsutil", "cp", fileName, fmt.Sprintf("gs://%s", os.Getenv("GCS_BUCKET"))).Run(); err != nil {
		panic(err)
	}
	//copying the results to the latest.json
	//nolint
	if err := exec.Command("gsutil", "cp", fmt.Sprintf("gs://%s/%s", os.Getenv("GCS_BUCKET"), fileName),
		fmt.Sprintf("gs://%s/latest.json", os.Getenv("GCS_BUCKET"))).Run(); err != nil {
		panic(err)
	}

	fmt.Println("Finished")
}
