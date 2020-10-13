package checker

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/google/go-github/v32/github"
)

type Checker struct {
	Ctx         context.Context
	Client      *github.Client
	HttpClient  *http.Client
	Owner, Repo string
	Logf        func(s string, f ...interface{})
}

type logger struct {
	message string
}

func (l *logger) Logf(s string, f ...interface{}) {
	l.message += fmt.Sprintf(s+"\n", f...)
}

type Runner struct {
	Checker Checker
}

func (r *Runner) Run(f CheckFn) CheckResult {
	var res CheckResult
	var l logger
	for retriesRemaining := 3; retriesRemaining > 0; retriesRemaining-- {
		checker := r.Checker
		l = logger{}
		checker.Logf = l.Logf
		res = f(checker)
		if res.ShouldRetry {
			log.Println(res.Error)
			continue
		}
		break
	}
	res.Details = l.message
	return res
}
