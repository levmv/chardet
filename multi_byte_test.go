package chardet

import "testing"

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
