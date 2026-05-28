// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

package aeat

import "encoding/xml"

const (
	NSLR      = "https://www2.agenciatributaria.gob.es/static_files/common/internet/dep/aplicaciones/es/aeat/tike/cont/ws/SuministroLR.xsd"
	NSInfo    = "https://www2.agenciatributaria.gob.es/static_files/common/internet/dep/aplicaciones/es/aeat/tike/cont/ws/SuministroInformacion.xsd"
	NSResp    = "https://www2.agenciatributaria.gob.es/static_files/common/internet/dep/aplicaciones/es/aeat/tike/cont/ws/RespuestaSuministro.xsd"
	NSSig     = "http://www.w3.org/2000/09/xmldsig#"
	SOAPAction = "http://www.agenciatributaria.gob.es/AEAT/VERIFACTU/RegistroFacturacion"
)

type Environment string

const (
	EnvTest Environment = "test"
	EnvProd Environment = "prod"
)

var Endpoints = map[Environment]string{
	EnvTest: "https://prewww1.aeat.es/wlpl/VERIFACTU-CONT/ws/VeriFactuSOAP",
	EnvProd: "https://www1.aeat.es/wlpl/VERIFACTU-CONT/ws/VeriFactuSOAP",
}

type SubmitResult struct {
	CSV         string
	Estado      string
	Descripcion string
	HTTPStatus  int
}

func (r SubmitResult) IsAccepted() bool {
	return r.Estado == "Correcto" || r.Estado == "ParcialmenteCorrecto" || r.Estado == "AceptadoConErrores"
}

type RegFactuSistemaFacturacion struct {
	XMLName   xml.Name `xml:"sfLR:RegFactuSistemaFacturacion"`
	NsLR      string   `xml:"xmlns:sfLR,attr"`
	NsInfo    string   `xml:"xmlns:sf,attr"`
	NsSig     string   `xml:"xmlns:ds,attr,omitempty"`
	Cabecera  CabeceraType          `xml:"sfLR:Cabecera"`
	Registros []RegistroFacturaType `xml:"sfLR:RegistroFactura"`
}

type CabeceraType struct {
	ObligadoEmision PersonaFisicaJuridicaESType `xml:"sf:ObligadoEmision"`
}

type PersonaFisicaJuridicaESType struct {
	NombreRazon string `xml:"sf:NombreRazon"`
	NIF         string `xml:"sf:NIF"`
}

type RegistroFacturaType struct {
	RegistroAlta *RegistroFacturacionAltaType `xml:"sf:RegistroAlta,omitempty"`
}

type RegistroFacturacionAltaType struct {
	IDVersion                  string                        `xml:"sf:IDVersion"`
	IDFactura                  IDFacturaExpedidaType         `xml:"sf:IDFactura"`
	RefExterna                 string                        `xml:"sf:RefExterna,omitempty"`
	NombreRazonEmisor          string                        `xml:"sf:NombreRazonEmisor"`
	Subsanacion                string                        `xml:"sf:Subsanacion,omitempty"`
	RechazoPrevio              string                        `xml:"sf:RechazoPrevio,omitempty"`
	TipoFactura                string                        `xml:"sf:TipoFactura"`
	TipoRectificativa          string                        `xml:"sf:TipoRectificativa,omitempty"`
	FechaOperacion             string                        `xml:"sf:FechaOperacion,omitempty"`
	DescripcionOperacion       string                        `xml:"sf:DescripcionOperacion"`
	FacturaSimplificadaArt7273 string                        `xml:"sf:FacturaSimplificadaArt7273,omitempty"`
	Macrodato                  string                        `xml:"sf:Macrodato,omitempty"`
	EmitidaPorTerceroODestinatario string                     `xml:"sf:EmitidaPorTerceroODestinatario,omitempty"`
	Desglose                   DesgloseType                  `xml:"sf:Desglose"`
	CuotaTotal                 string                        `xml:"sf:CuotaTotal"`
	ImporteTotal               string                        `xml:"sf:ImporteTotal"`
	Encadenamiento             EncadenamientoType             `xml:"sf:Encadenamiento"`
	SistemaInformatico         SistemaInformaticoType         `xml:"sf:SistemaInformatico"`
	FechaHoraHusoGenRegistro   string                        `xml:"sf:FechaHoraHusoGenRegistro"`
	NumRegistroAcuerdoFacturacion string                      `xml:"sf:NumRegistroAcuerdoFacturacion,omitempty"`
	IdAcuerdoSistemaInformatico  string                      `xml:"sf:IdAcuerdoSistemaInformatico,omitempty"`
	TipoHuella                 string                        `xml:"sf:TipoHuella"`
	Huella                     string                        `xml:"sf:Huella"`
	SignatureRaw               string                        `xml:",innerxml"`
}

type IDFacturaExpedidaType struct {
	IDEmisorFactura     string `xml:"sf:IDEmisorFactura"`
	NumSerieFactura     string `xml:"sf:NumSerieFactura"`
	FechaExpedicionFactura string `xml:"sf:FechaExpedicionFactura"`
}

type EncadenamientoType struct {
	PrimerRegistro  string                           `xml:"sf:PrimerRegistro,omitempty"`
	RegistroAnterior *EncadenamientoFacturaAnteriorType `xml:"sf:RegistroAnterior,omitempty"`
}

type EncadenamientoFacturaAnteriorType struct {
	IDEmisorFactura     string `xml:"sf:IDEmisorFactura"`
	NumSerieFactura     string `xml:"sf:NumSerieFactura"`
	FechaExpedicionFactura string `xml:"sf:FechaExpedicionFactura"`
	Huella              string `xml:"sf:Huella"`
}

type SistemaInformaticoType struct {
	NombreRazon                  string `xml:"sf:NombreRazon"`
	NIF                          string `xml:"sf:NIF"`
	NombreSistemaInformatico     string `xml:"sf:NombreSistemaInformatico"`
	IdSistemaInformatico         string `xml:"sf:IdSistemaInformatico"`
	Version                      string `xml:"sf:Version"`
	NumeroInstalacion            string `xml:"sf:NumeroInstalacion"`
	TipoUsoPosibleSoloVerifactu  string `xml:"sf:TipoUsoPosibleSoloVerifactu"`
	TipoUsoPosibleMultiOT        string `xml:"sf:TipoUsoPosibleMultiOT"`
	IndicadorMultiplesOT         string `xml:"sf:IndicadorMultiplesOT"`
}

type DesgloseType struct {
	DetalleDesglose []DetalleType `xml:"sf:DetalleDesglose"`
}

type DetalleType struct {
	Impuesto                 string `xml:"sf:Impuesto,omitempty"`
	ClaveRegimen             string `xml:"sf:ClaveRegimen,omitempty"`
	CalificacionOperacion    string `xml:"sf:CalificacionOperacion,omitempty"`
	OperacionExenta          string `xml:"sf:OperacionExenta,omitempty"`
	TipoImpositivo           string `xml:"sf:TipoImpositivo,omitempty"`
	BaseImponibleOimporteNoSujeto string `xml:"sf:BaseImponibleOimporteNoSujeto"`
	BaseImponibleACoste      string `xml:"sf:BaseImponibleACoste,omitempty"`
	CuotaRepercutida         string `xml:"sf:CuotaRepercutida,omitempty"`
	TipoRecargoEquivalencia  string `xml:"sf:TipoRecargoEquivalencia,omitempty"`
	CuotaRecargoEquivalencia string `xml:"sf:CuotaRecargoEquivalencia,omitempty"`
}

type soapResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		Respuesta struct {
			CSV         string `xml:"CSV"`
			EstadoEnvio string `xml:"EstadoEnvio"`
		} `xml:"RespuestaRegFactuSistemaFacturacion"`
	} `xml:"Body"`
}
