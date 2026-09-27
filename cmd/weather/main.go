package main

import (
	"fmt"
	"os"

	"github.com/kari/fmi"
)

var Version = "development"

func main() {
	if len(os.Args) <= 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s <paikka>\n", os.Args[0])
		os.Exit(2)
	}
	if os.Args[1] == "version" {
		fmt.Println("Version:", Version)
		return
	}

	weather, err := fmi.Weather(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(weather)
}
