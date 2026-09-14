package chardet

import (
	"bytes"
	"fmt"
	"testing"
)

func BenchmarkDetector(b *testing.B) {
	for _, mode := range []struct {
		name string
		d    *Detector
	}{
		{"Text", NewTextDetector()},
		{"HTML", NewHtmlDetector()},
	} {
		for _, input := range []struct {
			name string
			unit []byte
		}{
			{"ASCII", []byte("This is a short text about books. ")},
			{"UTF8", []byte("Café, Ελληνικά, 日本語, 😀. ")},
			{"UTF32", []byte{0, 0, 0x4e, 0x16}},
			{"Legacy", []byte("Un caf\xe9 avec une cr\xe8me br\xfbl\xe9e. ")},
			{"Markup", []byte("<p><b>Un caf\xe9 avec une cr\xe8me br\xfbl\xe9e.</b></p> ")},
		} {
			for _, size := range []int{64, 8192, 65536} {
				raw := bytes.Repeat(input.unit, (size+len(input.unit)-1)/len(input.unit))[:size]
				name := fmt.Sprintf("%s/%s/%d", mode.name, input.name, size)
				b.Run(name+"/All", func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						if _, err := mode.d.DetectAll(raw); err != nil {
							b.Fatal(err)
						}
					}
				})
				b.Run(name+"/Best", func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						if _, err := mode.d.DetectBest(raw); err != nil {
							b.Fatal(err)
						}
					}
				})
			}
		}
	}
}
