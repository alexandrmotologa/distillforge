package advisor

import (
	"fmt"
	"strings"
)

// HardwarePreset encapsulates hardware platform and compute capacities.
type HardwarePreset string

const (
	PresetAppleSilicon HardwarePreset = "apple_silicon"
	PresetGPU8GB        HardwarePreset = "gpu_8gb"
	PresetGPU16GB24GB   HardwarePreset = "gpu_16gb_24gb"
	PresetCPU           HardwarePreset = "cpu"
)

// DistillationRecipe contains recommended fine-tuning hyperparameters.
type DistillationRecipe struct {
	HardwareTarget       string  `json:"hardware_target"`
	RecommendedModel     string  `json:"recommended_model"`
	TrainingMethod       string  `json:"training_method"` // qlora, lora, mlx, cpu
	Rank                 int     `json:"r"`
	LoRAAlpha            int     `json:"lora_alpha"`
	LoRADropout          float64 `json:"lora_dropout"`
	LearningRate         float64 `json:"learning_rate"`
	BatchSize            int     `json:"batch_size"`
	GradientAccumulation int     `json:"gradient_accumulation_steps"`
	MaxSeqLength         int     `json:"max_seq_length"`
	QuantizationTarget   string  `json:"quantization_target"`
	EstimatedVRAMMB      int     `json:"estimated_vram_mb"`
	EstimatedTokPerSec   float64 `json:"estimated_tok_per_sec"`
	Rationale            string  `json:"rationale"`
}

// RecommendRecipe analyzes hardware target, dataset size, and task type.
func RecommendRecipe(hardware string, datasetSize int, taskType string) DistillationRecipe {
	h := strings.ToLower(strings.TrimSpace(hardware))

	switch h {
	case "apple_silicon", "mac", "mlx":
		return DistillationRecipe{
			HardwareTarget:       "Apple Silicon (M-series MLX)",
			RecommendedModel:     "meta-llama/Llama-3.2-1B-Instruct",
			TrainingMethod:       "mlx",
			Rank:                 16,
			LoRAAlpha:            32,
			LoRADropout:          0.05,
			LearningRate:         0.0002,
			BatchSize:            4,
			GradientAccumulation: 4,
			MaxSeqLength:         2048,
			QuantizationTarget:   "q4_k_m",
			EstimatedVRAMMB:      3200,
			EstimatedTokPerSec:   85.0,
			Rationale:            "Uses Apple Silicon unified memory with MLX framework, providing 85+ tok/sec localized inference with zero CUDA dependency.",
		}

	case "gpu_8gb", "rtx4060", "rtx3070", "consumer_gpu":
		return DistillationRecipe{
			HardwareTarget:       "Consumer GPU (8GB VRAM)",
			RecommendedModel:     "meta-llama/Llama-3.2-1B-Instruct",
			TrainingMethod:       "qlora",
			Rank:                 16,
			LoRAAlpha:            32,
			LoRADropout:          0.05,
			LearningRate:         0.0002,
			BatchSize:            2,
			GradientAccumulation: 8,
			MaxSeqLength:         2048,
			QuantizationTarget:   "q4_k_m",
			EstimatedVRAMMB:      4800,
			EstimatedTokPerSec:   65.0,
			Rationale:            "Fits comfortably inside 8GB VRAM using 4-bit NF4 bitsandbytes quantization, maintaining sub-20ms edge latency.",
		}

	case "gpu_16gb_24gb", "a10g", "rtx4090", "a5000", "enterprise_gpu":
		model := "meta-llama/Llama-3.2-3B-Instruct"
		if taskType == "classification" {
			model = "meta-llama/Llama-3.2-1B-Instruct"
		}
		return DistillationRecipe{
			HardwareTarget:       "Enterprise GPU (16GB - 24GB VRAM)",
			RecommendedModel:     model,
			TrainingMethod:       "qlora",
			Rank:                 32,
			LoRAAlpha:            64,
			LoRADropout:          0.05,
			LearningRate:         0.00015,
			BatchSize:            8,
			GradientAccumulation: 2,
			MaxSeqLength:         4096,
			QuantizationTarget:   "q8_0",
			EstimatedVRAMMB:      8500,
			EstimatedTokPerSec:   110.0,
			Rationale:            "Utilizes high memory headroom to train larger rank r=32 with Q8_0 precision for near-zero perplexity loss.",
		}

	case "cpu", "serverless", "minimal":
		fallthrough
	default:
		return DistillationRecipe{
			HardwareTarget:       "Standard CPU / Edge Gateway",
			RecommendedModel:     "HuggingFaceTB/SmolLM2-1.7B-Instruct",
			TrainingMethod:       "cpu",
			Rank:                 8,
			LoRAAlpha:            16,
			LoRADropout:          0.05,
			LearningRate:         0.0003,
			BatchSize:            2,
			GradientAccumulation: 4,
			MaxSeqLength:         1024,
			QuantizationTarget:   "q4_k_m",
			EstimatedVRAMMB:      1800,
			EstimatedTokPerSec:   25.0,
			Rationale:            "Designed for systems with no dedicated GPU, training on AVX-512 CPU instructions with low memory footprint.",
		}
	}
}

// PrintRecipeConsole formats the recipe for CLI display.
func PrintRecipeConsole(r DistillationRecipe) {
	fmt.Printf("\n--- distillforge Hardware Recipe: %s ---\n", r.HardwareTarget)
	fmt.Printf("  Recommended Base Model: %s\n", r.RecommendedModel)
	fmt.Printf("  Training Precision:     %s (rank: %d, alpha: %d)\n", r.TrainingMethod, r.Rank, r.LoRAAlpha)
	fmt.Printf("  Batch / Grad Accum:     %d / %d (LR: %e)\n", r.BatchSize, r.GradientAccumulation, r.LearningRate)
	fmt.Printf("  Quantization Target:    %s\n", r.QuantizationTarget)
	fmt.Printf("  Estimated Memory:       %d MB VRAM\n", r.EstimatedVRAMMB)
	fmt.Printf("  Estimated Throughput:   %.1f tokens/sec\n", r.EstimatedTokPerSec)
	fmt.Printf("  Engineering Rationale:  %s\n\n", r.Rationale)
}
