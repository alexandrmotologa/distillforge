package generator

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/alexandrmotologa/distillforge/internal/config"
)

// StrategyType defines the prompt evolution taxonomy.
type StrategyType string

const (
	StrategyInDepth     StrategyType = "in_depth"
	StrategyInBreadth   StrategyType = "in_breadth"
	StrategyAdversarial StrategyType = "adversarial"
	StrategyCoT         StrategyType = "cot"
)

// EvolInstructEngine handles mutating seed prompts into diverse synthetic training instances.
type EvolInstructEngine struct {
	client     TeacherClient
	strategies []config.EvolutionStrategy
	totalWeight float64
}

// NewEvolInstructEngine initializes an evolution engine with weighted strategies.
func NewEvolInstructEngine(client TeacherClient, strategies []config.EvolutionStrategy) *EvolInstructEngine {
	var total float64
	for _, s := range strategies {
		total += s.Weight
	}
	if total == 0 {
		strategies = []config.EvolutionStrategy{
			{Name: string(StrategyInDepth), Weight: 0.4},
			{Name: string(StrategyInBreadth), Weight: 0.3},
			{Name: string(StrategyAdversarial), Weight: 0.3},
		}
		total = 1.0
	}

	return &EvolInstructEngine{
		client:      client,
		strategies:  strategies,
		totalWeight: total,
	}
}

// SelectStrategy picks an evolution strategy based on configured weights.
func (e *EvolInstructEngine) SelectStrategy() string {
	r := rand.Float64() * e.totalWeight
	var accum float64
	for _, s := range e.strategies {
		accum += s.Weight
		if r <= accum {
			return s.Name
		}
	}
	return e.strategies[len(e.strategies)-1].Name
}

// EvolvePrompt transforms an input prompt using the selected evolution strategy.
func (e *EvolInstructEngine) EvolvePrompt(ctx context.Context, seedPrompt string, strategy string) (string, error) {
	systemPrompt, userPrompt := buildEvolutionPrompts(seedPrompt, strategy)
	mutated, err := e.client.Generate(ctx, systemPrompt, userPrompt)
	if err != nil {
		return "", fmt.Errorf("prompt evolution failed for strategy %s: %w", strategy, err)
	}

	cleaned := strings.TrimSpace(mutated)
	// Strip enclosing quotes or prefixes if present
	cleaned = strings.TrimPrefix(cleaned, "Evolved Prompt:")
	cleaned = strings.TrimPrefix(cleaned, "Evolved:")
	cleaned = strings.TrimSpace(cleaned)
	return cleaned, nil
}

func buildEvolutionPrompts(seed, strategy string) (string, string) {
	switch strategy {
	case "in_depth":
		sys := "You are an Evol-Instruct AI dataset specialist executing in-depth constraint evolution. Deepen the complexity of the domain instruction by adding realistic operational constraints, compliance criteria, or domain specific edge conditions. Return ONLY the evolved prompt with no meta-commentary."
		usr := fmt.Sprintf("Evolve the following instruction into a more complex, in-depth scenario while preserving its original core objective:\nOriginal:\n%s", seed)
		return sys, usr

	case "in_breadth":
		sys := "You are an Evol-Instruct AI dataset specialist executing in-breadth topic evolution. Generate a completely fresh instruction that belongs to the same domain and general complexity but covers a different scenario, user tone, or operational context. Return ONLY the evolved prompt with no meta-commentary."
		usr := fmt.Sprintf("Generate a parallel domain instruction with similar complexity but different scenario details:\nInput:\n%s", seed)
		return sys, usr

	case "adversarial":
		sys := "You are an Evol-Instruct AI dataset specialist executing adversarial edge-case evolution. Mutate the input prompt by injecting boundary conditions, typos, confusing language, subtle ambiguities, or urgent escalation triggers that models typically struggle with. Return ONLY the evolved prompt with no meta-commentary."
		usr := fmt.Sprintf("Make the following prompt an adversarial edge case to test model robustness:\nInput:\n%s", seed)
		return sys, usr

	case "cot":
		sys := "You are an Evol-Instruct AI dataset specialist. Rewrite the instruction to explicitly request step-by-step reasoning or internal chain-of-thought analysis before concluding with the final answer. Return ONLY the evolved prompt."
		usr := fmt.Sprintf("Rewrite this instruction to enforce structured step-by-step reasoning:\nInput:\n%s", seed)
		return sys, usr

	default:
		sys := "You are an Evol-Instruct AI dataset specialist. Rephrase the input instruction to improve lexical diversity while retaining its exact functional intent. Return ONLY the evolved prompt."
		usr := fmt.Sprintf("Rephrase this instruction:\nInput:\n%s", seed)
		return sys, usr
	}
}
