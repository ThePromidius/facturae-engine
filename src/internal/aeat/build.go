package aeat

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ThePromidius/facturae-engine/src/internal/chain"
	"github.com/ThePromidius/facturae-engine/src/internal/facturae"
)

func ExtractSignature(signedXML []byte) (string, error) {
	decoder := xml.NewDecoder(strings.NewReader(string(signedXML)))
	var buf strings.Builder
	depth := 0
	found := false
	hasDSNS := false

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("extracting signature: %w", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local == "Signature" && (t.Name.Space == "http://www.w3.org/2000/09/xmldsig#" || t.Name.Space == "ds") {
				found = true
				depth = 1
				fmt.Fprintf(&buf, "<ds:Signature")
				for _, attr := range t.Attr {
					if attr.Name.Space == "xmlns" && attr.Name.Local == "ds" {
						hasDSNS = true
					}
					writeAttr(&buf, attr)
				}
				if !hasDSNS {
					fmt.Fprintf(&buf, " xmlns:ds=\"%s\"", NSSig)
				}
				buf.WriteString(">")
				continue
			}
			if found {
				depth++
				encodeToken(&buf, t)
			}
		case xml.EndElement:
			if found {
				depth--
				if depth == 0 {
					buf.WriteString("</ds:Signature>")
					return buf.String(), nil
				}
				encodeToken(&buf, t)
			}
		case xml.CharData:
			if found {
				buf.Write(t)
			}
		case xml.Comment:
			if found {
				encodeToken(&buf, t)
			}
		case xml.ProcInst:
			if found {
				encodeToken(&buf, t)
			}
		case xml.Directive:
			if found {
				encodeToken(&buf, t)
			}
		}
	}
	return "", fmt.Errorf("extracting signature: ds:Signature element not found")
}

func writeAttr(buf *strings.Builder, a xml.Attr) {
	an := a.Name
	if an.Space == "xmlns" {
		fmt.Fprintf(buf, " xmlns:%s=%q", an.Local, a.Value)
	} else if an.Space != "" {
		fmt.Fprintf(buf, " %s:%s=%q", an.Space, an.Local, a.Value)
	} else {
		fmt.Fprintf(buf, " %s=%q", an.Local, a.Value)
	}
}

func encodeToken(buf *strings.Builder, t xml.Token) {
	switch v := t.(type) {
	case xml.StartElement:
		ns := v.Name.Space
		local := v.Name.Local
		if ns != "" {
			fmt.Fprintf(buf, "<%s:%s", ns, local)
		} else {
			fmt.Fprintf(buf, "<%s", local)
		}
		for _, a := range v.Attr {
			an := a.Name
			if an.Space != "" {
				fmt.Fprintf(buf, " %s:%s=%q", an.Space, an.Local, a.Value)
			} else {
				fmt.Fprintf(buf, " %s=%q", an.Local, a.Value)
			}
		}
		buf.WriteString(">")
	case xml.EndElement:
		ns := v.Name.Space
		local := v.Name.Local
		if ns != "" {
			fmt.Fprintf(buf, "</%s:%s>", ns, local)
		} else {
			fmt.Fprintf(buf, "</%s>", local)
		}
	case xml.CharData:
		buf.Write(v)
	case xml.Comment:
		buf.WriteString("<!--")
		buf.Write(v)
		buf.WriteString("-->")
	case xml.ProcInst:
		fmt.Fprintf(buf, "<?%s %s?>", v.Target, string(v.Inst))
	case xml.Directive:
		buf.WriteString("<!")
		buf.Write(v)
		buf.WriteString(">")
	}
}

type SuministroLRConfig struct {
	NombreSistemaInformatico string
	IdSistemaInformatico     string
	Version                  string
	NumeroInstalacion        string
	TipoUsoVerifactu         string
	TipoUsoMultiOT           string
	IndicadorMultiOT         string
}

func DefaultConfig() SuministroLRConfig {
	return SuministroLRConfig{
		NombreSistemaInformatico: "Vivex-SIF",
		IdSistemaInformatico:     "01",
		Version:                  "1.0.0",
		NumeroInstalacion:        "INST-001",
		TipoUsoVerifactu:         "S",
		TipoUsoMultiOT:           "N",
		IndicadorMultiOT:         "N",
	}
}

func BuildSuministroLR(
	f *facturae.FacturaE,
	inv facturae.Invoice,
	rec chain.Record,
	emisorCIF, emisorNombre string,
	signatureXML string,
	config SuministroLRConfig,
) *RegFactuSistemaFacturacion {
	hasPrev := rec.PreviousFingerprint != ""
	invNum := inv.InvoiceHeader.InvoiceNumber
	if s := inv.InvoiceHeader.InvoiceSeriesCode; s != "" {
		invNum = s + invNum
	}

	encadenamiento := EncadenamientoType{}
	if !hasPrev {
		encadenamiento.PrimerRegistro = "S"
	} else {
		encadenamiento.RegistroAnterior = &EncadenamientoFacturaAnteriorType{
			IDEmisorFactura:         emisorCIF,
			NumSerieFactura:         invNum,
			FechaExpedicionFactura:  formatFechaDDMMYYYY(inv.InvoiceIssueData.IssueDate),
			Huella:                  rec.PreviousFingerprint,
		}
	}

	detalles := make([]DetalleType, 0, len(inv.TaxesOutputs.Tax))
	for _, t := range inv.TaxesOutputs.Tax {
		detalles = append(detalles, DetalleType{
			Impuesto:                     mapTaxType(t.TaxTypeCode),
			ClaveRegimen:                 "01",
			CalificacionOperacion:        "S1",
			TipoImpositivo:               formatAmount(t.TaxRate, false),
			BaseImponibleOimporteNoSujeto: formatAmount(t.TaxableBase.TotalAmount, true),
			CuotaRepercutida:             formatAmount(t.TaxAmount.TotalAmount, true),
		})
	}

	sigField := ""
	if signatureXML != "" {
		sigField = signatureXML
	}

	return &RegFactuSistemaFacturacion{
		NsLR:   NSLR,
		NsInfo: NSInfo,
		NsSig:  NSSig,
		Cabecera: CabeceraType{
			ObligadoEmision: PersonaFisicaJuridicaESType{
				NombreRazon: emisorNombre,
				NIF:         emisorCIF,
			},
		},
		Registros: []RegistroFacturaType{
			{
				RegistroAlta: &RegistroFacturacionAltaType{
					IDVersion:            "1.0",
					IDFactura: IDFacturaExpedidaType{
						IDEmisorFactura:       emisorCIF,
						NumSerieFactura:       invNum,
						FechaExpedicionFactura: formatFechaDDMMYYYY(inv.InvoiceIssueData.IssueDate),
					},
					NombreRazonEmisor:    emisorNombre,
					TipoFactura:          mapInvoiceType(inv.InvoiceHeader.InvoiceDocumentType),
					DescripcionOperacion: "Factura " + inv.InvoiceHeader.InvoiceNumber,
					Desglose: DesgloseType{
						DetalleDesglose: detalles,
					},
					CuotaTotal:   formatAmount(inv.InvoiceTotals.TotalTaxOutputs, true),
					ImporteTotal: formatAmount(inv.InvoiceTotals.InvoiceTotal, true),
					Encadenamiento: encadenamiento,
					SistemaInformatico: SistemaInformaticoType{
						NombreRazon:                  emisorNombre,
						NIF:                          emisorCIF,
						NombreSistemaInformatico:     config.NombreSistemaInformatico,
						IdSistemaInformatico:         config.IdSistemaInformatico,
						Version:                      config.Version,
						NumeroInstalacion:            config.NumeroInstalacion,
						TipoUsoPosibleSoloVerifactu:  config.TipoUsoVerifactu,
						TipoUsoPosibleMultiOT:        config.TipoUsoMultiOT,
						IndicadorMultiplesOT:         config.IndicadorMultiOT,
					},
					FechaHoraHusoGenRegistro: time.Now().UTC().Format("2006-01-02T15:04:05Z"),
					TipoHuella:               "01",
					Huella:                   rec.Fingerprint,
					SignatureRaw:             sigField,
				},
			},
		},
	}
}

func formatAmount(v float64, signed bool) string {
	if signed {
		return strings.Replace(fmt.Sprintf("%.2f", v), ",", "", -1)
	}
	return strings.Replace(fmt.Sprintf("%.2f", v), ",", "", -1)
}

func formatFechaDDMMYYYY(date string) string {
	parts := strings.Split(date, "-")
	if len(parts) == 3 {
		return parts[2] + "-" + parts[1] + "-" + parts[0]
	}
	return date
}

func mapInvoiceType(docType string) string {
	switch docType {
	case "FC":
		return "F1"
	case "FA":
		return "F2"
	case "AF":
		return "R1"
	default:
		return "F1"
	}
}

func mapTaxType(taxCode string) string {
	switch taxCode {
	case "01":
		return "01"
	default:
		return "01"
	}
}


