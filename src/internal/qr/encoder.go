// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

package qr

import "fmt"

// encode generates a simplified QR code module matrix for the given text string.
// It supports versions 1â€“10 based on input length and draws finder patterns,
// timing patterns, and a dark module. The returned modules slice is a flat
// row-major boolean array where true represents a dark module.
func encode(text string) (modules []bool, size int, err error) {
	if len(text) > 300 {
		return nil, 0, fmt.Errorf("text too long for simplified encoder (%d chars, max 300)", len(text))
	}

	version := 1
	switch {
	case len(text) <= 17:
		version = 1
	case len(text) <= 32:
		version = 2
	case len(text) <= 53:
		version = 3
	case len(text) <= 78:
		version = 4
	case len(text) <= 106:
		version = 5
	case len(text) <= 134:
		version = 6
	case len(text) <= 154:
		version = 7
	case len(text) <= 192:
		version = 8
	case len(text) <= 230:
		version = 9
	default:
		version = 10
	}

	size = 17 + 4*version
	modules = make([]bool, size*size)

	set := func(row, col int, dark bool) {
		if row >= 0 && row < size && col >= 0 && col < size {
			modules[row*size+col] = dark
		}
	}

	drawFinder := func(r, c int) {
		for dr := -1; dr <= 7; dr++ {
			for dc := -1; dc <= 7; dc++ {
				dark := (dr == -1 || dr == 7 || dc == -1 || dc == 7) ||
					(dr >= 1 && dr <= 5 && dc >= 1 && dc <= 5 &&
						!(dr >= 2 && dr <= 4 && dc >= 2 && dc <= 4))
				set(r+dr, c+dc, dark)
			}
		}
	}
	drawFinder(0, 0)
	drawFinder(0, size-7)
	drawFinder(size-7, 0)

	for i := 8; i < size-8; i++ {
		set(6, i, i%2 == 0)
		set(i, 6, i%2 == 0)
	}

	set(8, size-8, true)

	data := []byte(text)
	dataIdx := 0
	bitIdx := 0

	for col := size - 1; col >= 1; col -= 2 {
		if col == 6 {
			col--
		}
		for rowBase := 0; rowBase < size; rowBase++ {
			for dx := 0; dx < 2; dx++ {
				c := col - dx
				r := rowBase
				if modules[r*size+c] {
					continue
				}
				var bit bool
				if dataIdx < len(data) {
					bit = (data[dataIdx]>>uint(7-bitIdx))&1 == 1
					bitIdx++
					if bitIdx == 8 {
						bitIdx = 0
						dataIdx++
					}
				}
				set(r, c, bit)
			}
		}
	}

	return modules, size, nil
}
