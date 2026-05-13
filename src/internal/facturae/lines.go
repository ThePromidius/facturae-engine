// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

package facturae

import (
	"fmt"
	"math"

	"github.com/ThePromidius/facturae-engine/src/internal/invoice"
)

// taxKey is an internal composite key used to group tax lines by type code and rate.
type taxKey struct {
	code string
	rate float64
}

// taxBucket accumulates the taxable base and tax amount for a given tax key during line processing.
type taxBucket struct {
	base   float64
	amount float64
}

// buildLines converts invoice line items into Facturae invoice lines, tax summary entries,
// and invoice totals. It groups line-level taxes by type and rate, then aggregates them.
func buildLines(lineas []invoice.Linea, moneda string) ([]InvoiceLine, []TaxOutput, *InvoiceTotals, error) {
	taxMap := make(map[taxKey]*taxBucket)
	var lines []InvoiceLine
	var grossTotal float64

	for i, l := range lineas {
		if l.Cantidad <= 0 {
			return nil, nil, nil, fmt.Errorf("linea %d: cantidad debe ser > 0", i)
		}
		lineGross := round2(l.Cantidad * l.PrecioUnitario)
		taxAmt := round2(lineGross * l.IVATipo / 100.0)

		key := taxKey{code: taxTypeIVA, rate: l.IVATipo}
		if _, ok := taxMap[key]; !ok {
			taxMap[key] = &taxBucket{}
		}
		taxMap[key].base += lineGross
		taxMap[key].amount += taxAmt

		lines = append(lines, InvoiceLine{
			ItemDescription:     l.Descripcion,
			Quantity:            l.Cantidad,
			UnitPriceWithoutTax: l.PrecioUnitario,
			TotalCost:           lineGross,
			GrossAmount:         lineGross,
			TaxesOutputs: LineTaxes{
				Tax: []LineTax{{
					TaxTypeCode: taxTypeIVA,
					TaxRate:     l.IVATipo,
					TaxableBase: Amount{TotalAmount: lineGross},
					TaxAmount:   Amount{TotalAmount: taxAmt},
				}},
			},
		})
		grossTotal += lineGross
	}

	var taxSummary []TaxOutput
	var totalTax float64
	for k, b := range taxMap {
		base := round2(b.base)
		amt := round2(b.amount)
		totalTax += amt
		taxSummary = append(taxSummary, TaxOutput{
			TaxTypeCode: k.code,
			TaxRate:     k.rate,
			TaxableBase: Amount{TotalAmount: base},
			TaxAmount:   Amount{TotalAmount: amt},
		})
	}

	gross := round2(grossTotal)
	invoiceTotal := round2(gross + totalTax)

	totals := &InvoiceTotals{
		TotalGrossAmount:            gross,
		TotalGrossAmountBeforeTaxes: gross,
		TotalTaxOutputs:             round2(totalTax),
		InvoiceTotal:                invoiceTotal,
		TotalOutstandingAmount:      invoiceTotal,
		TotalExecutableAmount:       invoiceTotal,
	}

	return lines, taxSummary, totals, nil
}

// round2 rounds a float64 to two decimal places using standard math.Round.
func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
