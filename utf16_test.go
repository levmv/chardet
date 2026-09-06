package chardet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
	"testing"
	"unicode/utf16"
	"unicode/utf8"
)

func utf16Bytes(text string, littleEndian bool) []byte {
	units := utf16.Encode([]rune(text))
	raw := make([]byte, len(units)*2)
	for i, c := range units {
		if littleEndian {
			binary.LittleEndian.PutUint16(raw[i*2:], c)
		} else {
			binary.BigEndian.PutUint16(raw[i*2:], c)
		}
	}
	return raw
}

func TestUTF16WithoutBOM(t *testing.T) {
	for _, text := range []string{
		"Hello, world!",
		"Un café avec une crème brûlée.",
		"Αυτό είναι ένα μικρό κείμενο για βιβλία.",
		"هذا نص قصير عن الكتب الجميلة.",
		"これは本の説明です。 新しい本です。 読書の時間です。 本を読みます。 ",
		"这是一本书的介绍。 书架上有新书。 今天一起读书。 阅读时间到了。 ",
		"Это небольшой текст о книгах и чтении.",
		"Books: 📚, travel: 🌍, launch: 🚀, smile: 😀.",
		"\tName\tValue\r\nbook\t42\r\n",
		"Title\r\n" + strings.Repeat("─", 60) + "\r\n" + strings.Repeat("Some more text. ", 20),
	} {
		for _, little := range []bool{false, true} {
			charset := "UTF-16BE"
			if little {
				charset = "UTF-16LE"
			}
			t.Run(charset+"/"+text, func(t *testing.T) {
				raw := utf16Bytes(text, little)
				for _, d := range []*Detector{NewTextDetector(), NewHtmlDetector()} {
					result, err := d.DetectBest(raw)
					if err != nil {
						t.Fatal(err)
					}
					if result.Charset != charset || result.Confidence != 99 {
						t.Fatalf("DetectBest(%x) = %+v, want %s at 99", raw, result, charset)
					}
				}
			})
		}
	}
}

func TestUTF16WithoutBOMNegativeEvidence(t *testing.T) {
	for _, raw := range [][]byte{
		nil, {0}, {0, 'a'}, {0, 'a', 0},
		[]byte("A regular ASCII text with no zero bytes."),
		[]byte("Café, Ελληνικά, 日本語, 😀."),
		[]byte("Un caf\xe9 avec une cr\xe8me br\xfbl\xe9e."),
		[]byte("\x82\xb1\x82\xf1\x82\xc9\x82\xbf\x82\xcd"), // Shift_JIS
		[]byte("\xc4\xe3\xba\xc3\xca\xc0\xbd\xe7"),         // GB18030
		bytes.Repeat([]byte{0}, 1024),
		bytes.Repeat([]byte{0xff}, 1024),
		bytes.Repeat([]byte("a-long-path.txt\x00"), 20),
		bytes.Repeat([]byte("another-long-file.txt\x00"), 20),
	} {
		for _, r := range []recognizer{newRecognizer_utf16be(), newRecognizer_utf16le()} {
			if got := r.Match(&recognizerInput{raw: raw}); got.Confidence != 0 {
				t.Errorf("%x: unexpected UTF-16 candidate %+v", raw, got)
			}
		}
	}
	for _, little := range []bool{false, true} {
		for _, text := range []string{
			"abc", "こんにちは", "世界您好", "😀😀😀😀",
			"Hello\x00world", "Hello\x01world", "Hello\u0085world", "Hello\ufffeworld",
		} {
			raw := utf16Bytes(text, little)
			if got := utf16Confidence(raw, little); got != 0 {
				t.Errorf("%q: confidence = %d, want 0", text, got)
			}
		}
		// Neither byte order should mistake supplementary characters for Latin.
		if got := utf16Confidence(utf16Bytes("😀😀😀😀", little), !little); got != 0 {
			t.Errorf("byte-swapped emoji: confidence = %d, want 0", got)
		}
	}
}

func TestUTF16SurrogatesAndTruncation(t *testing.T) {
	for _, little := range []bool{false, true} {
		prefix := utf16Bytes("Hello, 世界! ", little)
		for _, last := range []string{"é", "界", "😀", "\U0010ffff"} {
			tail := utf16Bytes(last, little)
			for n := 1; n <= len(tail); n++ {
				raw := append(append([]byte(nil), prefix...), tail[:n]...)
				want := 80
				if n == len(tail) {
					want = 99
				}
				if got := utf16Confidence(raw, little); got != want {
					t.Errorf("little=%t, tail=%x: confidence = %d, want %d", little, tail[:n], got, want)
				}
			}
		}
		for _, tail := range [][]uint16{
			{0xdc00}, {0xdfff}, {0xd800, 'a'}, {0xd800, 0xd800},
			{0xdbff, 0xe000}, {0xd800, 0xdc00, 0xdc00},
		} {
			raw := append([]byte(nil), prefix...)
			for _, c := range tail {
				if little {
					raw = append(raw, byte(c), byte(c>>8))
				} else {
					raw = append(raw, byte(c>>8), byte(c))
				}
			}
			if got := utf16Confidence(raw, little); got != 0 {
				t.Errorf("little=%t, malformed tail=%x: confidence = %d, want 0", little, tail, got)
			}
		}
	}
	for _, tail := range [][]byte{{0xdc}, {0xdf}, {0xd8, 0, 'a'}, {0xdb, 0xff, 0xd8}} {
		raw := append(utf16Bytes("Hello, 世界! ", false), tail...)
		if got := utf16Confidence(raw, false); got != 0 {
			t.Errorf("malformed partial BE tail=%x: confidence = %d, want 0", tail, got)
		}
	}
}

func TestUTF16BOMAndUTF32(t *testing.T) {
	for _, little := range []bool{false, true} {
		var r recognizer = newRecognizer_utf16be()
		if little {
			r = newRecognizer_utf16le()
		}
		for _, text := range []string{"\ufeff", "\ufeffこんにちは", "\ufeff😀", "\ufeff\x01"} {
			raw := utf16Bytes(text, little)
			if got := r.Match(&recognizerInput{raw: raw}).Confidence; got != 100 {
				t.Errorf("BOM little=%t %q: confidence = %d, want 100", little, text, got)
			}
		}
		for _, bom := range []string{"", "\ufeff"} {
			for _, text := range []string{
				"Hello, world!", "Καλημέρα κόσμε", "こんにちは", "😀🚀📚🌍",
				// These scalars also look like UTF-16 with tabs, LF or CR.
				"\U00090041\U00090042\U00090043\U00090044",
				"\U000a0041\U000a0042\U000a0043\U000a0044",
				"\U000d0041\U000d0042\U000d0043\U000d0044",
			} {
				var raw []byte
				for _, c := range bom + text {
					if little {
						raw = append(raw, byte(c), byte(c>>8), byte(c>>16), byte(c>>24))
					} else {
						raw = append(raw, byte(c>>24), byte(c>>16), byte(c>>8), byte(c))
					}
				}
				result, err := NewTextDetector().DetectBest(raw)
				want := "UTF-32BE"
				if little {
					want = "UTF-32LE"
				}
				if err != nil || result.Charset != want || result.Confidence != 100 {
					t.Errorf("UTF-32 little=%t %q: DetectBest = %+v, %v; want %s at 100", little, text, result, err, want)
				}
			}
		}
	}
}

func TestUTF16SampleBoundary(t *testing.T) {
	for _, little := range []bool{false, true} {
		for _, size := range []int{8, 9, 10, 1022, 1023, 1024, 1025, 1026, 8192} {
			t.Run(fmt.Sprintf("little=%t/size=%d", little, size), func(t *testing.T) {
				raw := utf16Bytes(strings.Repeat("a", (size+1)/2), little)[:size]
				want := 99
				if size < 1024 && size%2 != 0 {
					want = 80
				}
				if got := utf16Confidence(raw, little); got != want {
					t.Fatalf("confidence = %d, want %d", got, want)
				}
			})
		}
		// The confidence describes a bounded sample, not validity of the full input.
		raw := append(utf16Bytes(strings.Repeat("a", 512), little), 0, 0)
		if got := utf16Confidence(raw, little); got != 99 {
			t.Errorf("data after sample boundary: confidence = %d, want 99", got)
		}
	}
}

func FuzzUTF16Confidence(f *testing.F) {
	for _, text := range []string{"", "Hello", "Café Ελληνικά 日本語 😀", "\x00", "\ufeff", "\U0010ffff"} {
		f.Add(utf16Bytes(text, false))
		f.Add(utf16Bytes(text, true))
	}
	f.Fuzz(func(t *testing.T, raw []byte) {
		for _, little := range []bool{false, true} {
			got := utf16Confidence(raw, little)
			if got != 0 && got != 80 && got != 99 {
				t.Fatalf("unexpected confidence %d", got)
			}
			if got == 0 {
				continue
			}
			sample := raw
			if len(sample) > 1024 {
				sample = sample[:1024]
			}
			var order binary.ByteOrder = binary.BigEndian
			if little {
				order = binary.LittleEndian
			}
			var pending rune
			for len(sample) >= 2 {
				c := rune(order.Uint16(sample))
				sample = sample[2:]
				if pending != 0 {
					if utf16.DecodeRune(pending, c) == utf8.RuneError {
						t.Fatalf("invalid surrogate pair accepted in %x", raw)
					}
					pending = 0
				} else if utf16.IsSurrogate(c) {
					if c >= 0xdc00 {
						t.Fatalf("unpaired low surrogate accepted in %x", raw)
					}
					pending = c
				}
			}
			if got == 99 && (len(sample) != 0 || pending != 0) {
				t.Fatalf("incomplete UTF-16 sample has confidence 99: %x", raw)
			}
			if len(sample) != 0 && !little {
				canBeLow := sample[0] >= 0xdc && sample[0] <= 0xdf
				if canBeLow != (pending != 0) {
					t.Fatalf("invalid partial UTF-16BE unit accepted in %x", raw)
				}
			}
		}
	})
}
