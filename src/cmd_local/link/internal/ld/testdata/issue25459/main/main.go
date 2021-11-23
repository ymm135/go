package main

import "cmd_local/link/internal/ld/testdata/issue25459/a"

var Glob int

func main() {
	a.Another()
	Glob += a.ConstIf() + a.CallConstIf()
}
