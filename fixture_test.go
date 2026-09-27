package fmi

import (
	"context"
	_ "embed"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

//go:embed testdata/helsinki.xml
var helsinkiXML []byte

//go:embed testdata/pihtipudas.xml
var pihtipudasXML []byte

// fixtureTime is the timestamp of the winning observation in both
// fixtures: the newest timestep of the first station.
var fixtureTime = time.Date(2026, 9, 27, 6, 40, 0, 0, time.UTC)

// The fixtures were captured from the live API on 2026-09-27 and contain
// two stations with two 10-minute timesteps each. In both, the first
// station has no radiation measure and every station misses the hourly
// precipitation, so the expected observations pin how NaN measures are
// carried through extraction and formatting.
func helsinkiFixtureObservations() Observations {
	o := nanObservation()
	o.Temperature = 11.1
	o.WindSpeed = 3.9
	o.WindGust = 6.3
	o.WindDirection = 282
	o.Humidity = 86
	o.DewPoint = 8.8
	o.RainIntensity = 0.0
	o.SnowDepth = -1
	o.CloudCover = 0.0
	return o
}

func pihtipudasFixtureObservations() Observations {
	o := nanObservation()
	o.Temperature = 9.5
	o.WindSpeed = 3.1
	o.WindGust = 8.1
	o.WindDirection = 284
	o.Humidity = 85
	o.DewPoint = 7.1
	o.RainIntensity = 0.0
	o.SnowDepth = -1
	o.CloudCover = 7.0
	return o
}

func TestExtractFromFixtures(t *testing.T) {
	var tests = []struct {
		name string
		xml  []byte
		want Observations
	}{
		{name: "helsinki", xml: helsinkiXML, want: helsinkiFixtureObservations()},
		{name: "pihtipudas", xml: pihtipudasXML, want: pihtipudasFixtureObservations()},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			collection, err := parseFeatureCollection(test.xml)
			if err != nil {
				t.Fatalf("parseFeatureCollection() returned error: %v", err)
			}
			got, measuredAt, found := extractLatestObservations(collection)
			if !found {
				t.Fatal("extractLatestObservations() found no populated observation")
			}
			if !measuredAt.Equal(fixtureTime) {
				t.Errorf("measuredAt = %v, want %v", measuredAt, fixtureTime)
			}
			if !cmp.Equal(got, test.want, cmpopts.EquateNaNs()) {
				t.Errorf("observation mismatch (-got +want):\n%s", cmp.Diff(got, test.want, cmpopts.EquateNaNs()))
			}
		})
	}
}

// pointAPIAt redirects the package's API endpoint to a stub server for
// the duration of the test.
func pointAPIAt(t *testing.T, rawURL string) {
	t.Helper()

	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parsing stub server URL: %v", err)
	}

	old := apiURL
	apiURL = *u
	t.Cleanup(func() { apiURL = old })
}

// stubFMI serves canned responses keyed by the place query parameter and
// points apiURL at itself for the duration of the test.
func stubFMI(t *testing.T, fixtures map[string][]byte) {
	t.Helper()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, ok := fixtures[r.URL.Query().Get("place")]
		if !ok {
			http.Error(w, "unknown place", http.StatusBadRequest)
			return
		}
		_, _ = w.Write(data)
	}))
	t.Cleanup(ts.Close)

	pointAPIAt(t, ts.URL)
}

func TestCurrentFromFixtures(t *testing.T) {
	stubFMI(t, map[string][]byte{
		"Helsinki":   helsinkiXML,
		"Pihtipudas": pihtipudasXML,
	})

	var tests = []struct {
		place string
		want  Conditions
	}{
		{
			place: "Helsinki",
			want: Conditions{
				Place:        "Helsinki",
				Time:         fixtureTime,
				Observations: helsinkiFixtureObservations(),
			},
		},
		{
			place: "Pihtipudas",
			want: Conditions{
				Place:        "Pihtipudas",
				Time:         fixtureTime,
				Observations: pihtipudasFixtureObservations(),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.place, func(t *testing.T) {
			got, err := Current(context.Background(), test.place)
			if err != nil {
				t.Fatalf("Current(%q) returned error: %v", test.place, err)
			}
			if !cmp.Equal(got, test.want, cmpopts.EquateNaNs(), cmpopts.EquateApproxTime(time.Second)) {
				t.Errorf("Weather mismatch (-got +want):\n%s", cmp.Diff(got, test.want, cmpopts.EquateNaNs(), cmpopts.EquateApproxTime(time.Second)))
			}
		})
	}
}

func TestWeatherFromFixtures(t *testing.T) {
	stubFMI(t, map[string][]byte{
		"Helsinki":   helsinkiXML,
		"Pihtipudas": pihtipudasXML,
	})

	var tests = []struct {
		place string
		want  string
	}{
		{"Helsinki", "Viimeisimmät säähavainnot paikassa Helsinki: lämpötila 11.1°C (tuntuu kuin 8.1°C), selkeää, heikkoa länsituulta 3.9 m/s (6.3 m/s), ilmankosteus 86%"},
		{"Pihtipudas", "Viimeisimmät säähavainnot paikassa Pihtipudas: lämpötila 9.5°C (tuntuu kuin 6.7°C), melko pilvistä, heikkoa länsituulta 3.1 m/s (8.1 m/s), ilmankosteus 85%"},
	}

	for _, test := range tests {
		t.Run(test.place, func(t *testing.T) {
			got, err := Weather(test.place)
			if err != nil {
				t.Fatalf("Weather(%q) returned error: %v", test.place, err)
			}
			if got != test.want {
				t.Errorf("Weather(%q):\ngot  %s\nwant %s", test.place, got, test.want)
			}
		})
	}

	if _, err := Weather("Narnia"); !errors.Is(err, ErrUnknownPlace) {
		t.Errorf("Weather('Narnia') error = %v, want ErrUnknownPlace", err)
	}
}

// TestWeatherErrorPaths covers the failure modes the fixture stub does
// not serve: an unexpected HTTP status and a response with no members.
func TestWeatherErrorPaths(t *testing.T) {
	const emptyCollection = `<?xml version="1.0" encoding="UTF-8"?>
<wfs:FeatureCollection timeStamp="2026-09-27T06:44:07Z" numberMatched="0" numberReturned="0" xmlns:wfs="http://www.opengis.net/wfs/2.0"/>`

	var tests = []struct {
		name   string
		status int
		body   string
		want   error
	}{
		{name: "server error", status: http.StatusInternalServerError, body: "server error", want: ErrFetchFailed},
		{name: "empty feature collection", status: http.StatusOK, body: emptyCollection, want: ErrNoObservations},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(test.status)
				_, _ = w.Write([]byte(test.body))
			}))
			t.Cleanup(ts.Close)

			pointAPIAt(t, ts.URL)

			if _, err := Weather("Mordor"); !errors.Is(err, test.want) {
				t.Errorf("Weather('Mordor') error = %v, want %v", err, test.want)
			}
		})
	}
}

// TestCurrentErrors covers Current-specific behavior: the empty place
// check and a cancelled context.
func TestCurrentErrors(t *testing.T) {
	if _, err := Current(context.Background(), ""); !errors.Is(err, ErrNoPlace) {
		t.Errorf("Current(\"\") error = %v, want ErrNoPlace", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(ts.Close)
	pointAPIAt(t, ts.URL)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Current(ctx, "Mordor")
	if !errors.Is(err, ErrFetchFailed) || !errors.Is(err, context.Canceled) {
		t.Errorf("Current() with cancelled context error = %v, want ErrFetchFailed wrapping context.Canceled", err)
	}
}
