package signing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/ossf/scorecard-action/entrypoint" //nolint
	"github.com/ossf/scorecard-action/options"
	sigOpts "github.com/sigstore/cosign/cmd/cosign/cli/options"
	"github.com/sigstore/cosign/cmd/cosign/cli/sign"
)

// SignScorecardResult signs the results file and uploads the attestation to the Rekor transparency log.
func SignScorecardResult(scorecardResultsFile string) error {
	if err := os.Setenv("COSIGN_EXPERIMENTAL", "true"); err != nil {
		return fmt.Errorf("error setting COSIGN_EXPERIMENTAL env var: %w", err)
	}

	// Prepare settings for SignBlobCmd.
	rootOpts := &sigOpts.RootOptions{Timeout: sigOpts.DefaultTimeout} // Just the timeout.

	keyOpts := sign.KeyOpts{
		FulcioURL:    sigOpts.DefaultFulcioURL,     // Signing certificate provider.
		RekorURL:     sigOpts.DefaultRekorURL,      // Transparency log.
		OIDCIssuer:   sigOpts.DefaultOIDCIssuerURL, // OIDC provider to get ID token to auth for Fulcio.
		OIDCClientID: "sigstore",
	}
	regOpts := sigOpts.RegistryOptions{} // Not necessary so we leave blank.

	// This command will use the provided OIDCIssuer to authenticate into Fulcio, which will generate the
	// signing certificate on the scorecard result. This attestation is then uploaded to the Rekor transparency log.
	// The output bytes (signature) and certificate are discarded since verification can be done with just the payload.
	if _, err := sign.SignBlobCmd(rootOpts, keyOpts, regOpts, scorecardResultsFile, true, "", ""); err != nil {
		return fmt.Errorf("error signing payload: %w", err)
	}

	return nil
}

// GetJSONScorecardResults changes output settings to json and runs scorecard again.
// TODO: run scorecard only once and generate multiple formats together.
func GetJSONScorecardResults() ([]byte, error) {
	defer os.Setenv(options.EnvInputResultsFile, os.Getenv(options.EnvInputResultsFile))
	defer os.Setenv(options.EnvInputResultsFormat, os.Getenv(options.EnvInputResultsFormat))
	os.Setenv(options.EnvInputResultsFile, "results.json")
	os.Setenv(options.EnvInputResultsFormat, "json")

	actionJSON, err := entrypoint.New()
	if err != nil {
		return nil, fmt.Errorf("creating scorecard entrypoint: %w", err)
	}
	if err := actionJSON.Execute(); err != nil {
		return nil, fmt.Errorf("error during command execution: %w", err)
	}

	// Get json output data from file.
	jsonPayload, err := ioutil.ReadFile(os.Getenv(options.EnvInputResultsFile))
	if err != nil {
		return nil, fmt.Errorf("reading scorecard json results from file: %w", err)
	}

	return jsonPayload, nil
}

// ProcessSignature calls scorecard-api to process & upload signed scorecard results.
func ProcessSignature(jsonPayload []byte, repoName, repoRef string) error {
	// Prepare HTTP request body for scorecard-webapp-api call.
	resultsPayload := struct {
		JSONOutput string
	}{
		JSONOutput: string(jsonPayload),
	}

	payloadBytes, err := json.Marshal(resultsPayload)
	if err != nil {
		return fmt.Errorf("marshalling json results: %w", err)
	}

	// Call scorecard-webapp-api to process and upload signature.
	// Setup HTTP request and context.
	url := "https://api.securityscorecards.dev/verify"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes)) //nolint
	if err != nil {
		return fmt.Errorf("creating HTTP request: %w", err)
	}
	req.Header.Set("X-Repository", repoName)
	req.Header.Set("X-Branch", repoRef)

	ctx, cancel := context.WithTimeout(req.Context(), 10*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	// Execute request.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("executing scorecard-api call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("reading response body: %w", err)
		}
		return fmt.Errorf("http response %d, status: %v, error: %v", resp.StatusCode, resp.Status, string(bodyBytes)) //nolint
	}

	return nil
}
