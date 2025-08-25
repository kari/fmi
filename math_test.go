package fmi

import (
	"math"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

const tolerance = 0.0001

func TestHumidex(t *testing.T) {
	var tests = []struct {
		t, td, h float64
	}{
		{30, 15, 33.96940099074554},
		{30, 25, 42.33964388030867},
	}
	for _, test := range tests {
		got := Humidex(test.t, test.td)
		if !cmp.Equal(got, test.h, cmpopts.EquateApprox(0, tolerance)) {
			t.Errorf("Humidex(%.f, %.f) = %f; want %f", test.t, test.td, got, test.h)
		}
	}
}

func TestWindChill(t *testing.T) {
	var tests = []struct {
		t, v, c float64
	}{
		{-20, 1.389, -24.278786},
		{-20, 8.333, -32.567782},
		{9, 4, 6.759861},
	}
	for _, test := range tests {
		got := WindChill(test.t, test.v)
		if !cmp.Equal(got, test.c, cmpopts.EquateApprox(0, tolerance)) {
			t.Errorf("WindChill(%.f, %.f) = %f; want %f", test.t, test.v, got, test.c)
		}
	}
}

func TestSummerSimmer(t *testing.T) {
	var tests = []struct {
		t, rh, s float64
	}{
		{10, 50, 10},
		{14.5, 50, 14.5},
		{20, 50, 20.000000000000004},
		{20, 90, 21.685823754789276},
	}
	for _, test := range tests {
		got := SummerSimmer(test.t, test.rh)
		if !cmp.Equal(got, test.s, cmpopts.EquateApprox(0, tolerance)) {
			t.Errorf("SummerSimmer(%.f, %.f) = %f; want %f", test.t, test.rh, got, test.s)
		}
	}
}

func TestWindChillFMI(t *testing.T) {
	var tests = []struct {
		t, v, c float64
	}{
		{0, 0, 0},
		{10, 0, 10},
		{0, 5, -4.9352732517894129},
		{-5, 5, -11.190933253695526},
	}
	for _, test := range tests {
		got := WindChillFMI(test.t, test.v)
		if !cmp.Equal(got, test.c, cmpopts.EquateApprox(0, tolerance)) {
			t.Errorf("windChillFMI(%.f, %.f) = %f; want %f", test.t, test.v, got, test.c)
		}
	}
}

func TestFeelsLike(t *testing.T) {
	var tests = []struct {
		t, v, rh, rad, f float64
	}{
		{0, 0, 50, math.NaN(), 0},
		{10, 0, 50, math.NaN(), 9.999999999999996},
		{0, 5, 50, math.NaN(), -4.979998860306697},
		{-5, 5, 50, math.NaN(), -10.652971679267061},
		{25, 5, 50, math.NaN(), 23.384865234495123},
		{25, 5, 90, math.NaN(), 26.58793036859474},
		{0, 0, 50, 0, -0.250},
		{10, 0, 50, 50, 9.994999999999996},
		{0, 5, 50, 800, -2.6166655269733634},
		{25, 5, 50, 425, 24.523198567828455},
	}
	for _, test := range tests {
		got := FeelsLike(test.t, test.v, test.rh, test.rad)
		if !cmp.Equal(got, test.f, cmpopts.EquateApprox(0, tolerance)) {
			t.Errorf("FeelsLike(%.f, %.f, %.f, %.f) = %f; want %f", test.t, test.v, test.rh, test.rad, got, test.f)
		}
	}
}
