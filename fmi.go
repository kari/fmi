package main

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

// Weather returns current weather for a place
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
	name		label				measure
	t2m			Air Temperature 	degC
	ws_10min	Wind Speed			m/s
	wg_10min	Gust Speed			m/s
	wd_10min	Wind Direction		degrees
	rh			Relative humidity	%
	td			Dew-point temp.		degC
	r_1h		Precipitation amt	mm
	ri_10min	Precip. intensity	mm/h
	snow_aws	Snow depth			cm
	p_sea		Pressure (msl)		hPa
	vis			Visibility			m
	*/
	measures := []string{"t2m", "ws_10min", "wg_10min", "wd_10min", "rh", "r_1h", "ri_10min", "snow_aws"}
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
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error?
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
	fmt.Fprintf(&output, "lämpötila %.1f°C", latestObservations["t2m"].Value)
	if !math.IsNaN(latestObservations["ws_10min"].Value) {
		fmt.Fprintf(&output, ", %s %.f m/s (%.f m/s)", windSpeed(latestObservations["ws_10min"].Value, latestObservations["wd_10min"].Value), latestObservations["ws_10min"].Value, latestObservations["wg_10min"].Value)
	}
	fmt.Fprintf(&output, ", ilmankosteus %.f%%", latestObservations["rh"].Value)
	if !math.IsNaN(latestObservations["r_1h"].Value) && latestObservations["r_1h"].Value >= 0 {
		fmt.Fprintf(&output, ", sateen määrä %.1f mm (%.1f mm/h)", latestObservations["r_1h"].Value, latestObservations["ri_10min"].Value)
	}
	if !math.IsNaN(latestObservations["snow_aws"].Value) && latestObservations["snow_aws"].Value >= 0 {
		fmt.Fprintf(&output, ", lumen syvyys %.f cm", latestObservations["snow_aws"].Value)
	}

	return output.String()
}

func windSpeed(s float64, d float64) string {
	switch {
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
	return "tuulen nopeus"
}

func windDirection(d float64) string {
	switch {
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
		return "lounais"
	case d >= 337.5 && d <= 360:
		return "pohjois"
	}
	return ""
}

func main() {
	fmt.Println(Weather("helsinki"))
}
