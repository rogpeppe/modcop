package main

import (
	"flag"
	"fmt"
	"os"

	"gopkg.in/errgo.v2/fmt/errors"
)

var (
	exitCode = 0
	cwd      = "."
)

var commands = []*Command{
	listCommand,
}

func main() {
	os.Exit(main1())
}

func main1() int {
	if dir, err := os.Getwd(); err == nil {
		cwd = dir
	} else {
		errorf("cannot get current working directory: %v", err)
	}
	flag.Usage = func() {
		mainUsage(os.Stderr)
	}
	flag.Parse()
	if flag.NArg() == 0 {
		mainUsage(os.Stderr)
		return 2
	}
	cmdName := flag.Arg(0)
	args := flag.Args()[1:]
	if cmdName == "help" {
		return runHelp(args)
	}
	for _, cmd := range commands {
		if cmd.Name() != cmdName {
			continue
		}
		cmd.Flag.Usage = func() { cmd.Usage() }
		cmd.Flag.Parse(args)
		rcode := cmd.Run(cmd, cmd.Flag.Args())
		return max(exitCode, rcode)
	}
	errorf("modcop %s: unknown command\nRun 'modcop help' for usage\n", cmdName)
	return 2
}

const debug = false

func errorf(f string, a ...interface{}) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf(f, a...))
	if debug {
		for _, arg := range a {
			if err, ok := arg.(error); ok {
				fmt.Fprintf(os.Stderr, "error: %s\n", errors.Details(err))
			}
		}
	}
	exitCode = 1
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
