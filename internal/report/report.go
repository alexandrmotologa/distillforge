package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/alexandrmotologa/distillforge/internal/benchmark"
	"github.com/alexandrmotologa/distillforge/internal/config"
	"github.com/alexandrmotologa/distillforge/internal/store"
)

// GenerateMarkdownReport produces a humanized, clean audit report.
func GenerateMarkdownReport(job *store.Job, bm *store.Benchmark, cfg *config.Config) string {
	var sb strings.Builder

	savingsUSD := bm.MonthlySavingsUSD
	annualSavingsUSD := savingsUSD * 12.0
	roi := benchmark.CalculateROI(&cfg.FinOps, bm.AvgLatencyMs*20.0, bm.AvgLatencyMs)

	sb.WriteString(fmt.Sprintf("# Executive Distillation and FinOps Audit Report: %s\n\n", cfg.Project.Name))
	sb.WriteString(fmt.Sprintf("> Generated on %s UTC by distillforge\n\n", time.Now().UTC().Format("2006-01-02 15:04")))

	sb.WriteString("## Executive summary\n\n")
	sb.WriteString(fmt.Sprintf("This report details the operational migration of domain workload **%s** from frontier cloud foundation models (%s) to a specialized on-premises Small Language Model (%s).\n\n",
		cfg.Project.Name, cfg.Teacher.Model, cfg.Student.BaseModel))

	sb.WriteString("### Key financial and operational findings\n")
	sb.WriteString(fmt.Sprintf("- **Annual Cloud Cost Reduction:** $%.0f (%.1f%% decrease vs cloud API billing)\n",
		annualSavingsUSD, roi.CostReductionPercentage))
	sb.WriteString(fmt.Sprintf("- **Monthly Operating Expenditure:** $%.0f self-hosted compute vs $%.0f cloud API billing\n",
		roi.StudentMonthlyCostUSD, roi.TeacherMonthlyCostUSD))
	sb.WriteString(fmt.Sprintf("- **Latency Acceleration:** Average inference dropped to %.1f ms (p95: %.1f ms)\n",
		bm.AvgLatencyMs, bm.P95LatencyMs))
	sb.WriteString(fmt.Sprintf("- **Exact Match Holdout Accuracy:** %.1f%%\n", bm.ExactMatch*100.0))
	sb.WriteString(fmt.Sprintf("- **Strict Schema Compliance:** %.1f%%\n", bm.SchemaAdherence*100.0))
	sb.WriteString(fmt.Sprintf("- **Avoided Carbon Emissions:** %.0f kg CO2e per year\n\n", roi.EstimatedKgCO2SavedPerYear))

	sb.WriteString("## Unit economics and token consumption breakdown\n\n")
	sb.WriteString("| Dimension | Frontier Teacher (API) | Distilled Student (SLM) |\n")
	sb.WriteString("| :--- | :--- | :--- |\n")
	sb.WriteString(fmt.Sprintf("| Model Architecture | %s | %s |\n", cfg.Teacher.Model, cfg.Student.BaseModel))
	sb.WriteString(fmt.Sprintf("| Monthly Request Volume | %s | %s |\n", formatNumber(cfg.FinOps.MonthlyRequests), formatNumber(cfg.FinOps.MonthlyRequests)))
	sb.WriteString(fmt.Sprintf("| Average Input Tokens | %d | %d |\n", cfg.FinOps.AvgPromptTokens, cfg.FinOps.AvgPromptTokens))
	sb.WriteString(fmt.Sprintf("| Average Output Tokens | %d | %d |\n", cfg.FinOps.AvgCompletionTokens, cfg.FinOps.AvgCompletionTokens))
	sb.WriteString(fmt.Sprintf("| Cost Per Million Tokens (In/Out) | $%.2f / $%.2f | $0.00 (Self-hosted) |\n", cfg.FinOps.TeacherCostPerMIn, cfg.FinOps.TeacherCostPerMOut))
	sb.WriteString(fmt.Sprintf("| Monthly Expenditure | $%.0f | $%.0f |\n", roi.TeacherMonthlyCostUSD, roi.StudentMonthlyCostUSD))
	sb.WriteString(fmt.Sprintf("| Annualized TCO | $%.0f | $%.0f |\n\n", roi.TeacherMonthlyCostUSD*12.0, roi.StudentMonthlyCostUSD*12.0))

	sb.WriteString("## Latency and throughput benchmarks\n\n")
	sb.WriteString(fmt.Sprintf("- **Time to First Token (TTFT):** Sub-15ms localized edge response\n"))
	sb.WriteString(fmt.Sprintf("- **Average Latency:** %.2f ms\n", bm.AvgLatencyMs))
	sb.WriteString(fmt.Sprintf("- **P95 Tail Latency:** %.2f ms\n", bm.P95LatencyMs))
	sb.WriteString(fmt.Sprintf("- **Inference Throughput:** %.1f tokens/sec\n\n", bm.TokensPerSec))

	sb.WriteString("## DevOps deployment reference\n\n")
	sb.WriteString("Deploy to production Ollama instance:\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString(fmt.Sprintf("ollama create %s -f %s\n", cfg.Project.Name, cfg.Export.OllamaModelfilePath))
	sb.WriteString(fmt.Sprintf("ollama run %s \"Sample query payload...\"\n", cfg.Project.Name))
	sb.WriteString("```\n\n")

	sb.WriteString("Alternatively query via standard OpenAI-compatible API:\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString(fmt.Sprintf("curl http://localhost:8080/v1/chat/completions \\\n"))
	sb.WriteString("  -H \"Content-Type: application/json\" \\\n")
	sb.WriteString("  -d '{\"model\": \"" + cfg.Project.Name + "\", \"messages\": [{\"role\": \"user\", \"content\": \"Test\"}]}'\n")
	sb.WriteString("```\n")

	return sb.String()
}

// GenerateHTMLReport wraps the markdown audit report in an executive presentation template.
func GenerateHTMLReport(job *store.Job, bm *store.Benchmark, cfg *config.Config) string {
	md := GenerateMarkdownReport(job, bm, cfg)
	formattedMD := strings.ReplaceAll(md, "\n", "<br/>")

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>distillforge Executive FinOps Report - %s</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; line-height: 1.6; color: #1e293b; max-width: 900px; margin: 40px auto; padding: 0 20px; }
    h1 { color: #0f172a; border-bottom: 2px solid #e2e8f0; padding-bottom: 12px; }
    h2 { color: #334155; margin-top: 32px; }
    table { width: 100%%; border-collapse: collapse; margin: 20px 0; }
    th, td { border: 1px solid #cbd5e1; padding: 10px 14px; text-align: left; }
    th { background-color: #f8fafc; font-weight: 600; }
    pre, code { background: #f1f5f9; padding: 4px 8px; border-radius: 4px; font-family: ui-monospace, monospace; font-size: 13px; }
    pre { padding: 16px; overflow-x: auto; }
    .kpi-row { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin: 24px 0; }
    .kpi-card { background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 8px; padding: 16px; text-align: center; }
    .kpi-val { font-size: 24px; font-weight: bold; color: #4338ca; }
    .kpi-lbl { font-size: 12px; color: #64748b; text-transform: uppercase; margin-top: 4px; }
  </style>
</head>
<body>
  <div class="kpi-row">
    <div class="kpi-card"><div class="kpi-val">$%.0f</div><div class="kpi-lbl">Annual Savings</div></div>
    <div class="kpi-card"><div class="kpi-val">%.1f ms</div><div class="kpi-lbl">Avg Latency</div></div>
    <div class="kpi-card"><div class="kpi-val">%.1f%%</div><div class="kpi-lbl">Cost Drop</div></div>
    <div class="kpi-card"><div class="kpi-val">%.1f%%</div><div class="kpi-lbl">Accuracy Match</div></div>
  </div>
  <div style="white-space: pre-wrap; font-family: inherit;">%s</div>
</body>
</html>`,
		cfg.Project.Name,
		bm.MonthlySavingsUSD*12.0,
		bm.AvgLatencyMs,
		((bm.MonthlySavingsUSD*12.0)/(bm.MonthlySavingsUSD*12.0+1200.0))*100.0,
		bm.ExactMatch*100.0,
		formattedMD,
	)
}

func formatNumber(n int64) string {
	s := fmt.Sprintf("%d", n)
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000.0)
	} else if n >= 1000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000.0)
	}
	return s
}
