package main

import (
	"fmt"

	"github.com/koplec/monorail/result"
)

func main() {
	parse := func(s string) (int, error) {
		var n int
		_, err := fmt.Sscanf(s, "%d", &n)
		return n, err
	}

	out := result.
		AndThen(
			result.Map(result.Ok("100"), parse),
			func(n int) result.Result[int] {
				return result.Ok(n * 3)
			})

	if out.IsErr() {
		fmt.Println("error:", out.Error())
		return
	}
	fmt.Println("OK:", out.Unwrap()) // 300

}
