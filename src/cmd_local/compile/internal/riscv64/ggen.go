// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package riscv64

import (
	"cmd_local/compile/internal/gc"
	"cmd_local/internal/obj"
	"cmd_local/internal/obj/riscv"
)

func zeroRange(pp *gc.Progs, p *obj.Prog, off, cnt int64, _ *uint32) *obj.Prog {
	if cnt == 0 {
		return p
	}

	// Adjust the frame to account for LR.
	off += gc.Ctxt.FixedFrameSize()

	if cnt < int64(4*gc.Widthptr) {
		for i := int64(0); i < cnt; i += int64(gc.Widthptr) {
			p = pp.Appendpp(p, riscv.AMOV, obj.TYPE_REG, riscv.REG_ZERO, 0, obj.TYPE_MEM, riscv.REG_SP, off+i)
		}
		return p
	}

	if cnt <= int64(128*gc.Widthptr) {
		p = pp.Appendpp(p, riscv.AADDI, obj.TYPE_CONST, 0, off, obj.TYPE_REG, riscv.REG_A0, 0)
		p.Reg = riscv.REG_SP
		p = pp.Appendpp(p, obj.ADUFFZERO, obj.TYPE_NONE, 0, 0, obj.TYPE_MEM, 0, 0)
		p.To.Name = obj.NAME_EXTERN
		p.To.Sym = gc.Duffzero
		p.To.Offset = 8 * (128 - cnt/int64(gc.Widthptr))
		return p
	}

	// Loop, zeroing pointer width bytes at a time.
	// ADD	$(off), SP, T0
	// ADD	$(cnt), T0, T1
	// loop:
	// 	MOV	ZERO, (T0)
	// 	ADD	$Widthptr, T0
	//	BNE	T0, T1, loop
	p = pp.Appendpp(p, riscv.AADD, obj.TYPE_CONST, 0, off, obj.TYPE_REG, riscv.REG_T0, 0)
	p.Reg = riscv.REG_SP
	p = pp.Appendpp(p, riscv.AADD, obj.TYPE_CONST, 0, cnt, obj.TYPE_REG, riscv.REG_T1, 0)
	p.Reg = riscv.REG_T0
	p = pp.Appendpp(p, riscv.AMOV, obj.TYPE_REG, riscv.REG_ZERO, 0, obj.TYPE_MEM, riscv.REG_T0, 0)
	loop := p
	p = pp.Appendpp(p, riscv.AADD, obj.TYPE_CONST, 0, int64(gc.Widthptr), obj.TYPE_REG, riscv.REG_T0, 0)
	p = pp.Appendpp(p, riscv.ABNE, obj.TYPE_REG, riscv.REG_T0, 0, obj.TYPE_BRANCH, 0, 0)
	p.Reg = riscv.REG_T1
	gc.Patch(p, loop)
	return p
}
