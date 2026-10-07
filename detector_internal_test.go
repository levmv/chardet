package chardet

import (
	"reflect"
	"testing"
)

type fixedRecognizer recognizerOutput

func (r fixedRecognizer) Match(*recognizerInput) recognizerOutput {
	return recognizerOutput(r)
}

func TestDetectorResultOrder(t *testing.T) {
	d := &Detector{recognizers: []recognizer{
		fixedRecognizer{Charset: "second", Language: "es", Confidence: 40},
		fixedRecognizer{Charset: "first", Language: "en", Confidence: 100},
		fixedRecognizer{Charset: "second", Language: "fr", Confidence: 100},
		fixedRecognizer{Charset: "first", Language: "de", Confidence: 100},
		fixedRecognizer{Charset: "rejected", Confidence: 0},
	}}
	want := []Result{
		{Charset: "first", Language: "en", Confidence: 100},
		{Charset: "second", Language: "fr", Confidence: 100},
	}
	// Repeated calls exercise recognizers completing in different orders.
	for i := 0; i < 10; i++ {
		all, err := d.DetectAll(nil)
		if err != nil || !reflect.DeepEqual(all, want) {
			t.Fatalf("DetectAll() = %v, %v; want %v", all, err, want)
		}
		best, err := d.DetectBest(nil)
		if err != nil || best == nil || *best != want[0] {
			t.Fatalf("DetectBest() = %v, %v; want %v", best, err, want[0])
		}
	}
}

func TestDetectorNoResult(t *testing.T) {
	d := &Detector{recognizers: []recognizer{fixedRecognizer{Charset: "rejected"}}}
	if all, err := d.DetectAll(nil); all != nil || err != NotDetectedError {
		t.Fatalf("DetectAll() = %v, %v; want nil, NotDetectedError", all, err)
	}
	if best, err := d.DetectBest(nil); best != nil || err != NotDetectedError {
		t.Fatalf("DetectBest() = %v, %v; want nil, NotDetectedError", best, err)
	}
}

func TestDetectorUTF8ResultOrder(t *testing.T) {
	for _, tt := range []struct {
		name       string
		input      string
		competitor int
		utf8First  bool
		want       Result
	}{
		{"maximum tie", "Привет", 100, true, Result{Charset: "UTF-8", Confidence: 100}},
		{"tie below 100", "é", 80, true, Result{Charset: "UTF-8", Confidence: 80}},
		{"stronger candidate", "é", 100, true, Result{Charset: "other", Confidence: 100}},
		{"earlier candidate", "Привет", 100, false, Result{Charset: "other", Confidence: 100}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			rs := []recognizer{
				newRecognizer_utf8(),
				fixedRecognizer{Charset: "other", Confidence: tt.competitor},
			}
			if !tt.utf8First {
				rs[0], rs[1] = rs[1], rs[0]
			}
			d := &Detector{recognizers: rs}
			best, err := d.DetectBest([]byte(tt.input))
			if err != nil || best == nil || *best != tt.want {
				t.Fatalf("DetectBest() = %v, %v; want %+v", best, err, tt.want)
			}
		})
	}
}

func FuzzDetectBestMatchesDetectAll(f *testing.F) {
	for _, seed := range []string{
		"", "Hello, world!", "Привет", "é", "\xff", "Привет\xe2\x82",
		"<p><b>Привет, 世界! 😀</b></p><br><br>",
	} {
		f.Add([]byte(seed))
	}
	f.Fuzz(func(t *testing.T, raw []byte) {
		if len(raw) > 32768 {
			t.Skip()
		}
		for _, d := range []*Detector{NewTextDetector(), NewHtmlDetector()} {
			best, bestErr := d.DetectBest(raw)
			all, allErr := d.DetectAll(raw)
			if bestErr != allErr {
				t.Fatalf("error mismatch: DetectBest %v, DetectAll %v", bestErr, allErr)
			}
			if allErr == nil && (best == nil || len(all) == 0 || *best != all[0]) {
				t.Fatalf("result mismatch: DetectBest %v, DetectAll %v", best, all)
			}
		}
	})
}
