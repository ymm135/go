// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package s390x

import (
	"cmd_local/compile/internal/gc"
	"cmd_local/internal/obj/s390x"
)

func Init(arch *gc.Arch) {
	arch.LinkArch = &s390x.Links390x
	arch.REGSP = s390x.REGSP
	arch.MAXWIDTH = 1 << 50

	arch.ZeroRange = zerorange
	arch.Ginsnop = ginsnop
	arch.Ginsnopdefer = ginsnop

	arch.SSAMarkMoves = ssaMarkMoves
	arch.SSAGenValue = ssaGenValue
	arch.SSAGenBlock = ssaGenBlock
}
