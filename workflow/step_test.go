package workflow_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/koplec/monorail/result"
	"github.com/koplec/monorail/workflow"
)

func stepStringToLen(ctx context.Context, in string) result.Result[int] {
	return result.Ok(len(in))
}

func stepLenToParity(ctx context.Context, in int) result.Result[string] {
	if in%2 == 0 {
		return result.Err[string](fmt.Errorf("even length"))
	}
	return result.Ok("odd")
}

func TestThen_Success(t *testing.T) {
	chain1 := workflow.Start(stepStringToLen)
	chain2 := workflow.Then(chain1, func(ctx context.Context, in int) result.Result[string] {
		return result.Ok("len=" + string(rune('0'+in)))
	})

	res := workflow.Run(context.Background(), chain2, "go")
	if res.IsErr() {
		t.Fatalf("expected success, got error : %v", res.Error())
	}
	if res.Unwrap() != "len=2" {
		t.Fatalf("undexpected value:%s", res.Unwrap())
	}
}

func TestThen_ErrorPropagation(t *testing.T) {
	fst := workflow.Start(stepStringToLen)
	chain := workflow.Then(fst, stepLenToParity)

	res := workflow.Run(context.Background(), chain, "abcd")
	if res.IsOk() {
		t.Fatalf("expected error, got value: %v", res.Unwrap())
	}
	if got := res.Error(); got == nil || got.Error() != "even length" {
		t.Fatalf("unexpected error:%v", got)
	}
}

func TestThen_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	//すぐにキャンセルする
	cancel() //simulate cancel before running

	start := workflow.Start(stepStringToLen)
	next := workflow.Then(start, stepLenToParity)

	res := workflow.Run(ctx, next, "monorail")
	if res.IsOk() {
		t.Fatalf("expected cancelation error, got value %v", res.Unwrap())
	}
	if !errors.Is(res.Error(), context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", res.Error())
	}
}

func TestOnError_HandlerInvoked(t *testing.T) {
	start := workflow.Start(stepStringToLen)
	chain := workflow.Then(start, stepLenToParity).OnError(
		func(err error) error {
			return errors.New("wrapped: " + err.Error())
		})

	res := workflow.Run(context.Background(), chain, "mono")
	if res.IsOk() {
		t.Fatalf("expected error, got value %v", res.Unwrap())
	}
	if got := res.Error(); got == nil || got.Error() != "wrapped: even length" {
		t.Fatalf("unexpected transformed error: %v", got)
	}
}

func TestOnError_NilHandler(t *testing.T) {
	start := workflow.Start(stepStringToLen)
	chain := workflow.Then(start, stepLenToParity).OnError(nil)

	res := workflow.Run(context.Background(), chain, "mono")
	if res.IsOk() {
		t.Fatalf("expected error, got value %v", res.Unwrap())
	}
	if got := res.Error(); got == nil || got.Error() != "even length" {
		t.Fatalf("expected original error: %v", got)
	}
}

// Then next argument cannot infer 0 so this test is not necessary
// func TestThen_NilStepGuard(t *testing.T) {
// 	start := workflow.Start(stepStringToLen)
// 	chain := workflow.Then(start, nil)

// 	res := workflow.Run(context.Background(), chain, "data")
// 	if res.IsOk() {
// 		t.Fatalf("expected guard error, got value %v", res.Unwrap())
// 	}
// 	if res.Error() == nil || res.Error().Error() != "workflow: nil step" {
// 		t.Fatalf("unexpected guard error: %v", res.Error())
// 	}
// }

func TestRun_ContextCancelsMidChain(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	slowStep := func(ctx context.Context, in string) result.Result[int] {
		select {
		case <-time.After(5 * time.Millisecond):
			return result.Ok(len(in))
		case <-ctx.Done():
			return result.Err[int](ctx.Err())
		}
	}

	start := workflow.Start(slowStep)
	chain := workflow.Then(start, func(ctx context.Context, in int) result.Result[string] {
		return result.Ok("len ok")
	})

	res := workflow.Run(ctx, chain, "xxx")
	if res.IsOk() {
		t.Fatalf("expected cancellation, got value: %v", res.Unwrap())
	}
	if err := res.Error(); err == nil ||
		!errors.Is(err, context.DeadlineExceeded) &&
			!errors.Is(err, context.Canceled) {
		t.Fatalf("unexpected error: %v", err)
	}
}
