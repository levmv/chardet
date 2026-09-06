package chardet

import (
	"bytes"
	"encoding/binary"
	"unicode/utf8"
)

var utf8Bom = []byte{0xEF, 0xBB, 0xBF}

type recognizerUtf8 struct {
}

func newRecognizer_utf8() *recognizerUtf8 {
	return &recognizerUtf8{}
}

func (*recognizerUtf8) Match(input *recognizerInput) (output recognizerOutput) {
	output = recognizerOutput{
		Charset: "UTF-8",
	}
	hasBom := bytes.HasPrefix(input.raw, utf8Bom)
	var numValid, numInvalid int
	var truncated bool
	raw := input.raw
	// Skip the ASCII prefix in whole words, as in Go's UTF-8 validator.
	i := 0
	for len(raw)-i >= 8 && binary.LittleEndian.Uint64(raw[i:])&0x8080808080808080 == 0 {
		i += 8
	}
	// Check the byte ranges from RFC 3629, section 4, without decoding runes.
	for ; i < len(raw); i++ {
		c := raw[i]
		switch {
		case c < utf8.RuneSelf:
			continue
		case c < 0xe0:
			if c >= 0xc2 && i < len(raw)-1 && raw[i+1]&0xc0 == 0x80 {
				numValid++
				i++
				continue
			}
		case c < 0xf0:
			// Exclude overlong encodings and UTF-16 surrogates.
			if i < len(raw)-2 && raw[i+1]&0xc0 == 0x80 && raw[i+2]&0xc0 == 0x80 &&
				(c != 0xe0 || raw[i+1] >= 0xa0) && (c != 0xed || raw[i+1] < 0xa0) {
				numValid++
				i += 2
				continue
			}
		case c < 0xf5:
			// Exclude overlong encodings and values above U+10FFFF.
			if i < len(raw)-3 && raw[i+1]&0xc0 == 0x80 && raw[i+2]&0xc0 == 0x80 && raw[i+3]&0xc0 == 0x80 &&
				(c != 0xf0 || raw[i+1] >= 0x90) && (c != 0xf4 || raw[i+1] < 0x90) {
				numValid++
				i += 3
				continue
			}
		}
		if len(raw)-i < utf8.UTFMax && !utf8.FullRune(raw[i:]) {
			// A valid prefix at EOF may be a sample cut inside a rune.
			truncated = true
			break
		}
		numInvalid++
		if numInvalid > 5 {
			break
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
	} else if numValid == 0 && numInvalid == 0 && !truncated {
		// Plain ASCII
		output.Confidence = 10
	} else if numValid > numInvalid*10 {
		output.Confidence = 25
	}
	if truncated && output.Confidence > 80 {
		// Retain evidence from complete runes without claiming an intact input.
		output.Confidence = 80
	}
	return
}
