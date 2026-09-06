# chardet

chardet is library to automatically detect
[charset](http://en.wikipedia.org/wiki/Character_encoding) of texts for [Go
programming language](http://golang.org/). It's based on the algorithm and data
in [ICU](http://icu-project.org/)'s implementation.

## Documentation and Usage

See [pkgdoc](https://pkg.go.dev/github.com/levmv/chardet).

This project is a maintained fork of
[saintfish/chardet](https://github.com/saintfish/chardet). The original Git
history and license notices are preserved.

## Detection behavior

Each result has a confidence score from 1 to 100. `DetectAll` returns candidates
in descending confidence order; `DetectBest` returns the first candidate.
Confidence is a heuristic ranking score, not a probability. Scores need not add
up to 100, and detection does not generally guarantee that every byte can be
decoded.

The same bytes may be valid in several encodings. For example, ASCII text can
be decoded identically as UTF-8, Windows-1251, or Windows-1252. A high score
does not prove which encoding was originally used.

### UTF-8

For UTF-8, confidence 100 requires a complete, valid byte slice. If you pass
only a sample of a larger input, this requirement applies only to that sample.

An incomplete final character is tolerated when preceded by complete non-ASCII
characters or a UTF-8 BOM, with confidence capped at 80. This allows detection
of samples cut inside a character. Malformed bytes reduce confidence; partially
damaged text may still be returned as a candidate.

These examples show the confidence of the UTF-8 candidate, which may not be
the highest-ranked result:

| Input (Go expression) | UTF-8 confidence | Explanation |
| --- | --- | --- |
| `[]byte("Hello, world!")` | 10 | Valid ASCII, compatible with many encodings. |
| `[]byte("é")` | 80 | One complete character; 80 does not imply truncation. |
| `[]byte("こんにちは")` | 100 | Several complete non-ASCII characters. |
| `[]byte("こんにちは\xe2\x82")` | 80 | Complete text followed by an incomplete `€` character. |
