package main

import (
	"context"
	"fmt"
	"time"

	"github.com/koplec/monorail/parallel"
	"github.com/koplec/monorail/result"
)

func main() {
	ctx := context.Background()
	xs := []int{1, 2, 3, 4, 5}

	fn := func(ctx context.Context, i, v int) result.Result[int] {
		time.Sleep(time.Duration(200*v) * time.Millisecond)
		fmt.Printf("worker %d done (val=%d)\n", i, v)
		return result.Ok(v * v)
	}

	out := parallel.MapWithLimit(ctx, 2, xs, fn)
	if out.IsErr() {
		fmt.Println("error:", out.Error())
		return
	}

	fmt.Println("squares:", out.Unwrap())
}
