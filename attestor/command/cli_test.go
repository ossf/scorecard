// Copyright 2023 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package command

import (
	"testing"

	"github.com/spf13/cobra"
)

func Test_addSignFlags(t *testing.T) {
	type args struct {
		cmd *cobra.Command
	}
	testName := "Test addSignFlags"
	testArgs := args{
		cmd: &cobra.Command{},
	}

	t.Run(testName, func(t *testing.T) {
		addSignFlags(testArgs.cmd)
		// persistent flags of Image being set has to be tested in the integration test
		if testArgs.cmd.PersistentFlags().Lookup("image") == nil {
			t.Errorf("addSignFlags() did not add persistent flag 'image'")
		}
		if testArgs.cmd.PersistentFlags().Lookup("attestation-project") == nil {
			t.Errorf("addSignFlags() did not add persistent flag 'attestation-project'")
		}
		if testArgs.cmd.PersistentFlags().Lookup("overwrite") == nil {
			t.Errorf("addSignFlags() did not add persistent flag 'overwrite'")
		}
		if testArgs.cmd.PersistentFlags().Lookup("kms-key-name") == nil {
			t.Errorf("addSignFlags() did not add persistent flag 'kms-key-name'")
		}
		if testArgs.cmd.PersistentFlags().Lookup("kms-digest-alg") == nil {
			t.Errorf("addSignFlags() did not add persistent flag 'kms-digest-alg'")
		}
		if testArgs.cmd.PersistentFlags().Lookup("pgp-private-key") == nil {
			t.Errorf("addSignFlags() did not add persistent flag 'pgp-private-key'")
		}
		if testArgs.cmd.PersistentFlags().Lookup("pgp-passphrase") == nil {
			t.Errorf("addSignFlags() did not add persistent flag 'pgp-passphrase'")
		}
		if testArgs.cmd.PersistentFlags().Lookup("pkix-private-key") == nil {
			t.Errorf("addSignFlags() did not add persistent flag 'pkix-private-key'")
		}
		if testArgs.cmd.PersistentFlags().Lookup("pkix-alg") == nil {
			t.Errorf("addSignFlags() did not add persistent flag 'pkix-alg'")
		}
	})
}
