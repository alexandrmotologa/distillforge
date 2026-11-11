# CLI reference

distillforge provides a unified command line interface for executing each stage of the distillation lifecycle independently or as an automated end-to-end pipeline.

## Global options

All subcommands inherit these flags:
- `-c, --config string`: Path to YAML configuration file (default: `distillforge.yaml`).
- `--db string`: Path to SQLite database file for job and metric persistence (default: `distillforge.db`).
- `-h, --help`: Displays help information for the command.

## Subcommands

### distillforge studio

Starts the embedded Web Studio and REST API server.

```bash
distillforge studio --port 8080
```

Flags:
- `-p, --port int`: Port to listen on (default: `8080`).

Endpoints exposed:
- `http://localhost:8080`: Interactive Web Studio dashboard.
- `http://localhost:8080/api/health`: Health status.
- `http://localhost:8080/api/jobs`: Job list and execution triggers.
- `http://localhost:8080/api/jobs/{id}/events`: Server-Sent Events stream.

### distillforge generate

Expands seed examples into synthetic training pairs using Evol-Instruct.

```bash
distillforge generate --config distillforge.yaml --output ./data/dataset_raw.jsonl
```

Flags:
- `-o, --output string`: Output file path for raw generated JSONL (default: `dataset_raw.jsonl`).

### distillforge audit

Applies LLM-as-a-judge scoring, strict JSON schema validation, and 3-gram deduplication to clean the synthetic dataset.

```bash
distillforge audit --config distillforge.yaml --input ./data/dataset_raw.jsonl --output ./data/dataset_clean.jsonl
```

Flags:
- `-i, --input string`: Path to input raw dataset (default: `dataset_raw.jsonl`).
- `-o, --output string`: Path to output clean dataset (default: `dataset_clean.jsonl`).

### distillforge train

Launches LoRA or QLoRA fine-tuning on the clean dataset.

```bash
distillforge train --config distillforge.yaml --dataset ./data/dataset_clean.jsonl --output ./checkpoints/best
```

Flags:
- `-d, --dataset string`: Path to clean dataset (default: `dataset_clean.jsonl`).
- `-o, --output string`: Checkpoint destination directory (default: `./checkpoints/best`).
- `--simulation`: Forces simulated training execution for testing without GPU dependencies.

### distillforge evaluate

Runs holdout test cases through both Teacher and fine-tuned Student, producing accuracy, latency, and FinOps metrics.

```bash
distillforge evaluate --config distillforge.yaml --dataset ./data/dataset_clean.jsonl
```

Flags:
- `-d, --dataset string`: Path to holdout test set (default: `dataset_clean.jsonl`).

### distillforge export

Merges the trained LoRA adapter, quantizes weights into GGUF format, and generates a production Ollama Modelfile.

```bash
distillforge export --config distillforge.yaml --checkpoint ./checkpoints/best --output ./dist
```

Flags:
- `--checkpoint string`: Path to saved LoRA adapter checkpoint (default: `./checkpoints/best`).
- `-o, --output string`: Output directory for `.gguf` binary and `Modelfile` (default: `./dist`).
- `--simulation`: Simulates quantization in environments without llama.cpp binaries.

### distillforge run

Executes the entire 5-stage pipeline sequentially: generation, audit, fine-tuning, evaluation, and export.

```bash
distillforge run --config distillforge.yaml
```

Flags:
- `--simulation`: Runs mock generation, training, and export for testing.
