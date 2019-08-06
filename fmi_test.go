package fmi

import (
	"math"
	"strings"
	"testing"
)

func TestWeather(t *testing.T) {
	s := Weather("Helsinki")
	if !strings.Contains(s, "Helsinki") {
		t.Errorf("Weather('Helsinki') should contain 'Helsinki', instead got '%s'", s)
	}
	s2 := Weather("Narnia")
	if s2 != "Säähavaintoja ei löytynyt" {
		t.Errorf("Weather('Narnia') should return 'Säähavaintoja ei löytynyt', instead got '%s'", s2)
	}
}

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

func TestHumidexScale(t *testing.T) {
	var tests = []struct {
		h float64
		s string
	}{
		{0, ""},
		{-1, ""},
		{math.NaN(), ""},
		{20, "mukava"},
		{29, "lämmin"},
		{34, "kuuma"},
		{39, "tukala"},
		{40.1, "erittäin tukala"},
		{100, "erittäin tukala"},
	}
	for _, test := range tests {
		got := humidexScale(test.h)
		if got != test.s {
			t.Errorf("humidexScale(%.f) = '%s'; want '%s'", test.h, got, test.s)
		}
	}
}

func TestWindDirection(t *testing.T) {
	var tests = []struct {
		d float64
		s string
	}{
		{0, "pohjois"},
		{-1, ""},
		{math.NaN(), ""},
		{45, "koillis"},
		{90, "itä"},
		{135, "kaakkois"},
		{180, "etelä"},
		{225, "lounais"},
		{270, "länsi"},
		{315, "luoteis"},
		{360, "pohjois"},
	}
	for _, test := range tests {
		got := windDirection(test.d)
		if got != test.s {
			t.Errorf("windDirection(%.f) = '%s'; want '%s'", test.d, got, test.s)
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

func TestWindChillScale(t *testing.T) {
	var tests = []struct {
		w float64
		s string
	}{
		{0, ""},
		{-1, ""},
		{math.NaN(), ""},
		{-25, "erittäin kylmä"},
		{-35, "paleltumisvaara"},
		{-60, "suuri paleltumisvaara"},
		{-100, "suuri paleltumisvaara"},
	}
	for _, test := range tests {
		got := windChillScale(test.w)
		if got != test.s {
			t.Errorf("windChillScale(%.f) = '%s'; want '%s'", test.w, got, test.s)
		}
	}
}

func TestCloudCover(t *testing.T) {
	var tests = []struct {
		d float64
		s string
	}{
		{0, "selkeää"},
		{-1, ""},
		{math.NaN(), ""},
		{1, "selkeää"},
		{2, "melko selkeää"},
		{3, "melko selkeää"},
		{4, "puolipilvistä"},
		{5, "puolipilvistä"},
		{6, "melko pilvistä"},
		{7, "melko pilvistä"},
		{8, "pilvistä"},
		{9, "taivas ei näy"},
		{10, ""},
	}
	for _, test := range tests {
		got := cloudCover(test.d)
		if got != test.s {
			t.Errorf("cloudCover(%.f) = '%s'; want '%s'", test.d, got, test.s)
		}
	}
}

func TestWindSpeed(t *testing.T) {
	var tests = []struct {
		v, d float64
		s    string
	}{
		{0, 0, "tyyntä"},
		{-1, 0, ""},
		{math.NaN(), 0, ""},
		{1, 0, "heikkoa pohjoistuulta"},
		{4.1, 0, "kohtalaista pohjoistuulta"},
		{14, 0, "navakkaa pohjoistuulta"},
		{21, 0, "kovaa pohjoistuulta"},
		{32, 0, "myrskyä"},
		{33, 0, "hirmumyrskyä"},
		{100, 0, "hirmumyrskyä"},
	}
	for _, test := range tests {
		got := windSpeed(test.v, test.d)
		if got != test.s {
			t.Errorf("windSpeed(%.f, %.f) = '%s'; want '%s'", test.v, test.d, got, test.s)
		}
	}
}
