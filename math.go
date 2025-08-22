package fmi

import (
	"math"
)

// Humidex calculates the humidity index given air temperature t (degC)
// and dew point (degC) td.
// For reference, see https://en.wikipedia.org/wiki/Humidex
func Humidex(t float64, td float64) float64 {
	return t + 5.0/9.0*(6.11*math.Exp(5417.7530*(1/273.16-1/(273.15+td)))-10)
}

// WindChill calculates the wind chill effect given air temperature t (degC)
// and wind speed v (m/s) using a Canadian formula.
// The calculation works for air temperatures at or below 10C and wind speeds above 0.4 m/s.
// For reference see,
// https://fi.m.wikipedia.org/wiki/Pakkasen_purevuus#Uusi_kaava
func WindChill(t float64, v float64) float64 {
	return 13.12 + 0.6215*t - 13.956*math.Pow(v, 0.16) + 0.4867*t*math.Pow(v, 0.16)
}

// WindChillFmi calculates wind chill with FMI's formula
// For reference see,
// https://github.com/fmidev/smartmet-library-newbase/blob/0da9473163883089c35a4c7267ba4c8a8bb3e14f/newbase/NFmiMetMath.cpp#L380
// https://tietopyynto.fi/tietopyynto/ilmatieteen-laitoksen-kayttama-tuntuu-kuin-laskentakaava/
func WindChillFMI(t float64, v float64) float64 {
	var kmh = v * 3.6

	if kmh < 5 {
		return t + (-1.59+0.1345*t)/5*kmh
	}
	return 13.12 + 0.6215*t - 11.37*math.Pow(kmh, 0.16) + 0.3965*t*math.Pow(kmh, 0.16)
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

// FeelsLike calculates FMI's "feel like" temperature
// For reference see,
// https://github.com/fmidev/smartmet-library-newbase/blob/0da9473163883089c35a4c7267ba4c8a8bb3e14f/newbase/NFmiMetMath.cpp#L418
// https://tietopyynto.fi/tietopyynto/ilmatieteen-laitoksen-kayttama-tuntuu-kuin-laskentakaava/
func FeelsLike(t float64, v float64, rh float64, rad float64) float64 {
	const a = 15.0
	const t0 = 37.0

	var chill = a + (1-a/t0)*t + a/t0*math.Pow(v+1, 0.16)*(t-t0)

	var heat = SummerSimmer(t, rh)
	var feels = t + (chill - t) + (heat - t)

	if !math.IsNaN(rad) {
		const absorption = 0.07
		feels += 0.7*absorption*rad/(v+10) - 0.25
	}

	return feels
}
