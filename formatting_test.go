package fmi

import (
	"math"
	"testing"
)

func TestHumidexScale(t *testing.T) {
	var tests = []struct {
		h  float64
		s  string
		ok bool
	}{
		{0, "", false},
		{-1, "", false},
		{math.NaN(), "", false},
		{20, "mukava", true},
		{29, "lämmin", true},
		{34, "kuuma", true},
		{39, "tukala", true},
		{40.1, "erittäin tukala", true},
		{100, "erittäin tukala", true},
	}
	for _, test := range tests {
		got, ok := humidexScale(test.h)
		if got != test.s {
			t.Errorf("humidexScale(%.f) = '%s'; want '%s'", test.h, got, test.s)
		}
		if ok != test.ok {
			t.Errorf("humidexScale(%.f) = '%t'; want '%t'", test.h, ok, test.ok)
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

func TestWindChillScale(t *testing.T) {
	var tests = []struct {
		w  float64
		s  string
		ok bool
	}{
		{0, "", false},
		{-1, "", false},
		{math.NaN(), "", false},
		{-25, "erittäin kylmä", true},
		{-35, "paleltumisvaara", true},
		{-60, "suuri paleltumisvaara", true},
		{-100, "suuri paleltumisvaara", true},
	}
	for _, test := range tests {
		got, ok := windChillScale(test.w)
		if got != test.s {
			t.Errorf("windChillScale(%.f) = '%s'; want '%s'", test.w, got, test.s)
		}
		if ok != test.ok {
			t.Errorf("windChillScale(%.f) = '%t'; want '%t'", test.w, ok, test.ok)
		}
	}
}

func TestCloudCover(t *testing.T) {
	var tests = []struct {
		d  float64
		s  string
		ok bool
	}{
		{0, "selkeää", true},
		{-1, "", false},
		{math.NaN(), "", false},
		{1, "selkeää", true},
		{2, "melko selkeää", true},
		{3, "melko selkeää", true},
		{4, "puolipilvistä", true},
		{5, "puolipilvistä", true},
		{6, "melko pilvistä", true},
		{7, "melko pilvistä", true},
		{8, "pilvistä", true},
		{9, "taivas ei näy", true},
		{10, "", false},
	}
	for _, test := range tests {
		got, ok := cloudCover(test.d)
		if got != test.s {
			t.Errorf("cloudCover(%.f) = '%s'; want '%s'", test.d, got, test.s)
		}
		if ok != test.ok {
			t.Errorf("cloudCover(%.f) = '%t'; want '%t'", test.d, ok, test.ok)
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
