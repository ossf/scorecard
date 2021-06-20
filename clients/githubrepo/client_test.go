package githubrepo

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/go-github/v32/github"
	"github.com/ossf/scorecard/clients"
)

func TestInitRepo(t *testing.T) {
	ctx := context.Background()
	client := github.NewClient(http.DefaultClient)
	repoClient := CreateGithubRepoClient(ctx, client)
	err := repoClient.InitRepo("does", "not_exist")
	var e *clients.ErrRepoUnavailable
	if !errors.As(err, &e) {
		t.Errorf("expected: %v, got %v", e, err)
	}
}
