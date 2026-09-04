package chardet_test

import (
	"testing"

	"github.com/levmv/chardet"
)

func TestUTF32AcceptsMaximumCodePoint(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		charset string
	}{
		{name: "big endian", input: []byte{0x00, 0x00, 0xfe, 0xff, 0x00, 0x10, 0xff, 0xff}, charset: "UTF-32BE"},
		{name: "little endian", input: []byte{0xff, 0xfe, 0x00, 0x00, 0xff, 0xff, 0x10, 0x00}, charset: "UTF-32LE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := chardet.NewTextDetector().DetectBest(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			if result.Charset != tt.charset || result.Confidence != 100 {
				t.Fatalf("DetectBest() = %s at %d confidence, want %s at 100", result.Charset, result.Confidence, tt.charset)
			}
		})
	}
}
