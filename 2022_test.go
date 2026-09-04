package chardet_test

import (
	"testing"

	"github.com/levmv/chardet"
)

func TestISO2022Language(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		charset  string
		language string
	}{
		{name: "Japanese", input: []byte("\x1b$B"), charset: "ISO-2022-JP", language: "ja"},
		{name: "Korean", input: []byte("\x1b$)C"), charset: "ISO-2022-KR", language: "ko"},
		{name: "Chinese", input: []byte("\x1b$)A"), charset: "ISO-2022-CN", language: "zh"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := chardet.NewTextDetector().DetectAll(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			for _, result := range results {
				if result.Charset == tt.charset {
					if result.Language != tt.language {
						t.Fatalf("%s language = %s, want %s", result.Charset, result.Language, tt.language)
					}
					return
				}
			}
			t.Fatalf("DetectAll() did not return %s", tt.charset)
		})
	}
}
