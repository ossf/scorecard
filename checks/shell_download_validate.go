// Copyright 2021 Security Scorecard Authors
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

package checks

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

// ErrParsingDockerfile indicates a problem parsing the dockerfile.
var ErrParsingDockerfile = errors.New("file cannot be parsed")

// ErrParsingShellCommand indicates a problem parsing a shell command.
var ErrParsingShellCommand = errors.New("shell command cannot be parsed")

// List of interpreters.
var interpreters = []string{
	"sh", "bash", "dash", "ksh", "python",
	"perl", "ruby", "php", "node", "nodejs", "java",
	"exec",
}

func isBinaryName(expected, name string) bool {
	return strings.EqualFold(path.Base(name), expected)
}

// see https://serverfault.com/questions/226386/wget-a-script-and-run-it/890417.
func isDownloadUtility(cmd []string) bool {
	if len(cmd) == 0 {
		return false
	}
	// Note: we won't be catching those if developers have re-named
	// the utility.
	// Note: wget -O - <website>, but we don't check for that explicitly.
	utils := [3]string{"curl", "wget", "gsutil"}
	for _, b := range utils {
		if isBinaryName(b, cmd[0]) {
			return true
		}
	}

	// aws s3api get-object.
	// https://docs.aws.amazon.com/AmazonS3/latest/userguide/download-objects.html.
	if isBinaryName("aws", cmd[0]) {
		if len(cmd) >= 3 && strings.EqualFold("s3api", cmd[1]) && strings.EqualFold("get-object", cmd[2]) {
			return true
		}
	}
	return false
}

func getWgetOututFile(cmd []string) (pathfn string, ok bool, err error) {
	if isBinaryName("wget", cmd[0]) {
		for i := 1; i < len(cmd)-1; i++ {
			// Find -O output, or use the basename from url.
			if strings.EqualFold(cmd[i], "-O") {
				return cmd[i+1], true, nil
			}
		}

		// Could not find -O option, use the url's name instead.
		for i := 1; i < len(cmd); i++ {
			if !strings.HasPrefix(cmd[i], "http") {
				continue
			}

			u, err := url.Parse(cmd[i])
			if err != nil {
				return "", false, fmt.Errorf("url.Parse: %w", err)
			}
			return path.Base(u.Path), true, nil
		}
	}
	return "", false, nil
}

func getGsutilOututFile(cmd []string) (pathfn string, ok bool, err error) {
	if isBinaryName("gsutil", cmd[0]) {
		for i := 1; i < len(cmd)-1; i++ {
			if !strings.HasPrefix(cmd[i], "gs://") {
				continue
			}
			pathfn := cmd[i+1]
			if filepath.Clean(filepath.Dir(pathfn)) == filepath.Clean(pathfn) {
				// Directory.
				u, err := url.Parse(cmd[i])
				if err != nil {
					return "", false, fmt.Errorf("url.Parse: %w", err)
				}
				return filepath.Join(filepath.Dir(pathfn), path.Base(u.Path)), true, nil
			}

			// File provided.
			return pathfn, true, nil
		}
	}
	return "", false, nil
}

func getAwsOututFile(cmd []string) (pathfn string, ok bool, err error) {
	// https://docs.aws.amazon.com/AmazonS3/latest/userguide/download-objects.html.
	if isBinaryName("aws", cmd[0]) {
		if len(cmd) < 3 || !strings.EqualFold("s3api", cmd[1]) || !strings.EqualFold("get-object", cmd[2]) {
			return "", false, nil
		}

		// Just take the last 2 arguments.
		ifile := cmd[len(cmd)-2]
		ofile := cmd[len(cmd)-1]
		if filepath.Clean(filepath.Dir(ofile)) == filepath.Clean(ofile) {
			u, err := url.Parse(ifile)
			if err != nil {
				return "", false, fmt.Errorf("url.Parse: %w", err)
			}
			return filepath.Join(filepath.Dir(ofile), path.Base(u.Path)), true, nil
		}

		// File provided.
		return ofile, true, nil
	}
	return "", false, nil
}

func getOutputFile(cmd []string) (pathfn string, ok bool, err error) {
	if len(cmd) == 0 {
		return "", false, nil
	}

	// Wget.
	fn, b, err := getWgetOututFile(cmd)
	if err != nil || b {
		return fn, b, err
	}

	// Gsutil.
	fn, b, err = getGsutilOututFile(cmd)
	if err != nil || b {
		return fn, b, err
	}

	// Aws.
	fn, b, err = getAwsOututFile(cmd)
	if err != nil || b {
		return fn, b, err
	}

	return "", false, nil
}

func isInterpreter(cmd []string) bool {
	if len(cmd) == 0 {
		return false
	}

	for _, b := range interpreters {
		if isBinaryName(b, cmd[0]) {
			return true
		}
	}
	return false
}

func isInterpreterWithFile(cmd []string, fn string) bool {
	if len(cmd) == 0 {
		return false
	}

	for _, b := range interpreters {
		if isBinaryName(b, cmd[0]) {
			for _, arg := range cmd[1:] {
				if strings.EqualFold(filepath.Clean(arg), filepath.Clean(fn)) {
					return true
				}
			}
		}
	}
	return false
}

func isExecuteFile(cmd []string, fn string) bool {
	if len(cmd) == 0 {
		return false
	}

	return strings.EqualFold(filepath.Clean(cmd[0]), filepath.Clean(fn))
}

func extractCommand(cmd interface{}) ([]string, bool) {
	c, ok := cmd.(*syntax.CallExpr)
	if !ok {
		return nil, ok
	}

	var ret []string
	for _, w := range c.Args {
		if len(w.Parts) != 1 {
			continue
		}
		lit, ok := w.Parts[0].(*syntax.Lit)
		if ok && !strings.EqualFold(lit.Value, "sudo") {
			ret = append(ret, lit.Value)
		}
	}
	return ret, true
}

func isFetchPipeExecute(node syntax.Node, cmd, pathfn string,
	logf func(s string, f ...interface{})) bool {
	// BinaryCmd {Op=|, X=CallExpr{Args={curl, -s, url}}, Y=CallExpr{Args={bash,}}}.
	bc, ok := node.(*syntax.BinaryCmd)
	if !ok {
		return false
	}

	// Look for the pipe operator.
	if !strings.EqualFold(bc.Op.String(), "|") {
		return false
	}

	leftStmt, ok := extractCommand(bc.X.Cmd)
	if !ok {
		return false
	}
	rightStmt, ok := extractCommand(bc.Y.Cmd)
	if !ok {
		return false
	}

	if !isDownloadUtility(leftStmt) {
		return false
	}

	if !isInterpreter(rightStmt) {
		return false
	}

	logf("!! frozen-deps/fetch-execute - %v is fetching and executing non-pinned program '%v'",
		pathfn, cmd)
	return true
}

func getRedirectFile(red []*syntax.Redirect) (string, bool) {
	if len(red) == 0 {
		return "", false
	}
	for _, r := range red {
		if !strings.EqualFold(r.Op.String(), ">") {
			continue
		}

		if len(r.Word.Parts) != 1 {
			continue
		}

		lit, ok := r.Word.Parts[0].(*syntax.Lit)
		if ok {
			return lit.Value, true
		}
	}
	return "", false
}

func isExecuteFiles(node syntax.Node, cmd, pathfn string, files map[string]bool,
	logf func(s string, f ...interface{})) bool {
	ce, ok := node.(*syntax.CallExpr)
	if !ok {
		return false
	}

	c, ok := extractCommand(ce)
	if !ok {
		return false
	}

	ok = false
	for fn := range files {
		if isInterpreterWithFile(c, fn) || isExecuteFile(c, fn) {
			logf("!! frozen-deps/fetch-execute - %v is fetching and executing non-pinned program '%v'",
				pathfn, cmd)
			ok = true
		}
	}

	return ok
}

// Detect `fetch | exec`.
func validateCommandIsNotFetchPipeExecute(cmd, pathfn string, logf func(s string, f ...interface{})) (bool, error) {
	in := strings.NewReader(cmd)
	f, err := syntax.NewParser().Parse(in, "")
	if err != nil {
		return false, ErrParsingShellCommand
	}

	cmdValidated := true
	syntax.Walk(f, func(node syntax.Node) bool {
		if isFetchPipeExecute(node, cmd, pathfn, logf) {
			cmdValidated = false
		}

		// Continue walking the node graph.
		return true
	})

	return cmdValidated, nil
}

func recordFetchFileFromNode(node syntax.Node) (pathfn string, ok bool, err error) {
	ss, ok := node.(*syntax.Stmt)
	if !ok {
		return "", false, nil
	}

	cmd, ok := extractCommand(ss.Cmd)
	if !ok {
		return "", false, nil
	}

	if !isDownloadUtility(cmd) {
		return "", false, nil
	}

	fn, ok := getRedirectFile(ss.Redirs)
	if !ok {
		return getOutputFile(cmd)
	}

	return fn, true, nil
}

func validateCommandIsNotFetchToFileExecute(cmd, pathfn string, logf func(s string, f ...interface{})) (bool, error) {
	in := strings.NewReader(cmd)
	f, err := syntax.NewParser().Parse(in, "")
	if err != nil {
		return false, ErrParsingShellCommand
	}

	cmdValidated := true
	files := make(map[string]bool)
	syntax.Walk(f, func(node syntax.Node) bool {
		if err != nil {
			return false
		}
		var fn string
		var b bool
		// Record the file that is downloaded, if any.
		fn, b, err = recordFetchFileFromNode(node)
		if err != nil {
			return false
		} else if b {
			files[fn] = true
		}

		// Check if we're calling a file we previously downloaded.
		if isExecuteFiles(node, cmd, pathfn, files, logf) {
			cmdValidated = false
		}

		// Continue walking the node graph.
		return true
	})

	return cmdValidated, err
}

func isFetchStdinExecute(node syntax.Node, cmd, pathfn string,
	logf func(s string, f ...interface{})) bool {
	ce, ok := node.(*syntax.CallExpr)
	if !ok {
		return false
	}

	c, ok := extractCommand(ce)
	if !ok {
		return false
	}

	if !isInterpreter(c) {
		return false
	}

	// Now parse the process substitution part.
	// Example: `bash <(wget -qO- http://website.com/my-script.sh)`.
	l := 2
	if len(ce.Args) < l {
		return false
	}

	parts := ce.Args[1].Parts
	if len(parts) != 1 {
		return false
	}

	part := parts[0]
	p, ok := part.(*syntax.ProcSubst)
	if !ok {
		return false
	}

	if !strings.EqualFold(p.Op.String(), "<(") {
		return false
	}

	if len(p.Stmts) == 0 {
		return false
	}

	c, ok = extractCommand(p.Stmts[0].Cmd)
	if !ok {
		return false
	}

	if !isDownloadUtility(c) {
		return false
	}

	logf("!! frozen-deps/fetch-execute - %v is fetching and executing non-pinned program '%v'",
		pathfn, cmd)
	return true
}

func isSuCommand(cmd []string) bool {
	isSu := false
	for _, c := range cmd {
		if strings.EqualFold("su", c) {
			isSu = true
		} else if isSu && strings.HasPrefix(c, "-") && strings.Contains(c, "c") {
			return true
		}
	}
	return false
}

func extractSuCommand(args []*syntax.Word) (string, bool) {
	for _, arg := range args {
		if len(arg.Parts) != 1 {
			continue
		}
		part := arg.Parts[0]
		v, ok := part.(*syntax.DblQuoted)
		if !ok {
			continue
		}
		if len(v.Parts) != 1 {
			continue
		}

		lit, ok := v.Parts[0].(*syntax.Lit)
		if !ok {
			continue
		}
		return lit.Value, true
	}
	return "", false
}

func validateCommandIsNotNotSuFetchToStdinExecute(cmd, pathfn string,
	logf func(s string, f ...interface{})) (bool, error) {
	in := strings.NewReader(cmd)
	f, err := syntax.NewParser().Parse(in, "")
	if err != nil {
		return false, ErrParsingShellCommand
	}

	cmdValidated := true
	syntax.Walk(f, func(node syntax.Node) bool {
		if isFetchStdinExecute(node, cmd, pathfn, logf) {
			cmdValidated = false
		}

		// Continue walking the node graph.
		return true
	})

	return cmdValidated, nil
}

func isSuFetchStdinExecute(node syntax.Node, cmd, pathfn string,
	logf func(s string, f ...interface{})) (bool, error) {
	ce, ok := node.(*syntax.CallExpr)
	if !ok {
		return false, nil
	}

	c, ok := extractCommand(ce)
	if !ok {
		return false, nil
	}

	if !isSuCommand(c) {
		return false, nil
	}

	cs, ok := extractSuCommand(ce.Args)
	if !ok {
		return false, nil
	}

	// We have extracted the command. Now parse it.
	in := strings.NewReader(cs)
	f, err := syntax.NewParser().Parse(in, "")
	if err != nil {
		return false, fmt.Errorf("syntax.NewParser: %w", err)
	}

	isFetch := false
	syntax.Walk(f, func(node syntax.Node) bool {
		if isFetchStdinExecute(node, cmd, pathfn, logf) {
			isFetch = true
		}

		// Continue walking the node graph.
		return true
	})

	return isFetch, nil
}

func validateCommandIsNotSuFetchToStdinExecute(cmd, pathfn string,
	logf func(s string, f ...interface{})) (bool, error) {
	in := strings.NewReader(cmd)
	f, err := syntax.NewParser().Parse(in, "")
	if err != nil {
		return false, ErrParsingShellCommand
	}

	cmdValidated := true
	syntax.Walk(f, func(node syntax.Node) bool {
		if err != nil {
			return false
		}
		var b bool
		b, err = isSuFetchStdinExecute(node, cmd, pathfn, logf)
		if err != nil {
			return false
		} else if b {
			cmdValidated = false
		}

		// Continue walking the node graph.
		return true
	})

	return cmdValidated, err
}

func validateCommandIsNotFetchToStdinExecute(cmd, pathfn string,
	logf func(s string, f ...interface{})) (bool, error) {
	r1, err := validateCommandIsNotNotSuFetchToStdinExecute(cmd, pathfn, logf)
	if err != nil {
		return r1, err
	}

	r2, err := validateCommandIsNotSuFetchToStdinExecute(cmd, pathfn, logf)
	if err != nil {
		return r2, err
	}

	return r1 && r2, nil
}

func recordFetchFileFromString(cmd string) (fmap map[string]bool, ok bool, err error) {
	in := strings.NewReader(cmd)
	f, err := syntax.NewParser().Parse(in, "")
	if err != nil {
		return nil, false, ErrParsingShellCommand
	}

	files := make(map[string]bool)
	syntax.Walk(f, func(node syntax.Node) bool {
		if err != nil {
			return false
		}
		var b bool
		var fn string
		fn, b, err = recordFetchFileFromNode(node)
		if err != nil {
			return false
		} else if b {
			files[fn] = true
		}

		// Continue walking the node graph.
		return true
	})

	return files, true, err
}

func validateCommandIsNotFileExecute(cmd, pathfn string, files map[string]bool,
	logf func(s string, f ...interface{})) (bool, error) {
	in := strings.NewReader(cmd)
	f, err := syntax.NewParser().Parse(in, "")
	if err != nil {
		return false, ErrParsingShellCommand
	}

	cmdValidated := true
	syntax.Walk(f, func(node syntax.Node) bool {
		// Check if we're calling a file we previously downloaded.
		if isExecuteFiles(node, cmd, pathfn, files, logf) {
			cmdValidated = false
		}

		// Continue walking the node graph.
		return true
	})

	return cmdValidated, nil
}
