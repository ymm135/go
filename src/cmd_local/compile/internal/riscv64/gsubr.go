// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package riscv64

import (
	"cmd_local/compile/internal/gc"
	"cmd_local/internal/obj"
	"cmd_local/internal/obj/riscv"
)

func ginsnop(pp *gc.Progs) *obj.Prog {
	// Hardware nop is ADD $0, ZERO
	p := pp.Prog(riscv.AADD)
	p.From.Type = obj.TYPE_CONST
	p.Reg = riscv.REG_ZERO
	p.To = obj.Addr{Type: obj.TYPE_REG, Reg: riscv.REG_ZERO}
	return p
}
