package partial_test

import (
	"errors"
	"testing"

	"github.com/koplec/monorail/partial"
	"github.com/koplec/monorail/result"
)

func TestCombine_AllOK(t *testing.T) {
	rs := []result.Result[int]{
		result.Ok(1), result.Ok(2), result.Ok(3),
	}
	pr := partial.Combine(rs)

	if pr.HasError() {
		t.Fatalf("unexpected errors: %v", pr.Errors)
	}

	if pr.Ok() != 3 || pr.Failed() != 0 {
		t.Fatalf("summary mismatch: %v", pr.Summary())
	}

	if got := pr.ValuesOnly(); len(got) != 3 || got[0] != 1 || got[1] != 2 || got[2] != 3 {
		t.Fatalf("values-only mismatch: %#v", got)
	}
}

func TestCombine_Mixed(t *testing.T) {
	e2 := errors.New("boom2")
	rs := []result.Result[int]{
		result.Ok(10),
		result.Err[int](e2),
		result.Ok(30),
		result.Err[int](errors.New("boom3")),
	}

	pr := partial.Combine(rs)

	if !pr.HasError() || pr.Ok() != 2 || pr.Failed() != 2 {
		t.Fatalf("summary mismatch : %v", pr.Summary())
	}

	//indexが保存されている
	if pr.Values[0].Index != 0 || pr.Values[1].Index != 2 {
		t.Fatalf("value indexes lost: %+v", pr.Values)
	}
	if pr.Errors[0].Index != 1 || pr.Errors[1].Index != 3 {
		t.Fatalf("error indexes lost: %+v", pr.Errors)
	}

	//FirstError
	if pr.FirstError() == nil {
		t.Fatalf("expected first error")
	}

	//ErrorsOnly
	es := pr.ErrorsOnly()
	if len(es) != 2 || !errors.Is(es[0], e2) {
		t.Fatalf("errors-only mismatch: %+v", es)
	}
}

func TestReorder(t *testing.T) {
	rs := []result.Result[string]{
		result.Ok("a"), result.Err[string](errors.New("x")), result.Ok("c"),
	}
	pr := partial.Combine(rs)

	re := pr.Reorder(3)
	if re[0].IsErr() || re[1].IsOk() || re[2].IsErr() {
		t.Fatalf("reorder mismatch: %#v", re)
	}

	//値が戻っていること
	if re[0].Unwrap() != "a" || re[2].Unwrap() != "c" {
		t.Fatalf("reordered values mismatch")
	}
}

func TestElemError_Unwrap(t *testing.T) {
	base := errors.New("root")
	ee := partial.ElemError{Index: 5, Err: base}
	if !errors.Is(ee, base) {
		t.Fatalf("unwrap failed")
	}
}

func TestCombine_Empty(t *testing.T) {
	rs := []result.Result[int]{}
	pr := partial.Combine(rs)

	if pr.Ok() != 0 || pr.Failed() != 0 || pr.HasError() {
		t.Fatalf("empty mismatch: %+v", pr.Summary())
	}

	re := pr.Reorder(0)
	if len(re) != 0 {
		t.Fatalf("reorder empty mismatch")
	}
}

func TestReorder_LengthMustMatch(t *testing.T) {
	rs := []result.Result[int]{
		result.Ok(1),
		result.Err[int](errors.New("x")),
	}
	pr := partial.Combine(rs)

	out := pr.Reorder(len(rs))
	if out[0].IsErr() || out[1].IsOk() {
		t.Fatalf("reorder mismatch")
	}
}
