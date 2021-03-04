package main

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

// RepoURL parses and stores URL into fields.
type RepoURL struct {
	Host  string // Host where the repo is stored. Example GitHub.com
	Owner string // Owner of the repo. Example ossf.
	Repo  string // The actual repo. Example scorecard.
}

func (r *RepoURL) String() string {
	return fmt.Sprintf("%s/%s/%s", r.Host, r.Owner, r.Repo)
}

func (r *RepoURL) Set(s string) error {
	// Allow skipping scheme for ease-of-use, default to https.
	if !strings.Contains(s, "://") {
		s = "https://" + s
	}

	parsedURL, err := url.Parse(s)
	if err != nil {
		return errors.Wrap(err, "unable to parse the URL")
	}

	const splitLen = 2
	split := strings.SplitN(strings.Trim(parsedURL.Path, "/"), "/", splitLen)
	if len(split) != splitLen {
		log.Fatalf("invalid repo flag: [%s], pass the full repository URL", s)
	}

	r.Host, r.Owner, r.Repo = parsedURL.Host, split[0], split[1]
	return nil
}
