package parallel

import (
	"context"
	"fmt"

	"github.com/koplec/monorail/partial"
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

type recordedResult[U any] struct {
	res result.Result[U]
	set bool
}

func MapPartialWithLimit[T, U any](
	ctx context.Context,
	limit int,
	items []T,
	fn func(context.Context, int, T) result.Result[U],
) partial.PartialResult[U] {
	if len(items) == 0 {
		return partial.PartialResult[U]{}
	}

	out := make([]recordedResult[U], len(items))
	eg, ctx := errgroup.WithContext(ctx)

	if limit > 0 {
		eg.SetLimit(limit)
	}

	for i, item := range items {
		ii, it := i, item

		eg.Go(func() error {
			r := fn(ctx, ii, it)
			out[ii] = recordedResult[U]{res: r, set: true}
			if r.IsErr() {
				//エラーの場合はoutにErrをセット
				return r.Error()
			}
			//成功の場合はoutにValueをセット
			return nil
		})
	}

	//先に失敗したErrがあった場合はエラーに含ませる
	waitErr := eg.Wait()
	errors := make([]partial.ElemError, 0, len(items)+1)
	values := make([]partial.Elem[U], 0, len(items))
	if waitErr != nil {
		errors = append(errors, partial.ElemError{Index: -1, Err: waitErr})
	}

	//成功したValueと失敗したErrをPartialResultにまとめて返す
	for i, rec := range out {
		if !rec.set {
			//cancelや早期終了で未処理だったものを補足しておく
			errors = append(errors, partial.ElemError{Index: i, Err: fmt.Errorf("unprocessed")})
			continue
		}
		res := rec.res
		if res.IsOk() {
			values = append(values, partial.Elem[U]{
				Index: i,
				Value: res.Unwrap(),
			})
		} else {
			errors = append(errors, partial.ElemError{
				Index: i,
				Err:   res.Error(),
			})
		}
	}
	return partial.PartialResult[U]{
		Values: values,
		Errors: errors,
	}
}
