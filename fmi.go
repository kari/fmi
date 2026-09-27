// Package fmi fetches latest weather observations for a given place
// using FMI's open API
package fmi

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Errors reported by Weather and Current. Identify them with errors.Is.
var (
	// ErrNoPlace is returned when the place argument is empty.
	ErrNoPlace = errors.New("paikkaa ei syötetty")

	// ErrFetchFailed is returned when FMI cannot be reached or replies
	// with an unexpected HTTP status.
	ErrFetchFailed = errors.New("säähavaintoja ei saatu haettua")

	// ErrUnknownPlace is returned when FMI does not recognize the place.
	ErrUnknownPlace = errors.New("säähavaintopaikkaa ei löytynyt")

	// ErrNoObservations is returned when FMI has no usable observations
	// for the place.
	ErrNoObservations = errors.New("säähavaintoja ei löytynyt")
)

// simpleFeatureCollection is a struct in returned XML
type simpleFeatureCollection struct {
	Timestamp time.Time `xml:"timeStamp,attr"`
	Returned  int       `xml:"numberReturned,attr"`
	Matched   int       `xml:"numberMatched,attr"`
	Elements  []measure `xml:"member>BsWfsElement"`
}

// measure is a single raw measure (one parameter and its value) in the
// returned XML
type measure struct {
	Location  string    `xml:"Location>Point>pos"`
	Time      time.Time `xml:"Time"`
	Parameter string    `xml:"ParameterName"`
	Value     float64   `xml:"ParameterValue"`
}

// FMI parameter names returned by the weather observations API
const (
	paramTemperature   = "t2m"
	paramWindSpeed     = "ws_10min"
	paramWindGust      = "wg_10min"
	paramWindDirection = "wd_10min"
	paramHumidity      = "rh"
	paramDewPoint      = "td"
	paramPrecipitation = "r_1h"
	paramRainIntensity = "ri_10min"
	paramSnowDepth     = "snow_aws"
	paramCloudCover    = "n_man"
	paramRadiation     = "glob_u"
)

// Observations holds the weather measures of one station at one time.
// A NaN value means the measure is missing from the observation.
//
// The zero value is not meaningful: unset fields read as 0.0, which is a
// valid measure, so obtain Observations from Current instead of
// constructing one by hand.
type Observations struct {
	Temperature   float64 // degC
	WindSpeed     float64 // m/s
	WindGust      float64 // m/s
	WindDirection float64 // degrees
	Humidity      float64 // %
	DewPoint      float64 // degC
	Precipitation float64 // mm over the last hour
	RainIntensity float64 // mm/h
	SnowDepth     float64 // cm, -1 means no snow
	CloudCover    float64 // 1/8, 9 means the sky is not visible
	Radiation     float64 // W/m²
}

// nanObservation returns an Observations where every measure is missing
func nanObservation() Observations {
	return Observations{
		Temperature:   math.NaN(),
		WindSpeed:     math.NaN(),
		WindGust:      math.NaN(),
		WindDirection: math.NaN(),
		Humidity:      math.NaN(),
		DewPoint:      math.NaN(),
		Precipitation: math.NaN(),
		RainIntensity: math.NaN(),
		SnowDepth:     math.NaN(),
		CloudCover:    math.NaN(),
		Radiation:     math.NaN(),
	}
}

// set stores a raw measure in the matching field. Unknown parameters are
// ignored.
func (o *Observations) set(param string, value float64) {
	switch param {
	case paramTemperature:
		o.Temperature = value
	case paramWindSpeed:
		o.WindSpeed = value
	case paramWindGust:
		o.WindGust = value
	case paramWindDirection:
		o.WindDirection = value
	case paramHumidity:
		o.Humidity = value
	case paramDewPoint:
		o.DewPoint = value
	case paramPrecipitation:
		o.Precipitation = value
	case paramRainIntensity:
		o.RainIntensity = value
	case paramSnowDepth:
		o.SnowDepth = value
	case paramCloudCover:
		o.CloudCover = value
	case paramRadiation:
		o.Radiation = value
	}
}

// newObservation builds an observation from the raw measures of one
// station at one time; measures the station does not report stay NaN.
func newObservation(params map[string]float64) Observations {
	o := nanObservation()
	for param, value := range params {
		o.set(param, value)
	}
	return o
}

// hasValue reports whether at least one of the raw measures is not NaN
func hasValue(params map[string]float64) bool {
	for _, v := range params {
		if !math.IsNaN(v) {
			return true
		}
	}
	return false
}

// Conditions holds the latest observations of a place.
type Conditions struct {
	Place        string
	Time         time.Time // timestamp of the observations
	Observations Observations
}

// String returns the conditions as a written description.
func (c Conditions) String() string {
	var output strings.Builder

	caser := cases.Title(language.Finnish)

	fmt.Fprintf(&output, "Viimeisimmät säähavainnot paikassa %s: ", caser.String(strings.ToLower(c.Place)))
	formatTemperature(&output, c.Observations)
	formatCloudCover(&output, c.Observations)
	formatWindSpeed(&output, c.Observations)
	formatHumidity(&output, c.Observations)
	formatRain(&output, c.Observations)
	formatSnow(&output, c.Observations)

	return output.String()
}

// Current returns the latest observations of a place as structured
// Conditions. The error can be inspected with errors.Is against the
// package's Err values, with the underlying cause kept intact. Current
// is safe to call from multiple goroutines.
func Current(ctx context.Context, place string) (Conditions, error) {
	if place == "" {
		return Conditions{}, ErrNoPlace
	}

	collection, err := fetch(ctx, place)
	if err != nil {
		return Conditions{}, err
	}

	obs, measuredAt, ok := extractLatestObservations(collection)
	if !ok {
		return Conditions{}, ErrNoObservations
	}

	return Conditions{Place: place, Time: measuredAt, Observations: obs}, nil
}

// Weather returns the latest observations of a place as a written
// description. It is a convenience wrapper around Current, and safe to
// call from multiple goroutines.
func Weather(place string) (string, error) {
	w, err := Current(context.Background(), place)
	if err != nil {
		return "", err
	}
	return w.String(), nil
}

func parseFeatureCollection(data []byte) (simpleFeatureCollection, error) {
	var collection simpleFeatureCollection

	if err := xml.Unmarshal(data, &collection); err != nil {
		return simpleFeatureCollection{}, fmt.Errorf("virhe parsittaessa havaintoja: %w", err)
	}

	return collection, nil
}

// extractLatestObservations returns the most recent observation that has
// at least one measure, together with its timestamp. FMI may return
// observations where every measure is NaN and up to maxlocations
// stations, so the newest timestamp with a populated station wins.
// Stations sharing a timestamp are tried in the order FMI returned them,
// nearest to the place first.
func extractLatestObservations(collection simpleFeatureCollection) (Observations, time.Time, bool) {
	type stationTime struct {
		time     time.Time
		location string
	}

	// Group the raw measures by station and observation time. Keys are
	// kept in the order the stations first appear in the response.
	groups := make(map[stationTime]map[string]float64)
	var keys []stationTime
	for _, obs := range collection.Elements {
		key := stationTime{time: obs.Time, location: obs.Location}
		if groups[key] == nil {
			groups[key] = make(map[string]float64)
			keys = append(keys, key)
		}
		groups[key][obs.Parameter] = obs.Value
	}

	// Newest first. The stable sort keeps stations of the same timestamp
	// in FMI's order.
	slices.SortStableFunc(keys, func(a, b stationTime) int {
		return b.time.Compare(a.time)
	})

	for _, key := range keys {
		if hasValue(groups[key]) {
			return newObservation(groups[key]), key.time, true
		}
	}

	return Observations{}, time.Time{}, false
}

// apiURL is the endpoint of FMI's open data service. Tests override it
// to serve canned responses.
var apiURL = url.URL{
	Scheme: "https",
	Host:   "opendata.fmi.fi",
	Path:   "/wfs",
}

// fetch does a HTTP GET request against FMI's API and parses the
// response for a place
func fetch(ctx context.Context, place string) (simpleFeatureCollection, error) {
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
	measures := []string{
		paramTemperature,
		paramWindSpeed,
		paramWindGust,
		paramWindDirection,
		paramHumidity,
		paramPrecipitation,
		paramRainIntensity,
		paramSnowDepth,
		paramCloudCover,
		paramDewPoint,
		paramRadiation,
	}

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

	endpoint := apiURL
	endpoint.RawQuery = q.Encode()

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return simpleFeatureCollection{}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return simpleFeatureCollection{}, fmt.Errorf("%w: %w", ErrFetchFailed, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return simpleFeatureCollection{}, fmt.Errorf("virhe luettaessa havaintoja: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// FMI replies 400 with OperationParsingFailed when the place
		// cannot be parsed
		if resp.StatusCode == http.StatusBadRequest {
			return simpleFeatureCollection{}, ErrUnknownPlace
		}
		return simpleFeatureCollection{}, fmt.Errorf("%w (HTTP %d)", ErrFetchFailed, resp.StatusCode)
	}

	collection, err := parseFeatureCollection(body)
	if err != nil {
		return simpleFeatureCollection{}, err
	}
	if collection.Matched == 0 || collection.Returned == 0 {
		return simpleFeatureCollection{}, ErrNoObservations
	}

	return collection, nil
}
