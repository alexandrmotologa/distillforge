# Enterprise LLM FinOps and rightsizing guide

This guide outlines the financial modeling, compute cost amortization, and token accounting principles used by distillforge to quantify savings when transitioning repetitive domain tasks from frontier cloud APIs to small language models (SLMs).

## The unit economics of foundation model API calls

Foundation model providers charge per token. A typical customer service, invoice extraction, or classification prompt uses 200 to 500 input tokens and generates 50 to 150 output tokens.

Example rates for frontier models:
- GPT-4o: $5.00 per million input tokens, $15.00 per million output tokens.
- Claude 3.5 Sonnet: $3.00 per million input tokens, $15.00 per million output tokens.

For an organization processing 1,000,000 requests monthly with an average of 350 prompt tokens and 120 completion tokens:

```
Monthly Input Tokens:  1,000,000 * 350 = 350,000,000 tokens (350M)
Monthly Output Tokens: 1,000,000 * 120 = 120,000,000 tokens (120M)

Input Cost:  350 * $5.00  = $1,750
Output Cost: 120 * $15.00 = $1,800
Total Monthly API Cost:    $3,550 ($42,600 / year)
```

## The self-hosted small language model economics

A fine-tuned 1B to 3B model quantized to Q4_K_M occupies between 0.8GB and 2.2GB of memory. It can serve requests with sub-20ms latency on:
- AWS EC2 `g5.xlarge` (1x NVIDIA A10G 24GB): ~$1.00 per hour on-demand, or ~$110 to $140 per month with 1-year reserved instances or spot capacity.
- RunPod / Lambda Cloud Serverless: ~$0.20 per hour of active GPU runtime.
- On-premises hardware (Apple Silicon Mac Mini M4, 16GB unified memory): ~$599 one-time purchase, amortized to ~$25 monthly over 24 months with negligible electricity draw.

Comparing recurring expenditures:
- Frontier Cloud API: $3,550 monthly
- Self-hosted 1B SLM: $120 monthly
- Net Monthly Savings: $3,430 (96.6% reduction)
- Annual Recaptured Budget: $41,160

## Latency and throughput considerations

Cloud API roundtrips incur internet transit overhead, TLS handshakes, and foundation model queue delays:
- Cloud TTFT (Time to First Token): 600ms to 2,500ms
- On-prem 1B SLM TTFT (Ollama / vLLM): 12ms to 25ms
- Throughput: 65 to 110 tokens per second on consumer hardware

Replacing external API calls with local SLMs eliminates network roundtrip jitter and ensures consistent response times during peak production traffic.

## Data sovereignty and governance

Routing sensitive proprietary records (such as medical summaries, internal database schemas, or customer invoice details) outside company boundaries introduces compliance risks under GDPR, HIPAA, and SOC 2.

A localized SLM processes inferences strictly inside virtual private clouds or on-premises servers, removing data egress vectors entirely.
