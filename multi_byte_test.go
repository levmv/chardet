package chardet

import (
	"fmt"
	"testing"
)

func TestMultiByteConfidenceAtEndOfInput(t *testing.T) {
	type confidenceCase struct {
		name  string
		input []byte
		want  int
	}

	tests := []struct {
		name       string
		recognizer *recognizerMultiByte
		complete   []byte
		malformed  []byte
	}{
		{"Shift_JIS", newRecognizer_sjis(), []byte{0x82, 0xa0}, []byte{0x82, 0x20}},
		{"EUC-JP two byte", newRecognizer_euc_jp(), []byte{0xa4, 0xa2}, []byte{0xa4, 0x20}},
		{"EUC-JP three byte", newRecognizer_euc_jp(), []byte{0x8f, 0xa1, 0xa2}, []byte{0x8f, 0xa1, 0x20}},
		{"EUC-KR", newRecognizer_euc_kr(), []byte{0xb0, 0xa1}, []byte{0xb0, 0x20}},
		{"Big5", newRecognizer_big5(), []byte{0xa4, 0x40}, []byte{0xa4, 0x20}},
		{"GB18030 two byte", newRecognizer_gb_18030(), []byte{0xb5, 0xc4}, []byte{0xb5, 0x20}},
		{"GB18030 four byte", newRecognizer_gb_18030(), []byte{0x81, 0x30, 0x81, 0x30}, []byte{0x81, 0x30, 0x81, 0x20}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const ascii = "aaaaaaaaaa"
			cases := []confidenceCase{
				{"empty", nil, 0},
				{"complete character", tt.complete, 10},
				{"short ASCII", []byte(ascii[:9]), 0},
				{"ASCII threshold", []byte(ascii), 10},
				{"malformed final character", append([]byte(ascii), tt.malformed...), 0},
			}
			for n := 1; n < len(tt.complete); n++ {
				cases = append(cases, confidenceCase{
					fmt.Sprintf("truncated after %d bytes", n),
					append([]byte(ascii), tt.complete[:n]...),
					0,
				})
			}
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					got := tt.recognizer.Match(newRecognizerInput(tc.input, false)).Confidence
					if got != tc.want {
						t.Fatalf("Match(%x).Confidence = %d, want %d", tc.input, got, tc.want)
					}
				})
			}
		})
	}
}

func TestVariableWidthDecodersPreserveCompleteSequence(t *testing.T) {
	tests := []struct {
		name    string
		decoder charDecoder
		input   []byte
		want    uint32
	}{
		{name: "EUC three byte", decoder: charDecoder_euc{}, input: []byte{0x8f, 0xa1, 0xa2}, want: 0x8fa1a2},
		{name: "GB18030 four byte", decoder: charDecoder_gb_18030{}, input: []byte{0x81, 0x30, 0x81, 0x30}, want: 0x81308130},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, remaining, err := tt.decoder.DecodeOneChar(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			if len(remaining) != 0 {
				t.Fatalf("DecodeOneChar() left %d bytes, want none", len(remaining))
			}
			if got != tt.want {
				t.Fatalf("DecodeOneChar() = %#x, want %#x", got, tt.want)
			}
		})
	}
}
