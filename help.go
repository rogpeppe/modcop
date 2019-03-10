package main

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"strings"
)

// runHelp implements the 'help' command.
func runHelp(args []string) int {
	if len(args) == 0 {
		mainUsage(os.Stdout)
		return 0
	}
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "modcop help %s: too many arguments\n", strings.Join(args, " "))
		return 2
	}
	t := template.Must(template.New("").Parse(commandHelpTemplate))
	for _, c := range commands {
		if c.Name() == args[0] {
			if err := t.Execute(os.Stdout, c); err != nil {
				errorf("cannot write usage output: %v", err)
			}
			return 0
		}
	}
	fmt.Fprintf(os.Stderr, "modcop help %s: unknown command\n", args[0])
	return 2
}

func mainUsage(f io.Writer) {
	t := template.Must(template.New("").Parse(mainHelpTemplate))
	if err := t.Execute(f, commands); err != nil {
		errorf("cannot write usage output: %v", err)
	}
}

var mainHelpTemplate = `
The modcop command checks that package dependencies do not stray outside
well defined boundaries. It expects that the set of allowed dependencies
will be checked in along with the source code, and modcop will be called
as part of a CI check, so any addition to the set of allowed dependencies
will need to be explicitly vetted as part of the code review process.

The set of dependencies checked by modcop does not include dependencies
of test code outside of the packages being checked - it differs from
the module dependencies calculated by the go command in this respect.
It also does not consider versions - if it is desired to police allowed
module versions, a vetted Go proxy could be used.

Any package can include a go.dep file to limit its dependencies to
the packages listed in that; if a package does not have a go.dep file,
modcop falls back to checking the go.dep file in the top level directory
of the module.

Usage:

	modcop <command> [arguments]

The commands are:
{{range .}}
	{{.Name | printf "%-11s"}} {{.Short}}{{end}}

Use "modcop help <command>" for more information about a command.
`[1:]

var commandHelpTemplate = `
usage: {{.UsageLine}}

{{.Long}}`[1:]
