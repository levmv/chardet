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
