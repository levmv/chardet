package chardet

import (
	"encoding/binary"
	"strings"
	"testing"
)

func utf32Bytes(text string, order binary.ByteOrder) []byte {
	runes := []rune(text)
	raw := make([]byte, 4*len(runes))
	for i, c := range runes {
		order.PutUint32(raw[4*i:], uint32(c))
	}
	return raw
}

func TestUTF32TrailingBytes(t *testing.T) {
	for _, tt := range []struct {
		name    string
		order   binary.ByteOrder
		r       *recognizerUtf32
		invalid [][]byte
	}{
		{"BE", binary.BigEndian, newRecognizer_utf32be(), [][]byte{{0xff}, {0, 0x11}, {0, 0, 0xd8}, {0, 0, 0xdf}}},
		{"LE", binary.LittleEndian, newRecognizer_utf32le(), [][]byte{{0, 0, 0x11}, {0, 0xd8, 0}, {0, 0xdf, 0}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			confidence := func(raw []byte) int {
				return tt.r.Match(&recognizerInput{raw: raw}).Confidence
			}
			for _, bom := range []string{"", "\ufeff"} {
				prefix := utf32Bytes(bom+"Text", tt.order)
				// U+1D800 has low bytes that can still complete a valid LE scalar.
				for _, last := range []string{"界", "\U0001d800", "\U0010ffff"} {
					tail := utf32Bytes(last, tt.order)
					for n := 1; n <= 4; n++ {
						raw := append(append([]byte(nil), prefix...), tail[:n]...)
						want := 80
						if n == 4 {
							want = 100
						}
						if got := confidence(raw); got != want {
							t.Errorf("BOM=%t, tail=%x: confidence = %d, want %d", bom != "", tail[:n], got, want)
						}
						if n < 4 && confidence(tail[:n]) != 0 {
							t.Errorf("tail %x alone supplies no complete scalar", tail[:n])
						}
					}
				}
				for _, tail := range tt.invalid {
					raw := append(append([]byte(nil), prefix...), tail...)
					if got := confidence(raw); got != 0 {
						t.Errorf("BOM=%t, impossible tail=%x: confidence = %d, want 0", bom != "", tail, got)
					}
					// An impossible prefix counts as an error, with the existing ratio policy.
					raw = append(utf32Bytes(bom+strings.Repeat("A", 11), tt.order), tail...)
					want := 25
					if bom != "" {
						want = 80
					}
					if got := confidence(raw); got != want {
						t.Errorf("damaged text, BOM=%t, tail=%x: confidence = %d, want %d", bom != "", tail, got, want)
					}
				}
				// A possible truncation must not hide an earlier invalid scalar.
				raw := append(append([]byte(nil), prefix...), 0xff, 0xff, 0xff, 0xff, 0)
				if got := confidence(raw); got != 0 {
					t.Errorf("error before truncation, BOM=%t: confidence = %d, want 0", bom != "", got)
				}
			}
		})
	}
}
