package chardet

import (
	"bytes"
)

var (
	utf16beBom = []byte{0xFE, 0xFF}
	utf16leBom = []byte{0xFF, 0xFE}
	utf32beBom = []byte{0x00, 0x00, 0xFE, 0xFF}
	utf32leBom = []byte{0xFF, 0xFE, 0x00, 0x00}
)

type recognizerUtf16be struct {
}

func newRecognizer_utf16be() *recognizerUtf16be {
	return &recognizerUtf16be{}
}

func (*recognizerUtf16be) Match(input *recognizerInput) (output recognizerOutput) {
	output = recognizerOutput{
		Charset: "UTF-16BE",
	}
	if bytes.HasPrefix(input.raw, utf16beBom) {
		output.Confidence = 100
	} else {
		output.Confidence = utf16Confidence(input.raw, false)
	}
	return
}

type recognizerUtf16le struct {
}

func newRecognizer_utf16le() *recognizerUtf16le {
	return &recognizerUtf16le{}
}

func (*recognizerUtf16le) Match(input *recognizerInput) (output recognizerOutput) {
	output = recognizerOutput{
		Charset: "UTF-16LE",
	}
	if bytes.HasPrefix(input.raw, utf16leBom) {
		if !bytes.HasPrefix(input.raw, utf32leBom) {
			output.Confidence = 100
		}
	} else {
		output.Confidence = utf16Confidence(input.raw, true)
	}
	return
}

// utf16Confidence looks for repeated ASCII characters in a bounded UTF-16
// sample. Checking validity alone would accept many unrelated byte sequences.
// Text without ASCII characters needs a different, statistical recognizer.
func utf16Confidence(raw []byte, littleEndian bool) int {
	const sampleSize = 1024
	if len(raw) > sampleSize {
		raw = raw[:sampleSize]
	}
	if len(raw) < 8 || bytes.IndexByte(raw, 0) < 0 {
		return 0
	}

	var ascii, opposite int
	asciiBytes := true
	highSurrogate := false
	i := 0
	for ; i < len(raw)-1; i += 2 {
		hi, lo := raw[i], raw[i+1]
		if littleEndian {
			hi, lo = lo, hi
		}
		if asciiBytes && (!utf16ASCIIByte(hi) || !utf16ASCIIByte(lo)) {
			asciiBytes = false
		}
		c := uint16(hi)<<8 | uint16(lo)
		if highSurrogate {
			if c < 0xdc00 || c > 0xdfff {
				return 0
			}
			highSurrogate = false
			continue
		}
		switch {
		case c >= 0xd800 && c <= 0xdbff:
			highSurrogate = true
		case c >= 0xdc00 && c <= 0xdfff:
			return 0
		case c < 0x20 && c != '\t' && c != '\n' && c != '\r',
			c >= 0x7f && c <= 0x9f, c >= 0xfffe:
			// NUL is also common in UTF-32 and binary data. These controls
			// and noncharacters are negative text evidence, not invalid UTF-16.
			return 0
		case c < 0x7f:
			ascii++
		}
		if lo == 0 && utf16ASCIIByte(hi) {
			opposite++
		}
	}

	units := len(raw) / 2
	// Require at least four ASCII characters, at least 5% of the sample,
	// and at least twice as much evidence as in the other byte order.
	if ascii < 4 || ascii*20 < units || ascii < opposite*2 {
		return 0
	}
	// Sparse NUL separators in otherwise ASCII bytes (e.g. file lists)
	// should not turn ordinary byte pairs into supposed UTF-16 text.
	if asciiBytes && ascii*4 < units {
		return 0
	}
	if i < len(raw) && !littleEndian {
		// The high byte alone can already prove that a partial code unit
		// is an unpaired low surrogate, or cannot complete a pending pair.
		lowSurrogate := raw[i] >= 0xdc && raw[i] <= 0xdf
		if highSurrogate != lowSurrogate {
			return 0
		}
	}
	if i < len(raw) || highSurrogate {
		return 80
	}
	// Leave stronger UTF-8/UTF-32 matches ahead of this sampled heuristic.
	// Some valid UTF-32 sequences also look like UTF-16 with whitespace.
	return 99
}

// utf16ASCIIByte includes NUL because it is the high byte of UTF-16 ASCII.
func utf16ASCIIByte(b byte) bool {
	return b == 0 || b >= 0x20 && b <= 0x7e || b == '\t' || b == '\n' || b == '\r'
}

type recognizerUtf32 struct {
	name       string
	bom        []byte
	decodeChar func(input []byte) uint32
}

func decodeUtf32be(input []byte) uint32 {
	return uint32(input[0])<<24 | uint32(input[1])<<16 | uint32(input[2])<<8 | uint32(input[3])
}

func decodeUtf32le(input []byte) uint32 {
	return uint32(input[3])<<24 | uint32(input[2])<<16 | uint32(input[1])<<8 | uint32(input[0])
}

func newRecognizer_utf32be() *recognizerUtf32 {
	return &recognizerUtf32{
		"UTF-32BE",
		utf32beBom,
		decodeUtf32be,
	}
}

func newRecognizer_utf32le() *recognizerUtf32 {
	return &recognizerUtf32{
		"UTF-32LE",
		utf32leBom,
		decodeUtf32le,
	}
}

func (r *recognizerUtf32) Match(input *recognizerInput) (output recognizerOutput) {
	output = recognizerOutput{
		Charset: r.name,
	}
	hasBom := bytes.HasPrefix(input.raw, r.bom)
	var numValid, numInvalid uint32
	for b := input.raw; len(b) >= 4; b = b[4:] {
		if c := r.decodeChar(b); c > 0x10FFFF || (c >= 0xD800 && c <= 0xDFFF) {
			numInvalid++
		} else {
			numValid++
		}
	}
	if hasBom && numInvalid == 0 {
		output.Confidence = 100
	} else if hasBom && numValid > numInvalid*10 {
		output.Confidence = 80
	} else if numValid > 3 && numInvalid == 0 {
		output.Confidence = 100
	} else if numValid > 0 && numInvalid == 0 {
		output.Confidence = 80
	} else if numValid > numInvalid*10 {
		output.Confidence = 25
	}
	return
}
