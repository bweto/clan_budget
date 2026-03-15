package main

import (
"fmt"
"github.com/brianvoe/gofakeit/v7"
)

func main() {
	fmt.Println(gofakeit.Float64Range(-3.0, 3.0))
}
