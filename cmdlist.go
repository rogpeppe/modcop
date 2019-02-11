package main

import (
	"bufio"
	"os"
	"sort"
	"strings"
)

var listCommand = &Command{
	Short:     "list dependencies",
	UsageLine: "list [packages]",
	Long: `
List shows the set of current dependencies of the listed packages,
one per line. By default, only build dependencies are listed; if the
-test flag is provided, test dependencies will be listed too;
if the -testonly flag is provided, only dependencies that are
used exclusively by tests will be listed.
If the -m flag is provided, dependencies will be printed at the
module level and version information will be printed too,
in the same format used by "go list -m".

List accepts the same package wildcard patterns that are accepted
by the go list command. For more about specifying packages,
see 'go help packages'.
`[1:],
}

func init() {
	listCommand.Run = cmdList // break init cycle
}

var (
	listWithTest     = listCommand.Flag.Bool("test", false, "show test dependencies too")
	listWithTestOnly = listCommand.Flag.Bool("testonly", false, "show test dependencies only")
	listModules      = listCommand.Flag.Bool("m", false, "show dependencies at module level")
)

func cmdList(_ *Command, args []string) int {
	if len(args) == 0 {
		args = []string{"."}
	}
	pkgs, pkgDeps, err := deps(args, *listWithTest || *listWithTestOnly)
	if err != nil {
		errorf("%v", err)
		return 1
	}
	if *listModules {
		mods := make(dependencyMap)
		for ipName, testOnly := range pkgDeps {
			p := pkgs[ipName]
			if p.Module == nil && !p.Standard {
				continue
			}
			var pattern string
			switch {
			case p.Standard && strings.Contains(ipName, "test"):
				pattern = "stdtest"
			case p.Standard:
				pattern = "std"
			default:
				pattern = p.Module.Path + "/..."
			}
			mods.markVisited(pattern, testOnly)
		}
		pkgDeps = mods
	}
	listPkgs := make([]string, 0, len(pkgDeps))
	for ipName, testOnly := range pkgDeps {
		if *listWithTestOnly && !testOnly {
			continue
		}
		listPkgs = append(listPkgs, ipName)
	}
	sort.Strings(listPkgs)
	w := bufio.NewWriter(os.Stdout)
	for _, p := range listPkgs {
		w.WriteString(p)
		w.WriteByte('\n')
	}
	w.Flush()
	return 0
}
