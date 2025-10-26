package workflow

import (
	"context"
	"fmt"

	"github.com/koplec/monorail/result"
)

// Step expresses procedure function type
type Step[I, O any] func(context.Context, I) result.Result[O]

// Chain structure
type Chain[I, O any] struct {
	run Step[I, O]
}

// OnError is sugar syntax for function OnError
func (c Chain[I, O]) OnError(handler func(error) error) Chain[I, O] {
	return OnError(c, handler)
}

// Start returns a Chain wrapping the first Step
func Start[I, O any](step Step[I, O]) Chain[I, O] {
	return Chain[I, O]{run: step}
}

// Then returns a Chain to add Chain and next Step
func Then[I, Mid, O any](chain Chain[I, Mid], next Step[Mid, O]) Chain[I, O] {
	if next == nil {
		return Chain[I, O]{
			run: func(ctx context.Context, input I) result.Result[O] {
				return result.Err[O](fmt.Errorf("workflow: nil step"))
			},
		}
	}
	return Chain[I, O]{
		run: func(ctx context.Context, input I) result.Result[O] {
			if err := ctx.Err(); err != nil {
				return result.Err[O](err)
			}

			mid := chain.run(ctx, input)
			if err := mid.Error(); err != nil {
				return result.Err[O](err)
			}
			//実行後にcontextエラーがないかをチェック
			if err := ctx.Err(); err != nil {
				return result.Err[O](err)
			}

			// 次を呼ぶ
			return next(ctx, mid.Unwrap())
		},
	}
}

// OnError returns Chain which included error handler
func OnError[I, O any](chain Chain[I, O], handler func(error) error) Chain[I, O] {
	return Chain[I, O]{
		run: func(ctx context.Context, input I) result.Result[O] {

			if err := ctx.Err(); err != nil {
				transformed := transform(handler, err)
				return result.Err[O](transformed)
			}

			out := chain.run(ctx, input)

			if err := out.Error(); err != nil {
				transformed := transform(handler, err)
				return result.Err[O](transformed)
			}
			return out
		},
	}
}

// Run runs chain run actually and returns result.Result
func Run[I, O any](ctx context.Context, c Chain[I, O], input I) result.Result[O] {
	if err := ctx.Err(); err != nil {
		return result.Err[O](ctx.Err())
	}

	ret := c.run(ctx, input)

	if err := ctx.Err(); err != nil {
		return result.Err[O](ctx.Err())
	}

	return ret
}

// transform run handler and wrap error
func transform(handler func(error) error, err error) error {
	if handler == nil {
		return err
	}
	if transformed := handler(err); transformed != nil {
		return transformed
	}
	return err
}
