// Copyright 2020 Security Scorecard Authors
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
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"mvdan.cc/sh/v3/syntax"
)

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

func getWgetOututFile(cmd []string) (string, bool, error) {
	if isBinaryName("wget", cmd[0]) {
		for i := 1; i < len(cmd)-1; i++ {
			// Find -O output, or use the basename from url
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

func getGsutilOututFile(cmd []string) (string, bool, error) {
	if isBinaryName("gsutil", cmd[0]) {
		for i := 1; i < len(cmd)-1; i++ {
			if !strings.HasPrefix(cmd[i], "gs://") {
				continue
			}
			pathfn := cmd[i+1]
			if filepath.Dir(pathfn) == pathfn {
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

func getAwsOututFile(cmd []string) (string, bool, error) {
	// https://docs.aws.amazon.com/AmazonS3/latest/userguide/download-objects.html.
	if isBinaryName("aws", cmd[0]) {
		if len(cmd) < 3 || !strings.EqualFold("s3api", cmd[1]) || !strings.EqualFold("get-object", cmd[2]) {
			return "", false, nil
		}

		// Just take the last 2 arguments.
		ifile := cmd[len(cmd)-2]
		ofile := cmd[len(cmd)-1]
		if filepath.Dir(ofile) == ofile {
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

func getOutputFile(cmd []string) (string, bool, error) {
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

	shells := []string{
		"sh", "bash", "dash", "ksh", "python",
		"perl", "ruby", "php", "node", "nodejs", "java",
		"exec",
	}
	for _, b := range shells {
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

	shells := []string{
		"sh", "bash", "dash", "ksh", "python",
		"perl", "ruby", "php", "node", "nodejs", "java",
		"exec",
	}
	for _, b := range shells {
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
		// logf("not leftStmt: %v", leftStmt)
		return false
	}
	rightStmt, ok := extractCommand(bc.Y.Cmd)
	if !ok {
		// logf("not rightStmt: %v", rightStmt)
		return false
	}

	if !isDownloadUtility(leftStmt) {
		// logf("not isDownloadUtility: %v", leftStmt)
		return false
	}

	if !isInterpreter(rightStmt) {
		// logf("not isShell: %v", rightStmt)
		return false
	}

	logf("!! frozen-deps/fetch-execute - %v is fetching and executing non-pinned program '%v'",
		pathfn, cmd)
	// logf("we found a candidate: %v | %v", leftStmt, rightStmt)
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

func isFetchFileExecute(bc *syntax.BinaryCmd, cmd, pathfn string,
	logf func(s string, f ...interface{})) bool {

	// Check for && operator.
	if !strings.EqualFold(bc.Op.String(), "&&") {
		return false
	}

	leftStmt, ok := extractCommand(bc.X.Cmd)
	if !ok {
		// logf("not leftStmt: %v", leftStmt)
		return false
	}
	rightStmt, ok := extractCommand(bc.Y.Cmd)
	if !ok {
		// logf("not rightStmt: %v", rightStmt)
		return false
	}

	if !isDownloadUtility(leftStmt) {
		// logf("not isDownloadUtility: %v", leftStmt)
		return false
	}

	fn, ok := getRedirectFile(bc.X.Redirs)
	if !ok {
		return false
	}

	if !isInterpreterWithFile(rightStmt, fn) {
		// logf("not isShell: %v", rightStmt)
		return false
	}

	logf("!! frozen-deps/fetch-execute - %v is fetching and executing non-pinned program '%v'",
		pathfn, cmd)
	// logf("we found a candidate: %v | %v", leftStmt, rightStmt)
	return true
}

func isInterpreterWithFiles(node syntax.Node, cmd, pathfn string, files map[string]bool,
	logf func(s string, f ...interface{})) bool {
	// CallExpr command.
	ce, ok := node.(*syntax.CallExpr)
	if !ok {
		return false
	}

	// syntax.DebugPrint(os.Stdout, ce)
	c, ok := extractCommand(ce)
	if !ok {
		// logf("not leftStmt: %v", leftStmt)
		return false
	}

	ok = false
	for fn := range files {
		// logf("intepreter with %v %v", c, fn)
		if isInterpreterWithFile(c, fn) {
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
	// syntax.DebugPrint(os.Stdout, f)

	cmdValidated := true
	syntax.Walk(f, func(node syntax.Node) bool {
		// BinaryCmd {Op=|, X=CallExpr{Args={curl, -s, url}}, Y=CallExpr{Args={bash,}}}
		if isFetchPipeExecute(node, cmd, pathfn, logf) {
			cmdValidated = false
		}

		// Continue walking the node graph.
		return true
	})

	return cmdValidated, nil
}

func recordFetchFileFromNode(node syntax.Node) (string, bool, error) {
	ss, ok := node.(*syntax.Stmt)
	if !ok {
		return "", false, nil
	}

	cmd, ok := extractCommand(ss.Cmd)
	if !ok {
		// logf("not leftStmt: %v", leftStmt)
		return "", false, nil
	}

	if !isDownloadUtility(cmd) {
		// logf("not isDownloadUtility: %v", leftStmt)
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
	// syntax.DebugPrint(os.Stdout, f)

	cmdValidated := true
	files := make(map[string]bool)
	syntax.Walk(f, func(node syntax.Node) bool {
		if err != nil {
			return false
		}

		// Record the file that is downloaded, if any.
		fn, b, err := recordFetchFileFromNode(node)
		if err != nil {
			return false
		} else if b {
			files[fn] = true
			// logf("added %v from %v", fn, cmd)
		}

		// Check if we're calling a file we previously downloaded.
		if isInterpreterWithFiles(node, cmd, pathfn, files, logf) {
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

func validateCommandIsNotFetchToStdinExecute(cmd, pathfn string, logf func(s string, f ...interface{})) (bool, error) {
	in := strings.NewReader(cmd)
	f, err := syntax.NewParser().Parse(in, "")
	if err != nil {
		return false, ErrParsingShellCommand
	}
	// syntax.DebugPrint(os.Stdout, f)

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
	logf("cmd:%v", c)

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

func validateCommandIsNotSuFetchToStdinExecute(cmd, pathfn string,
	logf func(s string, f ...interface{})) (bool, error) {
	in := strings.NewReader(cmd)
	f, err := syntax.NewParser().Parse(in, "")
	if err != nil {
		return false, ErrParsingShellCommand
	}
	// syntax.DebugPrint(os.Stdout, f)

	cmdValidated := true
	syntax.Walk(f, func(node syntax.Node) bool {
		if err != nil {
			return false
		}
		b, err := isSuFetchStdinExecute(node, cmd, pathfn, logf)
		if err != nil {
			return false
		} else if !b {
			cmdValidated = false
		}

		// Continue walking the node graph.
		return true
	})

	return cmdValidated, err
}

func recordFetchFileFromString(cmd string) (map[string]bool, bool, error) {
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

		fn, b, err := recordFetchFileFromNode(node)
		if err != nil {
			return false
		} else if b {
			files[fn] = true
			// logf("added %v from %v", fn, cmd)
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
	// syntax.DebugPrint(os.Stdout, f)

	cmdValidated := true
	syntax.Walk(f, func(node syntax.Node) bool {
		// Check if we're calling a file we previously downloaded.
		if isInterpreterWithFiles(node, cmd, pathfn, files, logf) {
			cmdValidated = false
		}

		// Continue walking the node graph.
		return true
	})

	return cmdValidated, nil
}

func validateDockerfileDownloads(pathfn string, content []byte,
	logf func(s string, f ...interface{})) (bool, error) {
	contentReader := strings.NewReader(string(content))
	res, err := parser.Parse(contentReader)
	if err != nil {
		return false, fmt.Errorf("cannot read dockerfile content: %w", err)
	}

	ret := true
	files := make(map[string]bool)

	// Walk the Dockerfile's AST.
	for _, child := range res.AST.Children {
		cmdType := child.Value
		// Only look for the 'RUN' command.
		if cmdType != "run" {
			continue
		}

		var valueList []string
		for n := child.Next; n != nil; n = n.Next {
			valueList = append(valueList, n.Value)
		}

		logf("%v: %v", len(valueList), strings.Join(valueList, ","))
		if len(valueList) != 1 {
			return false, ErrParsingDockerfile
		}

		// Validate it's not downloading and piping into a shell, like
		// `curl | bash`.
		r, err := validateCommandIsNotFetchPipeExecute(valueList[0], pathfn, logf)
		if err != nil {
			return false, err
		} else if !r {
			ret = false
		}

		// Validate it is not a download command followed by
		// an execute: `curl > /tmp/file && /tmp/file`
		//			   `curl > /tmp/file && bash /tmp/file`
		//			   `curl > /tmp/file; bash /tmp/file`
		//			   `curl > /tmp/file; /tmp/file`
		r, err = validateCommandIsNotFetchToFileExecute(valueList[0], pathfn, logf)
		if err != nil {
			return false, err
		} else if !r {
			ret = false
		}

		// Validate it's not shelling out by redirecting input to stdin, like
		// `bash <(wget -qO- http://website.com/my-script.sh)`
		r, err = validateCommandIsNotFetchToStdinExecute(valueList[0], pathfn, logf)
		if err != nil {
			return false, err
		} else if !r {
			ret = false
		}

		// Validate it's not using su to execute a script, like
		// `sudo su -c "bash <(wget -qO- http://website.com/my-script.sh)" root`
		r, err = validateCommandIsNotSuFetchToStdinExecute(valueList[0], pathfn, logf)
		if err != nil {
			return false, err
		} else if !r {
			ret = false
		}

		// TODO(laurent): add check for cat file | bash
		// TODO(laurent): detect downloads of zip/tar files containing scripts.

		// Record the name of downloaded file, if any.
		fn, b, err := recordFetchFileFromString(valueList[0])
		if err != nil {
			return false, err
		} else if b {
			for f := range fn {
				// logf("added %v from %v", f, valueList[0])
				files[f] = true
			}
		}

		// Check if a previously-downloaded file is executed via
		// `bash <some-already-downloaded-file>`.
		r, err = validateCommandIsNotFileExecute(valueList[0], pathfn, files, logf)
		if err != nil {
			return false, err
		} else if !r {
			ret = false
		}
	}
	return ret, nil
}
