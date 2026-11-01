package scoring

import "testing"

func TestScoreReportsExactComponentsForKnownPath(t *testing.T) {
	result, err := Score(5, []int{90, 80})
	if err != nil {
		t.Fatal(err)
	}
	if result.ImpactTier != 5 || result.CompoundedConfidence != 72 || result.Score != 360 {
		t.Fatalf("unexpected score result: %#v", result)
	}
	if len(result.EdgeConfidences) != 2 || result.EdgeConfidences[0] != 90 || result.EdgeConfidences[1] != 80 {
		t.Fatalf("edge components not preserved: %#v", result.EdgeConfidences)
	}
	want := "Impact tier 5 × compounded confidence 72% (90% × 80%; floor after each edge) = 360 points"
	if result.Explanation != want {
		t.Fatalf("explanation = %q, want %q", result.Explanation, want)
	}
}

func TestScoreFloorsAfterEachEdgeAndAcceptsZeroConfidence(t *testing.T) {
	result, err := Score(3, []int{99, 99, 0})
	if err != nil {
		t.Fatal(err)
	}
	if result.CompoundedConfidence != 0 || result.Score != 0 {
		t.Fatalf("expected a zero-confidence score: %#v", result)
	}
	result, err = Score(2, []int{99, 99})
	if err != nil {
		t.Fatal(err)
	}
	if result.CompoundedConfidence != 98 || result.Score != 196 {
		t.Fatalf("expected explicit stepwise floor: %#v", result)
	}
}

func TestScoreRejectsInvalidInput(t *testing.T) {
	for _, test := range []struct {
		name        string
		impact      int
		confidences []int
	}{
		{name: "impact too small", impact: 0, confidences: []int{100}},
		{name: "impact too large", impact: 6, confidences: []int{100}},
		{name: "no edges", impact: 1},
		{name: "confidence too small", impact: 1, confidences: []int{-1}},
		{name: "confidence too large", impact: 1, confidences: []int{101}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := Score(test.impact, test.confidences); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
