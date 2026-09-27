package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kari/fmi"
)

var place = flag.String("place", "Helsinki", "search weather for place")

func main() {
	flag.Parse()

	weather, err := fmi.Weather(*place)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(weather)
}
