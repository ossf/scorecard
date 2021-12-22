package deps

import (
	"fmt"
	"os/exec"
	"strings"
)

type golang struct{}

func NewGolangDeps() Deps {
	return golang{}
}

// FetchDependecies parses the dependencies in the go.mod using the `go list command`
// This functions expects the directory to contain the go.mod file.
func (g golang) FetchDependencies(directory string) ([]string, error) {
	modquery := `
	go list -m -f '{{if not (or  .Main)}}{{.Path}}{{end}}' all \
	| grep "^github" \
	| sort -u \
	| cut -d/ -f1-3 \
	| awk '{print $1}' \
	| tr '\n' ',' | head -c -1
	`
	// Runs the modquery to generate the dependencies
	c := exec.Command("bash", "-c", fmt.Sprintf("cd %s;", directory)+modquery)
	data, err := c.Output()
	if err != nil {
		return nil, err
	}
	parameters := []string{}
	return append(parameters, strings.Split(string(data), ",")...), nil
}
