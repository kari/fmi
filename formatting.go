package fmi

import (
	"fmt"
	"io"
	"math"
	"strings"
)

func formatTemperature(output io.Writer, o Observations) {
	temp := o.Temperature
	if math.IsNaN(temp) {
		fmt.Fprint(output, "lämpötilatiedot puuttuvat")
		return
	}

	fmt.Fprintf(output, "lämpötila %.1f°C", temp)

	feels := math.NaN()
	if !math.IsNaN(o.WindSpeed) && !math.IsNaN(o.Humidity) {
		feels = FeelsLike(temp, o.WindSpeed, o.Humidity, o.Radiation)
	}

	// Humidex classifies heat and wind chill cold; both share the
	// parenthesis with the feels-like temperature.
	label := ""
	if td := o.DewPoint; !math.IsNaN(td) && temp > 20 {
		label, _ = humidexScale(Humidex(temp, td))
	} else if ws := o.WindSpeed; !math.IsNaN(ws) && temp <= 10 {
		label, _ = windChillScale(WindChillFMI(temp, ws))
	}

	var parts []string
	if label != "" {
		parts = append(parts, label)
	}
	if !math.IsNaN(feels) {
		parts = append(parts, fmt.Sprintf("tuntuu kuin %.1f°C", feels))
	}
	if len(parts) > 0 {
		fmt.Fprintf(output, " (%s)", strings.Join(parts, ", "))
	}
}

func formatCloudCover(output io.Writer, o Observations) {
	if cover, ok := cloudCover(o.CloudCover); ok {
		fmt.Fprintf(output, ", %s", cover)
	}
}

func formatWindSpeed(output io.Writer, o Observations) {
	ws := o.WindSpeed
	if math.IsNaN(ws) {
		return
	}
	fmt.Fprintf(output, ", %s %.1f m/s", windSpeed(ws, o.WindDirection), ws)
	if wg := o.WindGust; !math.IsNaN(wg) {
		fmt.Fprintf(output, " (%.1f m/s)", wg)
	}
}

func formatHumidity(output io.Writer, o Observations) {
	if rh := o.Humidity; !math.IsNaN(rh) {
		fmt.Fprintf(output, ", ilmankosteus %.f%%", rh)
	}
}

func formatRain(output io.Writer, o Observations) {
	if r := o.Precipitation; !math.IsNaN(r) && r >= 0 {
		fmt.Fprintf(output, ", sateen määrä %.1f mm", r)
		if ri := o.RainIntensity; !math.IsNaN(ri) {
			fmt.Fprintf(output, " (%.1f mm/h)", ri)
		}
	}
}

func formatSnow(output io.Writer, o Observations) {
	if snow := o.SnowDepth; !math.IsNaN(snow) && snow >= 0 {
		fmt.Fprintf(output, ", lumen syvyys %.f cm", snow)
	}
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
	case d > 8:
		// 9 means the sky is not visible
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
