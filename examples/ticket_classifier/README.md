# Support ticket classifier distillation

This example demonstrates distilling GPT-4o into a 1B parameter small language model (Llama 3.2 1B Instruct) to classify incoming enterprise support tickets into 12 operational categories.

## Domain requirements

Customer support operations handle thousands of tickets daily. Routing each ticket through frontier cloud models (such as GPT-4o) costs substantial monthly expenditure and introduces 800ms to 2000ms latency.

A fine-tuned 1B parameter model runs locally with under 20ms latency while adhering to a strict JSON schema:
- `category`: One of 12 predefined classification tags.
- `confidence`: Numeric confidence score between 0.0 and 1.0.
- `requires_escalation`: Boolean indicator for urgent technical escalation.
- `sla_hours`: Target response SLA in hours.

## Running the distillation pipeline

Execute the full pipeline from the repository root:

```bash
# Run end-to-end distillation with the ticket classifier config
distillforge run --config ./examples/ticket_classifier/distillforge.yaml
```

Or execute stage by stage:

```bash
# 1. Generate synthetic dataset from seed tickets
distillforge generate --config ./examples/ticket_classifier/distillforge.yaml --output ./data/tickets_raw.jsonl

# 2. Audit and filter with LLM-as-a-judge
distillforge audit --config ./examples/ticket_classifier/distillforge.yaml --input ./data/tickets_raw.jsonl --output ./data/tickets_clean.jsonl

# 3. Fine-tune with LoRA
distillforge train --config ./examples/ticket_classifier/distillforge.yaml --dataset ./data/tickets_clean.jsonl --output ./checkpoints/ticket_classifier

# 4. Evaluate against holdout test cases
distillforge evaluate --config ./examples/ticket_classifier/distillforge.yaml --dataset ./data/tickets_clean.jsonl

# 5. Export to quantized GGUF and Ollama Modelfile
distillforge export --config ./examples/ticket_classifier/distillforge.yaml --checkpoint ./checkpoints/ticket_classifier --output ./dist/ticket_classifier
```

## FinOps economics

At 1,500,000 monthly requests (180 prompt tokens, 65 completion tokens):
- GPT-4o cloud API cost: ~$2,812 per month ($33,744 annually)
- Self-hosted 1B model (AWS g5.xlarge or on-prem mini PC): ~$110 per month ($1,320 annually)
- Net annual financial savings: $32,424 (96.1% reduction)
- Average latency drop: From 850ms down to 18.2ms (46x faster)
