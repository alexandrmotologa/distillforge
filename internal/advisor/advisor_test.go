package advisor

import (
	"testing"
)

func TestRecommendRecipe(t *testing.T) {
	tests := []struct {
		hardware    string
		wantMethod  string
		wantRank    int
		wantModel   string
	}{
		{"apple_silicon", "mlx", 16, "meta-llama/Llama-3.2-1B-Instruct"},
		{"gpu_8gb", "qlora", 16, "meta-llama/Llama-3.2-1B-Instruct"},
		{"gpu_16gb_24gb", "qlora", 32, "meta-llama/Llama-3.2-3B-Instruct"},
		{"cpu", "cpu", 8, "HuggingFaceTB/SmolLM2-1.7B-Instruct"},
	}

	for _, tt := range tests {
		t.Run(tt.hardware, func(t *testing.T) {
			recipe := RecommendRecipe(tt.hardware, 1000, "extraction")
			if recipe.TrainingMethod != tt.wantMethod {
				t.Errorf("expected method %s, got %s", tt.wantMethod, recipe.TrainingMethod)
			}
			if recipe.Rank != tt.wantRank {
				t.Errorf("expected rank %d, got %d", tt.wantRank, recipe.Rank)
			}
			if recipe.RecommendedModel != tt.wantModel {
				t.Errorf("expected model %s, got %s", tt.wantModel, recipe.RecommendedModel)
			}
		})
	}
}
