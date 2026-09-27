package fmi_test

import (
	"errors"
	"fmt"
	"log"

	"github.com/kari/fmi"
)

// ExampleWeather shows how to fetch and print the latest observations.
func ExampleWeather() {
	weather, err := fmi.Weather("Turku")
	if errors.Is(err, fmi.ErrUnknownPlace) {
		log.Fatal("tuntematon paikka")
	} else if err != nil {
		log.Fatal(err)
	}
	fmt.Println(weather)
}
