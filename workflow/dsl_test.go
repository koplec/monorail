package workflow_test

import (
	"context"
	"errors"
	"testing"

	"github.com/koplec/monorail/result"
	"github.com/koplec/monorail/workflow"
)

func TestStepOf_PanicsOnNil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("StepOf should panic when given nil step")
		}
	}()
	workflow.StepOf[string, string](nil)
}

func TestSequential_SingleStep(t *testing.T) {
	ctx := context.Background()
	step := workflow.StepOf(func(_ context.Context, in string) result.Result[string] {
		return result.Ok(in + "!")
	})

	res := workflow.Sequential(ctx, "go", step)
	if res.IsErr() {
		t.Fatalf("expected ok, got error: %v", res.Error())
	}
	if got := res.Unwrap(); got != "go!" {
		t.Fatalf("unexpected value: %s", got)
	}
}

func TestSequential_MultiStep(t *testing.T) {
	ctx := context.Background()
	step1 := workflow.StepOf(func(_ context.Context, in string) result.Result[string] {
		return result.Ok(in + "1")
	})
	step2 := workflow.StepOf(func(_ context.Context, in string) result.Result[string] {
		return result.Ok(in + "2")
	})
	step3 := workflow.StepOf(func(_ context.Context, in string) result.Result[string] {
		return result.Ok(in + "3")
	})

	res := workflow.Sequential(ctx, "x", step1, step2, step3)
	if res.IsErr() {
		t.Fatalf("expected ok, got error: %v", res.Error())
	}
	if got := res.Unwrap(); got != "x123" {
		t.Fatalf("unexpected value: %s", got)
	}
}

func TestSequential_ShortCircuitOnError(t *testing.T) {
	ctx := context.Background()

	okStep := workflow.StepOf(func(_ context.Context, in string) result.Result[string] {
		return result.Ok(in + "-ok")
	})
	errStep := workflow.StepOf(func(_ context.Context, in string) result.Result[string] {
		return result.Err[string](context.Canceled)
	})
	unreachable := workflow.StepOf(func(_ context.Context, in string) result.Result[string] {
		t.Fatalf("should not be called after error")
		return result.Ok("noop")
	})

	res := workflow.Sequential(ctx, "start", okStep, errStep, unreachable)
	if res.IsOk() {
		t.Fatalf("expected error, got value: %v", res.Unwrap())
	}
	if !errors.Is(res.Error(), context.Canceled) {
		t.Fatalf("unexpected error: %v", res.Error())
	}
}
