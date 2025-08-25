// Package fmi fetches latest weather observations for a given place
// using FMI's open API
package fmi

import (
	"context"
	"encoding/xml"
	"errors"
	"io"
	"math"
	"net/http"
	"net/url"
	"slices"
	"sort"
	"strings"
	"time"
)

// simpleFeatureCollection is a struct in returned XML
type simpleFeatureCollection struct {
	Timestamp time.Time     `xml:"timeStamp,attr"`
	Returned  int           `xml:"numberReturned,attr"`
	Matched   int           `xml:"numberMatched,attr"`
	Elements  []observation `xml:"member>BsWfsElement"`
}

// observation is a struct in returned XML
type observation struct {
	Location  string    `xml:"Location>Point>pos"`
	Time      time.Time `xml:"Time"`
	Parameter string    `xml:"ParameterName"`
	Value     float64   `xml:"ParameterValue"`
}

// observations holds observations for a place as a map
type observations map[string]float64

// Weather returns current weather for a place as a written description
func Weather(place string) (string, error) {

	if place == "" {
		return "", errors.New("paikkaa ei syötetty")
	}

	obs, err := getObservations(place)
	if err != nil {
		return "", err
	}

	weather := formatObservations(place, obs)

	return weather, nil
}

func parseFeatureCollection(data []byte) (simpleFeatureCollection, error) {
	var collection simpleFeatureCollection

	if err := xml.Unmarshal(data, &collection); err != nil {
		return simpleFeatureCollection{}, errors.New("virhe parsittaessa havaintoja")
	}

	return collection, nil
}

func extractLatestObservations(collection simpleFeatureCollection, measures []string) observations {
	observations := make(map[time.Time]map[string]map[string]float64)
	times := make([]time.Time, 0)
	locations := make([]string, 0)

	for _, obs := range collection.Elements {
		if observations[obs.Time] == nil {
			times = append(times, obs.Time)
			observations[obs.Time] = make(map[string]map[string]float64)
		}
		if observations[obs.Time][obs.Location] == nil {
			if !slices.Contains(locations, obs.Location) {
				locations = append(locations, obs.Location)
			}
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

	return latestObs
}

// getObservations does a HTTP GET request against FMI's API to fetch data
// for a place
func getObservations(place string) (observations, error) {
	/*  Parameters:
	name		label				measure
	t2m			Air Temperature		degC
	ws_10min	Wind Speed			m/s
	wg_10min	Gust Speed			m/s
	wd_10min	Wind Direction		degrees
	rh			Relative humidity	%
	td			Dew-point temp.		degC
	r_1h		Precipitation amt	mm
	ri_10min	Precip. intensity	mm/h
	snow_aws	Snow depth			cm
				-1 = no snow, 0 = snow in vicinity
	p_sea		Pressure (msl)		hPa
	vis			Visibility			m
	n_man		Cloud cover			1/8
	wawa		Present weather		code (00-99)
				see: https://www.wmo.int/pages/prog/www/WMOCodes/WMO306_vI1/Publications/2017update/Sel9.pdf
	*/
	measures := []string{"t2m", "ws_10min", "wg_10min", "wd_10min", "rh", "r_1h", "ri_10min", "snow_aws", "n_man", "td", "glob_u"}

	q := url.Values{}
	q.Set("service", "WFS")
	q.Set("version", "2.0.0")
	q.Set("request", "getFeature")
	q.Set("storedquery_id", "fmi::observations::weather::simple")

	q.Set("place", place)
	q.Set("maxlocations", "2")
	q.Set("parameters", strings.Join(measures, ","))

	// There should be data every 10 mins
	q.Set("timestep", "10")
	endTime := time.Now().UTC().Truncate(10 * time.Minute)
	startTime := endTime.Add(-10 * time.Minute)
	q.Set("starttime", startTime.Format(time.RFC3339))
	q.Set("endtime", endTime.Format(time.RFC3339))

	endpoint := url.URL{
		Scheme: "http",
		Host:   "opendata.fmi.fi",
		Path:   "/wfs",
	}

	endpoint.RawQuery = q.Encode()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.New("säähavaintoja ei saatu haettua")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("virhe luettaessa havaintoja")
	}

	if resp.StatusCode != http.StatusOK {
		// If place parsing fails, returns 400 with OperationParsingFailed
		return nil, errors.New("säähavaintopaikkaa ei löytynyt")
	}

	collection, err := parseFeatureCollection(body)
	if err != nil || collection.Matched == 0 || collection.Returned == 0 {
		return nil, errors.New("säähavaintoja ei löytynyt")
	}

	latestObs := extractLatestObservations(collection, measures)
	if len(latestObs) == 0 {
		return nil, errors.New("säähavaintoja ei löytynyt")
	}

	return latestObs, nil
}

func countNanMeasures(obs observations, measures []string) int {
	count := 0
	for _, measure := range measures {
		if math.IsNaN(obs[measure]) {
			count++
		}
	}
	return count
}
