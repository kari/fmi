package fmi

import (
	"context"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestWeather(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live FMI API test in short mode")
	}

	s, err := Weather("Helsinki")
	if err != nil || !strings.Contains(s, "Helsinki") {
		t.Errorf("Weather('Helsinki') should contain 'Helsinki', instead got '%s'", s)
	}
	s2, err := Weather("Narnia")
	if err == nil {
		t.Errorf("Weather('Narnia') should return with an error, instead got '%s'", s2)
	}
	s3, err := Weather("")
	if err == nil {
		t.Errorf("Weather('') should return with an error, instead got '%s'", s3)
	}
	s4, err := Weather("ajfhjasdf")
	if err == nil {
		t.Errorf("Weather('ajfhjasdf') should return an error, instead got '%s'", s4)
	}
	s5, err := Weather("Pihtipudas")
	if err != nil || !strings.Contains(s5, "Pihtipudas") {
		t.Errorf("Weather('Pihtipudas') should contain 'Pihtipudas', instead got '%s'", s5)
	}

	w, err := Current(context.Background(), "Helsinki")
	if err != nil || w.Place != "Helsinki" || w.Time.IsZero() {
		t.Errorf("Current('Helsinki') should return the place and a timestamp, instead got %+v, %v", w, err)
	}
	if s := w.String(); !strings.Contains(s, "Helsinki") {
		t.Errorf("Current('Helsinki').String() should contain 'Helsinki', instead got '%s'", s)
	}
}

func TestExtractLatestObservations(t *testing.T) {
	const (
		helsinki = "Helsinki"
		vantaa   = "Vantaa"
	)
	newer := time.Date(2024, 6, 1, 12, 10, 0, 0, time.UTC)
	older := newer.Add(-10 * time.Minute)

	newestTemp := nanObservation()
	newestTemp.Temperature = 10

	olderTemp := nanObservation()
	olderTemp.Temperature = 5

	vantaaTemp := nanObservation()
	vantaaTemp.Temperature = 7

	windAndCloud := nanObservation()
	windAndCloud.WindSpeed = 3.5
	windAndCloud.WindDirection = 225
	windAndCloud.CloudCover = 2

	var tests = []struct {
		name      string
		elements  []measure
		want      Observations
		wantTime  time.Time
		wantFound bool
	}{
		{
			name: "newest populated observation wins over older ones",
			elements: []measure{
				{Time: older, Location: helsinki, Parameter: paramTemperature, Value: 5},
				{Time: newer, Location: helsinki, Parameter: paramTemperature, Value: 10},
			},
			want:      newestTemp,
			wantTime:  newer,
			wantFound: true,
		},
		{
			name: "all-NaN newest observation falls back to an older one",
			elements: []measure{
				{Time: older, Location: helsinki, Parameter: paramTemperature, Value: 5},
				{Time: newer, Location: helsinki, Parameter: paramTemperature, Value: math.NaN()},
			},
			want:      olderTemp,
			wantTime:  older,
			wantFound: true,
		},
		{
			name: "all-NaN station falls back to another station at the same time",
			elements: []measure{
				{Time: newer, Location: helsinki, Parameter: paramTemperature, Value: math.NaN()},
				{Time: newer, Location: vantaa, Parameter: paramTemperature, Value: 7},
			},
			want:      vantaaTemp,
			wantTime:  newer,
			wantFound: true,
		},
		{
			name: "first station wins over later stations at the same time",
			elements: []measure{
				{Time: newer, Location: vantaa, Parameter: paramTemperature, Value: 7},
				{Time: newer, Location: helsinki, Parameter: paramTemperature, Value: 5},
			},
			want:      vantaaTemp,
			wantTime:  newer,
			wantFound: true,
		},
		{
			name: "measures map to matching fields and unknown parameters are dropped",
			elements: []measure{
				{Time: newer, Location: helsinki, Parameter: paramWindSpeed, Value: 3.5},
				{Time: newer, Location: helsinki, Parameter: paramWindDirection, Value: 225},
				{Time: newer, Location: helsinki, Parameter: paramCloudCover, Value: 2},
				{Time: newer, Location: helsinki, Parameter: "p_sea", Value: 1005.5},
			},
			want:      windAndCloud,
			wantTime:  newer,
			wantFound: true,
		},
		{
			name: "only all-NaN observations yields none",
			elements: []measure{
				{Time: newer, Location: helsinki, Parameter: paramTemperature, Value: math.NaN()},
				{Time: older, Location: vantaa, Parameter: paramWindSpeed, Value: math.NaN()},
			},
			wantFound: false,
		},
		{
			name:      "no elements yields none",
			elements:  nil,
			wantFound: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, measuredAt, found := extractLatestObservations(simpleFeatureCollection{Elements: test.elements})
			if found != test.wantFound {
				t.Fatalf("found = %t, want %t", found, test.wantFound)
			}
			if found && !measuredAt.Equal(test.wantTime) {
				t.Errorf("measuredAt = %v, want %v", measuredAt, test.wantTime)
			}
			if !cmp.Equal(got, test.want, cmpopts.EquateNaNs()) {
				t.Errorf("observation mismatch (-got +want):\n%s", cmp.Diff(got, test.want, cmpopts.EquateNaNs()))
			}
		})
	}
}
