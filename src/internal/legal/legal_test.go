package legal

import (
	"strings"
	"testing"
)

func TestGenerateDeclaracionResponsable(t *testing.T) {
	info := SoftwareInfo{
		Name:        "AwesomePOS",
		Version:     "2.4.1",
		ProducerNIF: "B87654321",
	}

	got := GenerateDeclaracionResponsable(info)

	requiredPhrases := []string{
		"DECLARACIÓN RESPONSABLE",
		"AwesomePOS",
		"2.4.1",
		"B87654321",
		"integridad, conservación, accesibilidad",
		"Veri*factu",
	}

	for _, phrase := range requiredPhrases {
		if !strings.Contains(got, phrase) {
			t.Errorf("Declaration missing phrase: %q", phrase)
		}
	}
}
