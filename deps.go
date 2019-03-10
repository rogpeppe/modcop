package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/errgo.v2/fmt/errors"
)

type depCtx struct {
	pkgs      map[string]*goListPackage
	withTests bool

	// deps holds dependency map of all the packages
	// found so far.
	deps dependencyMap

	// When true, standard library dependencies will be
	// traversed recursively too.
	addStandard bool
}

// deps returns the list of build and test-only dependencies of the
// packages with the given import paths. Paths may contain pattern
// wildcards as supported by the Go tool.
//
// The full dependency graph is pruned by omitting any dependencies of
// external test dependencies. The dependency traversal is also bounded
// by the standard library. An entry will returned for each standard
// library package directly referred to from non-standard-library code -
// dependencies are not traversed further than that.
//
// The packages directly matched by pkgNames will not themselves be added.
func deps(pkgNames []string, withTests bool) (map[string]*goListPackage, dependencyMap, error) {
	pkgs, err := goListDeps(pkgNames, withTests)
	if err != nil {
		return nil, nil, errors.Wrap(err)
	}
	ctx := &depCtx{
		withTests: withTests,
		pkgs:      make(map[string]*goListPackage),
		deps:      make(dependencyMap),
	}
	// Add all packages to the packages map.
	for _, p := range pkgs {
		ctx.pkgs[p.ImportPath] = p
	}

	// Find the set of dependencies of all listed packages
	// and add them to ctx.testOnly.
	for _, p := range pkgs {
		if !p.DepOnly && !isTestGeneratedPackage(p) {
			if err := ctx.addDeps(p, isTestGeneratedPackage(p), true); err != nil {
				return nil, nil, err
			}
		}
	}
	// add the dependencies we've found to the appropriate
	// list for returning.
	for ipName := range ctx.deps {
		p := ctx.pkgs[ipName]
		if !p.DepOnly || isTestGeneratedPackage(p) || isInternal(p.ImportPath) {
			// It's one of the initial package or it's
			// an internal package we want to ignore,
			// so remove it from the dependency graph.
			// TODO perhaps leave internal imports
			// from root modules alone?
			delete(ctx.deps, ipName)
		}
	}
	return ctx.pkgs, ctx.deps, nil
}

// isTestGeneratedPackage reports whether the given package
// is generated just for tests and this should not be considered
// part of the user-visible dependency tree.
func isTestGeneratedPackage(p *goListPackage) bool {
	if p.ForTest != "" {
		return true
	}
	if !strings.HasSuffix(p.ImportPath, ".test") {
		return false
	}
	// The pseudo-package created by the go tool
	// has a Go file that looks like an absolute path and
	// doesn't have a go suffix.
	// TODO is there a better way to check this?
	for _, f := range p.GoFiles {
		if filepath.IsAbs(f) || !strings.HasSuffix(f, ".go") {
			return true
		}
	}
	return false
}

func isInternal(importPath string) bool {
	// TODO more efficient check.
	return strings.HasPrefix(importPath, "internal/") || strings.Contains(importPath, "/internal/") || strings.HasSuffix(importPath, "/internal")
}

// addDeps recursively adds an entry for the given package and all its dependencies
// to ctx.testOnly. isTest holds whether the package has been reached by test
// dependencies only; isRoot holds whether the dependency is a top level
// dependency passed to addDeps.
func (ctx *depCtx) addDeps(p *goListPackage, isTest, isRoot bool) error {
	importPath := p.ImportPath
	if p.Module != nil && p.Module.Replace != nil {
		// The module for the current package has been
		// replaced, so add it as a dependency using its replaced
		// name rather than the import path that was used.
		importPath1 := strings.TrimPrefix(importPath, p.Module.Path)
		if len(importPath1) == len(importPath) {
			return fmt.Errorf("package %s does not seem to be in module %s", importPath, p.Module.Path)
		}
		importPath = p.Module.Replace.Path + importPath1
		if _, ok := ctx.pkgs[importPath]; !ok {
			// There's no entry for the replacement package in the
			// dependencies, so add it.
			ctx.pkgs[importPath] = p
		}
	}
	changed := ctx.deps.markVisited(importPath, isTest)
	if !changed {
		return nil
	}
	addImports := func(ipNames []string, isTest bool) error {
		for _, ipName := range ipNames {
			ip := ctx.pkgs[ipName]
			if ip == nil {
				return errors.Notef(nil, nil, "cannot find package entry for %v", ipName)
			}
			if err := ctx.addDeps(ip, isTest, false); err != nil {
				return errors.Wrap(err)
			}
		}
		return nil
	}
	// Avoid traversing further into the standard library than explicit
	// dependencies because we want to be able to use the same go.dep
	// file with multiple Go versions.
	if p.Standard && !ctx.addStandard {
		return nil
	}
	if err := addImports(p.Imports, isTest); err != nil {
		return errors.Wrap(err)
	}
	if ctx.withTests && isRoot {
		if err := addImports(p.TestImports, true); err != nil {
			return errors.Wrap(err)
		}
		if err := addImports(p.XTestImports, true); err != nil {
			return errors.Wrap(err)
		}
	}
	return nil
}

// A dependency map is used to represent a set
// of package dependencies keyed by package name.
//
// Each entry represents a dependency; true for
// test-only dependencies and false for build
// dependencies (they might also be test dependencies,
// but they're not exclusively test dependencies)
type dependencyMap map[string]bool

// isTestOnly reports whether the given importPath has
// been marked as a test-only dependency.
func (m dependencyMap) isTestOnly(importPath string) bool {
	return m[importPath]
}

// markVisited marks that the given package has been visited as
// a test dependency. It reports whether the visited entry has
// changed.
func (m dependencyMap) markVisited(importPath string, isTest bool) (changed bool) {
	testOnly, ok := m[importPath]
	switch {
	case !ok:
		m[importPath] = isTest
		return true
	case testOnly && !isTest:
		// It was only in tests but now it's moved
		// to the inner circle.
		m[importPath] = false
		return true
	}
	return false
}
