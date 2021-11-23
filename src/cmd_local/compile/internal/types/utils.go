// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"cmd_local/internal/obj"
	"fmt"
)

const BADWIDTH = -1000000000

// The following variables must be initialized early by the frontend.
// They are here to break import cycles.
// TODO(gri) eliminate these dependencies.
var (
	Widthptr    int
	Dowidth     func(*Type)
	Fatalf      func(string, ...interface{})
	Sconv       func(*Sym, int, int) string       // orig: func sconv(s *Sym, flag FmtFlag, mode fmtMode) string
	Tconv       func(*Type, int, int) string      // orig: func tconv(t *Type, flag FmtFlag, mode fmtMode) string
	FormatSym   func(*Sym, fmt.State, rune, int)  // orig: func symFormat(sym *Sym, s fmt.State, verb rune, mode fmtMode)
	FormatType  func(*Type, fmt.State, rune, int) // orig: func typeFormat(t *Type, s fmt.State, verb rune, mode fmtMode)
	TypeLinkSym func(*Type) *obj.LSym
	Ctxt        *obj.Link

	FmtLeft     int
	FmtUnsigned int
	FErr        int
)

func (s *Sym) String() string {
	return Sconv(s, 0, FErr)
}

func (sym *Sym) Format(s fmt.State, verb rune) {
	FormatSym(sym, s, verb, FErr)
}

func (t *Type) String() string {
	// The implementation of tconv (including typefmt and fldconv)
	// must handle recursive types correctly.
	return Tconv(t, 0, FErr)
}

// ShortString generates a short description of t.
// It is used in autogenerated method names, reflection,
// and itab names.
func (t *Type) ShortString() string {
	return Tconv(t, FmtLeft, FErr)
}

// LongString generates a complete description of t.
// It is useful for reflection,
// or when a unique fingerprint or hash of a type is required.
func (t *Type) LongString() string {
	return Tconv(t, FmtLeft|FmtUnsigned, FErr)
}

func (t *Type) Format(s fmt.State, verb rune) {
	FormatType(t, s, verb, FErr)
}

type bitset8 uint8

func (f *bitset8) set(mask uint8, b bool) {
	if b {
		*(*uint8)(f) |= mask
	} else {
		*(*uint8)(f) &^= mask
	}
}
