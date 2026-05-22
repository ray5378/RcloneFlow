package runnercli

import (
	"context"
	"sync"
	"testing"
)

func TestRegisterCancelFn(t *testing.T) {
	runCancelFns = sync.Map{}
	ctx, cancel := context.WithCancel(context.Background())
	registerCancelFn(1, cancel)
	_ = ctx

	v, ok := runCancelFns.Load(int64(1))
	if !ok {
		t.Fatal("expected cancel function to be registered")
	}
	if _, ok := v.(context.CancelFunc); !ok {
		t.Fatal("expected registered value to be context.CancelFunc")
	}
}

func TestUnregisterCancelFn(t *testing.T) {
	runCancelFns = sync.Map{}
	ctx, cancel := context.WithCancel(context.Background())
	registerCancelFn(2, cancel)
	unregisterCancelFn(2)

	if _, ok := runCancelFns.Load(int64(2)); ok {
		t.Fatal("expected cancel function to be unregistered")
	}

	cancel()
	_ = ctx
}

func TestUnregisterCancelFn_NotFound(t *testing.T) {
	runCancelFns = sync.Map{}
	unregisterCancelFn(99)
}

func TestCancelRun_Exists(t *testing.T) {
	runCancelFns = sync.Map{}
	ctx, cancel := context.WithCancel(context.Background())
	registerCancelFn(3, cancel)

	if !CancelRun(3) {
		t.Fatal("expected CancelRun to return true for registered run")
	}

	if ctx.Err() == nil {
		t.Fatal("expected context to be cancelled after CancelRun")
	}
}

func TestCancelRun_NotFound(t *testing.T) {
	runCancelFns = sync.Map{}
	if CancelRun(99) {
		t.Fatal("expected CancelRun to return false for unregistered run")
	}
}

func TestCancelRun_MultipleRuns(t *testing.T) {
	runCancelFns = sync.Map{}
	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx2, cancel2 := context.WithCancel(context.Background())
	registerCancelFn(10, cancel1)
	registerCancelFn(20, cancel2)

	CancelRun(10)
	if ctx1.Err() == nil {
		t.Fatal("expected ctx1 to be cancelled")
	}
	if ctx2.Err() != nil {
		t.Fatal("expected ctx2 to NOT be cancelled")
	}
}

func TestCancelRun_AlreadyCancelled(t *testing.T) {
	runCancelFns = sync.Map{}
	ctx, cancel := context.WithCancel(context.Background())
	registerCancelFn(5, cancel)

	CancelRun(5)
	_ = ctx
	if !CancelRun(5) {
		t.Fatal("expected CancelRun to return true on already cancelled context")
	}
}