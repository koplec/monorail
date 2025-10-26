package workflow

import (
	"context"

	"github.com/koplec/monorail/result"
)

type SequentialStep[I, O any] struct {
	fn Step[I, O]
}

func StepOf[I, O any](fn Step[I, O]) SequentialStep[I, O] {
	if fn == nil {
		panic("workflow: nil step")
	}
	return SequentialStep[I, O]{
		fn: fn,
	}

}

// まずは単一型で検討
// TODO: 型が変わるステップ
func Sequential[T any](ctx context.Context, initial T, first SequentialStep[T, T], rest ...SequentialStep[T, T]) result.Result[T] {
	chain := Start(first.fn)
	for _, step := range rest {
		chain = Then(chain, step.fn)
	}
	return Run(ctx, chain, initial)
}

func Parallel[I, O any](ctx context.Context, input I, branches ...Step[I, O]) result.Result[[]O] {
	//TODO notimplemented
	panic("workflow: Parallel not implemented yet")
}
