package chardet

import (
	"bytes"
	"fmt"
	"testing"
)

var benchmarkUTF16BE, benchmarkUTF16LE recognizerOutput

func utf16BenchmarkInputs() []struct {
	name string
	unit []byte
} {
	return []struct {
		name string
		unit []byte
	}{
		{"ASCII", []byte("This is a short text about books. ")},
		{"UTF8", []byte("Café, Ελληνικά, 日本語, 😀. ")},
		{"Windows1252", []byte("Un caf\xe9 avec une cr\xe8me br\xfbl\xe9e. ")},
		{"NULSeparators", []byte("a-long-path.txt\x00")},
		{"Zeroes", []byte{0}},
		{"LatinBE", utf16Bytes("Un café avec une crème brûlée. ", false)},
		{"LatinLE", utf16Bytes("Un café avec une crème brûlée. ", true)},
		{"GreekLE", utf16Bytes("Αυτό είναι ένα μικρό κείμενο για βιβλία. ", true)},
		{"JapaneseBE", utf16Bytes("これは本の説明です。 新しい本です。 読書の時間です。 本を読みます。 ", false)},
		{"EmojiLE", utf16Bytes("Books 📚, travel 🌍, launch 🚀, smile 😀. ", true)},
		{"BOMBE", utf16Bytes("\ufeffUn café avec une crème brûlée. ", false)},
		{"BOMLE", utf16Bytes("\ufeffUn café avec une crème brûlée. ", true)},
		{"UTF32LE", []byte{'b', 0, 0, 0, 'o', 0, 0, 0, 'o', 0, 0, 0, 'k', 0, 0, 0}},
	}
}

// BenchmarkUTF16 measures both byte orders, as DetectAll always runs both.
func BenchmarkUTF16(b *testing.B) {
	be, le := newRecognizer_utf16be(), newRecognizer_utf16le()
	for _, text := range utf16BenchmarkInputs() {
		for _, size := range []int{64, 8192, 65536} {
			raw := bytes.Repeat(text.unit, (size+len(text.unit)-1)/len(text.unit))[:size]
			input := &recognizerInput{raw: raw}
			b.Run(fmt.Sprintf("%s/%d", text.name, size), func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					benchmarkUTF16BE = be.Match(input)
					benchmarkUTF16LE = le.Match(input)
				}
			})
		}
	}
}
