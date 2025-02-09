// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package modcmd

import (
	"context"
	"fmt"
	"strings"

	"cmd_local/go/internal/base"
	"cmd_local/go/internal/imports"
	"cmd_local/go/internal/modload"

	"golang.org/x/mod/module"
)

var cmdWhy = &base.Command{
	UsageLine: "go mod why [-m] [-vendor] packages...",
	Short:     "explain why packages or modules are needed",
	Long: `
Why shows a shortest path in the import graph from the main module to
each of the listed packages. If the -m flag is given, why treats the
arguments as a list of modules and finds a path to any package in each
of the modules.

By default, why queries the graph of packages matched by "go list all",
which includes tests for reachable packages. The -vendor flag causes why
to exclude tests of dependencies.

The output is a sequence of stanzas, one for each package or module
name on the command line, separated by blank lines. Each stanza begins
with a comment line "# package" or "# module" giving the target
package or module. Subsequent lines give a path through the import
graph, one package per line. If the package or module is not
referenced from the main module, the stanza will display a single
parenthesized note indicating that fact.

For example:

	$ go mod why golang.org/x/text/language golang.org/x/text/encoding
	# golang.org/x/text/language
	rsc.io/quote
	rsc.io/sampler
	golang.org/x/text/language

	# golang.org/x/text/encoding
	(main module does not need package golang.org/x/text/encoding)
	$

See https://golang.org/ref/mod#go-mod-why for more about 'go mod why'.
	`,
}

var (
	whyM      = cmdWhy.Flag.Bool("m", false, "")
	whyVendor = cmdWhy.Flag.Bool("vendor", false, "")
)

func init() {
	cmdWhy.Run = runWhy // break init cycle
	base.AddModCommonFlags(&cmdWhy.Flag)
}

func runWhy(ctx context.Context, cmd *base.Command, args []string) {
	modload.ForceUseModules = true
	modload.RootMode = modload.NeedRoot

	loadOpts := modload.PackageOpts{
		Tags:          imports.AnyTags(),
		LoadTests:     !*whyVendor,
		SilenceErrors: true,
		UseVendorAll:  *whyVendor,
	}

	if *whyM {
		listU := false
		listVersions := false
		listRetractions := false
		for _, arg := range args {
			if strings.Contains(arg, "@") {
				base.Fatalf("go mod why: module query not allowed")
			}
		}
		mods := modload.ListModules(ctx, args, listU, listVersions, listRetractions)
		byModule := make(map[module.Version][]string)
		_, pkgs := modload.LoadPackages(ctx, loadOpts, "all")
		for _, path := range pkgs {
			m := modload.PackageModule(path)
			if m.Path != "" {
				byModule[m] = append(byModule[m], path)
			}
		}
		sep := ""
		for _, m := range mods {
			best := ""
			bestDepth := 1000000000
			for _, path := range byModule[module.Version{Path: m.Path, Version: m.Version}] {
				d := modload.WhyDepth(path)
				if d > 0 && d < bestDepth {
					best = path
					bestDepth = d
				}
			}
			why := modload.Why(best)
			if why == "" {
				vendoring := ""
				if *whyVendor {
					vendoring = " to vendor"
				}
				why = "(main module does not need" + vendoring + " module " + m.Path + ")\n"
			}
			fmt.Printf("%s# %s\n%s", sep, m.Path, why)
			sep = "\n"
		}
	} else {
		// Resolve to packages.
		matches, _ := modload.LoadPackages(ctx, loadOpts, args...)

		modload.LoadPackages(ctx, loadOpts, "all") // rebuild graph, from main module (not from named packages)

		sep := ""
		for _, m := range matches {
			for _, path := range m.Pkgs {
				why := modload.Why(path)
				if why == "" {
					vendoring := ""
					if *whyVendor {
						vendoring = " to vendor"
					}
					why = "(main module does not need" + vendoring + " package " + path + ")\n"
				}
				fmt.Printf("%s# %s\n%s", sep, path, why)
				sep = "\n"
			}
		}
	}
}
