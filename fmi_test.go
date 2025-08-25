package fmi

import (
	"strings"
	"testing"
)

func TestWeather(t *testing.T) {
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
}
