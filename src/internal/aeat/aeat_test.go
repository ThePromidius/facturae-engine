package aeat

import (
	"encoding/xml"
	"strings"
	"testing"
)

func TestXMLMarshalling(t *testing.T) {
	reg := Registro{
		IDFactura: IDFactura{
			IDEmisorFactura: IDEmisor{NIF: "A12345678"},
			NumSerieFactura: "S123",
			FechaExpedicion: "2026-05-13",
		},
		TipoFactura: "F1",
		Huella:      "CURRENT-HASH",
		RegistroAnterior: &RegistroAnterior{
			IDEmisorFactura: IDEmisor{NIF: "A12345678"},
			NumSerieFactura: "S122",
			FechaExpedicion: "2026-05-12",
			Huella:          "PREV-HASH",
		},
	}

	sum := SuministroLR{
		XmlnsSum: "http://www.agenciatributaria.gob.es/AEAT/VERIFACTU/SistemaFacturacion.xsd",
		Registro: reg,
	}

	output, err := xml.MarshalIndent(sum, "", "  ")
	if err != nil {
		t.Fatalf("xml.Marshal failed: %v", err)
	}

	xmlStr := string(output)
	
	// Check for critical nodes and namespaces
	expectedNodes := []string{
		"<sum:RegFactuSistemaFacturacion",
		"<sum:IDFactura>",
		"<sum:RegistroAnterior>",
		"<sum:Huella>PREV-HASH</sum:Huella>",
	}

	for _, node := range expectedNodes {
		if !strings.Contains(xmlStr, node) {
			t.Errorf("XML output missing node: %s\nOutput: %s", node, xmlStr)
		}
	}
}
