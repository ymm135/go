// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"cmd_local/internal/objabi"
	"cmd_local/internal/sys"
	"cmd_local/link/internal/amd64"
	"cmd_local/link/internal/arm"
	"cmd_local/link/internal/arm64"
	"cmd_local/link/internal/ld"
	"cmd_local/link/internal/mips"
	"cmd_local/link/internal/mips64"
	"cmd_local/link/internal/ppc64"
	"cmd_local/link/internal/riscv64"
	"cmd_local/link/internal/s390x"
	"cmd_local/link/internal/wasm"
	"cmd_local/link/internal/x86"
	"fmt"
	"os"
)

// The bulk of the linker implementation lives in cmd_local/link/internal/ld.
// Architecture-specific code lives in cmd_local/link/internal/GOARCH.
//
// Program initialization:
//
// Before any argument parsing is done, the Init function of relevant
// architecture package is called. The only job done in Init is
// configuration of the architecture-specific variables.
//
// Then control flow passes to ld.Main, which parses flags, makes
// some configuration decisions, and then gives the architecture
// packages a second chance to modify the linker's configuration
// via the ld.Arch.Archinit function.

func main() {
	var arch *sys.Arch
	var theArch ld.Arch

	switch objabi.GOARCH {
	default:
		fmt.Fprintf(os.Stderr, "link: unknown architecture %q\n", objabi.GOARCH)
		os.Exit(2)
	case "386":
		arch, theArch = x86.Init()
	case "amd64":
		arch, theArch = amd64.Init()
	case "arm":
		arch, theArch = arm.Init()
	case "arm64":
		arch, theArch = arm64.Init()
	case "mips", "mipsle":
		arch, theArch = mips.Init()
	case "mips64", "mips64le":
		arch, theArch = mips64.Init()
	case "ppc64", "ppc64le":
		arch, theArch = ppc64.Init()
	case "riscv64":
		arch, theArch = riscv64.Init()
	case "s390x":
		arch, theArch = s390x.Init()
	case "wasm":
		arch, theArch = wasm.Init()
	}
	ld.Main(arch, theArch)
}
