package analytics

import (
	"testing"

	"github.com/alexandrmotologa/distillforge/internal/store"
)

func TestDatasetAnalytics(t *testing.T) {
	samples := []store.Sample{
		{
			Prompt:     "How do I reset my password?",
			Response:   `{"category": "Authentication", "confidence": 0.98}`,
			Strategy:   "in_depth",
			JudgeScore: 0.95,
			IsAccepted: true,
		},
		{
			Prompt:     "My internet has been disconnected for three hours after high winds.",
			Response:   `{"category": "Network Outage", "confidence": 0.99}`,
			Strategy:   "adversarial",
			JudgeScore: 0.92,
			IsAccepted: true,
		},
		{
			Prompt:     "Short query",
			Response:   `{"category": "Billing", "confidence": 0.65}`,
			Strategy:   "in_breadth",
			JudgeScore: 0.65,
			IsAccepted: false,
		},
	}

	res := ComputeDatasetAnalytics(samples)

	if res.TotalSamples != 3 {
		t.Fatalf("expected 3 total samples, got %d", res.TotalSamples)
	}
	if res.AcceptedSamples != 2 {
		t.Fatalf("expected 2 accepted samples, got %d", res.AcceptedSamples)
	}
	if res.RejectedSamples != 1 {
		t.Fatalf("expected 1 rejected sample, got %d", res.RejectedSamples)
	}
	if res.TotalVocabularyTokens <= 0 || res.UniqueVocabularyTokens <= 0 {
		t.Fatalf("expected positive vocabulary counts, got %d / %d", res.TotalVocabularyTokens, res.UniqueVocabularyTokens)
	}
	if res.LexicalDiversityTTR <= 0 || res.LexicalDiversityTTR > 1.0 {
		t.Fatalf("expected TTR between 0 and 1, got %f", res.LexicalDiversityTTR)
	}
	if len(res.PromptLengthHistogram) == 0 {
		t.Fatal("expected prompt length histogram buckets")
	}
	if len(res.StrategyBreakdown) != 3 {
		t.Fatalf("expected 3 strategies, got %d", len(res.StrategyBreakdown))
	}
}
