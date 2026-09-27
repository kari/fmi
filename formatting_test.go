package fmi

import (
	"bytes"
	"math"
	"testing"
)

// The format functions read only the fields they format, but every field
// they read must be set explicitly: unset fields are 0.0, which counts as
// a valid measure, so missing measures are NaN.
func TestFormatTemperature(t *testing.T) {
	var tests = []struct {
		obs Observations
		s   string
	}{
		{Observations{Temperature: math.NaN()}, "lämpötilatiedot puuttuvat"},
		{Observations{Temperature: 12.9, WindSpeed: math.NaN(), Humidity: math.NaN(), DewPoint: math.NaN(), Radiation: math.NaN()}, "lämpötila 12.9°C"},
		{Observations{Temperature: 12.9, WindSpeed: 5, Humidity: 50, DewPoint: math.NaN(), Radiation: math.NaN()}, "lämpötila 12.9°C (tuntuu kuin 9.7°C)"},
		{Observations{Temperature: 12.9, WindSpeed: 5, Humidity: 50, DewPoint: math.NaN(), Radiation: 500}, "lämpötila 12.9°C (tuntuu kuin 11.0°C)"},
		{Observations{Temperature: 22.9, WindSpeed: 5, Humidity: 70, DewPoint: 15, Radiation: math.NaN()}, "lämpötila 22.9°C (lämmin, tuntuu kuin 22.3°C)"},
		{Observations{Temperature: 22.9, WindSpeed: 5, Humidity: math.NaN(), DewPoint: 15, Radiation: math.NaN()}, "lämpötila 22.9°C (lämmin)"},
		{Observations{Temperature: -22.9, WindSpeed: 15, Humidity: 20, DewPoint: math.NaN(), Radiation: math.NaN()}, "lämpötila -22.9°C (paleltumisvaara, tuntuu kuin -36.5°C)"},
		{Observations{Temperature: -22.9, WindSpeed: 15, Humidity: math.NaN(), DewPoint: math.NaN(), Radiation: math.NaN()}, "lämpötila -22.9°C (paleltumisvaara)"},
	}

	buf := new(bytes.Buffer)
	for _, test := range tests {
		formatTemperature(buf, test.obs)
		if buf.String() != test.s {
			t.Errorf("got '%s', wanted '%s'", buf.String(), test.s)
		}
		buf.Reset()
	}
}

func TestFormatCloudCover(t *testing.T) {
	var tests = []struct {
		obs Observations
		s   string
	}{
		{Observations{CloudCover: math.NaN()}, ""},
		{Observations{CloudCover: 1}, ", selkeää"},
	}

	buf := new(bytes.Buffer)
	for _, test := range tests {
		formatCloudCover(buf, test.obs)
		if buf.String() != test.s {
			t.Errorf("got '%s', wanted '%s'", buf.String(), test.s)
		}
		buf.Reset()
	}
}

func TestFormatWindSpeed(t *testing.T) {
	var tests = []struct {
		obs Observations
		s   string
	}{
		{Observations{WindSpeed: math.NaN(), WindDirection: math.NaN(), WindGust: math.NaN()}, ""},
		{Observations{WindSpeed: 1.1, WindDirection: math.NaN(), WindGust: math.NaN()}, ", heikkoa tuulta 1.1 m/s"},
		{Observations{WindSpeed: 1.1, WindDirection: 225, WindGust: math.NaN()}, ", heikkoa lounaistuulta 1.1 m/s"},
		{Observations{WindSpeed: 1.1, WindDirection: 225, WindGust: 3.2}, ", heikkoa lounaistuulta 1.1 m/s (3.2 m/s)"},
	}

	buf := new(bytes.Buffer)
	for _, test := range tests {
		formatWindSpeed(buf, test.obs)
		if buf.String() != test.s {
			t.Errorf("got '%s', wanted '%s'", buf.String(), test.s)
		}
		buf.Reset()
	}
}

func TestFormatHumidity(t *testing.T) {
	var tests = []struct {
		obs Observations
		s   string
	}{
		{Observations{Humidity: math.NaN()}, ""},
		{Observations{Humidity: 65}, ", ilmankosteus 65%"},
	}

	buf := new(bytes.Buffer)
	for _, test := range tests {
		formatHumidity(buf, test.obs)
		if buf.String() != test.s {
			t.Errorf("got '%s', wanted '%s'", buf.String(), test.s)
		}
		buf.Reset()
	}
}

func TestFormatRain(t *testing.T) {
	var tests = []struct {
		obs Observations
		s   string
	}{
		{Observations{Precipitation: math.NaN(), RainIntensity: math.NaN()}, ""},
		{Observations{Precipitation: 1.1, RainIntensity: math.NaN()}, ", sateen määrä 1.1 mm"},
		{Observations{Precipitation: 1.1, RainIntensity: 0.5}, ", sateen määrä 1.1 mm (0.5 mm/h)"},
	}

	buf := new(bytes.Buffer)
	for _, test := range tests {
		formatRain(buf, test.obs)
		if buf.String() != test.s {
			t.Errorf("got '%s', wanted '%s'", buf.String(), test.s)
		}
		buf.Reset()
	}
}

func TestFormatSnow(t *testing.T) {
	var tests = []struct {
		obs Observations
		s   string
	}{
		{Observations{SnowDepth: math.NaN()}, ""},
		{Observations{SnowDepth: 7}, ", lumen syvyys 7 cm"},
	}

	buf := new(bytes.Buffer)
	for _, test := range tests {
		formatSnow(buf, test.obs)
		if buf.String() != test.s {
			t.Errorf("got '%s', wanted '%s'", buf.String(), test.s)
		}
		buf.Reset()
	}
}

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
		{10, "taivas ei näy", true},
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
