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

package data

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"

	"github.com/jszwec/csvutil"

	"github.com/ossf/scorecard/v4/clients/githubrepo"
)

// Iterator interface is used to iterate through list of input repos for the cron job.
type Iterator interface {
	HasNext() bool
	Next() (RepoFormat, error)
}

// MakeIteratorFrom returns an implementation of Iterator interface.
// Currently returns an instance of csvIterator.
func MakeIteratorFrom(reader io.Reader) (Iterator, error) {
	csvReader := csv.NewReader(reader)
	csvReader.Comment = '#'
	dec, err := csvutil.NewDecoder(csvReader)
	if err != nil {
		return nil, fmt.Errorf("error in csvutil.NewDecoder: %w", err)
	}
	return &csvIterator{decoder: dec}, nil
}

type csvIterator struct {
	decoder *csvutil.Decoder
	err     error
	next    RepoFormat
}

func (reader *csvIterator) HasNext() bool {
	reader.err = reader.decoder.Decode(&reader.next)
	return !errors.Is(reader.err, io.EOF)
}

func (reader *csvIterator) Next() (RepoFormat, error) {
	if reader.err != nil {
		return reader.next, fmt.Errorf("reader has error: %w", reader.err)
	}
	// Sanity check valid GitHub URL.
	if _, err := githubrepo.MakeGithubRepo(reader.next.Repo); err != nil {
		return reader.next, fmt.Errorf("invalid GitHub URL: %w", err)
	}
	return reader.next, nil
}
