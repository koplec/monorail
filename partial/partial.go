package partial

import "github.com/koplec/monorail/result"

// どの要素で失敗したか
type ElemError struct {
	Index int
	Err   error
}

//errorに対応
func (e ElemError) Error() string {
	return e.Err.Error()
}

// errorsパッケージはUnwrapがあると、さらに中のエラーにさかのぼる
// errors.Isなどで利用
func (e ElemError) Unwrap() error {
	return e.Err
}

// どの要素で成功したか
type Elem[T any] struct {
	Index int
	Value T
}

// 成功値と失敗を同時に返す
type PartialResult[T any] struct {
	Values []Elem[T]
	Errors []ElemError
}

func (p PartialResult[T]) HasError() bool {
	return len(p.Errors) > 0
}

func (p PartialResult[T]) FirstError() error {
	if len(p.Errors) == 0 {
		return nil
	}
	return p.Errors[0].Err
}

func (p PartialResult[T]) Ok() int {
	return len(p.Values)
}

func (p PartialResult[T]) Failed() int {
	return len(p.Errors)
}

//値だけ取り出す
func (p PartialResult[T]) ValuesOnly() []T {
	vs := make([]T, 0, len(p.Values))
	for _, iv := range p.Values {
		vs = append(vs, iv.Value)
	}
	return vs
}

//エラーだけ取り出す
func (p PartialResult[T]) ErrorsOnly() []error {
	es := make([]error, 0, len(p.Errors))
	for _, ie := range p.Errors {
		es = append(es, ie.Err)
	}
	return es
}

// 件数サマリ
type Summary struct {
	Total  int
	Ok     int
	Failed int
}

func (p PartialResult[T]) Summary() Summary {
	oks := p.Ok()
	ers := p.Failed()
	return Summary{
		Total:  oks + ers,
		Ok:     oks,
		Failed: ers,
	}
}

// Combine
// ResultのスライスをPartialResultにまとめる
func Combine[T any](rs []result.Result[T]) PartialResult[T] {
	pr := PartialResult[T]{
		Values: make([]Elem[T], 0, len(rs)), //appendの再割り当て抑止のため、cap(len(rs))
		Errors: make([]ElemError, 0, len(rs)),
	}
	for i, r := range rs {
		if r.IsErr() {
			pr.Errors = append(
				pr.Errors,
				ElemError{
					Index: i,
					Err:   r.Error(),
				},
			)
		} else {
			pr.Values = append(
				pr.Values,
				Elem[T]{
					Index: i,
					Value: r.Unwrap(),
				},
			)
		}
	}

	return pr
}

func (pr PartialResult[T]) Reorder(
	total int, //元の長さがないとフィルタの後などに困る
) []result.Result[T] {
	if total < 0 {
		total = 0
	}
	rs := make([]result.Result[T], total)

	for _, er := range pr.Errors {
		rs[er.Index] = result.Err[T](er.Err)
	}

	for _, iv := range pr.Values {
		rs[iv.Index] = result.Ok[T](iv.Value)
	}

	return rs
}
