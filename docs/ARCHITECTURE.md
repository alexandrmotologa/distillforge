# System architecture

distillforge automates the transfer of task-specific capabilities from large cloud foundation models (Teacher) into lightweight 1B to 3B parameter models (Student) deployed on edge or local hardware.

## System components

The system consists of five primary layers coordinated by a compiled Go CLI and an embedded React 19 studio:

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

## Synthetic data generation and Evol-Instruct

Seed datasets often contain only 20 to 50 genuine enterprise examples. Fine-tuning an SLM requires 1,000 to 10,000 pairs covering edge cases.

The Evol-Instruct engine expands seed examples across four mutation dimensions:
1. In-depth evolution: Adds realistic operational constraints, compliance rules, or multi-step conditions to the prompt.
2. In-breadth evolution: Creates parallel scenarios covering distinct user tones, phrasing styles, and topic variations within the same task domain.
3. Adversarial evolution: Introduces ambiguous inputs, typographical noise, or edge cases that frequently trip foundation models.
4. Chain-of-thought evolution: Enforces internal reasoning steps before generating the final structured answer.

## Multi-criteria quality filter

Generated data is filtered before reaching the training pipeline:
1. Length and structure heuristics: Drops empty responses, truncated text, or refusal statements ("As an AI language model...").
2. JSON Schema enforcement: If a schema is configured, the response must parse as valid JSON matching all required properties.
3. Deduplication: Calculates token 3-gram Jaccard similarity against all previously accepted prompts. Pairs with similarity exceeding the threshold are discarded.
4. LLM-as-a-judge: Evaluates instruction following, format cleanliness, and groundedness. Only samples exceeding the quality threshold pass into the training set.

## Training engine and telemetry

The fine-tuning engine supports multiple execution backends:
- QLoRA (4-bit NF4 bitsandbytes): Runs on consumer GPUs with 4GB to 8GB of VRAM.
- Standard LoRA (FP16 / BF16): Runs on datacenter GPUs or Apple Silicon unified memory.
- MLX Adapter: Native training on M-series Apple Silicon chips.
- CPU fallback mode: Enables execution on machines without dedicated accelerators.

During training, stdout is monitored for metric events. Loss, perplexity, learning rate, epoch, and memory footprint are parsed and broadcast via Server-Sent Events (SSE) to the web studio in real time.

## Evaluation and FinOps calculator

Holdout test cases are executed across both the Teacher model and the fine-tuned Student model. The profiler tracks:
- Exact match rate on ground truth targets.
- Schema adherence rate.
- Time to first token (TTFT) and throughput (tokens per second).
- Monthly and annual cloud expenditure reduction calculated against input and output token rates.
