// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main_test

import (
	"bytes"
	"os"
	"testing"

	"cmd_local/go/internal/help"
	"cmd_local/go/internal/modload"
)

func TestDocsUpToDate(t *testing.T) {
	t.Parallel()

	if !modload.Enabled() {
		t.Skipf("help.Help in GOPATH mode is configured by main.main")
	}

	buf := new(bytes.Buffer)
	// Match the command in mkalldocs.sh that generates alldocs.go.
	help.Help(buf, []string{"documentation"})
	data, err := os.ReadFile("alldocs.go")
	if err != nil {
		t.Fatalf("error reading alldocs.go: %v", err)
	}
	if !bytes.Equal(data, buf.Bytes()) {
		t.Errorf("alldocs.go is not up to date; run mkalldocs.sh to regenerate it")
	}
}
