// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package fix implements the ``go fix'' command.
package fix

import (
	"cmd_local/go/internal/base"
	"cmd_local/go/internal/cfg"
	"cmd_local/go/internal/load"
	"cmd_local/go/internal/modload"
	"cmd_local/go/internal/str"
	"context"
	"fmt"
	"os"
)

var CmdFix = &base.Command{
	Run:       runFix,
	UsageLine: "go fix [packages]",
	Short:     "update packages to use new APIs",
	Long: `
Fix runs the Go fix command on the packages named by the import paths.

For more about fix, see 'go doc cmd_local/fix'.
For more about specifying packages, see 'go help packages'.

To run fix with specific options, run 'go tool fix'.

See also: go fmt, go vet.
	`,
}

func runFix(ctx context.Context, cmd *base.Command, args []string) {
	pkgs := load.PackagesAndErrors(ctx, args)
	w := 0
	for _, pkg := range pkgs {
		if pkg.Error != nil {
			base.Errorf("%v", pkg.Error)
			continue
		}
		pkgs[w] = pkg
		w++
	}
	pkgs = pkgs[:w]

	printed := false
	for _, pkg := range pkgs {
		if modload.Enabled() && pkg.Module != nil && !pkg.Module.Main {
			if !printed {
				fmt.Fprintf(os.Stderr, "go: not fixing packages in dependency modules\n")
				printed = true
			}
			continue
		}
		// Use pkg.gofiles instead of pkg.Dir so that
		// the command only applies to this package,
		// not to packages in subdirectories.
		files := base.RelPaths(pkg.InternalAllGoFiles())
		base.Run(str.StringList(cfg.BuildToolexec, base.Tool("fix"), files))
	}
}
