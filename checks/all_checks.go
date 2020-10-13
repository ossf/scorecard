package checks

import "github.com/dlorenc/scorecard/checker"

var AllChecks = []checker.NamedCheck{}

func registerCheck(name string, fn checker.CheckFn) {
	AllChecks = append(AllChecks, checker.NamedCheck{
		Name: name,
		Fn:   fn,
	})
}
