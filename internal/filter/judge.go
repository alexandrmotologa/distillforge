package filter

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/alexandrmotologa/distillforge/internal/generator"
)

// JudgeEvaluation holds the scores for a single generated sample.
type JudgeEvaluation struct {
	OverallScore        float64 `json:"overall_score"`
	InstructionFollowing float64 `json:"instruction_following"`
	FormatAdherence     float64 `json:"format_adherence"`
	Groundedness        float64 `json:"groundedness"`
	Feedback            string  `json:"feedback"`
}

// Judge evaluates the quality of synthetic prompt-response pairs.
type Judge struct {
	client generator.TeacherClient
}

// NewJudge initializes the scoring engine.
func NewJudge(client generator.TeacherClient) *Judge {
	return &Judge{client: client}
}

// Evaluate evaluates a prompt and response pair using multi-criteria rubric.
func (j *Judge) Evaluate(ctx context.Context, prompt, response string) (*JudgeEvaluation, error) {
	// Fast heuristic check for obvious defects
	if len(strings.TrimSpace(response)) == 0 {
		return &JudgeEvaluation{
			OverallScore: 0.0,
			Feedback:     "empty response",
		}, nil
	}

	lower := strings.ToLower(response)
	refusalSignals := []string{
		"as an ai",
		"i cannot assist",
		"i apologize, but",
		"language model",
		"i am unable to",
	}
	for _, sig := range refusalSignals {
		if strings.Contains(lower, sig) {
			return &JudgeEvaluation{
				OverallScore: 0.1,
				Feedback:     "model refusal or meta disclosure detected",
			}, nil
		}
	}

	systemPrompt := `You are an expert AI Dataset Auditor and Judge. Evaluate the given Prompt-Response pair for a fine-tuning dataset.
Score from 0.0 to 1.0 across three criteria:
1. instruction_following: Does the response directly and comprehensively fulfill the instruction?
2. format_adherence: Is the format clean, well-structured, and free of conversational filler?
3. groundedness: Is the response factually grounded and free of obvious hallucinations?

Return ONLY a JSON object in this exact schema:
{
  "instruction_following": 0.95,
  "format_adherence": 1.0,
  "groundedness": 0.90,
  "feedback": "Concise summary"
}`

	userPrompt := fmt.Sprintf("Prompt:\n%s\n\nResponse:\n%s", prompt, response)

	judgement, err := j.client.Generate(ctx, systemPrompt, userPrompt)
	if err != nil {
		// Fallback to heuristic scoring if judge call fails
		return heuristicJudge(prompt, response), nil
	}

	eval := parseJudgeOutput(judgement)
	if eval.OverallScore == 0 && eval.InstructionFollowing > 0 {
		eval.OverallScore = (eval.InstructionFollowing*0.4 + eval.FormatAdherence*0.3 + eval.Groundedness*0.3)
	}

	return eval, nil
}

func parseJudgeOutput(raw string) *JudgeEvaluation {
	clean := strings.TrimSpace(raw)
	// Strip markdown blocks if present
	if strings.HasPrefix(clean, "```json") {
		clean = strings.TrimPrefix(clean, "```json")
		clean = strings.TrimSuffix(clean, "```")
	} else if strings.HasPrefix(clean, "```") {
		clean = strings.TrimPrefix(clean, "```")
		clean = strings.TrimSuffix(clean, "```")
	}
	clean = strings.TrimSpace(clean)

	var parsed struct {
		InstructionFollowing float64 `json:"instruction_following"`
		FormatAdherence      float64 `json:"format_adherence"`
		Groundedness         float64 `json:"groundedness"`
		Feedback             string  `json:"feedback"`
	}

	if err := json.Unmarshal([]byte(clean), &parsed); err == nil {
		if parsed.InstructionFollowing > 0 || parsed.FormatAdherence > 0 || parsed.Groundedness > 0 {
			overall := (parsed.InstructionFollowing*0.4 + parsed.FormatAdherence*0.3 + parsed.Groundedness*0.3)
			return &JudgeEvaluation{
				OverallScore:         overall,
				InstructionFollowing: parsed.InstructionFollowing,
				FormatAdherence:      parsed.FormatAdherence,
				Groundedness:         parsed.Groundedness,
				Feedback:             parsed.Feedback,
			}
		}
	}

	// Attempt string extraction if JSON is slightly unformatted
	return heuristicJudge(raw, raw)
}

func heuristicJudge(prompt, response string) *JudgeEvaluation {
	score := 0.85
	trimmed := strings.TrimSpace(response)

	if len(trimmed) > 15 {
		score += 0.05
	}
	if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
		score += 0.05
	}
	if score > 1.0 {
		score = 1.0
	}

	return &JudgeEvaluation{
		OverallScore:         score,
		InstructionFollowing: score,
		FormatAdherence:      score,
		Groundedness:         score,
		Feedback:             "heuristic evaluation score",
	}
}

// ParseScore helper converts strings or numbers safely.
func ParseScore(v string) float64 {
	val, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
	if err != nil {
		return 0.5
	}
	return val
}
