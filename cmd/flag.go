package cmd

import (
	"fmt"
	"log"
	"regexp"
	"strings"
)

type repoFlag struct {
	host, owner, repo string
}

func (r *repoFlag) String() string {
	return fmt.Sprintf("%s/%s/%s", r.host, r.owner, r.repo)
}

func (r *repoFlag) Type() string {
	return "repo"
}

func (r *repoFlag) Set(s string) error {
	rgx, _ := regexp.Compile("^https?://")
	repo = rgx.ReplaceAllString(repo, "")
	split := strings.SplitN(s, "/", 3)
	if len(split) != 3 {
		log.Fatalf("invalid repo flag: [%s], pass the full repository URL", s)
	}
	r.host, r.owner, r.repo = split[0], split[1], split[2]

	switch r.host {
	case "github.com":
		return nil
	default:
		return fmt.Errorf("unsupported host: %s", r.host)
	}
}
