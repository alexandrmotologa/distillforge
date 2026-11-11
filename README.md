# distillforge

Turnkey LLM-to-SLM distillation, synthetic data pipeline, and edge fine-tuning studio.

Companies frequently route simple, repetitive domain tasks to large frontier foundation models such as GPT-4o or Claude 3.5 Sonnet. Using a 70B to 400B model for narrow classification, structured extraction, or query generation creates three problems: high API costs that scale with volume, 800ms to 2500ms response latencies, and data privacy risks when sending internal data to third-party endpoints.

distillforge automates the transfer of domain capabilities from frontier models into specialized 1B to 3B small language models (such as Llama 3.2 1B or 3B, Qwen 2.5 1.5B or 3B, and SmolLM2). The pipeline generates synthetic training datasets with Evol-Instruct, filters samples using LLM-as-a-judge, fine-tunes with LoRA or QLoRA, and exports quantized GGUF binaries and Ollama Modelfiles for local deployment.

```
+---------------------------------------------------------------------------------+
|                             distillforge Engine                                 |
+---------------------------------------------------------------------------------+
                                         |
     +-----------------------------------+-----------------------------------+
     |                                                                       |
     v                                                                       v
+-----------------------------+                         +-----------------------------+
|    Seed Ingestion Engine    |                         |   Declarative YAML Parser   |
| Reads raw prompt-target JSON|                         | Enforces constraints, schema|
+-----------------------------+                         +-----------------------------+
                                         |
                                         v
                      +-------------------------------------+
                      |   Synthetic Data Generator          |
                      |   Evol-Instruct mutation strategies |
                      |   In-depth, breadth, adversarial    |
                      +-------------------------------------+
                                         |
                                         v
                      +-------------------------------------+
                      |   Quality & Safety Guard Filter     |
                      |   LLM-as-a-judge multi-criteria     |
                      |   Strict JSON Schema verification   |
                      |   N-gram Jaccard deduplication      |
                      +-------------------------------------+
                                         |
                                         v
                      +-------------------------------------+
                      |   LoRA Fine-Tuning Runner           |
                      |   PyTorch SFTTrainer / Unsloth      |
                      |   Apple Silicon MLX adapter script  |
                      |   Real-time SSE metric telemetry    |
                      +-------------------------------------+
                                         |
                                         v
                      +-------------------------------------+
                      |   Head-to-Head Holdout Benchmarker  |
                      |   Accuracy, TTFT, token latency     |
                      |   Monthly cloud FinOps calculator   |
                      +-------------------------------------+
                                         |
                                         v
                      +-------------------------------------+
                      |   Quantization & Export Packaging   |
                      |   Quantized GGUF (Q4_K_M, Q8_0)     |
                      |   Automated Ollama Modelfile writer |
                      +-------------------------------------+
```

## Features

- Synthetic dataset generation: Expands 20 to 50 seed examples into 1,000 to 10,000 training pairs using weighted Evol-Instruct mutations (in-depth constraint deepening, in-breadth topic variations, adversarial edge cases, and step-by-step reasoning tokens).
- Multi-criteria quality filter: LLM-as-a-judge scoring engine evaluates instruction following, format cleanliness, and factual consistency. Enforces strict JSON Schema validation and purges near-duplicate samples via token n-gram Jaccard similarity.
- Adaptive training engine: Fine-tunes using PyTorch and Hugging Face PEFT with 4-bit QLoRA, FP16 LoRA, Apple Silicon MLX, or CPU fallback. Tracks loss, learning rate, perplexity, and VRAM in real time.
- Holdout benchmarking: Runs comparative tests between the Teacher model and the fine-tuned Student model, measuring exact match, schema compliance, time to first token (TTFT), and throughput.
- FinOps savings model: Calculates token burn, cloud API costs, self-hosted compute costs, annual budget reduction, and carbon footprint reduction.
- Local edge deployment: Converts trained adapters into quantized GGUF binaries (Q4_K_M, Q8_0) and writes production Ollama Modelfiles configured with appropriate system prompts, templates, and stop tokens.
- Embedded web studio: React 19 interface served directly by the Go binary with interactive training curves, head-to-head output diffs, dataset inspector, FinOps calculator, and interactive test playground.

## Installation

### Prerequisites
- Go 1.22 or newer
- Python 3.10 or newer (with PyTorch and Hugging Face transformers for GPU training)
- Optional: Node.js 20+ (only required if modifying the web studio source)

### Building the binary

Clone the repository and compile the orchestrator binary:

```bash
git clone https://github.com/alexandrmotologa/distillforge.git
cd distillforge

# Build the distillforge executable
go build -o distillforge ./cmd/distillforge
```

Install Python dependencies for fine-tuning:

```bash
pip install -r python/requirements.txt
```

## Quickstart

### 1. Launch the web studio

Run the studio server to access the dashboard and interactive wizard:

```bash
distillforge studio --port 8080
```

Open `http://localhost:8080` in your browser.

### 2. Run the end-to-end pipeline

Execute all five stages (generation, filtering, training, benchmarking, export) in a single command using the support ticket classifier example:

```bash
distillforge run --config ./examples/ticket_classifier/distillforge.yaml
```

### 3. Run individual stages

Each stage can be executed independently:

```bash
# Generate synthetic dataset from seeds
distillforge generate --config distillforge.yaml --output ./data/dataset_raw.jsonl

# Audit and filter with LLM-as-a-judge and schema checks
distillforge audit --config distillforge.yaml --input ./data/dataset_raw.jsonl --output ./data/dataset_clean.jsonl

# Fine-tune the student model with LoRA
distillforge train --config distillforge.yaml --dataset ./data/dataset_clean.jsonl --output ./checkpoints/best

# Benchmark student against teacher on holdout tests
distillforge evaluate --config distillforge.yaml --dataset ./data/dataset_clean.jsonl

# Export to quantized GGUF and write Ollama Modelfile
distillforge export --config distillforge.yaml --checkpoint ./checkpoints/best --output ./dist
```

## Configuration specification

Pipeline jobs are defined in a declarative YAML file (`distillforge.yaml`):

```yaml
version: "1"
project:
  name: "ticket-classifier"
  description: "Distill GPT-4o into Llama-3.2-1B for real-time ticket categorization"

teacher:
  provider: "openai" # openai, anthropic, ollama, custom, or mock
  model: "gpt-4o"
  temperature: 0.7
  max_tokens: 512
  api_key_env: "OPENAI_API_KEY"

student:
  base_model: "meta-llama/Llama-3.2-1B-Instruct"
  quantization_target: "q4_k_m"

seed_data:
  file: "./examples/ticket_classifier/data/raw_tickets.jsonl"
  input_key: "prompt"
  target_key: "response"

generation:
  target_dataset_size: 1000
  concurrency: 4
  evolution_strategies:
    - name: "in_depth"
      weight: 0.4
    - name: "in_breadth"
      weight: 0.3
    - name: "adversarial"
      weight: 0.3
  quality_threshold: 0.80

filter:
  schema_file: "./examples/ticket_classifier/schema.json"
  dedup_threshold: 0.85
  min_response_chars: 10
  max_response_chars: 2000

training:
  method: "qlora" # qlora, lora, mlx, or cpu
  r: 16
  lora_alpha: 32
  learning_rate: 0.0002
  epochs: 3
  batch_size: 4
  gradient_accumulation_steps: 4

evaluation:
  test_split: 0.2
  sample_size: 50
  metrics:
    - "exact_match"
    - "schema_adherence"
    - "latency_p95"

finops:
  monthly_requests: 1000000
  avg_prompt_tokens: 350
  avg_completion_tokens: 120
  teacher_cost_per_m_in: 5.00
  teacher_cost_per_m_out: 15.00
  self_hosted_monthly_infra_cost: 120.00

export:
  formats:
    - "gguf"
    - "ollama"
  ollama_modelfile_path: "./dist/Modelfile"
  output_dir: "./dist"
```

## Deploying to Ollama

After running `distillforge export`, deploy the quantized model to a local Ollama instance:

```bash
ollama create ticket-classifier -f ./dist/Modelfile
ollama run ticket-classifier "Customer ticket: ONT fiber light is red after power outage."
```

## Included domain examples

The `examples/` directory contains complete reference configurations, JSON schemas, and seed datasets:
- Support ticket classifier (`examples/ticket_classifier`): 12-class enterprise support routing with escalation triggers and SLA calculations.
- Invoice entity extractor (`examples/invoice_extractor`): Parses unstructured invoice OCR text into nested JSON records with line items.
- SQL query generator (`examples/sql_generator`): Translates natural language questions into relational SQL queries with table dependencies.

## Architecture and documentation

- [System architecture](docs/ARCHITECTURE.md): Component breakdown, Evol-Instruct mutation taxonomy, quality filtering logic, and training backends.
- [CLI reference](docs/CLI_REFERENCE.md): Detailed parameter reference for all subcommands.
- [FinOps and rightsizing guide](docs/FINOPS_GUIDE.md): Token unit economics, infrastructure amortization, and ROI equations.

## License

MIT
