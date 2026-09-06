package chardet_test

import (
	"encoding/hex"
	"testing"

	"github.com/levmv/chardet"
)

func TestWindows1251Language(t *testing.T) {
	// "Это небольшой русский текст о книгах и чтении." encoded as Windows-1251.
	input, err := hex.DecodeString("ddf2ee20ede5e1eeebfcf8eee920f0f3f1f1eae8e920f2e5eaf1f220ee20eaede8e3e0f520e820f7f2e5ede8e82e")
	if err != nil {
		t.Fatal(err)
	}
	result, err := chardet.NewTextDetector().DetectBest(input)
	if err != nil {
		t.Fatal(err)
	}
	if result.Charset != "windows-1251" || result.Language != "ru" {
		t.Fatalf("DetectBest() = %s/%s, want windows-1251/ru", result.Charset, result.Language)
	}
}

func TestISO88592Detection(t *testing.T) {
	tests := []struct {
		name     string
		language string
		hexInput string
	}{
		{
			name:     "Czech",
			language: "cs",
			hexInput: "50f8ed6c69b920be6c75bb6f75e86bfd206bf9f220fa70ec6c20efe162656c736be920f364792e20546f746f206a65206b72e1746bfd20e865736bfd2074657874206f206b6e6968e16368206120e874656eed2e",
		},
		{
			name:     "Hungarian",
			language: "hu",
			hexInput: "c17276ed7a74fb72f52074fc6bf67266fa72f367e9702e20457a206567792072f6766964206d616779617220737af6766567206bf66e7976656b72f56c20e973206f6c766173e17372f36c2e",
		},
		{
			name:     "Polish",
			language: "pl",
			hexInput: "5a61bff3b3e62067eab66cb1206a61bcf12e20546f206a657374206b72f3746b6920706f6c736b692074656b7374206f206b7369b1bf6b616368206f72617a20637a7974616e69752e",
		},
		{
			name:     "Romanian",
			language: "ro",
			hexInput: "ce6e63e320756e207465787420726f6de26e657363206465737072652063e372fe6920ba6920646573707265206369746972652070656e74727520616365617374e320ee6e636572636172652e",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, err := hex.DecodeString(tt.hexInput)
			if err != nil {
				t.Fatal(err)
			}
			result, err := chardet.NewTextDetector().DetectBest(input)
			if err != nil {
				t.Fatal(err)
			}
			if result.Charset != "ISO-8859-2" || result.Language != tt.language {
				t.Fatalf("DetectBest() = %s/%s, want ISO-8859-2/%s", result.Charset, result.Language, tt.language)
			}
		})
	}
}
