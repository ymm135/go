// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mips

import (
	"cmd_local/compile/internal/gc"
	"cmd_local/compile/internal/ssa"
	"cmd_local/internal/obj/mips"
	"cmd_local/internal/objabi"
)

func Init(arch *gc.Arch) {
	arch.LinkArch = &mips.Linkmips
	if objabi.GOARCH == "mipsle" {
		arch.LinkArch = &mips.Linkmipsle
	}
	arch.REGSP = mips.REGSP
	arch.MAXWIDTH = (1 << 31) - 1
	arch.SoftFloat = (objabi.GOMIPS == "softfloat")
	arch.ZeroRange = zerorange
	arch.Ginsnop = ginsnop
	arch.Ginsnopdefer = ginsnop
	arch.SSAMarkMoves = func(s *gc.SSAGenState, b *ssa.Block) {}
	arch.SSAGenValue = ssaGenValue
	arch.SSAGenBlock = ssaGenBlock
}
