package fmi

import (
	"fmt"
	"io"
	"math"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func formatTemperature(output io.Writer, observations observations) {
	if temp, ok := observations["t2m"]; ok && !math.IsNaN(temp) {
		fmt.Fprintf(output, "lämpötila %.1f°C", temp)

		feels := math.NaN()
		if ws, ok := observations["ws_10min"]; ok {
			if rh, ok := observations["rh"]; ok {
				if rad, ok := observations["glob_u"]; ok {
					feels = FeelsLikeTemperature(temp, ws, rh, rad)
				} else {
					feels = FeelsLikeTemperature(temp, ws, rh, math.NaN())
				}
			}
		}

		if td, ok := observations["td"]; ok && temp > 20 {
			if h, ok := humidexScale(humidex(temp, td)); ok {
				if !math.IsNaN(feels) {
					fmt.Fprintf(output, " (%s, tuntuu kuin %.1f°C)", h, feels)
				} else {
					fmt.Fprintf(output, " (%s)", h)
				}
			} else {
				if !math.IsNaN(feels) {
					fmt.Fprintf(output, " (tuntuu kuin %.1f°C)", feels)
				}
			}
		} else if ws, ok := observations["ws_10min"]; ok && temp <= 10 {
			if wc, ok := windChillScale(windChillFmi(temp, ws)); ok {
				if !math.IsNaN(feels) {
					fmt.Fprintf(output, " (%s, tuntuu kuin %.1f°C)", wc, feels)
				} else {
					fmt.Fprintf(output, " (%s)", wc)
				}
			} else {
				if !math.IsNaN(feels) {
					fmt.Fprintf(output, " (tuntuu kuin %.1f°C)", feels)
				}
			}
		} else if !math.IsNaN(feels) {
			fmt.Fprintf(output, " (tuntuu kuin %.1f°C)", feels)
		}
	} else {
		fmt.Fprint(output, "lämpötilatiedot puuttuvat")
	}
}

func formatCloudCover(output io.Writer, observations observations) {
	if cc, ok := observations["n_man"]; ok {
		if cover, ok := cloudCover(cc); ok {
			fmt.Fprintf(output, ", %s", cover)
		}
	}
}

func formatWindSpeed(output io.Writer, observations observations) {
	if ws, ok := observations["ws_10min"]; ok {
		if wd, ok := observations["wd_10min"]; ok {
			fmt.Fprintf(output, ", %s %.1f m/s", windSpeed(ws, wd), ws)
		} else {
			fmt.Fprintf(output, ", %s %.1f m/s", windSpeed(ws, math.NaN()), ws)
		}
		if wg, ok := observations["wg_10min"]; ok && !math.IsNaN(wg) {
			fmt.Fprintf(output, " (%.1f m/s)", wg)
		}
	}
}

func formatHumidity(output io.Writer, observations observations) {
	if rh, ok := observations["rh"]; ok && !math.IsNaN(rh) {
		fmt.Fprintf(output, ", ilmankosteus %.f%%", rh)
	}
}

func formatRain(output io.Writer, observations observations) {
	if r, ok := observations["r_1h"]; ok && r >= 0 {
		fmt.Fprintf(output, ", sateen määrä %.1f mm", r)
		if ri, ok := observations["ri_10min"]; ok {
			fmt.Fprintf(output, " (%.1f mm/h)", ri)
		}
	}
}

func formatSnow(output io.Writer, observations observations) {
	if snow, ok := observations["snow_aws"]; ok && snow >= 0 {
		fmt.Fprintf(output, ", lumen syvyys %.f cm", snow)
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
func cloudCover(d float64) (string, bool) {
	switch {
	case d < 0:
		return "", false
	case d >= 0 && d <= 1:
		return "selkeää", true
	case d <= 3:
		return "melko selkeää", true
	case d <= 5:
		return "puolipilvistä", true
	case d <= 7:
		return "melko pilvistä", true
	case d <= 8:
		return "pilvistä", true
	case d == 9:
		return "taivas ei näy", true
	}
	return "", false
}

// humidexScale converts humidex index h to a textual classification using
// definitions from:
// https://web.archive.org/web/20150319113439/http://ilmatieteenlaitos.fi/tietoa-helteen-tukaluudesta
func humidexScale(h float64) (string, bool) {
	switch {
	case h < 20:
		return "", false
	case h <= 26:
		return "mukava", true
	case h <= 30:
		return "lämmin", true
	case h <= 34:
		return "kuuma", true
	case h <= 40:
		return "tukala", true
	case h > 40:
		return "erittäin tukala", true
	}
	return "", false
}

// windChillScale converts windChill index w to a textual representation using
// classifications from https://fi.wikipedia.org/wiki/Pakkasen_purevuus
func windChillScale(w float64) (string, bool) {
	switch {
	case w > -25:
		return "", false
	case w <= -60:
		return "suuri paleltumisvaara", true
	case w <= -35:
		return "paleltumisvaara", true
	case w <= -25:
		return "erittäin kylmä", true
	}
	return "", false
}
