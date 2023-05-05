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

package data

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"reflect"

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
	header, err := csvutil.Header(RepoFormat{}, "csv")
	if err != nil {
		return nil, fmt.Errorf("error in csvutil.Header: %w", err)
	}
	dec, err := csvutil.NewDecoder(csvReader, header...)
	if err != nil {
		return nil, fmt.Errorf("error in csvutil.NewDecoder: %w", err)
	}
	return &csvIterator{decoder: dec}, nil
}

type csvIterator struct {
	decoder     *csvutil.Decoder
	err         error
	next        RepoFormat
	afterHeader bool
}

// returns true on the first call if the most recently decoded record is a header.
// always returns false on subsequent calls, as this is only intended to evaluate the first line.
func (reader *csvIterator) isHeader() bool {
	if reader.afterHeader {
		return false
	}
	header, err := csvutil.Header(RepoFormat{}, "csv")
	if err != nil {
		reader.err = err
		return false
	}
	lastRead := reader.decoder.Record()
	reader.afterHeader = true
	return reflect.DeepEqual(header, lastRead)
}

func (reader *csvIterator) HasNext() bool {
	reader.err = reader.decoder.Decode(&reader.next)
	if reader.isHeader() {
		reader.err = reader.decoder.Decode(&reader.next)
	}
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

func MakeNestedIterator(iterators []Iterator) (Iterator, error) {
	return &nestedIterator{iterators: iterators}, nil
}

type nestedIterator struct {
	iterators []Iterator
	current   int
}

func (i *nestedIterator) HasNext() bool {
	for i.current < len(i.iterators) {
		if i.iterators[i.current].HasNext() {
			return true
		}
		i.current++
	}
	return false
}

func (i *nestedIterator) Next() (RepoFormat, error) {
	r, err := i.iterators[i.current].Next()
	if err != nil {
		return RepoFormat{}, fmt.Errorf("nestedIterator.Next(): %w", err)
	}
	return r, nil
}
