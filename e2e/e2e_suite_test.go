package e2e

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/google/go-github/v32/github"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/ossf/scorecard/roundtripper"
	"github.com/shurcooL/githubv4"
	"go.uber.org/zap"
)

var ghClient *github.Client
var graphClient *githubv4.Client
var client *http.Client

type log struct {
	messages []string
}

func (l *log) Logf(s string, f ...interface{}) {
	l.messages = append(l.messages, fmt.Sprintf(s, f...))
}

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2e Suite")
}

var _ = BeforeSuite(func() {
	// making sure the GITHUB_AUTH_TOKEN is set prior to running e2e tests
	token, contains := os.LookupEnv("GITHUB_AUTH_TOKEN")

	Expect(contains).ShouldNot(BeFalse(),
		"GITHUB_AUTH_TOKEN env variable is not set.The GITHUB_AUTH_TOKEN env variable has to be set to run e2e test.")
	Expect(len(token)).ShouldNot(BeZero(), "Length of the GITHUB_AUTH_TOKEN env variable is zero.")

	ctx := context.TODO()

	logLevel := zap.LevelFlag("verbosity", zap.InfoLevel, "override the default log level")
	cfg := zap.NewProductionConfig()
	cfg.Level.SetLevel(*logLevel)
	logger, err := cfg.Build()
	Expect(err).Should(BeNil())

	sugar := logger.Sugar()
	// nolint
	defer logger.Sync() // flushes buffer, if any
	Expect(sugar).ShouldNot(BeNil())

	rt := roundtripper.NewTransport(ctx, sugar)

	client = &http.Client{
		Transport: rt,
	}

	ghClient = github.NewClient(client)
	graphClient = githubv4.NewClient(client)

})

var _ = AfterSuite(func() {
})
