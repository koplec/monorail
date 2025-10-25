package result_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/koplec/monorail/result"
)

func TestOkErrAndBasicOps(t *testing.T) {
	ok := result.Ok(42)
	if !ok.IsOk() || ok.IsErr() {
		t.Fatalf("expected Ok")
	}
	if v := ok.Unwrap(); v != 42 {
		t.Fatalf("Unwrap mismatch")
	}

	if err := ok.Error(); err != nil {
		t.Fatalf("unexpected error")
	}

	e := result.Err[int](errors.New("int error"))
	if !e.IsErr() || e.IsOk() {
		t.Fatalf("expected Err")
	}

	v, err := e.Value()
	if err == nil {
		t.Fatalf("expected error")
	}
	if v != 0 {
		t.Fatalf("expected zero value")
	}
}

func TestErrPanicsOnNil(t *testing.T) {
	defer func() {
		// panicが発生したら、recover()の値として、panicの引数が返る
		// panicが発生すると、nilでないはず
		if recover() == nil {
			t.Fatalf("Err(nil) must panic")
		}
	}()

	_ = result.Err[int](nil)
}

func TestUnwrapOrAndMapFlatMap(t *testing.T) {
	r := result.Ok("100")
	parsed := result.Map(r, func(s string) (int, error) {
		var n int
		_, err := fmt.Sscanf(s, "%d", &n)
		return n, err
	})
	if parsed.IsErr() || parsed.Unwrap() != 100 {
		t.Fatalf("map failed")
	}

	doubled := result.AndThen(parsed, func(num int) result.Result[int] {
		return result.Ok(num * 2)
	})
	if doubled.IsErr() || doubled.Unwrap() != 200 {
		t.Fatalf("AndThen failed")
	}

	errR := result.Map(result.Ok("oops"), func(string) (int, error) {
		return 0, errors.New("ng")
	})
	if !errR.IsErr() || errR.UnwrapOr(9) != 9 {
		t.Fatalf("UnwrapOr fallback failed")
	}
}
