package fmi

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type featureCollection struct {
	Timestamp time.Time     `xml:"timeStamp,attr"`
	Results   int           `xml:"numberReturned,attr"`
	Matches   int           `xml:"numberMatched,attr"`
	Elements  []observation `xml:"member>BsWfsElement"`
}

type observation struct {
	Time      time.Time `xml:"Time"`
	Parameter string    `xml:"ParameterName"`
	Value     float64   `xml:"ParameterValue"`
}

// Weather returns current weather for a place as a written description
func Weather(place string) string {

	endpoint := url.URL{
		Scheme:   "http",
		Host:     "opendata.fmi.fi",
		Path:     "/wfs",
		RawQuery: "service=WFS&version=2.0.0&request=getFeature&storedquery_id=fmi::observations::weather::simple",
	}

	q := endpoint.Query()
	q.Set("place", place)

	/* Parameters:
	   name		    label				measure
	   t2m		    Air Temperature 	degC
	   ws_10min	    Wind Speed			m/s
	   wg_10min	    Gust Speed			m/s
	   wd_10min	    Wind Direction		degrees
	   rh		    Relative humidity	%
	   td		    Dew-point temp.		degC
	   r_1h		    Precipitation amt	mm
	   ri_10min	    Precip. intensity	mm/h
	   snow_aws	    Snow depth			cm
	   p_sea	    Pressure (msl)		hPa
	   vis		    Visibility			m
	   n_man	    Cloud cover			1/8
	   wawa		    Present weather?	?
	*/
	measures := []string{"t2m", "ws_10min", "wg_10min", "wd_10min", "rh", "r_1h", "ri_10min", "snow_aws", "n_man", "td"}
	q.Set("parameters", strings.Join(measures, ","))

	// There should be data every 10 mins
	endTime := time.Now().UTC().Truncate(10 * time.Minute)
	startTime := endTime.Add(-10 * time.Minute)
	q.Set("starttime", startTime.Format(time.RFC3339))
	q.Set("endtime", endTime.Format(time.RFC3339))

	endpoint.RawQuery = q.Encode()

	resp, err := http.Get(endpoint.String())
	if err != nil {
		// handle error
		return "Säähavaintoja ei saatu haettua"
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error?
		return "Virhe luettaessa havaintoja"
	}

	var collection featureCollection
	xml.Unmarshal(body, &collection)

	if collection.Matches == 0 || collection.Results == 0 {
		return "Säähavaintoja ei löytynyt"
	}

	// fmt.Println(collection)

	latestObservations := make(map[string]observation)

	for _, obs := range collection.Elements {
		v, ok := latestObservations[obs.Parameter]
		if !ok || v.Time.Before(obs.Time) {
			latestObservations[obs.Parameter] = observation{
				Time:  obs.Time,
				Value: obs.Value,
			}
		}
	}

	var output strings.Builder
	fmt.Fprintf(&output, "Viimeisimmät säähavainnot paikassa %s: ", strings.Title(strings.ToLower(place)))
	if !math.IsNaN(latestObservations["t2m"].Value) {
		fmt.Fprintf(&output, "lämpötila %.1f°C", latestObservations["t2m"].Value)
		switch {
		case latestObservations["t2m"].Value > 20 && !math.IsNaN(latestObservations["td"].Value):
			h := humidex(latestObservations["t2m"].Value, latestObservations["td"].Value)
			if humidexScale(h) != "" {
				fmt.Fprintf(&output, " (%s, tuntuu kuin %.1f°C)", humidexScale(h), h)
			} else {
				fmt.Fprintf(&output, " (tuntuu kuin %.1f°C)", h)
			}
		case latestObservations["t2m"].Value <= 10 && !math.IsNaN(latestObservations["ws_10min"].Value):
			wc := windChill(latestObservations["t2m"].Value, latestObservations["ws_10min"].Value)
			if windChillScale(wc) != "" {
				fmt.Fprintf(&output, " (%s, tuntuu kuin %.1f°C)", windChillScale(wc), wc)
			} else {
				fmt.Fprintf(&output, " (tuntuu kuin %.1f°C)", wc)
			}
		}
	} else {
		fmt.Fprint(&output, "lämpötilatiedot puuttuvat")
	}
	if !math.IsNaN(latestObservations["n_man"].Value) {
		fmt.Fprintf(&output, ", %s", cloudCover(latestObservations["n_man"].Value))
	}
	if !math.IsNaN(latestObservations["ws_10min"].Value) {
		fmt.Fprintf(&output, ", %s %.f m/s (%.f m/s)", windSpeed(latestObservations["ws_10min"].Value, latestObservations["wd_10min"].Value), latestObservations["ws_10min"].Value, latestObservations["wg_10min"].Value)
	}
	if !math.IsNaN(latestObservations["rh"].Value) {
		fmt.Fprintf(&output, ", ilmankosteus %.f%%", latestObservations["rh"].Value)
	}
	if !math.IsNaN(latestObservations["r_1h"].Value) && latestObservations["r_1h"].Value >= 0 {
		fmt.Fprintf(&output, ", sateen määrä %.1f mm (%.1f mm/h)", latestObservations["r_1h"].Value, latestObservations["ri_10min"].Value)
	}
	if !math.IsNaN(latestObservations["snow_aws"].Value) && latestObservations["snow_aws"].Value >= 0 {
		fmt.Fprintf(&output, ", lumen syvyys %.f cm", latestObservations["snow_aws"].Value)
	}

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

// humidex calculates the humidity index given air temperature t (degC)
// and dew point (degC) td.
// For reference, see https://en.wikipedia.org/wiki/Humidex
func humidex(t float64, td float64) float64 {
	return t + 5.0/9.0*(6.11*math.Exp(5417.7530*(1/273.16-1/(273.15+td)))-10)
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

// windChill calculates the wind chill effect given air temperature t (degC)
// and wind speed (m/s). The calculation works for air temperatures at or
// below 10C. For reference see,
// https://fi.m.wikipedia.org/wiki/Pakkasen_purevuus#Uusi_kaava
func windChill(t float64, v float64) float64 {
	return 13.12 + 0.6215*t - 13.956*math.Pow(v, 0.16) + 0.4867*t*math.Pow(v, 0.16)
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
