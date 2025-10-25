package parallel_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/koplec/monorail/parallel"
	"github.com/koplec/monorail/result"
)

func TestMap_OrderPreserved(t *testing.T) {
	ctx := context.Background()
	xs := []int{1, 2, 3, 4, 5}

	out := parallel.Map(
		ctx, xs,
		func(_ context.Context, i int, v int) result.Result[int] {
			//並行でもインデックスに対応した場所へ描かれることを確認
			return result.Ok(v * 10)
		},
	)

	if out.IsErr() {
		t.Fatalf("unexpected err: %v", out.Error())
	}

	got := out.Unwrap()
	want := []int{10, 20, 30, 40, 50}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order mismatch at %d: got=%v, want=%v", i, got[i], want[i])
		}
	}
}

func TestMap_ErrorPropagation(t *testing.T) {
	ctx := context.Background()
	xs := []int{1, 2, 3, 4}
	out := parallel.Map(
		ctx, xs, func(_ context.Context, _ int, v int) result.Result[int] {
			if v == 3 {
				return result.Err[int](errors.New("boom"))
			}
			return result.Ok(v)
		},
	)
	if out.IsOk() {
		t.Fatalf("expected error")
	}
}

func TestMapWithLimit_Equivalence(t *testing.T) {
	ctx := context.Background()
	xs := make([]int, 200)
	for i := range xs {
		xs[i] = i
	}

	fn := func(_ context.Context, _ int, v int) result.Result[int] {
		return result.Ok(v * v)
	}

	a := parallel.MapWithLimit(ctx, 0, xs, fn) //無制限
	b := parallel.MapWithLimit(ctx, 4, xs, fn) //制限あり
	if a.IsErr() || b.IsErr() {
		t.Fatalf("unexpected error:%v %v", a.Error(), b.Error())
	}

	ga, gb := a.Unwrap(), b.Unwrap()
	for i := range ga {
		if ga[i] != gb[i] {
			t.Fatalf("limit changed result at %d: %d vs %d", i, ga[i], gb[i])
		}
	}
}

func TestMap_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	xs := make([]int, 1000)
	for i := range xs {
		xs[i] = i
	}

	var started int32
	fn := func(ctx context.Context, _ int, _ int) result.Result[int] {
		atomic.AddInt32(&started, 1)

		select {
		case <-ctx.Done():
			return result.Err[int](ctx.Err())
		case <-time.After(10 * time.Millisecond):
			return result.Ok(1)
		}
	}

	go func() {
		time.Sleep(2 * time.Millisecond)
		cancel()
	}()

	out := parallel.MapWithLimit(ctx, 32, xs, fn)
	if out.IsOk() {
		t.Fatalf("expected cancellation error")
	}

	//ある程度のgoroutineは開始している
	if atomic.LoadInt32(&started) == 0 {
		t.Fatalf("no goroutine started?")
	}
}

func TestMap_Empty(t *testing.T) {
	ctx := context.Background()
	out := parallel.Map[int, int](
		ctx, nil,
		func(context.Context, int, int) result.Result[int] {
			return result.Ok(1)
		})
	if out.IsErr() {
		t.Fatalf("nil should slice be ok: %v", out.Error())
	}

	if len(out.Unwrap()) != 0 {
		t.Fatalf("want empty result")
	}
}
