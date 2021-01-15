package fmi

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type simpleFeatureCollection struct {
	Timestamp time.Time     `xml:"timeStamp,attr"`
	Returned  int           `xml:"numberReturned,attr"`
	Matched   int           `xml:"numberMatched,attr"`
	Elements  []observation `xml:"member>BsWfsElement"`
}

type observation struct {
	Location  string    `xml:"Location>Point>pos"`
	Time      time.Time `xml:"Time"`
	Parameter string    `xml:"ParameterName"`
	Value     float64   `xml:"ParameterValue"`
}

func getFeature(query string, parameters url.Values) (*http.Response, error) {
	endpoint := url.URL{
		Scheme:   "http",
		Host:     "opendata.fmi.fi",
		Path:     "/wfs",
		RawQuery: "service=WFS&version=2.0.0&request=getFeature",
	}

	q := endpoint.Query()
	q.Set("storedquery_id", query)

	for k, vs := range parameters {
		for _, v := range vs {
			q.Add(k, v)
		}
	}

	endpoint.RawQuery = q.Encode()

	return http.Get(endpoint.String())
}

func formatObservations(place string, observations map[string]float64) string {
	var output strings.Builder

	fmt.Fprintf(&output, "Viimeisimmät säähavainnot paikassa %s: ", strings.Title(strings.ToLower(place)))
	if !math.IsNaN(observations["t2m"]) {
		fmt.Fprintf(&output, "lämpötila %.1f°C", observations["t2m"])
		switch {
		case observations["t2m"] > 20 && !math.IsNaN(observations["td"]):
			h := humidex(observations["t2m"], observations["td"])
			if humidexScale(h) != "" {
				fmt.Fprintf(&output, " (%s, tuntuu kuin %.1f°C)", humidexScale(h), h)
			} else {
				fmt.Fprintf(&output, " (tuntuu kuin %.1f°C)", h)
			}
		case observations["t2m"] <= 10 && !math.IsNaN(observations["ws_10min"]):
			wc := windChillFmi(observations["t2m"], observations["ws_10min"])
			if windChillScale(wc) != "" {
				fmt.Fprintf(&output, " (%s, tuntuu kuin %.1f°C)", windChillScale(wc), wc)
			} else {
				fmt.Fprintf(&output, " (tuntuu kuin %.1f°C)", wc)
			}
		}
	} else {
		fmt.Fprint(&output, "lämpötilatiedot puuttuvat")
	}
	if !math.IsNaN(observations["n_man"]) {
		fmt.Fprintf(&output, ", %s", cloudCover(observations["n_man"]))
	}
	if !math.IsNaN(observations["ws_10min"]) {
		fmt.Fprintf(&output, ", %s %.f m/s (%.f m/s)", windSpeed(observations["ws_10min"], observations["wd_10min"]), observations["ws_10min"], observations["wg_10min"])
	}
	if !math.IsNaN(observations["rh"]) {
		fmt.Fprintf(&output, ", ilmankosteus %.f%%", observations["rh"])
	}
	if !math.IsNaN(observations["r_1h"]) && observations["r_1h"] >= 0 {
		fmt.Fprintf(&output, ", sateen määrä %.1f mm (%.1f mm/h)", observations["r_1h"], observations["ri_10min"])
	}
	if !math.IsNaN(observations["snow_aws"]) && observations["snow_aws"] >= 0 {
		fmt.Fprintf(&output, ", lumen syvyys %.f cm", observations["snow_aws"])
	}

	return output.String()
}

// Weather returns current weather for a place as a written description
func Weather(place string) string {

	if place == "" {
		return "Paikkaa ei syötetty"
	}

	return simpleObservations(place)
}

func appendIfMssing(slice []string, s string) []string {
	for _, el := range slice {
		if el == s {
			return slice
		}
	}
	return append(slice, s)
}

func simpleObservations(place string) string {
	q := url.Values{}
	q.Set("place", place)
	q.Set("maxlocations", "2")

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
	       wawa		    Present weather 	code (00-99)
	                                        see: https://www.wmo.int/pages/prog/www/WMOCodes/WMO306_vI1/Publications/2017update/Sel9.pdf
	*/
	measures := []string{"t2m", "ws_10min", "wg_10min", "wd_10min", "rh", "r_1h", "ri_10min", "snow_aws", "n_man", "td"}
	q.Set("parameters", strings.Join(measures, ","))

	// There should be data every 10 mins
	q.Set("timestep", "10")
	endTime := time.Now().UTC().Truncate(10 * time.Minute)
	startTime := endTime.Add(-10 * time.Minute)
	q.Set("starttime", startTime.Format(time.RFC3339))
	q.Set("endtime", endTime.Format(time.RFC3339))

	resp, err := getFeature("fmi::observations::weather::simple", q)
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

	if resp.StatusCode != 200 {
		// If place parsing fails, returns 400 with OperationParsingFailed
		return "Säähavaintopaikkaa ei löytynyt"
	}

	var collection simpleFeatureCollection
	xml.Unmarshal(body, &collection)

	if collection.Matched == 0 || collection.Returned == 0 {
		return "Säähavaintoja ei löytynyt"
	}

	observations := make(map[time.Time]map[string]map[string]float64)
	times := make([]time.Time, 0)
	locations := make([]string, 0)

	for _, obs := range collection.Elements {
		if observations[obs.Time] == nil {
			times = append(times, obs.Time)
			observations[obs.Time] = make(map[string]map[string]float64)
		}
		if observations[obs.Time][obs.Location] == nil {
			locations = appendIfMssing(locations, obs.Location)
			observations[obs.Time][obs.Location] = make(map[string]float64)
		}
		observations[obs.Time][obs.Location][obs.Parameter] = obs.Value
	}

	sort.Slice(times, func(i, j int) bool {
		return times[i].After(times[j])
	})

	latestObs := make(map[string]float64)
	for _, timeIndex := range times {
		for _, locationIndex := range locations {
			if countNanMeasures(observations[timeIndex][locationIndex], measures) != len(measures) {
				latestObs = observations[timeIndex][locationIndex]
			}
		}
	}
	if len(latestObs) == 0 {
		return "Säähavaintoja ei löytynyt"
	}

	return formatObservations(place, latestObs)
}

func countNanMeasures(obs map[string]float64, measures []string) int {
	count := 0
	for _, measure := range measures {
		if math.IsNaN(obs[measure]) {
			count++
		}
	}
	return count
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
// and wind speed v (m/s) using a Canadian formula.
// The calculation works for air temperatures at or below 10C and wind speeds above 0.4 m/s.
// For reference see,
// https://fi.m.wikipedia.org/wiki/Pakkasen_purevuus#Uusi_kaava
func windChill(t float64, v float64) float64 {
	return 13.12 + 0.6215*t - 13.956*math.Pow(v, 0.16) + 0.4867*t*math.Pow(v, 0.16)
}

// WindChillFmi calculates wind chill with FMI's formula
// For reference see,
// https://github.com/fmidev/smartmet-library-newbase/blob/0da9473163883089c35a4c7267ba4c8a8bb3e14f/newbase/NFmiMetMath.cpp#L380
// https://tietopyynto.fi/tietopyynto/ilmatieteen-laitoksen-kayttama-tuntuu-kuin-laskentakaava/
func windChillFmi(t float64, v float64) float64 {
	var kmh = v * 3.6

	if kmh < 5 {
		return t + (-1.59+0.1345*t)/5*kmh
	}
	return 13.12 + 0.6215*t - 11.37*math.Pow(v, 0.16) + 0.3965*t*math.Pow(v, 0.16)
}

// SummerSimmer calculates the Summer Simmer index
// For reference see,
// https://github.com/fmidev/smartmet-library-newbase/blob/0da9473163883089c35a4c7267ba4c8a8bb3e14f/newbase/NFmiMetMath.cpp#L335
// http://www.summersimmer.com/home.htm
func SummerSimmer(t float64, rh float64) float64 {
	const simmerLimit = 14.5
	const rhRef = 50.0 / 100.0

	if t <= simmerLimit {
		return t
	}

	var r = rh / 100.0

	return (1.8*t - 0.55*(1-r)*(1.8*t-26) - 0.55*(1-rhRef)*26) / (1.8 * (1 - 0.55*(1-rhRef)))
}

// FeelsLikeTemperature calculates FMI's "feel like" temperature
// For reference see,
// https://github.com/fmidev/smartmet-library-newbase/blob/0da9473163883089c35a4c7267ba4c8a8bb3e14f/newbase/NFmiMetMath.cpp#L418
// https://tietopyynto.fi/tietopyynto/ilmatieteen-laitoksen-kayttama-tuntuu-kuin-laskentakaava/
func FeelsLikeTemperature(t float64, v float64, rh float64, rad float64) float64 {
	const a = 15.0
	const t0 = 37.0

	var chill = a + (1-a/t0)*t + a/t0*math.Pow(v+1, 0.16)*(t-t0)

	var heat = SummerSimmer(t, rh)
	var feels = t + (chill - t) + (heat - t)

	if rad != -1 {
		const absorption = 0.07
		feels += 0.7*absorption*rad/(v+10) - 0.25
	}

	return feels
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
