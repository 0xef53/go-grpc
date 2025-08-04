package utils

import (
	"testing"
)

func TestHostportNormalization(t *testing.T) {
	type value struct {
		Orig string
		Want string
	}

	values := []value{
		{"   A.B.C.D", "A.B.C.D:9191"},
		{" 192.168.0.1   ", "192.168.0.1:9191"},
		{"192.168.0.1:5555", "192.168.0.1:5555"},
		{" 2a01:5560:1001:e9fe:21::37 ", "[2a01:5560:1001:e9fe:21::37]:9191"},
		{"[2a01:5560:1001:e9fe:21::37]  ", "[2a01:5560:1001:e9fe:21::37]:9191"},
		{" [2a01:5560:1001:e9fe:21::37]:1234  ", "[2a01:5560:1001:e9fe:21::37]:1234"},
		{"[2a01::]:1234 ", "[2a01::]:1234"},
		{" [2a01::]   ", "[2a01::]:9191"},
		{"2a01::   ", "[2a01::]:9191"},
	}

	for idx, v := range values {
		got := NormalizeHostport(v.Orig)
		if got != v.Want {
			t.Fatalf("got invalid result (idx == %d):\nwant:\t%q\ngot:\t%q", idx, v.Want, got)
		}
	}
}
