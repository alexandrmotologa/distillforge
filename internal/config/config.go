package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the top-level declarative distillation job specification.
type Config struct {
	Version    string           `yaml:"version" json:"version"`
	Project    ProjectConfig    `yaml:"project" json:"project"`
	Teacher    TeacherConfig    `yaml:"teacher" json:"teacher"`
	Student    StudentConfig    `yaml:"student" json:"student"`
	SeedData   SeedDataConfig   `yaml:"seed_data" json:"seed_data"`
	Generation GenerationConfig `yaml:"generation" json:"generation"`
	Filter     FilterConfig     `yaml:"filter" json:"filter"`
	Training   TrainingConfig   `yaml:"training" json:"training"`
	Evaluation EvaluationConfig `yaml:"evaluation" json:"evaluation"`
	FinOps     FinOpsConfig     `yaml:"finops" json:"finops"`
	Export     ExportConfig     `yaml:"export" json:"export"`
}

type ProjectConfig struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
}

type TeacherConfig struct {
	Provider    string  `yaml:"provider" json:"provider"` // openai, anthropic, ollama, custom, mock
	Model       string  `yaml:"model" json:"model"`
	BaseURL     string  `yaml:"base_url,omitempty" json:"base_url,omitempty"`
	Temperature float64 `yaml:"temperature" json:"temperature"`
	MaxTokens   int     `yaml:"max_tokens" json:"max_tokens"`
	APIKeyEnv   string  `yaml:"api_key_env,omitempty" json:"api_key_env,omitempty"`
}

type StudentConfig struct {
	BaseModel          string `yaml:"base_model" json:"base_model"`
	QuantizationTarget string `yaml:"quantization_target" json:"quantization_target"` // q4_k_m, q8_0, etc.
}

type SeedDataConfig struct {
	File      string `yaml:"file" json:"file"`
	InputKey  string `yaml:"input_key" json:"input_key"`
	TargetKey string `yaml:"target_key" json:"target_key"`
}

type EvolutionStrategy struct {
	Name   string  `yaml:"name" json:"name"` // in_depth, in_breadth, adversarial, cot
	Weight float64 `yaml:"weight" json:"weight"`
}

type GenerationConfig struct {
	TargetDatasetSize   int                 `yaml:"target_dataset_size" json:"target_dataset_size"`
	Concurrency         int                 `yaml:"concurrency" json:"concurrency"`
	EvolutionStrategies []EvolutionStrategy `yaml:"evolution_strategies" json:"evolution_strategies"`
	QualityThreshold    float64             `yaml:"quality_threshold" json:"quality_threshold"`
}

type FilterConfig struct {
	SchemaFile       string  `yaml:"schema_file,omitempty" json:"schema_file,omitempty"`
	DedupThreshold   float64 `yaml:"dedup_threshold" json:"dedup_threshold"`
	MinResponseChars int     `yaml:"min_response_chars" json:"min_response_chars"`
	MaxResponseChars int     `yaml:"max_response_chars" json:"max_response_chars"`
}

type TrainingConfig struct {
	Method                   string  `yaml:"method" json:"method"` // qlora, lora, cpu
	R                        int     `yaml:"r" json:"r"`
	LoRAAlpha                int     `yaml:"lora_alpha" json:"lora_alpha"`
	LoRADropout              float64 `yaml:"lora_dropout" json:"lora_dropout"`
	LearningRate             float64 `yaml:"learning_rate" json:"learning_rate"`
	Epochs                   int     `yaml:"epochs" json:"epochs"`
	BatchSize                int     `yaml:"batch_size" json:"batch_size"`
	GradientAccumulationSteps int    `yaml:"gradient_accumulation_steps" json:"gradient_accumulation_steps"`
	MaxSeqLength             int     `yaml:"max_seq_length" json:"max_seq_length"`
	Optimizer                string  `yaml:"optimizer" json:"optimizer"`
	WarmupRatio              float64 `yaml:"warmup_ratio" json:"warmup_ratio"`
}

type EvaluationConfig struct {
	TestSplit  float64  `yaml:"test_split" json:"test_split"`
	SampleSize int      `yaml:"sample_size" json:"sample_size"`
	Metrics    []string `yaml:"metrics" json:"metrics"`
}

type FinOpsConfig struct {
	MonthlyRequests            int64   `yaml:"monthly_requests" json:"monthly_requests"`
	AvgPromptTokens            int     `yaml:"avg_prompt_tokens" json:"avg_prompt_tokens"`
	AvgCompletionTokens        int     `yaml:"avg_completion_tokens" json:"avg_completion_tokens"`
	TeacherCostPerMIn          float64 `yaml:"teacher_cost_per_m_in" json:"teacher_cost_per_m_in"`
	TeacherCostPerMOut         float64 `yaml:"teacher_cost_per_m_out" json:"teacher_cost_per_m_out"`
	SelfHostedMonthlyInfraCost float64 `yaml:"self_hosted_monthly_infra_cost" json:"self_hosted_monthly_infra_cost"`
}

type ExportConfig struct {
	Formats             []string `yaml:"formats" json:"formats"` // gguf, ollama, hf
	OllamaModelfilePath string   `yaml:"ollama_modelfile_path" json:"ollama_modelfile_path"`
	OutputDir           string   `yaml:"output_dir" json:"output_dir"`
}

// Load reads and parses a YAML configuration file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %q: %w", path, err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse yaml config: %w", err)
	}

	cfg.SetDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// SetDefaults assigns sensible default values for omitted optional settings.
func (c *Config) SetDefaults() {
	if c.Version == "" {
		c.Version = "1"
	}

	if c.Teacher.Provider == "" {
		c.Teacher.Provider = "openai"
	}
	if c.Teacher.Model == "" {
		c.Teacher.Model = "gpt-4o"
	}
	if c.Teacher.Temperature == 0 {
		c.Teacher.Temperature = 0.7
	}
	if c.Teacher.MaxTokens == 0 {
		c.Teacher.MaxTokens = 1024
	}
	if c.Teacher.APIKeyEnv == "" {
		c.Teacher.APIKeyEnv = "OPENAI_API_KEY"
	}

	if c.Student.BaseModel == "" {
		c.Student.BaseModel = "meta-llama/Llama-3.2-1B-Instruct"
	}
	if c.Student.QuantizationTarget == "" {
		c.Student.QuantizationTarget = "q4_k_m"
	}

	if c.SeedData.InputKey == "" {
		c.SeedData.InputKey = "prompt"
	}
	if c.SeedData.TargetKey == "" {
		c.SeedData.TargetKey = "response"
	}

	if c.Generation.TargetDatasetSize == 0 {
		c.Generation.TargetDatasetSize = 1000
	}
	if c.Generation.Concurrency <= 0 {
		c.Generation.Concurrency = 4
	}
	if len(c.Generation.EvolutionStrategies) == 0 {
		c.Generation.EvolutionStrategies = []EvolutionStrategy{
			{Name: "in_depth", Weight: 0.4},
			{Name: "in_breadth", Weight: 0.3},
			{Name: "adversarial", Weight: 0.3},
		}
	}
	if c.Generation.QualityThreshold == 0 {
		c.Generation.QualityThreshold = 0.80
	}

	if c.Filter.DedupThreshold == 0 {
		c.Filter.DedupThreshold = 0.85
	}
	if c.Filter.MinResponseChars == 0 {
		c.Filter.MinResponseChars = 5
	}
	if c.Filter.MaxResponseChars == 0 {
		c.Filter.MaxResponseChars = 8000
	}

	if c.Training.Method == "" {
		c.Training.Method = "qlora"
	}
	if c.Training.R == 0 {
		c.Training.R = 16
	}
	if c.Training.LoRAAlpha == 0 {
		c.Training.LoRAAlpha = 32
	}
	if c.Training.LoRADropout == 0 {
		c.Training.LoRADropout = 0.05
	}
	if c.Training.LearningRate == 0 {
		c.Training.LearningRate = 0.0002
	}
	if c.Training.Epochs == 0 {
		c.Training.Epochs = 3
	}
	if c.Training.BatchSize == 0 {
		c.Training.BatchSize = 4
	}
	if c.Training.GradientAccumulationSteps == 0 {
		c.Training.GradientAccumulationSteps = 4
	}
	if c.Training.MaxSeqLength == 0 {
		c.Training.MaxSeqLength = 2048
	}
	if c.Training.Optimizer == "" {
		c.Training.Optimizer = "adamw_torch"
	}
	if c.Training.WarmupRatio == 0 {
		c.Training.WarmupRatio = 0.03
	}

	if c.Evaluation.TestSplit == 0 {
		c.Evaluation.TestSplit = 0.2
	}
	if c.Evaluation.SampleSize == 0 {
		c.Evaluation.SampleSize = 50
	}
	if len(c.Evaluation.Metrics) == 0 {
		c.Evaluation.Metrics = []string{
			"exact_match",
			"schema_adherence",
			"latency_p95",
			"cost_per_million",
		}
	}

	if c.FinOps.MonthlyRequests == 0 {
		c.FinOps.MonthlyRequests = 1000000
	}
	if c.FinOps.AvgPromptTokens == 0 {
		c.FinOps.AvgPromptTokens = 350
	}
	if c.FinOps.AvgCompletionTokens == 0 {
		c.FinOps.AvgCompletionTokens = 120
	}
	if c.FinOps.TeacherCostPerMIn == 0 {
		c.FinOps.TeacherCostPerMIn = 5.00
	}
	if c.FinOps.TeacherCostPerMOut == 0 {
		c.FinOps.TeacherCostPerMOut = 15.00
	}
	if c.FinOps.SelfHostedMonthlyInfraCost == 0 {
		c.FinOps.SelfHostedMonthlyInfraCost = 120.00
	}

	if len(c.Export.Formats) == 0 {
		c.Export.Formats = []string{"gguf", "ollama"}
	}
	if c.Export.OllamaModelfilePath == "" {
		c.Export.OllamaModelfilePath = "./Modelfile"
	}
	if c.Export.OutputDir == "" {
		c.Export.OutputDir = "./dist"
	}
}

// Validate checks business invariants and configuration constraints.
func (c *Config) Validate() error {
	if c.Project.Name == "" {
		return fmt.Errorf("project.name is required")
	}

	validProviders := map[string]bool{
		"openai":    true,
		"anthropic": true,
		"ollama":    true,
		"custom":    true,
		"mock":      true,
	}
	if !validProviders[c.Teacher.Provider] {
		return fmt.Errorf("unsupported teacher provider %q (allowed: openai, anthropic, ollama, custom, mock)", c.Teacher.Provider)
	}

	if c.Student.BaseModel == "" {
		return fmt.Errorf("student.base_model is required")
	}

	if c.SeedData.File == "" {
		return fmt.Errorf("seed_data.file is required")
	}

	validMethods := map[string]bool{
		"qlora": true,
		"lora":  true,
		"cpu":   true,
		"mlx":   true,
	}
	if !validMethods[c.Training.Method] {
		return fmt.Errorf("unsupported training method %q (allowed: qlora, lora, cpu, mlx)", c.Training.Method)
	}

	if c.Generation.QualityThreshold < 0 || c.Generation.QualityThreshold > 1 {
		return fmt.Errorf("generation.quality_threshold must be between 0.0 and 1.0")
	}

	if c.Evaluation.TestSplit <= 0 || c.Evaluation.TestSplit >= 1 {
		return fmt.Errorf("evaluation.test_split must be between 0.0 and 1.0 exclusive")
	}

	return nil
}
