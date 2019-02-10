package main

import (
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/errgo.v2/fmt/errors"
)

func deps(pkgNames []string, withTests bool) (build, test []*goListPackage, err error) {
	pkgs, err := goListDeps(pkgNames, withTests)
	if err != nil {
		return nil, nil, errors.Wrap(err)
	}
	pkgMap := make(map[string]*goListPackage)
	for _, p := range pkgs {
		pkgMap[p.ImportPath] = p
	}
	// testOnly holds a map from package import path to whether that package
	// is used for testing only (true) or just the build (false).
	testOnly := make(map[string]bool)
	for _, p := range pkgs {
		if !p.DepOnly {
			if err := addDeps(p, pkgMap, isTestPackage(p), true, withTests, testOnly); err != nil {
				return nil, nil, err
			}
		}
	}
	for ipName, isTest := range testOnly {
		p := pkgMap[ipName]
		if !p.DepOnly {
			// It's one of the initial packages, so we don't count it
			// as a dependency.
			continue
		}
		if isInternal(p.ImportPath) {
			// It's an internal package - ignore it.
			continue
		}
		if isTest {
			test = append(test, p)
		} else {
			build = append(build, p)
		}
	}
	sort.Slice(build, func(i, j int) bool {
		return build[i].ImportPath < build[j].ImportPath
	})
	sort.Slice(test, func(i, j int) bool {
		return test[i].ImportPath < test[j].ImportPath
	})
	return build, test, nil
}

func isTestPackage(p *goListPackage) bool {
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

func addDeps(p *goListPackage, pkgs map[string]*goListPackage, isTest, isRoot bool, withTests bool, testOnly map[string]bool) error {
	if isTest {
		if _, ok := testOnly[p.ImportPath]; ok {
			// It's already present as a build or test dependency.
			return nil
		}
	} else {
		if isTest, ok := testOnly[p.ImportPath]; ok && !isTest {
			// It's already present as a build dependency.
			return nil
		}
	}
	testOnly[p.ImportPath] = isTest
	addImports := func(ipNames []string, isTest bool) error {
		for _, ipName := range ipNames {
			if ipName == "unsafe" && p.Standard {
				// The standard library is always allowed access to unsafe.
				continue
			}
			ip := pkgs[ipName]
			if ip == nil {
				return errors.Notef(nil, nil, "cannot find package entry for %v", ipName)
			}
			if err := addDeps(ip, pkgs, isTest, false, withTests, testOnly); err != nil {
				return errors.Wrap(err)
			}
		}
		return nil
	}
	if err := addImports(p.Imports, isTest); err != nil {
		return errors.Wrap(err)
	}
	if withTests && isRoot {
		if err := addImports(p.TestImports, true); err != nil {
			return errors.Wrap(err)
		}
		if err := addImports(p.XTestImports, true); err != nil {
			return errors.Wrap(err)
		}
	}
	return nil
}
