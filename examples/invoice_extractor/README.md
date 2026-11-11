# Invoice entity extractor distillation

This example demonstrates distilling Claude 3.5 Sonnet into a 1.5B parameter model (Qwen 2.5 1.5B Instruct) to parse unstructured invoice OCR text into strictly validated JSON schemas.

## Problem statement

Financial document processing often routes invoices and receipts through heavy multimodal cloud foundation models. At 500,000 monthly invoices, paying frontier per-token pricing incurs steep operational costs and data compliance concerns under banking regulations.

Distilling to a dedicated 1.5B model preserves 100% on-premises data isolation while achieving sub-25ms parsing times with guaranteed JSON Schema compliance.

## Quick start

```bash
# Run full pipeline with invoice configuration
distillforge run --config ./examples/invoice_extractor/distillforge.yaml
```
