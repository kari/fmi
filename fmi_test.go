package fmi

import (
	"fmt"
	"testing"
)

func TestWeather(t *testing.T) {
	Weather("Helsinki")
	fmt.Printf(Weather("lappi"))
}
