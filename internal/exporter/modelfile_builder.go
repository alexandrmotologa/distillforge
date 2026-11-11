package exporter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ModelfileOptions defines parameters for the Ollama Modelfile.
type ModelfileOptions struct {
	GGUFPath     string
	BaseModel    string
	SystemPrompt string
	Temperature  float64
	TopP         float64
	StopTokens   []string
}

// BuildModelfileContent formats the Ollama Modelfile text based on model architecture.
func BuildModelfileContent(opts ModelfileOptions) string {
	var sb strings.Builder

	// 1. FROM clause
	sb.WriteString(fmt.Sprintf("FROM %s\n\n", filepath.ToSlash(opts.GGUFPath)))

	// 2. TEMPLATE clause (Llama-3 / ChatML formatting)
	if strings.Contains(strings.ToLower(opts.BaseModel), "llama-3") {
		sb.WriteString(`TEMPLATE """{{ if .System }}<|start_header_id|>system<|end_header_id|>

{{ .System }}<|eot_id|>{{ end }}{{ if .Prompt }}<|start_header_id|>user<|end_header_id|>

{{ .Prompt }}<|eot_id|>{{ end }}<|start_header_id|>assistant<|end_header_id|>

{{ .Response }}<|eot_id|>"""
`)
	} else {
		// Generic ChatML default
		sb.WriteString(`TEMPLATE """{{ if .System }}<|im_start|>system
{{ .System }}<|im_end|>
{{ end }}{{ if .Prompt }}<|im_start|>user
{{ .Prompt }}<|im_end|>
{{ end }}<|im_start|>assistant
{{ .Response }}<|im_end|>"""
`)
	}
	sb.WriteString("\n")

	// 3. PARAMETER settings
	temp := opts.Temperature
	if temp == 0 {
		temp = 0.7
	}
	topP := opts.TopP
	if topP == 0 {
		topP = 0.9
	}
	sb.WriteString(fmt.Sprintf("PARAMETER temperature %.2f\n", temp))
	sb.WriteString(fmt.Sprintf("PARAMETER top_p %.2f\n", topP))

	stops := opts.StopTokens
	if len(stops) == 0 {
		if strings.Contains(strings.ToLower(opts.BaseModel), "llama-3") {
			stops = []string{"<|eot_id|>", "<|start_header_id|>", "<|end_header_id|>"}
		} else {
			stops = []string{"<|im_end|>", "<|im_start|>"}
		}
	}
	for _, st := range stops {
		sb.WriteString(fmt.Sprintf("PARAMETER stop %q\n", st))
	}
	sb.WriteString("\n")

	// 4. SYSTEM prompt
	sys := opts.SystemPrompt
	if sys == "" {
		sys = "You are a specialized enterprise Small Language Model distilled with distillforge. Provide direct, high-precision structured answers."
	}
	sb.WriteString(fmt.Sprintf("SYSTEM \"\"\"%s\"\"\"\n", sys))

	return sb.String()
}

// WriteModelfile creates the Modelfile on disk.
func WriteModelfile(outputPath string, opts ModelfileOptions) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}
	content := BuildModelfileContent(opts)
	return os.WriteFile(outputPath, []byte(content), 0644)
}
