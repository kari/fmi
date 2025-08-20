package main

import (
	"fmt"
	"os"

	"github.com/kari/fmi"
)

var Version = "development"

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("Usage:", os.Args[0], "<paikka>")
	} else if os.Args[1] == "version" {
		fmt.Println("Version:", Version)
	} else {
		fmt.Println(fmi.Weather(os.Args[1]))
	}
}
