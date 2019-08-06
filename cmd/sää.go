package main

import (
	"fmt"
	"os"

	"github.com/kari/fmi"
)

func main() {
	fmt.Println(fmi.Weather(os.Args[1]))
}
