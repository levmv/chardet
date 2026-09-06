package chardet

import (
	"bytes"
	"fmt"
	"testing"
)

var benchmarkUTF8Result recognizerOutput

func BenchmarkUTF8(b *testing.B) {
	r := newRecognizer_utf8()
	for _, text := range []struct {
		name string
		unit string
		tail string
	}{
		{"ASCII", "This is a short text about books. ", ""},
		{"Latin", "Un café avec une crème brûlée. ", ""},
		{"Cyrillic", "Это короткий текст о книгах. ", ""},
		{"CJK", "これは古い本のタイトルです。", ""},
		{"Emoji", "😀🚀📚🌍", ""},
		{"Windows1251", "\xdd\xf2\xee \xea\xee\xf0\xee\xf2\xea\xe8\xe9 \xf2\xe5\xea\xf1\xf2 \xee \xea\xed\xe8\xe3\xe0\xf5. ", ""},
		{"MixedTruncated", "Café, Ελληνικά, 日本語, 😀. ", "\xf0\x9f\x98"},
		{"MixedDamaged", "Café, Ελληνικά, 日本語, 😀. ", "\xff"},
	} {
		for _, size := range []int{64, 8192, 65536} {
			bodySize := size - len(text.tail)
			raw := bytes.Repeat([]byte(text.unit), bodySize/len(text.unit))
			for len(raw) < bodySize {
				raw = append(raw, ' ')
			}
			raw = append(raw, text.tail...)
			input := &recognizerInput{raw: raw}
			b.Run(fmt.Sprintf("%s/%d", text.name, size), func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(len(raw)))
				for i := 0; i < b.N; i++ {
					benchmarkUTF8Result = r.Match(input)
				}
			})
		}
	}
}
