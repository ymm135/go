// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !js

package main

import (
	"cmd_local/internal/traceviewer"
	"context"
	"internal_local/trace"
	"io"
	rtrace "runtime/trace"
	"strings"
	"sync"
	"testing"
	"time"
)

// stacks is a fake stack map populated for test.
type stacks map[uint64][]*trace.Frame

// add adds a stack with a single frame whose Fn field is
// set to the provided fname and returns a unique stack id.
func (s *stacks) add(fname string) uint64 {
	if *s == nil {
		*s = make(map[uint64][]*trace.Frame)
	}

	id := uint64(len(*s))
	(*s)[id] = []*trace.Frame{{Fn: fname}}
	return id
}

// TestGoroutineCount tests runnable/running goroutine counts computed by generateTrace
// remain in the valid range.
//   - the counts must not be negative. generateTrace will return an error.
//   - the counts must not include goroutines blocked waiting on channels or in syscall.
func TestGoroutineCount(t *testing.T) {
	w := trace.NewWriter()
	w.Emit(trace.EvBatch, 0, 0)  // start of per-P batch event [pid, timestamp]
	w.Emit(trace.EvFrequency, 1) // [ticks per second]

	var s stacks

	// In this test, we assume a valid trace contains EvGoWaiting or EvGoInSyscall
	// event for every blocked goroutine.

	// goroutine 10: blocked
	w.Emit(trace.EvGoCreate, 1, 10, s.add("pkg.f1"), s.add("main.f1")) // [timestamp, new goroutine id, new stack id, stack id]
	w.Emit(trace.EvGoWaiting, 1, 10)                                   // [timestamp, goroutine id]

	// goroutine 20: in syscall
	w.Emit(trace.EvGoCreate, 1, 20, s.add("pkg.f2"), s.add("main.f2"))
	w.Emit(trace.EvGoInSyscall, 1, 20) // [timestamp, goroutine id]

	// goroutine 30: runnable
	w.Emit(trace.EvGoCreate, 1, 30, s.add("pkg.f3"), s.add("main.f3"))

	w.Emit(trace.EvProcStart, 2, 0) // [timestamp, thread id]

	// goroutine 40: runnable->running->runnable
	w.Emit(trace.EvGoCreate, 1, 40, s.add("pkg.f4"), s.add("main.f4"))
	w.Emit(trace.EvGoStartLocal, 1, 40)          // [timestamp, goroutine id]
	w.Emit(trace.EvGoSched, 1, s.add("main.f4")) // [timestamp, stack]

	res, err := trace.Parse(w, "")
	if err != nil {
		t.Fatalf("failed to parse test trace: %v", err)
	}
	res.Stacks = s // use fake stacks.

	params := &traceParams{
		parsed:  res,
		endTime: int64(1<<63 - 1),
	}

	// Use the default viewerDataTraceConsumer but replace
	// consumeViewerEvent to intercept the ViewerEvents for testing.
	c := viewerDataTraceConsumer(io.Discard, 0, 1<<63-1)
	c.consumeViewerEvent = func(ev *traceviewer.Event, _ bool) {
		if ev.Name == "Goroutines" {
			cnt := ev.Arg.(*goroutineCountersArg)
			if cnt.Runnable+cnt.Running > 2 {
				t.Errorf("goroutine count=%+v; want no more than 2 goroutines in runnable/running state", cnt)
			}
			t.Logf("read %+v %+v", ev, cnt)
		}
	}

	// If the counts drop below 0, generateTrace will return an error.
	if err := generateTrace(params, c); err != nil {
		t.Fatalf("generateTrace failed: %v", err)
	}
}

func TestGoroutineFilter(t *testing.T) {
	// Test that we handle state changes to selected goroutines
	// caused by events on goroutines that are not selected.

	var s stacks

	w := trace.NewWriter()
	w.Emit(trace.EvBatch, 0, 0)  // start of per-P batch event [pid, timestamp]
	w.Emit(trace.EvFrequency, 1) // [ticks per second]

	// goroutine 10: blocked
	w.Emit(trace.EvGoCreate, 1, 10, s.add("pkg.f1"), s.add("main.f1")) // [timestamp, new goroutine id, new stack id, stack id]
	w.Emit(trace.EvGoWaiting, 1, 10)                                   // [timestamp, goroutine id]

	// goroutine 20: runnable->running->unblock 10
	w.Emit(trace.EvGoCreate, 1, 20, s.add("pkg.f2"), s.add("main.f2"))
	w.Emit(trace.EvGoStartLocal, 1, 20)                    // [timestamp, goroutine id]
	w.Emit(trace.EvGoUnblockLocal, 1, 10, s.add("pkg.f2")) // [timestamp, goroutine id, stack]
	w.Emit(trace.EvGoEnd, 1)                               // [timestamp]

	// goroutine 10: runnable->running->block
	w.Emit(trace.EvGoStartLocal, 1, 10)         // [timestamp, goroutine id]
	w.Emit(trace.EvGoBlock, 1, s.add("pkg.f3")) // [timestamp, stack]

	res, err := trace.Parse(w, "")
	if err != nil {
		t.Fatalf("failed to parse test trace: %v", err)
	}
	res.Stacks = s // use fake stacks

	params := &traceParams{
		parsed:  res,
		endTime: int64(1<<63 - 1),
		gs:      map[uint64]bool{10: true},
	}

	c := viewerDataTraceConsumer(io.Discard, 0, 1<<63-1)
	if err := generateTrace(params, c); err != nil {
		t.Fatalf("generateTrace failed: %v", err)
	}
}

func TestPreemptedMarkAssist(t *testing.T) {
	w := trace.NewWriter()
	w.Emit(trace.EvBatch, 0, 0)  // start of per-P batch event [pid, timestamp]
	w.Emit(trace.EvFrequency, 1) // [ticks per second]

	var s stacks
	// goroutine 9999: running -> mark assisting -> preempted -> assisting -> running -> block
	w.Emit(trace.EvGoCreate, 1, 9999, s.add("pkg.f1"), s.add("main.f1")) // [timestamp, new goroutine id, new stack id, stack id]
	w.Emit(trace.EvGoStartLocal, 1, 9999)                                // [timestamp, goroutine id]
	w.Emit(trace.EvGCMarkAssistStart, 1, s.add("main.f1"))               // [timestamp, stack]
	w.Emit(trace.EvGoPreempt, 1, s.add("main.f1"))                       // [timestamp, stack]
	w.Emit(trace.EvGoStartLocal, 1, 9999)                                // [timestamp, goroutine id]
	w.Emit(trace.EvGCMarkAssistDone, 1)                                  // [timestamp]
	w.Emit(trace.EvGoBlock, 1, s.add("main.f2"))                         // [timestamp, stack]

	res, err := trace.Parse(w, "")
	if err != nil {
		t.Fatalf("failed to parse test trace: %v", err)
	}
	res.Stacks = s // use fake stacks

	params := &traceParams{
		parsed:  res,
		endTime: int64(1<<63 - 1),
	}

	c := viewerDataTraceConsumer(io.Discard, 0, 1<<63-1)

	marks := 0
	c.consumeViewerEvent = func(ev *traceviewer.Event, _ bool) {
		if strings.Contains(ev.Name, "MARK ASSIST") {
			marks++
		}
	}
	if err := generateTrace(params, c); err != nil {
		t.Fatalf("generateTrace failed: %v", err)
	}

	if marks != 2 {
		t.Errorf("Got %v MARK ASSIST events, want %v", marks, 2)
	}
}

func TestFoo(t *testing.T) {
	prog0 := func() {
		ctx, task := rtrace.NewTask(context.Background(), "ohHappyDay")
		rtrace.Log(ctx, "", "log before task ends")
		task.End()
		rtrace.Log(ctx, "", "log after task ends") // log after task ends
	}
	if err := traceProgram(t, prog0, "TestFoo"); err != nil {
		t.Fatalf("failed to trace the program: %v", err)
	}
	res, err := parseTrace()
	if err != nil {
		t.Fatalf("failed to parse the trace: %v", err)
	}
	annotRes, _ := analyzeAnnotations()
	var task *taskDesc
	for _, t := range annotRes.tasks {
		if t.name == "ohHappyDay" {
			task = t
			break
		}
	}
	if task == nil {
		t.Fatal("failed to locate expected task event")
	}

	params := &traceParams{
		parsed:    res,
		mode:      modeTaskOriented,
		startTime: task.firstTimestamp() - 1,
		endTime:   task.lastTimestamp() + 1,
		tasks:     []*taskDesc{task},
	}

	c := viewerDataTraceConsumer(io.Discard, 0, 1<<63-1)

	var logBeforeTaskEnd, logAfterTaskEnd bool
	c.consumeViewerEvent = func(ev *traceviewer.Event, _ bool) {
		if ev.Name == "log before task ends" {
			logBeforeTaskEnd = true
		}
		if ev.Name == "log after task ends" {
			logAfterTaskEnd = true
		}
	}
	if err := generateTrace(params, c); err != nil {
		t.Fatalf("generateTrace failed: %v", err)
	}
	if !logBeforeTaskEnd {
		t.Error("failed to find 'log before task ends'")
	}
	if !logAfterTaskEnd {
		t.Error("failed to find 'log after task ends'")
	}

}

func TestDirectSemaphoreHandoff(t *testing.T) {
	prog0 := func() {
		var mu sync.Mutex
		var wg sync.WaitGroup
		mu.Lock()
		// This is modeled after src/sync/mutex_test.go to trigger Mutex
		// starvation mode, in which the goroutine that calls Unlock hands off
		// both the semaphore and its remaining time slice. See issue 36186.
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < 100; i++ {
					mu.Lock()
					time.Sleep(100 * time.Microsecond)
					mu.Unlock()
				}
			}()
		}
		mu.Unlock()
		wg.Wait()
	}
	if err := traceProgram(t, prog0, "TestDirectSemaphoreHandoff"); err != nil {
		t.Fatalf("failed to trace the program: %v", err)
	}
	_, err := parseTrace()
	if err != nil {
		t.Fatalf("failed to parse the trace: %v", err)
	}
}
