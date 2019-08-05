package main

import (
	"flag"
	"fmt"

	"github.com/kari/fmi"
)

var place = flag.String("place", "Helsinki", "search weather for place")

func main() {
	flag.Parse()

	fmt.Println(fmi.Weather(*place))
}
