package checker

import (
	"context"
	"net/http"

	"github.com/google/go-github/v32/github"
)

type Checker struct {
	Ctx         context.Context
	Client      *github.Client
	HttpClient  *http.Client
	Owner, Repo string
}
