// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Export guts for testing.
// Since testing imports os and os imports internal_local/poll,
// the internal_local/poll tests can not be in package poll.

package poll

var Consume = consume

type FDMutex struct {
	fdMutex
}

func (mu *FDMutex) Incref() bool {
	return mu.incref()
}

func (mu *FDMutex) IncrefAndClose() bool {
	return mu.increfAndClose()
}

func (mu *FDMutex) Decref() bool {
	return mu.decref()
}

func (mu *FDMutex) RWLock(read bool) bool {
	return mu.rwlock(read)
}

func (mu *FDMutex) RWUnlock(read bool) bool {
	return mu.rwunlock(read)
}
