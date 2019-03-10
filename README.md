# modcop: policing your Go dependencies

WORK IN PROGRESS: this README represents an end goal - it's
not implemented yet!

The modcop command checks that module dependencies do not stray outside
well defined boundaries. It expects that the set of allowed dependencies
will be checked in along with the source code, and modcop will be called
as part of a CI check, so any addition to the set of allowed dependencies
will need to be explicitly vetted as part of the code review process.

The set of dependencies checked by modcop does not include dependencies
of test code outside of the module being checked - it differs from
the module dependencies calculated by the go command in this respect.
It also does not consider versions - if it is desired to police allowed
module versions, a vetted Go proxy could be used.

Any package can include a go.dep file to limit its dependencies to
the packages listed in that; if a package does not have a go.dep file,
modcop falls back to checking the go.dep file in the top level directory
of the module.

Note: This leaves us in the situation that it's not possible to have a
go.dep file that applies *only* to the top level package in a module. An
alternative approach might be to define that any go.dep file applies to
all the packages hierarchically beneath it.

## modcop subcommands

modcop allow [-m] [packages]

	Allow adds the given package dependencies to the
	local package's go.dep file, or to the module's top level go.dep
	file if the -m flag is specified. The go.dep file will be
	created if it does not already exist.

	The packages will be added to the test or the build section
	of the go.dep file depending on how they are actually used
	within the module. For example, running "modcop allow gopkg.in/check.v1"
	on a package that only imports that package in tests will add
	it to the test section, not the build section.

	The packages support a very limited subset of the patterns
	allowed by the Go tool: a package name may end in "/...",
	which will allow all packages that would be matched by
	the Go tool with that pattern. The "/..." may only appear
	at the end of a package name.
	
	The special name "std" can be used to allow all packages
	in the standard library that don't contain the string "test".
	This excludes packages like testing and net/http/httptest,
	which must be mentioned explicitly.

	If no packages are specified, any dependencies that are
	used but not mentioned in the go.dep file will be added,
	and any dependencies in the go.dep file that aren't used
	will be removed. If the -m flag is given, all additions to the
	go.dep file will also allow all packages in the imported module.

modcop check [-with file]...

	Check checks that the dependencies of all packages in the
	current module are conformant with their go.dep files.
	If there's a package without a go.dep file, the module's root-level
	go.dep file will be used; if that does not exist, no checking
	will be done.
	
	If the -with flag is specified, its file should be a go.dep file that
	is also used to check the dependencies. The dependencies
	should be OK with both go.dep files for check to succeed.
	This can be used to provide a global "whitelist" of acceptable packages
	or modules.
	
	For example, if an organization provides this global whitelist in /org/go.dep:
	
		build (
			std
			golang.org/x/tools/go/...
		)
		test (
			testing
		)
	
	and a package's go.dep file is:
	
		build (
			golang.org/x/tools/...
		)
		test (
			testing
		)
	
	Then a call of "modcop check -with /org/go.dep" would fail if the package's
	build dependencies include golang.org/x/tools/blog, or the package's test
	dependencies include net/httptest.

modcop list [-m] [-test] [packages]

	List shows the set of current dependencies of the listed packages.
	By default, only build dependencies are listed; if the
	-test flag is provided, test dependencies will be listed too.
	If the -m flag is provided, dependencies will be printed at the
	module level and version information will be printed too
	(in the same format used by "go list -m").

	List accepts the same package wildcard patterns that are accepted
	by the go list command.


## The go.dep file format:

A go.dep file has a similar syntax to the go.mod file.
It is line-oriented, with // comments but no /* */ comments.
Each line holds a single directive, made up of a verb followed
by arguments. For example:

	build github.com/juju/utils/parallel
	build golang.org/x/tools/...
	build std
	test github.com/frankban/quicktest
	test github.com/google/go-cmp/...

The verbs are:

	build, to allow a package to be depended on by a build.
	test, to allow a package to be depended on by a test build.

A package name may refer to a single package, or to all packages
hierarchically beneath it (with the `/...` suffix). Like the Go command,
this pattern does not include submodules.

The leading verb can be factored out of adjacent lines to create a block,
like in Go imports:

	build (
		github.com/juju/utils/parallel
		golang.org/x/tools/...
		std
	)
	test (
		github.com/frankban/quicktest
		github.com/google/go-cmp/...
	)
