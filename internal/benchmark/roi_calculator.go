package benchmark

import (
	"github.com/alexandrmotologa/distillforge/internal/config"
)

// ROIProjection contains calculated cloud savings and operational FinOps metrics.
type ROIProjection struct {
	MonthlyRequests            int64   `json:"monthly_requests"`
	TotalMonthlyInputTokens    int64   `json:"total_monthly_input_tokens"`
	TotalMonthlyOutputTokens   int64   `json:"total_monthly_output_tokens"`
	TeacherMonthlyCostUSD      float64 `json:"teacher_monthly_cost_usd"`
	StudentMonthlyCostUSD      float64 `json:"student_monthly_cost_usd"`
	MonthlySavingsUSD          float64 `json:"monthly_savings_usd"`
	AnnualSavingsUSD           float64 `json:"annual_savings_usd"`
	CostReductionPercentage    float64 `json:"cost_reduction_percentage"`
	EstimatedLatencyReduction  float64 `json:"estimated_latency_reduction_percent"`
	EstimatedKgCO2SavedPerYear float64 `json:"estimated_kg_co2_saved_per_year"`
}

// CalculateROI computes financial and carbon savings based on FinOps configurations.
func CalculateROI(cfg *config.FinOpsConfig, teacherAvgLatencyMs, studentAvgLatencyMs float64) ROIProjection {
	inTokens := cfg.MonthlyRequests * int64(cfg.AvgPromptTokens)
	outTokens := cfg.MonthlyRequests * int64(cfg.AvgCompletionTokens)

	teacherInCost := (float64(inTokens) / 1_000_000.0) * cfg.TeacherCostPerMIn
	teacherOutCost := (float64(outTokens) / 1_000_000.0) * cfg.TeacherCostPerMOut
	teacherTotalMonthly := teacherInCost + teacherOutCost

	studentTotalMonthly := cfg.SelfHostedMonthlyInfraCost
	monthlySavings := teacherTotalMonthly - studentTotalMonthly
	if monthlySavings < 0 {
		monthlySavings = 0
	}
	annualSavings := monthlySavings * 12.0

	var costReductionPct float64
	if teacherTotalMonthly > 0 {
		costReductionPct = (monthlySavings / teacherTotalMonthly) * 100.0
	}

	var latencyReductionPct float64
	if teacherAvgLatencyMs > 0 && studentAvgLatencyMs > 0 {
		latencyReductionPct = ((teacherAvgLatencyMs - studentAvgLatencyMs) / teacherAvgLatencyMs) * 100.0
		if latencyReductionPct < 0 {
			latencyReductionPct = 0
		}
	} else {
		// Standard default assumption: 800ms API roundtrip vs 18ms edge SLM
		latencyReductionPct = 95.0
	}

	// Approximate carbon footprint: 1M cloud frontier tokens ~= 0.4 kg CO2e, on-prem SLM quantized ~= 0.04 kg CO2e
	totalMTokensYearly := (float64(inTokens+outTokens) / 1_000_000.0) * 12.0
	kgCO2Saved := totalMTokensYearly * (0.40 - 0.04)

	return ROIProjection{
		MonthlyRequests:            cfg.MonthlyRequests,
		TotalMonthlyInputTokens:    inTokens,
		TotalMonthlyOutputTokens:   outTokens,
		TeacherMonthlyCostUSD:      teacherTotalMonthly,
		StudentMonthlyCostUSD:      studentTotalMonthly,
		MonthlySavingsUSD:          monthlySavings,
		AnnualSavingsUSD:           annualSavings,
		CostReductionPercentage:    costReductionPct,
		EstimatedLatencyReduction:  latencyReductionPct,
		EstimatedKgCO2SavedPerYear: kgCO2Saved,
	}
}
