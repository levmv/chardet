package chardet

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestUTF8Confidence(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"empty", "", 10},
		{"ASCII", "Hello, world!", 10},
		{"BOM only", "\ufeff", 100},
		{"BOM and ASCII", "\ufeffHello", 100},
		{"one multibyte rune", "é", 80},
		{"three multibyte runes", "é€😀", 80},
		{"four multibyte runes", "é€😀界", 100},
		{"Cyrillic", "Привет", 100},
		{"replacement character", "\ufffd", 80},
		{"scalar boundaries", "\u0080\u07ff\u0800\ud7ff\ue000\uffff\U00010000\U0010ffff", 100},
		{"stray continuation", "\x80", 0},
		{"two byte overlong", strings.Repeat("\xc0\x80", 4), 0},
		{"three byte overlong", strings.Repeat("\xe0\x9f\xbf", 4), 0},
		{"four byte overlong", strings.Repeat("\xf0\x8f\xbf\xbf", 4), 0},
		{"surrogate", strings.Repeat("\xed\xa0\x80", 4), 0},
		{"above maximum", strings.Repeat("\xf4\x90\x80\x80", 4), 0},
		{"invalid lead", strings.Repeat("\xf5\x80\x80\x80", 4), 0},
		{"invalid byte after BOM", "\ufeff\xff", 0},
		{"error ratio boundary", strings.Repeat("é", 10) + "\xff", 0},
		{"damaged text", strings.Repeat("é", 11) + "\xff", 25},
		{"damaged text with BOM", "\ufeff" + strings.Repeat("é", 10) + "\xff", 80},
		{"resume after invalid lead", "\xff" + strings.Repeat("é", 11), 25},
		{"resume after invalid continuation", "\xe2" + strings.Repeat("é", 11), 25},
		{"partial surrogate is already invalid", "Καλημέρα\xed\xa0", 0},
		{"partial overlong is already invalid", "こんにちは\xe0\x9f", 0},
		{"partial above maximum is already invalid", "😀🚀📚🌍\xf4\x90", 0},
		{"truncation does not hide earlier errors", "déjà vu à Noël\xff\xe2\x82", 0},
		{"damaged text and truncation", strings.Repeat("é", 11) + "\xff\xe2\x82", 25},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newRecognizer_utf8().Match(&recognizerInput{raw: []byte(tt.input)}).Confidence
			if got != tt.want {
				t.Fatalf("UTF-8 confidence for %x = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestUTF8TruncatedFinalRune(t *testing.T) {
	for _, last := range []string{"é", "€", "😀", "\U0010ffff"} {
		for n := 1; n < len(last); n++ {
			for _, prefix := range []struct {
				name string
				text string
				want int
			}{
				{"no prefix", "", 0},
				{"ASCII", "Hello", 0},
				{"one multibyte rune", "é", 80},
				{"two byte runes", "Καλημέρα", 80},
				{"three byte runes", "こんにちは", 80},
				{"four byte runes", "😀🚀📚🌍", 80},
				{"BOM", "\ufeff", 80},
			} {
				t.Run(fmt.Sprintf("%x/%d/%s", last, n, prefix.name), func(t *testing.T) {
					raw := []byte(prefix.text + last[:n])
					got := newRecognizer_utf8().Match(&recognizerInput{raw: raw}).Confidence
					if got != prefix.want {
						t.Fatalf("UTF-8 confidence for %x = %d, want %d", raw, got, prefix.want)
					}
				})
			}
		}
	}
}

func TestUTF8ASCIIPrefix(t *testing.T) {
	for _, n := range []int{0, 1, 7, 8, 9, 15, 16, 17, 31, 32, 33, 8192} {
		for _, tail := range []struct {
			name string
			text string
			want int
		}{
			{"ASCII only", "", 10},
			{"one multibyte rune", "é", 80},
			{"two byte runes", "Καλημέρα", 100},
			{"three byte runes", "こんにちは", 100},
			{"four byte runes", "😀🚀📚🌍", 100},
			{"invalid byte", "\xff", 0},
			{"incomplete rune only", "\xe2\x82", 0},
			{"incomplete final rune", "café 😀 日本語\xe2\x82", 80},
		} {
			t.Run(fmt.Sprintf("%d/%s", n, tail.name), func(t *testing.T) {
				raw := []byte(strings.Repeat("a", n) + tail.text)
				got := newRecognizer_utf8().Match(&recognizerInput{raw: raw}).Confidence
				if got != tail.want {
					t.Fatalf("UTF-8 confidence after %d ASCII bytes and %x = %d, want %d", n, tail.text, got, tail.want)
				}
			})
		}
	}
}

func FuzzUTF8Confidence(f *testing.F) {
	for _, seed := range []string{
		"", "Hello", "Привет", "é€😀", "\ufeff", "\ufffd", "\U0010ffff",
		"Καλημέρα", "こんにちは", "😀🚀📚🌍",
		"\xc0\x80", "\xed\xa0\x80", "\xf4\x90\x80\x80", "\xe2\x82",
		"こんにちは\xe2\x82", "\xff" + strings.Repeat("é", 11),
		"aaaaaaaa\xff", "aaaaaaaaΚαλημέρα\xe2\x82",
	} {
		f.Add([]byte(seed))
	}
	f.Fuzz(func(t *testing.T, raw []byte) {
		got := newRecognizer_utf8().Match(&recognizerInput{raw: raw}).Confidence
		if !utf8.Valid(raw) {
			if got == 100 {
				t.Fatalf("invalid UTF-8 %x has confidence 100", raw)
			}
			return
		}

		// Preserve the existing confidence levels for all complete, valid input.
		nonASCII := 0
		for _, r := range string(raw) {
			if r >= utf8.RuneSelf {
				nonASCII++
			}
		}
		want := 10
		if bytes.HasPrefix(raw, []byte("\ufeff")) || nonASCII > 3 {
			want = 100
		} else if nonASCII > 0 {
			want = 80
		}
		if got != want {
			t.Fatalf("valid UTF-8 %x has confidence %d, want %d", raw, got, want)
		}
	})
}
