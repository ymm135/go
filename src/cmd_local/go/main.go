// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate ./mkalldocs.sh

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"cmd_local/go/internal/base"
	"cmd_local/go/internal/bug"
	"cmd_local/go/internal/cfg"
	"cmd_local/go/internal/clean"
	"cmd_local/go/internal/doc"
	"cmd_local/go/internal/envcmd"
	"cmd_local/go/internal/fix"
	"cmd_local/go/internal/fmtcmd"
	"cmd_local/go/internal/generate"
	"cmd_local/go/internal/get"
	"cmd_local/go/internal/help"
	"cmd_local/go/internal/list"
	"cmd_local/go/internal/modcmd"
	"cmd_local/go/internal/modfetch"
	"cmd_local/go/internal/modget"
	"cmd_local/go/internal/modload"
	"cmd_local/go/internal/run"
	"cmd_local/go/internal/test"
	"cmd_local/go/internal/tool"
	"cmd_local/go/internal/trace"
	"cmd_local/go/internal/version"
	"cmd_local/go/internal/vet"
	"cmd_local/go/internal/work"
)

func init() {
	base.Go.Commands = []*base.Command{
		bug.CmdBug,
		work.CmdBuild,
		clean.CmdClean,
		doc.CmdDoc,
		envcmd.CmdEnv,
		fix.CmdFix,
		fmtcmd.CmdFmt,
		generate.CmdGenerate,
		modget.CmdGet,
		work.CmdInstall,
		list.CmdList,
		modcmd.CmdMod,
		run.CmdRun,
		test.CmdTest,
		tool.CmdTool,
		version.CmdVersion,
		vet.CmdVet,

		help.HelpBuildConstraint,
		help.HelpBuildmode,
		help.HelpC,
		help.HelpCache,
		help.HelpEnvironment,
		help.HelpFileType,
		modload.HelpGoMod,
		help.HelpGopath,
		get.HelpGopathGet,
		modfetch.HelpGoproxy,
		help.HelpImportPath,
		modload.HelpModules,
		modget.HelpModuleGet,
		modfetch.HelpModuleAuth,
		help.HelpPackages,
		modfetch.HelpPrivate,
		test.HelpTestflag,
		test.HelpTestfunc,
		modget.HelpVCS,
	}
}

func main() {
	_ = go11tag
	flag.Usage = base.Usage
	flag.Parse()
	log.SetFlags(0)

	args := flag.Args()
	if len(args) < 1 {
		base.Usage()
	}

	if args[0] == "get" || args[0] == "help" {
		if !modload.WillBeEnabled() {
			// Replace module-aware get with GOPATH get if appropriate.
			*modget.CmdGet = *get.CmdGet
		}
	}

	cfg.CmdName = args[0] // for error messages
	if args[0] == "help" {
		help.Help(os.Stdout, args[1:])
		return
	}

	// Diagnose common mistake: GOPATH==GOROOT.
	// This setting is equivalent to not setting GOPATH at all,
	// which is not what most people want when they do it.
	if gopath := cfg.BuildContext.GOPATH; filepath.Clean(gopath) == filepath.Clean(runtime.GOROOT()) {
		fmt.Fprintf(os.Stderr, "warning: GOPATH set to GOROOT (%s) has no effect\n", gopath)
	} else {
		for _, p := range filepath.SplitList(gopath) {
			// Some GOPATHs have empty directory elements - ignore them.
			// See issue 21928 for details.
			if p == "" {
				continue
			}
			// Note: using HasPrefix instead of Contains because a ~ can appear
			// in the middle of directory elements, such as /tmp/git-1.8.2~rc3
			// or C:\PROGRA~1. Only ~ as a path prefix has meaning to the shell.
			if strings.HasPrefix(p, "~") {
				fmt.Fprintf(os.Stderr, "go: GOPATH entry cannot start with shell metacharacter '~': %q\n", p)
				os.Exit(2)
			}
			if !filepath.IsAbs(p) {
				if cfg.Getenv("GOPATH") == "" {
					// We inferred $GOPATH from $HOME and did a bad job at it.
					// Instead of dying, uninfer it.
					cfg.BuildContext.GOPATH = ""
				} else {
					fmt.Fprintf(os.Stderr, "go: GOPATH entry is relative; must be absolute path: %q.\nFor more details see: 'go help gopath'\n", p)
					os.Exit(2)
				}
			}
		}
	}

	if fi, err := os.Stat(cfg.GOROOT); err != nil || !fi.IsDir() {
		fmt.Fprintf(os.Stderr, "go: cannot find GOROOT directory: %v\n", cfg.GOROOT)
		os.Exit(2)
	}

	// Set environment (GOOS, GOARCH, etc) explicitly.
	// In theory all the commands we invoke should have
	// the same default computation of these as we do,
	// but in practice there might be skew
	// This makes sure we all agree.
	cfg.OrigEnv = os.Environ()
	cfg.CmdEnv = envcmd.MkEnv()
	for _, env := range cfg.CmdEnv {
		if os.Getenv(env.Name) != env.Value {
			os.Setenv(env.Name, env.Value)
		}
	}

BigCmdLoop:
	for bigCmd := base.Go; ; {
		for _, cmd := range bigCmd.Commands {
			if cmd.Name() != args[0] {
				continue
			}
			if len(cmd.Commands) > 0 {
				bigCmd = cmd
				args = args[1:]
				if len(args) == 0 {
					help.PrintUsage(os.Stderr, bigCmd)
					base.SetExitStatus(2)
					base.Exit()
				}
				if args[0] == "help" {
					// Accept 'go mod help' and 'go mod help foo' for 'go help mod' and 'go help mod foo'.
					help.Help(os.Stdout, append(strings.Split(cfg.CmdName, " "), args[1:]...))
					return
				}
				cfg.CmdName += " " + args[0]
				continue BigCmdLoop
			}
			if !cmd.Runnable() {
				continue
			}
			cmd.Flag.Usage = func() { cmd.Usage() }
			if cmd.CustomFlags {
				args = args[1:]
			} else {
				base.SetFromGOFLAGS(&cmd.Flag)
				cmd.Flag.Parse(args[1:])
				args = cmd.Flag.Args()
			}
			ctx := maybeStartTrace(context.Background())
			ctx, span := trace.StartSpan(ctx, fmt.Sprint("Running ", cmd.Name(), " command"))
			cmd.Run(ctx, cmd, args)
			span.Done()
			base.Exit()
			return
		}
		helpArg := ""
		if i := strings.LastIndex(cfg.CmdName, " "); i >= 0 {
			helpArg = " " + cfg.CmdName[:i]
		}
		fmt.Fprintf(os.Stderr, "go %s: unknown command\nRun 'go help%s' for usage.\n", cfg.CmdName, helpArg)
		base.SetExitStatus(2)
		base.Exit()
	}
}

func init() {
	base.Usage = mainUsage
}

func mainUsage() {
	help.PrintUsage(os.Stderr, base.Go)
	os.Exit(2)
}

func maybeStartTrace(pctx context.Context) context.Context {
	if cfg.DebugTrace == "" {
		return pctx
	}

	ctx, close, err := trace.Start(pctx, cfg.DebugTrace)
	if err != nil {
		base.Fatalf("failed to start trace: %v", err)
	}
	base.AtExit(func() {
		if err := close(); err != nil {
			base.Fatalf("failed to stop trace: %v", err)
		}
	})

	return ctx
}
