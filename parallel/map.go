package parallel

import (
	"context"

	"github.com/koplec/monorail/result"
	"golang.org/x/sync/errgroup"
)

func Map[A, B any](
	ctx context.Context,
	xs []A,
	fn func(context.Context, int, A) result.Result[B],
) result.Result[[]B] {
	return MapWithLimit(ctx, 0, xs, fn)
}

func MapWithLimit[A, B any](
	ctx context.Context,
	limit int,
	xs []A,
	fn func(context.Context, int, A) result.Result[B],
) result.Result[[]B] {
	out := make([]result.Result[B], len(xs))
	eg, ctx := errgroup.WithContext(ctx)

	if limit > 0 {
		eg.SetLimit(limit)
	}

	// var mu sync.Mutex //outへの競合防止

	for i, x := range xs {
		ii, xx := i, x
		eg.Go(func() error {
			r := fn(ctx, ii, xx)
			// mu.Lock() 同じout[ii]への書き込みが重複しないので、lock不要
			out[ii] = r //iへのアクセスだけど、ロック必要？
			// mu.Unlock()
			if r.IsErr() {
				return r.Error()
			}
			return nil
		})
	}

	//先に失敗したErrがあった場合は返す
	if err := eg.Wait(); err != nil {
		return result.Err[[]B](err)
	}

	return result.Collect(out)

}
