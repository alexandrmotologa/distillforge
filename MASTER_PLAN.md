# distillforge Architecture and Implementation Plan
> Turnkey LLM to SLM Distillation, Synthetic Data Pipeline, and Edge Fine-Tuning Studio

## 1. Overview and problem statement

### 1.1 The foundation model cost problem
Modern enterprise systems frequently route simple, repetitive domain tasks to large frontier models such as GPT-4o or Claude 3.5 Sonnet at high token costs. These tasks include ticket categorization, structured entity extraction, natural language to SQL translation, and policy verification.

Routing routine tasks to 70B through 405B parameter models introduces three core problems:
1. High operational expenditure: API costs scale directly with request volume, resulting in substantial recurring bills.
2. Latency overhead: Cloud API roundtrips frequently range from 500ms to 2500ms time to first token.
3. Data privacy concerns: Proprietary enterprise inputs leave internal network boundaries.

### 1.2 Small language models on the edge
A fine-tuned 1B to 3B small language model (such as Llama 3.2 1B or 3B, Qwen 2.5 1.5B or 3B, or SmolLM2) can match frontier model task accuracy on well-defined domains while delivering sub-20ms inference latency, over 90 percent cost reduction, and private local deployment.

Historically, distillation requires manual engineering across multiple disconnected tools: generating synthetic datasets, filtering hallucinations, formatting LoRA datasets, configuring hyperparameters, evaluating holdout sets, and quantizing models to GGUF format.

distillforge unifies this lifecycle into a single Go CLI binary, a Python training engine, and an embedded React 19 web studio.

## 2. System architecture

```
+-----------------------------------------------------------------------------------+
|                              distillforge Platform                                |
+-----------------------------------------------------------------------------------+
                                         |
     +-----------------------------------+-----------------------------------+
     |                                                                       |
     v                                                                       v
+-----------------------------+                         +-----------------------------+
|    Seed Examples (JSONL)    |                         |  Declarative Job Spec YAML  |
| - 20 to 50 domain prompts   |                         | - Teacher & Student configs |
| - Edge cases and targets    |                         | - Schema, LoRA & FinOps     |
+-----------------------------+                         +-----------------------------+
     |                                                                       |
     +-----------------------------------+-----------------------------------+
                                         |
                                         v
                      +-------------------------------------+
                      |   Synthetic Data Generator Engine   |
                      |   - Evol-Instruct mutations         |
                      |   - In-depth constraint deepening   |
                      |   - In-breadth topic variations     |
                      |   - Adversarial edge-case generator |
                      |   - Step-by-step CoT reasoning      |
                      +-------------------------------------+
                                         |
                                         v
                      +-------------------------------------+
                      |     Quality & Safety Filter Guard   |
                      |   - LLM-as-a-judge multi-criteria   |
                      |   - Strict JSON Schema validator    |
                      |   - N-gram cosine deduplication     |
                      |   - Heuristic length and safety checks
                      +-------------------------------------+
                                         |
                                         v
                      +-------------------------------------+
                      |      Supervised Dataset Exporter    |
                      |   - Alpaca, ShareGPT, OpenAI JSONL  |
                      |   - Train / Validation holdout split|
                      +-------------------------------------+
                                         |
                                         v
                      +-------------------------------------+
                      |    LoRA and QLoRA Training Engine   |
                      |   - PyTorch SFTTrainer / Unsloth    |
                      |   - Apple Silicon MLX adapter       |
                      |   - Real-time loss and VRAM via SSE |
                      +-------------------------------------+
                                         |
                                         v
                      +-------------------------------------+
                      |     Benchmark and FinOps Evaluator  |
                      |   - Teacher vs Student holdout diff |
                      |   - TTFT, latency, tokens per second|
                      |   - Monthly cloud dollar ROI report |
                      +-------------------------------------+
                                         |
                                         v
                      +-------------------------------------+
                      |       Quantize, Package and Deploy  |
                      |   - Quantized GGUF (Q4_K_M, Q8_0)   |
                      |   - Ollama Modelfile generator      |
                      |   - Merged FP16 adapter export      |
                      +-------------------------------------+
                                         |
                                         v
                      +-------------------------------------+
                      |      Embedded Web Studio UI         |
                      |   - React 19 + Tailwind CSS         |
                      |   - Training curves and loss charts |
                      |   - Head-to-head comparison diff    |
                      |   - Interactive FinOps calculator   |
                      |   - Real-time model test playground |
                      +-------------------------------------+
```

## 3. Declarative distillation configuration format

```yaml
version: "1"
project:
  name: "ticket-classifier"
  description: "Distill GPT-4o into Llama-3.2-1B for real-time support ticket categorization"

teacher:
  provider: "openai" # openai, anthropic, ollama, custom, or mock
  model: "gpt-4o"
  temperature: 0.7
  max_tokens: 1024
  api_key_env: "OPENAI_API_KEY"

student:
  base_model: "meta-llama/Llama-3.2-1B-Instruct"
  quantization_target: "q4_k_m"

seed_data:
  file: "./data/raw_tickets.jsonl"
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
  schema_file: "./schema.json"
  dedup_threshold: 0.85
  min_response_chars: 10
  max_response_chars: 4000

training:
  method: "qlora" # qlora, lora, or cpu
  r: 16
  lora_alpha: 32
  lora_dropout: 0.05
  learning_rate: 0.0002
  epochs: 3
  batch_size: 4
  gradient_accumulation_steps: 4
  max_seq_length: 2048
  optimizer: "adamw_torch"

evaluation:
  test_split: 0.2
  sample_size: 50
  metrics:
    - "exact_match"
    - "schema_adherence"
    - "latency_p95"
    - "cost_per_million"

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
  ollama_modelfile_path: "./Modelfile"
```

## 4. Implementation plan and tasks

### Task 1: Repository setup and configuration engine
- Scaffold Go module `github.com/alexandrmotologa/distillforge`.
- Create YAML job spec parser and schema validator with default templates.
- Write unit tests for configuration parsing, boundary validation, and default overrides.

### Task 2: Persistence and storage layer
- Implement lightweight SQLite database driver using `modernc.org/sqlite`.
- Define tables for jobs, datasets, synthetic samples, training metrics, benchmarks, and artifacts.
- Write unit tests verifying migrations, sample indexing, and query pagination.

### Task 3: Synthetic data generator with Evol-Instruct
- Implement Teacher client interface supporting OpenAI, Anthropic, Ollama, and an offline simulation provider for deterministic tests.
- Implement Evol-Instruct mutation algorithms:
  - In-depth evolution: adds domain constraints and technical conditions.
  - In-breadth evolution: creates varied rephrasings and parallel scenarios.
  - Adversarial evolution: injects boundary cases and noisy inputs.
  - Chain-of-thought evolution: incorporates step-by-step reasoning tokens.
- Implement concurrent generation worker pool with rate limit retry backoff.
- Write unit tests for mutation transforms, prompt template expansions, and batch execution.

### Task 4: Multi-criteria quality filter and deduplication
- Implement LLM-as-a-judge evaluator calculating adherence, correctness, and reasoning scores.
- Implement JSON Schema validator enforcing strict output formats.
- Implement N-gram Jaccard and TF-IDF cosine similarity deduplication.
- Implement export formats: Alpaca, ShareGPT, and Hugging Face JSONL.
- Write unit tests for filter acceptance thresholds, deduplication pruning, and schema rejections.

### Task 5: LoRA fine-tuning execution engine
- Create Python training script (`python/train_lora.py`) using PyTorch, Hugging Face PEFT, TRL, and bitsandbytes.
- Implement Apple Silicon MLX adapter script (`python/train_mlx.py`).
- Create Go process monitor that executes training subprocesses, captures stdout and stderr, parses loss metrics, and streams real-time telemetry over SSE.
- Write tests for training runner arguments, metric parsers, and error handling.

### Task 6: Head-to-head benchmarking and FinOps calculator
- Implement evaluation test harness executing holdout datasets against Teacher and Student models.
- Profile latency: time to first token (TTFT), inter-token duration, and tokens per second throughput.
- Implement FinOps calculator: compute monthly dollar savings, latency reduction ratio, and cost per million tokens.
- Write unit tests for benchmark metric aggregation and financial projection equations.

### Task 7: Quantization and model packaging
- Implement GGUF conversion and quantization invoking `llama.cpp` tools or Python scripts.
- Implement automated Ollama `Modelfile` generator configured with model prompt templates, system rules, and stop tokens.
- Write unit tests verifying Modelfile syntax and GGUF quantization command generation.

### Task 8: Embedded React 19 web studio
- Create modern Vite + React 19 application in `web/` with Tailwind CSS.
- Build interactive views:
  - Dashboard: active projects, model runs, dataset stats.
  - New Distillation wizard: configure YAML spec, preview Evol-Instruct mutations.
  - Live Training monitor: real-time loss and perplexity curves, VRAM tracking, epoch timer.
  - Head-to-Head diff viewer: side-by-side Teacher vs Student comparison on holdout test set.
  - FinOps savings calculator: dynamic sliders for traffic volume and hardware amortization.
  - Interactive test playground: send live prompts to the distilled model.
- Build production assets and embed into Go binary via `embed.FS`.

### Task 9: CLI entrypoint and orchestrator commands
- Build unified CLI with subcommands:
  - `distillforge studio`: launches the web studio server.
  - `distillforge generate`: runs synthetic data generation.
  - `distillforge audit`: runs quality filtering and deduplication.
  - `distillforge train`: launches fine-tuning pipeline.
  - `distillforge evaluate`: runs head-to-head benchmarking.
  - `distillforge export`: quantizes model and creates Ollama Modelfile.
  - `distillforge run`: executes the entire pipeline end to end in one command.

### Task 10: Complete real-world domain examples
- Support ticket classifier: 12-class customer support routing with confidence score.
- Invoice entity extractor: structured JSON extraction with nested items and validation schema.
- SQL query generator: natural language to SQL translation across relational schemas.

### Task 11: CI workflows and documentation
- GitHub Actions CI workflow for Go build, tests, and linter.
- Humanized documentation: `README.md`, `docs/ARCHITECTURE.md`, `docs/CLI_REFERENCE.md`, `docs/FINOPS_GUIDE.md`.

### Task 12: Usability, file ingestion and Human-in-the-Loop (HITL) curation
- Drag-and-drop file ingestion endpoint (`POST /api/upload`) supporting `.jsonl` and `.csv` seed datasets with schema validation and key discovery.
- Inline dataset curation in Studio: edit prompt/response pairs, add manual golden seeds, bulk accept/reject, and on-demand re-scoring with LLM-as-a-judge.
- Split-screen dual streaming playground executing Teacher and Student in parallel with side-by-side token diff and latency comparisons.

### Task 13: Synthetic data diversity and AI research analytics
- Implement analytics engine computing token length distributions, Type-Token Ratio (TTR) lexical richness, and strategy balance.
- Add category distribution checks to detect and prevent mode collapse during prompt mutation.
- Implement training anomaly and loss spike guard detecting divergent loss jumps and perplexity stalls.

### Task 14: Enterprise FinOps executive report and recipe advisor
- Executive FinOps report generator producing self-contained Markdown and HTML audit reports with latency percentiles (p50, p90, p95, p99), financial ROI, and emissions models.
- Hardware and distillation recipe advisor recommending optimal LoRA rank `r`, learning rates, batch sizes, and quantization targets based on deployment hardware.

### Task 15: Local OpenAI-compatible serving and multi-quantization benchmarking
- Build OpenAI-compatible chat completions server (`POST /v1/chat/completions`) and CLI command `distillforge serve` for zero-code client migration.
- Multi-quantization benchmarking harness comparing `Q4_K_M`, `Q5_K_M`, and `Q8_0` memory footprint, disk size, and token latency.
