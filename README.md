# chardet

chardet detects character encodings for Go, using algorithms and data from
[ICU](http://icu-project.org/).

This project is a maintained fork of
[saintfish/chardet](https://github.com/saintfish/chardet).

## Usage

Pass the original text bytes in `data` (`[]byte`):

```go
result, err := chardet.NewTextDetector().DetectBest(data)
if err != nil {
	return err
}

fmt.Printf("%s (confidence %d)\n", result.Charset, result.Confidence)
```

Use `NewHtmlDetector` for HTML input. See the
[API documentation](https://pkg.go.dev/github.com/levmv/chardet) for details.

## Detection behavior

Each result has a confidence score from 1 to 100. `DetectAll` returns candidates
in descending confidence order; `DetectBest` returns the first candidate.
Confidence is a heuristic ranking score, not a probability. Scores need not add
up to 100.

Short inputs are often ambiguous: for example, ASCII text is compatible with
several encodings. A high score does not prove the original encoding, and
detection generally does not guarantee that the entire input can be decoded.

For UTF-8, confidence 100 requires a complete, valid byte slice. If you pass
only a sample of a larger input, this requirement applies only to that sample.

UTF-16 can be detected without a byte order mark (BOM), but short texts and
texts with few or no ASCII characters may go undetected.
