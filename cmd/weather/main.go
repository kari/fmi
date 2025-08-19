package main

import (
	"fmt"
	"os"

	"github.com/kari/fmi"
)

func main() {
	if len(os.Args) > 1 {
		fmt.Println(fmi.Weather(os.Args[1]))
	} else {
		fmt.Println("Usage:", os.Args[0], "<paikka>")
	}
}
