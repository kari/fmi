package main

import (
	"flag"
	"fmt"

	"github.com/kari/fmi"
)

var place = flag.String("place", "Helsinki", "search weather for place")

func main() {
	flag.Parse()

	if weather, err := fmi.Weather(*place); err == nil {
		fmt.Println(weather)
	}

}
