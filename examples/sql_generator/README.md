# Natural language to SQL distillation

This example demonstrates distilling GPT-4o into a 1.7B parameter small model (SmolLM2 1.7B Instruct) for fast, secure enterprise Text-to-SQL generation.

## Key benefits

Internal analytics queries frequently contain sensitive database table names, column schemas, and internal identifiers. Routing ad-hoc employee queries to external third-party cloud APIs presents security and data leak risks.

Distilling to an on-premises 1.7B model provides:
- Zero data egress: Database schema stays within internal network boundaries.
- Low latency: Queries generate in 16ms to 22ms.
- High precision: Fine-tuned directly on internal dialect conventions.

## Running the distillation

```bash
# Execute end-to-end distillation
distillforge run --config ./examples/sql_generator/distillforge.yaml
```
