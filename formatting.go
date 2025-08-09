package fmi

import (
	"fmt"
	"math"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func formatTemperature(output *strings.Builder, observations observations) {
	if !math.IsNaN(observations["t2m"]) {
		fmt.Fprintf(output, "lämpötila %.1f°C", observations["t2m"])
		switch {
		case observations["t2m"] > 20 && !math.IsNaN(observations["td"]):
			h := humidex(observations["t2m"], observations["td"])
			if humidexScale(h) != "" {
				fmt.Fprintf(output, " (%s, tuntuu kuin %.1f°C)", humidexScale(h), h)
			} else {
				fmt.Fprintf(output, " (tuntuu kuin %.1f°C)", h)
			}
		case observations["t2m"] <= 10 && !math.IsNaN(observations["ws_10min"]):
			wc := windChillFmi(observations["t2m"], observations["ws_10min"])
			if windChillScale(wc) != "" {
				fmt.Fprintf(output, " (%s, tuntuu kuin %.1f°C)", windChillScale(wc), wc)
			} else {
				fmt.Fprintf(output, " (tuntuu kuin %.1f°C)", wc)
			}
		}
	} else {
		fmt.Fprint(output, "lämpötilatiedot puuttuvat")
	}

}

func formatCloudCover(output *strings.Builder, observations observations) {
	if !math.IsNaN(observations["n_man"]) {
		fmt.Fprintf(output, ", %s", cloudCover(observations["n_man"]))
	}
}

func formatWindSpeed(output *strings.Builder, observations observations) {
	if !math.IsNaN(observations["ws_10min"]) {
		fmt.Fprintf(output, ", %s %.f m/s (%.f m/s)", windSpeed(observations["ws_10min"], observations["wd_10min"]), observations["ws_10min"], observations["wg_10min"])
	}
}

func formatHumidity(output *strings.Builder, observations observations) {
	if !math.IsNaN(observations["rh"]) {
		fmt.Fprintf(output, ", ilmankosteus %.f%%", observations["rh"])
	}
}

func formatRain(output *strings.Builder, observations observations) {
	if !math.IsNaN(observations["r_1h"]) && observations["r_1h"] >= 0 {
		fmt.Fprintf(output, ", sateen määrä %.1f mm (%.1f mm/h)", observations["r_1h"], observations["ri_10min"])
	}
}

func formatSnow(output *strings.Builder, observations observations) {
	if !math.IsNaN(observations["snow_aws"]) && observations["snow_aws"] >= 0 {
		fmt.Fprintf(output, ", lumen syvyys %.f cm", observations["snow_aws"])
	}
}

// formatObservations returns a string representation of weather observations
// at a place
func formatObservations(place string, observations observations) string {
	var output strings.Builder

	c := cases.Title(language.Finnish)

	fmt.Fprintf(&output, "Viimeisimmät säähavainnot paikassa %s: ", c.String(strings.ToLower(place)))
	formatTemperature(&output, observations)
	formatCloudCover(&output, observations)
	formatWindSpeed(&output, observations)
	formatHumidity(&output, observations)
	formatRain(&output, observations)
	formatSnow(&output, observations)

	return output.String()
}

// windSpeed takes wind speed s (m/s) and direction d (angle) and
// returns a textual representation of them.
// For reference, see: https://ilmatieteenlaitos.fi/tuulet
func windSpeed(s float64, d float64) string {
	switch {
	case s < 0:
		return ""
	case s < 1:
		return "tyyntä"
	case s <= 4:
		return fmt.Sprintf("heikkoa %stuulta", windDirection(d))
	case s <= 8:
		return fmt.Sprintf("kohtalaista %stuulta", windDirection(d))
	case s <= 14:
		return fmt.Sprintf("navakkaa %stuulta", windDirection(d))
	case s <= 21:
		return fmt.Sprintf("kovaa %stuulta", windDirection(d))
	case s < 33:
		return "myrskyä"
	case s >= 33:
		return "hirmumyrskyä"
	}
	return ""
}

// windDirection takes a wind direction d in angles (0-360) and converts
// it to a string representation. For reference, see:
// https://ilmatieteenlaitos.fi/tuulet
func windDirection(d float64) string {
	switch {
	case d < 0:
		return ""
	case d >= 0 && d <= 22.5:
		return "pohjois"
	case d < 67.5:
		return "koillis"
	case d <= 112.5:
		return "itä"
	case d < 157.5:
		return "kaakkois"
	case d <= 202.5:
		return "etelä"
	case d < 247.5:
		return "lounais"
	case d <= 292.5:
		return "länsi"
	case d < 337.5:
		return "luoteis"
	case d >= 337.5 && d <= 360:
		return "pohjois"
	}
	return ""
}

// cloudCover converts the cloud cover measure (1/8) to textual format
// using definitions at https://ilmatieteenlaitos.fi/pilvisyys
func cloudCover(d float64) string {
	switch {
	case d < 0:
		return ""
	case d >= 0 && d <= 1:
		return "selkeää"
	case d <= 3:
		return "melko selkeää"
	case d <= 5:
		return "puolipilvistä"
	case d <= 7:
		return "melko pilvistä"
	case d <= 8:
		return "pilvistä"
	case d == 9:
		return "taivas ei näy"
	}
	return ""
}

// humidexScale converts humidex index h to a textual classification using
// definitions from:
// https://web.archive.org/web/20150319113439/http://ilmatieteenlaitos.fi/tietoa-helteen-tukaluudesta
func humidexScale(h float64) string {
	switch {
	case h < 20:
		return ""
	case h <= 26:
		return "mukava"
	case h <= 30:
		return "lämmin"
	case h <= 34:
		return "kuuma"
	case h <= 40:
		return "tukala"
	case h > 40:
		return "erittäin tukala"
	}
	return ""
}

// windChillScale converts windChill index w to a textual representation using
// classifications from https://fi.wikipedia.org/wiki/Pakkasen_purevuus
func windChillScale(w float64) string {
	switch {
	case w > -25:
		return ""
	case w <= -60:
		return "suuri paleltumisvaara"
	case w <= -35:
		return "paleltumisvaara"
	case w <= -25:
		return "erittäin kylmä"
	}
	return ""
}
