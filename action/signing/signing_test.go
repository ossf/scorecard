package signing

import (
	"io/ioutil"
	"testing"
)

// TODO: For this test to work, fake the OIDC token retrieval with something like.
//nolint // https://github.com/sigstore/cosign/blob/286bb0c58757009e99ab7080c720b30e51d08855/cmd/cosign/cli/fulcio/fulcio_test.go

// func Test_SignScorecardResult(t *testing.T) {
// 	t.Parallel()
// 	// Generate random bytes to use as our payload. This is done because signing identical payloads twice
// 	// just creates multiple entries under it, so we are keeping this test simple and not comparing timestamps.
// 	fmt.Println("ACTIONS_ID_TOKEN_REQUEST_TOKEN:")
// 	fmt.Println(os.Getenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN"))
// 	scorecardResultsFile := "./sign-random-data.txt"
// 	randomData := make([]byte, 20)
// 	if _, err := rand.Read(randomData); err != nil {
// 		t.Errorf("signScorecardResult() error generating random bytes, %v", err)
// 		return
// 	}
// 	if err := ioutil.WriteFile(scorecardResultsFile, randomData, 0o600); err != nil {
// 		t.Errorf("signScorecardResult() error writing random bytes to file, %v", err)
// 		return
// 	}

// 	// Sign example scorecard results file.
// 	err := SignScorecardResult(scorecardResultsFile)
// 	if err != nil {
// 		t.Errorf("signScorecardResult() error, %v", err)
// 		return
// 	}

// 	// Verify that the signature was created and uploaded to the Rekor tlog by looking up the payload.
// 	ctx := context.Background()
// 	rekorClient, err := rekor.NewClient(options.DefaultRekorURL)
// 	if err != nil {
// 		t.Errorf("signScorecardResult() error getting Rekor client, %v", err)
// 		return
// 	}
// 	scorecardResultData, err := ioutil.ReadFile(scorecardResultsFile)
// 	if err != nil {
// 		t.Errorf("signScorecardResult() error reading scorecard result file, %v", err)
// 		return
// 	}
// 	uuids, err := cosign.FindTLogEntriesByPayload(ctx, rekorClient, scorecardResultData)
// 	if err != nil {
// 		t.Errorf("signScorecardResult() error getting tlog entries, %v", err)
// 		return
// 	}

// 	if len(uuids) != 1 {
// 		t.Errorf("signScorecardResult() error finding signature in Rekor tlog, %v", err)
// 		return
// 	}
// }

// Test using scorecard results that have already been signed & uploaded.
func Test_ProcessSignature(t *testing.T) {
	t.Parallel()

	sarifPayload, serr := ioutil.ReadFile("testdata/results.sarif")
	jsonPayload, jerr := ioutil.ReadFile("testdata/results.json")
	repoName := "rohankh532/scorecard-OIDC-test"
	repoRef := "refs/heads/main"

	if serr != nil || jerr != nil {
		t.Errorf("Error reading testdata:, %v, %v", serr, jerr)
	}

	if err := ProcessSignature(sarifPayload, jsonPayload, repoName, repoRef); err != nil {
		t.Errorf("ProcessSignature() error:, %v", err)
		return
	}
}
