package fmi

import (
	"math"
	"testing"
)

func TestHumidex(t *testing.T) {
	var tests = []struct {
		t, td, h float64
	}{
		{30, 15, 34},
		{30, 25, 42},
	}
	for _, test := range tests {
		got := humidex(test.t, test.td)
		if math.Round(got) != test.h {
			t.Errorf("humidex(%.f, %.f) = %.2f; want %.f", test.t, test.td, got, test.h)
		}
	}
}

func TestWindChill(t *testing.T) {
	var tests = []struct {
		t, v, c float64
	}{
		{-20, 1.389, -24},
		{-20, 8.333, -33},
		{9, 4, 7},
	}
	for _, test := range tests {
		got := windChill(test.t, test.v)
		if math.Round(got) != test.c {
			t.Errorf("windChill(%.f, %.f) = %.2f; want %.f", test.t, test.v, got, test.c)
		}
	}
}

func TestSummerSimmer(t *testing.T) {
	got := SummerSimmer(14, 0)
	if got != 14 {
		t.Errorf("SummerSimmer should return temperature if below limit")
	}
	// TODO: Add test cases
}

func TestWindChillFmi(t *testing.T) {
	got := windChillFmi(-10.0, 0)
	if got != -10.0 {
		t.Errorf("Wind chill should equal temperature when wind speed is zero, got %.2f", got)
	}
	got = windChillFmi(-10.0, 1)
	if got >= -10.0 {
		t.Errorf("Wind chill should be less than temperature when wind speed is below 5 km/h (1.4 m/s), got %.2f", got)
	}
	got = windChillFmi(-10.0, 2)
	if got >= -10.0 {
		t.Errorf("Wind chill should be less than temperature when wind speed is above 5 km/h (1.4 m/s), got %.2f", got)
	}
	// TODO: Add test cases
}
