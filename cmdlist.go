package main

import (
	"os"

	"github.com/rogpeppe/modcop/depfile"
)

var listCommand = &Command{
	Short:     "list dependencies",
	UsageLine: "list [packages]",
	Long: `
List shows the set of current dependencies of the listed packages.
By default, only build dependencies are listed; if the
-test flag is provided, test dependencies will be listed too.
If the -m flag is provided, dependencies will be printed at the
module level and version information will be printed too,
in the same format used by "go list -m".

List accepts the same package wildcard patterns that are accepted
by the go list command.
`[1:],
}

func init() {
	listCommand.Run = cmdList // break init cycle
}

var (
	listWithTest = listCommand.Flag.Bool("test", false, "show test dependencies too")
	listModules  = listCommand.Flag.Bool("m", false, "show dependencies at module level")
)

func cmdList(_ *Command, args []string) int {
	if len(args) == 0 {
		args = []string{"."}
	}
	build, test, err := deps(args, *listWithTest)
	if err != nil {
		errorf("%v", err)
		return 1
	}
	buildPkgs := make([]string, len(build))
	for i, p := range build {
		buildPkgs[i] = p.ImportPath
	}
	testPkgs := make([]string, len(test))
	for i, p := range test {
		testPkgs[i] = p.ImportPath
	}
	os.Stdout.Write(depfile.Format(&depfile.File{
		Build: buildPkgs,
		Test:  testPkgs,
	}))
	return 0
}
