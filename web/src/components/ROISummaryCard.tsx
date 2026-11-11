import React, { useState } from 'react'
import { DollarSign, TrendingDown, Leaf, ShieldAlert, Cpu } from 'lucide-react'

export const ROISummaryCard: React.FC = () => {
  const [monthlyRequests, setMonthlyRequests] = useState(1000000)
  const [promptTokens, setPromptTokens] = useState(350)
  const [completionTokens, setCompletionTokens] = useState(120)
  const [teacherCostIn, setTeacherCostIn] = useState(5.0) // $5 / 1M input tokens
  const [teacherCostOut, setTeacherCostOut] = useState(15.0) // $15 / 1M output tokens
  const [selfHostedCost, setSelfHostedCost] = useState(120.0) // $120/mo for compute instance

  // FinOps Calculations
  const monthlyInTokens = (monthlyRequests * promptTokens) / 1000000
  const monthlyOutTokens = (monthlyRequests * completionTokens) / 1000000

  const teacherMonthlyCost = monthlyInTokens * teacherCostIn + monthlyOutTokens * teacherCostOut
  const studentMonthlyCost = selfHostedCost
  const monthlySavings = Math.max(0, teacherMonthlyCost - studentMonthlyCost)
  const annualSavings = monthlySavings * 12
  const costReductionPct = teacherMonthlyCost > 0 ? (monthlySavings / teacherMonthlyCost) * 100 : 0
  const kgCO2SavedPerYear = (monthlyInTokens + monthlyOutTokens) * 12 * 0.36

  return (
    <div className="space-y-6">
      {/* FinOps Header */}
      <div className="p-6 rounded-xl border border-slate-800 bg-slate-900/60 backdrop-blur">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <div className="flex items-center gap-2">
              <DollarSign className="w-5 h-5 text-emerald-400" />
              <h2 className="text-base font-bold text-white">Interactive LLM FinOps ROI Simulator</h2>
            </div>
            <p className="text-xs text-slate-400 mt-1">
              Project financial savings and carbon footprint reduction by replacing cloud APIs with on-prem SLMs
            </p>
          </div>

          <div className="text-right">
            <div className="text-xs text-slate-400 uppercase tracking-wider">Projected Annual Savings</div>
            <div className="text-2xl font-bold font-mono text-emerald-400">
              ${annualSavings.toLocaleString('en-US', { maximumFractionDigits: 0 })}
            </div>
          </div>
        </div>
      </div>

      {/* FinOps KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="p-5 rounded-xl border border-slate-800 bg-slate-900/40">
          <div className="flex items-center justify-between text-xs text-slate-400 mb-2">
            <span>Teacher Monthly Bill</span>
            <ShieldAlert className="w-4 h-4 text-rose-400" />
          </div>
          <div className="text-xl font-bold text-white mb-1 font-mono">
            ${teacherMonthlyCost.toLocaleString('en-US', { maximumFractionDigits: 0 })}
          </div>
          <p className="text-[11px] text-slate-400">Frontier API token consumption</p>
        </div>

        <div className="p-5 rounded-xl border border-slate-800 bg-slate-900/40">
          <div className="flex items-center justify-between text-xs text-slate-400 mb-2">
            <span>Student On-Prem Compute</span>
            <Cpu className="w-4 h-4 text-indigo-400" />
          </div>
          <div className="text-xl font-bold text-white mb-1 font-mono">
            ${studentMonthlyCost.toFixed(0)}
          </div>
          <p className="text-[11px] text-slate-400">Fixed instance / hardware cost</p>
        </div>

        <div className="p-5 rounded-xl border border-slate-800 bg-slate-900/40">
          <div className="flex items-center justify-between text-xs text-slate-400 mb-2">
            <span>Monthly Cost Drop</span>
            <TrendingDown className="w-4 h-4 text-emerald-400" />
          </div>
          <div className="text-xl font-bold text-emerald-400 mb-1 font-mono">
            {costReductionPct.toFixed(1)}%
          </div>
          <p className="text-[11px] text-slate-400">Over 90% budget recaptured</p>
        </div>

        <div className="p-5 rounded-xl border border-slate-800 bg-slate-900/40">
          <div className="flex items-center justify-between text-xs text-slate-400 mb-2">
            <span>CO2 Avoided Yearly</span>
            <Leaf className="w-4 h-4 text-emerald-400" />
          </div>
          <div className="text-xl font-bold text-white mb-1 font-mono">
            {kgCO2SavedPerYear.toFixed(0)} kg
          </div>
          <p className="text-[11px] text-slate-400">Reduced datacenter energy footprint</p>
        </div>
      </div>

      {/* Sliders and Configuration */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="p-6 rounded-xl border border-slate-800 bg-slate-900/60 backdrop-blur space-y-5">
          <h3 className="text-sm font-semibold text-white">Workload Parameters</h3>

          <div>
            <div className="flex justify-between text-xs text-slate-300 mb-1">
              <span>Monthly Production Requests</span>
              <span className="font-mono text-indigo-400 font-bold">
                {monthlyRequests.toLocaleString('en-US')}
              </span>
            </div>
            <input
              type="range"
              min="100000"
              max="10000000"
              step="100000"
              value={monthlyRequests}
              onChange={(e) => setMonthlyRequests(Number(e.target.value))}
              className="w-full accent-indigo-500"
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <div className="flex justify-between text-xs text-slate-300 mb-1">
                <span>Avg Prompt Tokens</span>
                <span className="font-mono text-slate-400">{promptTokens}</span>
              </div>
              <input
                type="number"
                value={promptTokens}
                onChange={(e) => setPromptTokens(Number(e.target.value))}
                className="w-full px-3 py-1.5 rounded-lg bg-slate-800 border border-slate-700 text-xs text-white"
              />
            </div>

            <div>
              <div className="flex justify-between text-xs text-slate-300 mb-1">
                <span>Avg Output Tokens</span>
                <span className="font-mono text-slate-400">{completionTokens}</span>
              </div>
              <input
                type="number"
                value={completionTokens}
                onChange={(e) => setCompletionTokens(Number(e.target.value))}
                className="w-full px-3 py-1.5 rounded-lg bg-slate-800 border border-slate-700 text-xs text-white"
              />
            </div>
          </div>
        </div>

        <div className="p-6 rounded-xl border border-slate-800 bg-slate-900/60 backdrop-blur space-y-5">
          <h3 className="text-sm font-semibold text-white">Rate & Infrastructure Economics</h3>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="text-xs text-slate-300 block mb-1">Teacher Input Cost ($/1M)</label>
              <input
                type="number"
                step="0.5"
                value={teacherCostIn}
                onChange={(e) => setTeacherCostIn(Number(e.target.value))}
                className="w-full px-3 py-1.5 rounded-lg bg-slate-800 border border-slate-700 text-xs text-white"
              />
            </div>

            <div>
              <label className="text-xs text-slate-300 block mb-1">Teacher Output Cost ($/1M)</label>
              <input
                type="number"
                step="0.5"
                value={teacherCostOut}
                onChange={(e) => setTeacherCostOut(Number(e.target.value))}
                className="w-full px-3 py-1.5 rounded-lg bg-slate-800 border border-slate-700 text-xs text-white"
              />
            </div>
          </div>

          <div>
            <label className="text-xs text-slate-300 block mb-1">Self-Hosted Monthly Infra Cost ($/mo)</label>
            <input
              type="number"
              step="10"
              value={selfHostedCost}
              onChange={(e) => setSelfHostedCost(Number(e.target.value))}
              className="w-full px-3 py-1.5 rounded-lg bg-slate-800 border border-slate-700 text-xs text-white"
            />
            <p className="text-[11px] text-slate-400 mt-1">
              Covers AWS EC2 g5.xlarge (A10G GPU), RunPod serverless, or amortized Apple Silicon Mac Mini.
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}
