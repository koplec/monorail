package result_test

import (
	"errors"
	"testing"

	"github.com/koplec/monorail/result"
)

func TestCollectAndTraverse(t *testing.T) {
	ok := []result.Result[int]{
		result.Ok(1),
		result.Ok(2),
		result.Ok(3),
	}
	got := result.Collect(ok)
	if got.IsErr() || len(got.Unwrap()) != 3 {
		t.Fatalf("collect failed")
	}

	errs := []result.Result[int]{
		result.Ok(1),
		result.Err[int](errors.New("xxx")),
	}
	if result.Collect(errs).IsOk() {
		t.Fatal("collect expected error")
	}

	tr := result.Traverse(
		[]int{1, 2, 3},
		func(n int) result.Result[int] {
			return result.Ok(n * n)
		},
	)
	if tr.IsErr() || tr.Unwrap()[2] != 9 {
		t.Fatalf("traverse failed")
	}
}
