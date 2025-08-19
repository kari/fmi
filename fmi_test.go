package fmi

import (
	"strings"
	"testing"
)

func TestWeather(t *testing.T) {
	s := Weather("Helsinki")
	if !strings.Contains(s, "Helsinki") {
		t.Errorf("Weather('Helsinki') should contain 'Helsinki', instead got '%s'", s)
	}
	s2 := Weather("Narnia")
	if s2 != "säähavaintoja ei löytynyt" {
		t.Errorf("Weather('Narnia') should return 'säähavaintoja ei löytynyt', instead got '%s'", s2)
	}
	s3 := Weather("")
	if s3 != "Paikkaa ei syötetty" {
		t.Errorf("Weather('') should return 'Paikkaa ei syötetty', instead got '%s'", s3)
	}
	s4 := Weather("ajfhjasdf")
	if s4 != "säähavaintopaikkaa ei löytynyt" {
		t.Errorf("Weather('ajfhjasdf') should return 'säähavaintopaikkaa ei löytynyt', instead got '%s'", s4)
	}
	s5 := Weather("Pihtipudas")
	if !strings.Contains(s5, "Pihtipudas") {
		t.Errorf("Weather('Pihtipudas') should contain 'Pihtipudas', instead got '%s'", s5)
	}

}
