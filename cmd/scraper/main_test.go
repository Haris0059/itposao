package main

import "testing"

func TestInBiH(t *testing.T) {
	for loc, want := range map[string]bool{
		"Sarajevo, Bosnia and Herzegovina": true,
		"Bosnia & Herzegovina":             true,
		"Banja Luka":                       true,
		"Sarajevo, BiH":                    true,
		"Tuzla":                            true,
		"Belgrade, Serbia":                 false,
		"Zagreb, Croatia":                  false,
		"Remote":                           false,
		"":                                 false,
	} {
		if got := inBiH(loc); got != want {
			t.Errorf("inBiH(%q) = %v, want %v", loc, got, want)
		}
	}
}
