package qr

import (
	"net/url"
	"testing"
)

func TestVerificationURL(t *testing.T) {
	params := VerifactuParams{
		EmisorCIF: "A12345678",
		Serie:     "SER",
		Numero:    "456",
		Fecha:     "13-05-2026",
		Total:     150.75,
	}

	got := VerificationURL(params)
	
	u, err := url.Parse(got)
	if err != nil {
		t.Fatalf("Failed to parse generated URL: %v", err)
	}

	expectedBase := "https://www2.agenciatributaria.gob.es/static_files/common/internet/dep/aplicaciones/es/aeat/tike/cont/qr.html"
	if u.Scheme+"://"+u.Host+u.Path != expectedBase {
		t.Errorf("Base URL = %v, want %v", u.Scheme+"://"+u.Host+u.Path, expectedBase)
	}

	q := u.Query()
	tests := []struct {
		param string
		want  string
	}{
		{"nif", "A12345678"},
		{"numserie", "SER456"},
		{"fecha", "13-05-2026"},
		{"importe", "150.75"},
	}

	for _, tt := range tests {
		if gotVal := q.Get(tt.param); gotVal != tt.want {
			t.Errorf("URL param %s = %v, want %v", tt.param, gotVal, tt.want)
		}
	}
}
