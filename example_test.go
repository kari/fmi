package fmi_test

import (
	"context"
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

// ExampleCurrent shows how to read the observations as structured data.
func ExampleCurrent() {
	w, err := fmi.Current(context.Background(), "Turku")
	if errors.Is(err, fmi.ErrUnknownPlace) {
		log.Fatal("tuntematon paikka")
	} else if err != nil {
		log.Fatal(err)
	}
	fmt.Println(w)
	fmt.Printf("lämpötila %.1f°C\n", w.Observations.Temperature)
}
