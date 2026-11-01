package scoring

import (
	"errors"
	"fmt"
	"strings"
)

type Result struct {
	ImpactTier           int   `json:"impact_tier"`
	EdgeConfidences      []int `json:"edge_confidences"`
	CompoundedConfidence int   `json:"compounded_confidence"`
	Score                int   `json:"score"`
	Explanation          string `json:"explanation"`
}

func Score(impact int, edgeConfidences []int) (Result, error) {
	if impact < 1 || impact > 5 {
		return Result{}, errors.New("impact tier must be from 1 to 5")
	}
	if len(edgeConfidences) == 0 {
		return Result{}, errors.New("a scored path must contain at least one edge confidence")
	}
	confidence := 100
	parts := make([]string, len(edgeConfidences))
	for index, value := range edgeConfidences {
		if value < 0 || value > 100 {
			return Result{}, fmt.Errorf("edge confidence %d must be from 0 to 100", index+1)
		}
		confidence = confidence * value / 100
		parts[index] = fmt.Sprintf("%d%%", value)
	}
	score := impact * confidence
	explanation := fmt.Sprintf("Impact tier %d × compounded confidence %d%% (%s; floor after each edge) = %d points", impact, confidence, strings.Join(parts, " × "), score)
	return Result{ImpactTier: impact, EdgeConfidences: append([]int(nil), edgeConfidences...), CompoundedConfidence: confidence, Score: score, Explanation: explanation}, nil
}
