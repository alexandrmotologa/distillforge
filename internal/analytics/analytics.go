package analytics

import (
	"encoding/json"
	"math"
	"sort"
	"strings"
	"unicode"

	"github.com/alexandrmotologa/distillforge/internal/store"
)

// DistributionBucket represents a range and its item count.
type DistributionBucket struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

// DatasetAnalytics encapsulates diversity, token metrics, and balance scores.
type DatasetAnalytics struct {
	TotalSamples           int                  `json:"total_samples"`
	AcceptedSamples        int                  `json:"accepted_samples"`
	RejectedSamples        int                  `json:"rejected_samples"`
	AcceptanceRate         float64              `json:"acceptance_rate"`
	AvgPromptTokens        float64              `json:"avg_prompt_tokens"`
	AvgCompletionTokens    float64              `json:"avg_completion_tokens"`
	PromptLengthHistogram  []DistributionBucket `json:"prompt_length_histogram"`
	ResponseLengthHistogram []DistributionBucket `json:"response_length_histogram"`
	StrategyBreakdown      []DistributionBucket `json:"strategy_breakdown"`
	ScoreHistogram         []DistributionBucket `json:"score_histogram"`
	TotalVocabularyTokens  int                  `json:"total_vocabulary_tokens"`
	UniqueVocabularyTokens int                  `json:"unique_vocabulary_tokens"`
	LexicalDiversityTTR    float64              `json:"lexical_diversity_ttr"` // Type-Token Ratio (0.0 - 1.0)
	BalanceIndex           float64              `json:"balance_index"`         // Simpson Diversity Index
	HealthStatus           string               `json:"health_status"`        // optimal, repetitive, biased
}

// ComputeDatasetAnalytics evaluates the statistical health of samples.
func ComputeDatasetAnalytics(samples []store.Sample) *DatasetAnalytics {
	total := len(samples)
	if total == 0 {
		return &DatasetAnalytics{HealthStatus: "empty"}
	}

	var acceptedCount int
	var totalPromptToks int
	var totalRespToks int

	promptBuckets := map[string]int{"<50": 0, "50-100": 0, "100-250": 0, "250-500": 0, ">500": 0}
	respBuckets := map[string]int{"<25": 0, "25-50": 0, "50-100": 0, "100-200": 0, ">200": 0}
	scoreBuckets := map[string]int{"90-100%": 0, "80-89%": 0, "70-79%": 0, "<70%": 0}
	strategyCounts := make(map[string]int)
	categoryCounts := make(map[string]int)

	wordFrequency := make(map[string]int)
	totalWords := 0

	for _, s := range samples {
		if s.IsAccepted {
			acceptedCount++
		}

		// Token estimation
		pTokens := estimateTokens(s.Prompt)
		rTokens := estimateTokens(s.Response)
		totalPromptToks += pTokens
		totalRespToks += rTokens

		// Prompt bucket
		switch {
		case pTokens < 50:
			promptBuckets["<50"]++
		case pTokens <= 100:
			promptBuckets["50-100"]++
		case pTokens <= 250:
			promptBuckets["100-250"]++
		case pTokens <= 500:
			promptBuckets["250-500"]++
		default:
			promptBuckets[">500"]++
		}

		// Response bucket
		switch {
		case rTokens < 25:
			respBuckets["<25"]++
		case rTokens <= 50:
			respBuckets["25-50"]++
		case rTokens <= 100:
			respBuckets["50-100"]++
		case rTokens <= 200:
			respBuckets["100-200"]++
		default:
			respBuckets[">200"]++
		}

		// Score bucket
		scorePct := s.JudgeScore * 100
		switch {
		case scorePct >= 90:
			scoreBuckets["90-100%"]++
		case scorePct >= 80:
			scoreBuckets["80-89%"]++
		case scorePct >= 70:
			scoreBuckets["70-79%"]++
		default:
			scoreBuckets["<70%"]++
		}

		// Strategy
		strat := s.Strategy
		if strat == "" {
			strat = "seed"
		}
		strategyCounts[strat]++

		// Category extraction from JSON response if present
		var parsed map[string]any
		if err := json.Unmarshal([]byte(s.Response), &parsed); err == nil {
			if cat, ok := parsed["category"].(string); ok {
				categoryCounts[cat]++
			}
		}

		// Vocabulary tokens
		words := tokenizeWords(s.Prompt + " " + s.Response)
		for _, w := range words {
			totalWords++
			wordFrequency[w]++
		}
	}

	uniqueWords := len(wordFrequency)
	var ttr float64
	if totalWords > 0 {
		ttr = float64(uniqueWords) / float64(totalWords)
	}

	// Calculate Gini-Simpson diversity index for categories
	var balanceIdx float64 = 1.0
	if len(categoryCounts) > 1 {
		var sumSq float64
		for _, cnt := range categoryCounts {
			p := float64(cnt) / float64(total)
			sumSq += p * p
		}
		balanceIdx = 1.0 - sumSq
	}

	health := "optimal"
	if ttr < 0.15 {
		health = "repetitive"
	} else if len(categoryCounts) > 1 && balanceIdx < 0.4 {
		health = "biased"
	}

	return &DatasetAnalytics{
		TotalSamples:           total,
		AcceptedSamples:        acceptedCount,
		RejectedSamples:        total - acceptedCount,
		AcceptanceRate:         (float64(acceptedCount) / float64(total)) * 100.0,
		AvgPromptTokens:        float64(totalPromptToks) / float64(total),
		AvgCompletionTokens:    float64(totalRespToks) / float64(total),
		PromptLengthHistogram:  mapToBuckets(promptBuckets, []string{"<50", "50-100", "100-250", "250-500", ">500"}),
		ResponseLengthHistogram: mapToBuckets(respBuckets, []string{"<25", "25-50", "50-100", "100-200", ">200"}),
		StrategyBreakdown:      sortMapByCount(strategyCounts),
		ScoreHistogram:         mapToBuckets(scoreBuckets, []string{"90-100%", "80-89%", "70-79%", "<70%"}),
		TotalVocabularyTokens:  totalWords,
		UniqueVocabularyTokens: uniqueWords,
		LexicalDiversityTTR:    math.Round(ttr*1000) / 1000,
		BalanceIndex:           math.Round(balanceIdx*100) / 100,
		HealthStatus:           health,
	}
}

func estimateTokens(text string) int {
	words := strings.Fields(text)
	toks := int(float64(len(words)) * 1.3)
	if toks == 0 && len(text) > 0 {
		return 1
	}
	return toks
}

func tokenizeWords(text string) []string {
	var words []string
	var current strings.Builder
	for _, r := range strings.ToLower(text) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(r)
		} else if current.Len() > 0 {
			words = append(words, current.String())
			current.Reset()
		}
	}
	if current.Len() > 0 {
		words = append(words, current.String())
	}
	return words
}

func mapToBuckets(m map[string]int, order []string) []DistributionBucket {
	var out []DistributionBucket
	for _, key := range order {
		out = append(out, DistributionBucket{Label: key, Count: m[key]})
	}
	return out
}

func sortMapByCount(m map[string]int) []DistributionBucket {
	var out []DistributionBucket
	for k, v := range m {
		out = append(out, DistributionBucket{Label: k, Count: v})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Count > out[j].Count
	})
	return out
}
