package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rogpeppe/go-internal/goproxytest"
	"github.com/rogpeppe/go-internal/gotooltest"
	"github.com/rogpeppe/go-internal/testscript"
	"gopkg.in/errgo.v2/fmt/errors"
)

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"modcop": main1,
	}))
}

func TestScript(t *testing.T) {
	p := testscript.Params{
		Dir:      "testdata",
		TestWork: true,
	}
	if err := gotooltest.Setup(&p); err != nil {
		t.Fatalf("failed to setup go tool for: %v", err)
	}
	origSetup := p.Setup
	p.Setup = func(env *testscript.Env) error {
		origSetup(env)
		// Set up the proxy server if there are
		modDir := filepath.Join(env.WorkDir, ".gomodproxy")
		if _, err := os.Stat(modDir); err != nil {
			return errors.Newf("no modules found in %s", modDir)
			// Ensure that the proxy directory exists so that goproxytest
			// doesn't error.
			if err := os.Mkdir(modDir, 0777); err != nil {
				return errors.Wrap(err)
			}
		}
		srv, err := goproxytest.NewServer(modDir, "")
		if err != nil {
			return errors.Wrap(err)
		}
		env.Defer(srv.Close)
		env.Vars = append(env.Vars, "GOPROXY="+srv.URL)
		return nil
	}
	testscript.Run(t, p)
}
