package main

import (
	"fmt"
	"os"

	"github.com/kari/fmi"
)

func main() {
	if len(os.Args) > 1 {
		fmt.Println(fmi.Weather(os.Args[1]))
	}
}
