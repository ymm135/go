// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !386
// +build !amd64

package cpu

// Name returns the CPU name given by the vendor
// if it can be read directly from memory or by CPU instructions.
// If the CPU name can not be determined an empty string is returned.
//
// Implementations that use the Operating System (e.g. sysctl or /sys/)
// to gather CPU information for display should be placed in internal_local/sysinfo.
func Name() string {
	// "A CPU has no name".
	return ""
}
