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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"github.com/ossf/scorecard/v3/checker"
	sce "github.com/ossf/scorecard/v3/errors"
)

var (
	// supportedShells is the list of shells that are supported by mvdan.cc/sh/v3/syntax.
	supportedShells = []string{
		"sh", "bash", "mksh",
	}
	// otherShells are not supported by our parser.
	otherShells = []string{
		"dash", "ksh",
	}
	shellNames         = append(supportedShells, otherShells...)
	pythonInterpreters = []string{"python", "python3", "python2.7"}
	shellInterpreters  = append([]string{"exec", "su"}, shellNames...)
	otherInterpreters  = []string{"perl", "ruby", "php", "node", "nodejs", "java"}
	interpreters       = append(otherInterpreters,
		append(shellInterpreters, append(shellNames, pythonInterpreters...)...)...)
)

// Note: aws is handled separately because it uses different
// cli options.
var downloadUtils = []string{
	"curl", "wget", "gsutil",
}

func isBinaryName(expected, name string) bool {
	return strings.EqualFold(path.Base(name), expected)
}

func isExecuteFile(cmd []string, fn string) bool {
	if len(cmd) == 0 {
		return false
	}

	return strings.EqualFold(filepath.Clean(cmd[0]), filepath.Clean(fn))
}

// see https://serverfault.com/questions/226386/wget-a-script-and-run-it/890417.
func isDownloadUtility(cmd []string) bool {
	if len(cmd) == 0 {
		return false
	}
	// Note: we won't be catching those if developers have re-named
	// the utility.
	// Note: wget -O - <website>, but we don't check for that explicitly.
	for _, b := range downloadUtils {
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

func getWgetOutputFile(cmd []string) (pathfn string, ok bool, err error) {
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
				return "", false, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("url.Parse: %v", err))
			}
			return path.Base(u.Path), true, nil
		}
	}
	return "", false, nil
}

func getGsutilOutputFile(cmd []string) (pathfn string, ok bool, err error) {
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
					return "", false, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("url.Parse: %v", err))
				}
				return filepath.Join(filepath.Dir(pathfn), path.Base(u.Path)), true, nil
			}

			// File provided.
			return pathfn, true, nil
		}
	}
	return "", false, nil
}

func getAWSOutputFile(cmd []string) (pathfn string, ok bool, err error) {
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
				return "", false, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("url.Parse: %v", err))
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
	fn, b, err := getWgetOutputFile(cmd)
	if err != nil || b {
		return fn, b, err
	}

	// Gsutil.
	fn, b, err = getGsutilOutputFile(cmd)
	if err != nil || b {
		return fn, b, err
	}

	// Aws.
	fn, b, err = getAWSOutputFile(cmd)
	if err != nil || b {
		return fn, b, err
	}

	// TODO(laurent): add other cloud services' utilities
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

func isShellInterpreterOrCommand(cmd []string) bool {
	if len(cmd) == 0 {
		return false
	}

	if isPythonCommand(cmd) {
		return false
	}

	for _, b := range otherInterpreters {
		if isBinaryName(b, cmd[0]) {
			return false
		}
	}
	return true
}

func extractInterpreterAndCommand(cmd []string) (string, bool) {
	if len(cmd) == 0 {
		return "", false
	}

	for _, b := range interpreters {
		if isCommand(cmd, b) {
			return b, true
		}
	}
	return "", false
}

func isInterpreterWithFile(cmd []string, fn string) bool {
	if len(cmd) == 0 {
		return false
	}

	for _, b := range interpreters {
		if !isBinaryName(b, cmd[0]) {
			continue
		}
		for _, arg := range cmd[1:] {
			if strings.EqualFold(filepath.Clean(arg), filepath.Clean(fn)) {
				return true
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
		switch v := w.Parts[0].(type) {
		default:
			continue
		case *syntax.SglQuoted:
			ret = append(ret, "'"+v.Value+"'")
		case *syntax.DblQuoted:
			if len(v.Parts) != 1 {
				continue
			}
			lit, ok := v.Parts[0].(*syntax.Lit)
			if ok {
				ret = append(ret, "\""+lit.Value+"\"")
			}
		case *syntax.Lit:
			if !strings.EqualFold(v.Value, "sudo") {
				ret = append(ret, v.Value)
			}
		}
	}

	return ret, true
}

func isFetchPipeExecute(node syntax.Node, cmd, pathfn string,
	dl checker.DetailLogger) bool {
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

	dl.Warn3(&checker.LogMessage{
		Path:    pathfn,
		Type:    checker.FileTypeSource,
		Offset:  0, // TODO: add line numbers
		Snippet: cmd,
		Text:    "insecure (not pinned by hash) download detected",
	})
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
	dl checker.DetailLogger) bool {
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
			dl.Warn3(&checker.LogMessage{
				Path:    pathfn,
				Type:    checker.FileTypeSource,
				Offset:  0, // TODO: add line numbers
				Snippet: cmd,
				Text:    "insecure (not pinned by hash) download detected",
			})
			ok = true
		}
	}

	return ok
}

// Npm install docs are here.
// https://docs.npmjs.com/cli/v7/commands/npm-install
func isNpmUnpinnedDownload(cmd []string) bool {
	if len(cmd) == 0 {
		return false
	}

	if !isBinaryName("npm", cmd[0]) {
		return false
	}

	// `Npm install` will automatically look up the
	// package.json and package-lock.json, so we don't flag it.

	if len(cmd) <= 2 {
		return false
	}

	found := false
	for i := 1; i < len(cmd)-1; i++ {
		// Search for get and install commands.
		if strings.EqualFold(cmd[i], "install") ||
			strings.EqualFold(cmd[i], "i") {
			found = true
		}

		if !found {
			continue
		}

		nextArg := cmd[i+1]

		// If this is a option to install, skip over it.
		if nextArg[0] == '-' {
			continue
		}

		// We've skipped all options, `nextArg` holds the package.
		return true
	}

	return false
}

func isGoUnpinnedDownload(cmd []string) bool {
	if len(cmd) == 0 {
		return false
	}

	if !isBinaryName("go", cmd[0]) {
		return false
	}

	// `Go install` will automatically look up the
	// go.mod and go.sum, so we don't flag it.

	if len(cmd) <= 2 {
		return false
	}

	found := false
	hashRegex := regexp.MustCompile("^[A-Fa-f0-9]{40,}$")
	for i := 1; i < len(cmd)-1; i++ {
		// Search for get and install commands.
		if strings.EqualFold(cmd[i], "install") ||
			strings.EqualFold(cmd[i], "get") {
			found = true
		}

		if !found {
			continue
		}

		pkg := cmd[i+1]
		// Verify pkg = name@hash
		parts := strings.Split(pkg, "@")

		if len(parts) != 2 {
			continue
		}
		hash := parts[1]
		if hashRegex.Match([]byte(hash)) {
			return false
		}
	}

	return found
}

func isUnpinnedPipInstall(cmd []string) bool {
	if !isBinaryName("pip", cmd[0]) && !isBinaryName("pip3", cmd[0]) {
		return false
	}

	isInstall := false
	hasRequireHashes := false
	hasAdditionalArgs := false
	hasWheel := false
	for i := 1; i < len(cmd); i++ {
		// Search for install commands.
		if strings.EqualFold(cmd[i], "install") {
			isInstall = true
			continue
		}

		if !isInstall {
			break
		}

		// https://github.com/ossf/scorecard/issues/1306#issuecomment-974539197.
		if strings.EqualFold(cmd[i], "--require-hashes") {
			hasRequireHashes = true
			break
		}

		// Exclude *.whl as they're mostly used
		// for tests. See https://github.com/ossf/scorecard/pull/611.
		if strings.HasSuffix(cmd[i], ".whl") {
			// We continue because a command may contain
			// multiple packages to install, not just `.whl` files.
			hasWheel = true
			continue
		}

		hasAdditionalArgs = true
	}

	// If hashes are required, it's pinned.
	if hasRequireHashes {
		return false
	}

	// With additional arguments, it's unpinned.
	// Example: `pip install bla.whl pkg1`
	if hasAdditionalArgs {
		return true
	}

	// No additional arguments and hashes are not required.
	// The only pinned command is `pip install *.whl`
	if hasWheel {
		return false
	}

	// Any other form of install is unpinned,
	// e.g. `pip install`.
	return isInstall
}

func isPythonCommand(cmd []string) bool {
	for _, pi := range pythonInterpreters {
		if isBinaryName(pi, cmd[0]) {
			return true
		}
	}
	return false
}

func extractPipCommand(cmd []string) ([]string, bool) {
	if len(cmd) == 0 {
		return nil, false
	}

	for i := 1; i < len(cmd); i++ {
		// Search for pip module.
		if strings.EqualFold(cmd[i], "-m") &&
			i < len(cmd)-1 &&
			strings.EqualFold(cmd[i+1], "pip") {
			return cmd[i+1:], true
		}
	}
	return nil, false
}

func isUnpinnedPythonPipInstall(cmd []string) bool {
	if !isPythonCommand(cmd) {
		return false
	}
	pipCommand, ok := extractPipCommand(cmd)
	if !ok {
		return false
	}
	return isUnpinnedPipInstall(pipCommand)
}

func isPipUnpinnedDownload(cmd []string) bool {
	if len(cmd) == 0 {
		return false
	}

	if isUnpinnedPipInstall(cmd) {
		return true
	}

	if isUnpinnedPythonPipInstall(cmd) {
		return true
	}

	return false
}

func isUnpinnedPakageManagerDownload(node syntax.Node, cmd, pathfn string,
	dl checker.DetailLogger) bool {
	ce, ok := node.(*syntax.CallExpr)
	if !ok {
		return false
	}

	c, ok := extractCommand(ce)
	if !ok {
		return false
	}

	// Go get/install.
	if isGoUnpinnedDownload(c) {
		dl.Warn3(&checker.LogMessage{
			Path:    pathfn,
			Type:    checker.FileTypeSource,
			Offset:  0, // TODO: add line numbers
			Snippet: cmd,
			Text:    "insecure (not pinned by hash) download detected",
		})
		return true
	}

	// Pip install.
	if isPipUnpinnedDownload(c) {
		dl.Warn3(&checker.LogMessage{
			Path:    pathfn,
			Type:    checker.FileTypeSource,
			Offset:  0, // TODO: add line numbers
			Snippet: cmd,
			Text:    "insecure (not pinned by hash) download detected",
		})
		return true
	}

	// Npm install.
	if isNpmUnpinnedDownload(c) {
		dl.Warn3(&checker.LogMessage{
			Path:    pathfn,
			Type:    checker.FileTypeSource,
			Offset:  0, // TODO: add line numbers
			Snippet: cmd,
			Text:    "insecure (not pinned by hash) download detected",
		})
		return true
	}

	// TODO(laurent): add other package managers.

	return false
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

func isFetchProcSubsExecute(node syntax.Node, cmd, pathfn string,
	dl checker.DetailLogger) bool {
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

	dl.Warn3(&checker.LogMessage{
		Path:    pathfn,
		Type:    checker.FileTypeSource,
		Offset:  0, // TODO: add line numbers
		Snippet: cmd,
		Text:    "insecure (not pinned by hash) download detected",
	})
	return true
}

func isCommand(cmd []string, b string) bool {
	isBin := false
	for _, c := range cmd {
		if isBinaryName(b, c) {
			isBin = true
		} else if isBin && strings.HasPrefix(c, "-") && strings.Contains(c, "c") {
			return true
		}
	}
	return false
}

func extractInterpreterCommandFromArgs(args []*syntax.Word) (string, bool) {
	for _, arg := range args {
		if len(arg.Parts) != 1 {
			continue
		}
		part := arg.Parts[0]
		switch v := part.(type) {
		case *syntax.DblQuoted:
			if len(v.Parts) != 1 {
				continue
			}

			lit, ok := v.Parts[0].(*syntax.Lit)
			if !ok {
				continue
			}
			return lit.Value, true

		case *syntax.SglQuoted:
			return v.Value, true
		}
	}

	return "", false
}

func extractInterpreterAndCommandFromNode(node syntax.Node) (interpreter, command string, yes bool) {
	ce, ok := node.(*syntax.CallExpr)
	if !ok {
		return "", "", false
	}

	c, ok := extractCommand(ce)
	if !ok {
		return "", "", false
	}

	i, ok := extractInterpreterAndCommand(c)
	if !ok {
		return "", "", false
	}

	cs, ok := extractInterpreterCommandFromArgs(ce.Args)
	if !ok {
		return "", "", false
	}

	return i, cs, true
}

func nodeToString(p *syntax.Printer, node syntax.Node) (string, error) {
	// https://github.com/mvdan/sh/blob/24dd9930bc1cfc7be025f8b75b2e9e9f04524012/syntax/printer.go#L135.
	var buf bytes.Buffer
	err := p.Print(&buf, node)
	// This is ugly, but the parser does not have a defined error type :/.
	if err != nil && !strings.Contains(err.Error(), "unsupported node type") {
		return "", sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("syntax.Printer.Print: %v", err))
	}
	return buf.String(), nil
}

func validateShellFileAndRecord(pathfn string, content []byte, files map[string]bool,
	dl checker.DetailLogger) (bool, error) {
	in := strings.NewReader(string(content))
	f, err := syntax.NewParser().Parse(in, "")
	if err != nil {
		// Note: this is caught by internal caller and only printed
		// to avoid failing on shell scripts that our parser does not understand.
		// Example: https://github.com/openssl/openssl/blob/master/util/shlib_wrap.sh.in
		return false, sce.WithMessage(sce.ErrorShellParsing, err.Error())
	}

	printer := syntax.NewPrinter()
	validated := true

	syntax.Walk(f, func(node syntax.Node) bool {
		cmdStr, e := nodeToString(printer, node)
		if e != nil {
			err = e
			return false
		}

		// interpreter -c "CMD".
		i, c, ok := extractInterpreterAndCommandFromNode(node)
		// TODO: support other interpreters.
		// Example: https://github.com/apache/airflow/blob/main/scripts/ci/kubernetes/ci_run_kubernetes_tests.sh#L75
		// HOST_PYTHON_VERSION=$(python3 -c 'import sys; print(f"{sys.version_info[0]}.{sys.version_info[1]}")')``
		// nolinter
		if ok && isShellInterpreterOrCommand([]string{i}) {
			ok, e := validateShellFileAndRecord(pathfn, []byte(c), files, dl)
			validated = ok
			if e != nil {
				err = e
				return true
			}
		}

		// `curl | bash` (supports `sudo`).
		if isFetchPipeExecute(node, cmdStr, pathfn, dl) {
			validated = false
		}

		// Check if we're calling a file we previously downloaded.
		// Includes `curl > /tmp/file [&&|;] [bash] /tmp/file`
		if isExecuteFiles(node, cmdStr, pathfn, files, dl) {
			validated = false
		}

		// `bash <(wget -qO- http://website.com/my-script.sh)`. (supports `sudo`).
		if isFetchProcSubsExecute(node, cmdStr, pathfn, dl) {
			validated = false
		}

		// Package manager's unpinned installs.
		if isUnpinnedPakageManagerDownload(node, cmdStr, pathfn, dl) {
			validated = false
		}
		// TODO(laurent): add check for cat file | bash.
		// TODO(laurent): detect downloads of zip/tar files containing scripts.
		// TODO(laurent): detect command being an env variable.
		// TODO(laurent): detect unpinned git clone.

		// Record the file that is downloaded, if any.
		fn, b, e := recordFetchFileFromNode(node)
		if e != nil {
			err = e
			return false
		} else if b {
			files[fn] = true
		}

		// Continue walking the node graph.
		return true
	})

	return validated, err
}

// The functions below are the only ones that should be called by other files.
// There needs to be a call to extractInterpreterCommandFromString() prior
// to calling other functions.
func isSupportedShell(shellName string) bool {
	for _, name := range supportedShells {
		if isBinaryName(name, shellName) {
			return true
		}
	}
	return false
}

func isShellScriptFile(pathfn string, content []byte) bool {
	return isMatchingShellScriptFile(pathfn, content, shellInterpreters)
}

// isSupportedShellScriptFile returns true if this file is one of the shell scripts we can parse. If a shebang
// is present in the file, the decision is based entirely on that, otherwise the file extension is used to decide.

func isSupportedShellScriptFile(pathfn string, content []byte) bool {
	return isMatchingShellScriptFile(pathfn, content, supportedShells)
}

func isMatchingShellScriptFile(pathfn string, content []byte, shellsToMatch []string) bool {
	// Determine if it matches the file extension first.
	hasShellFileExtension := false
	for _, name := range shellsToMatch {
		// Look at the prefix.
		if strings.HasSuffix(pathfn, "."+name) {
			hasShellFileExtension = true
			break
		}
	}

	// Look at file content.
	r := strings.NewReader(string(content))
	scanner := bufio.NewScanner(r)
	// TODO: support perl scripts with embedded shell scripts:
	// https://github.com/openssl/openssl/blob/master/test/recipes/15-test_dsaparam.t.

	// Only look at first line.
	if !scanner.Scan() {
		return hasShellFileExtension
	}
	line := scanner.Text()

	//  #!/bin/XXX, #!XXX, #!/usr/bin/env XXX, #!env XXX
	if !strings.HasPrefix(line, "#!") {
		// If there's no shebang, go off the file extension.
		return hasShellFileExtension
	}

	line = line[2:]
	for _, name := range shellsToMatch {
		parts := strings.Split(line, " ")
		// #!/bin/bash, #!bash -e
		if len(parts) >= 1 && isBinaryName(name, parts[0]) {
			return true
		}

		// #!/bin/env bash
		if len(parts) >= 2 &&
			isBinaryName("env", parts[0]) &&
			isBinaryName(name, parts[1]) {
			return true
		}
	}

	return false // It has a shebang, but it's not one of our matching shells.
}

func validateShellFile(pathfn string, content []byte, dl checker.DetailLogger) (bool, error) {
	files := make(map[string]bool)
	r, err := validateShellFileAndRecord(pathfn, content, files, dl)
	if err != nil && errors.Is(err, sce.ErrorShellParsing) {
		// Discard and print this particular error for now.
		dl.Debug(err.Error())
		err = nil
	}
	return r, err
}
