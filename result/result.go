package result

import "fmt"

type Result[T any] struct {
	value T
	err   error
}

func Ok[T any](v T) Result[T] {
	return Result[T]{
		value: v, err: nil,
	}
}

func Err[T any](err error) Result[T] {
	if err == nil {
		panic("result.Err called with nil error")
	}
	var zero T //Tのゼロ値を入れる
	return Result[T]{
		value: zero, err: err,
	}
}

func (r Result[T]) IsOk() bool {
	return r.err == nil
}
func (r Result[T]) IsErr() bool {
	return r.err != nil
}

func (r Result[T]) Unwrap() T {
	if r.err != nil {
		panic(fmt.Sprintf("Unwrap on Err: %v", r.err))
	}
	return r.value
}

func (r Result[T]) UnwrapOr(def T) T {
	if r.err != nil {
		return def
	}
	return r.value
}

func (r Result[T]) Error() error {
	return r.err
}

func (r Result[T]) Value() (T, error) {
	return r.value, r.err
}

//
func Map[T, U any](r Result[T], f func(T) (U, error)) Result[U] {
	if r.err != nil {
		return Err[U](r.err)
	}
	u, err := f(r.value)
	if err != nil {
		return Err[U](err)
	}
	return Ok(u)
}

func FlatMap[T, U any](r Result[T], f func(T) Result[U]) Result[U] {
	if r.err != nil {
		return Err[U](r.err)
	}
	return f(r.value)
}

// FlatMapのsyntax sugar
// 正興なら、次に進むという意味
func AndThen[T, U any](r Result[T], f func(T) Result[U]) Result[U] {
	return FlatMap(r, f)
}

// Resultのスライスをまとめて、1つでもErrがあればそのErrを返す
func Collect[T any](rs []Result[T]) Result[[]T] {
	out := make([]T, len(rs))
	for i, r := range rs {
		if r.err != nil {
			return Err[[]T](r.err)
		}
		out[i] = r.value
	}
	return Ok(out)
}

//　AのスライスをからBのスライスに変換して、Resultで返す
func Traverse[A, B any](xs []A, fn func(A) Result[B]) Result[[]B] {
	ret := make([]Result[B], len(xs))
	for i, x := range xs {
		ret[i] = fn(x)
	}
	return Collect(ret)
}
