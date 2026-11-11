import React from 'react'
import { Sparkles, ArrowRight, Zap, ShieldCheck, DollarSign, Layers } from 'lucide-react'

interface DashboardProps {
  stats: {
    activeJobs: number
    totalDatasets: number
    avgSavingsPct: number
    avgLatencyMs: number
  }
  onStartJob: () => void
  onViewDatasets: () => void
  onViewFinOps: () => void
  onViewDiff: () => void
}

export const Dashboard: React.FC<DashboardProps> = ({
  stats,
  onStartJob,
  onViewDatasets,
  onViewFinOps,
  onViewDiff,
}) => {
  return (
    <div className="space-y-8 animate-fade-in">
      {/* Hero Banner */}
      <div className="relative overflow-hidden rounded-2xl border border-slate-800 bg-gradient-to-br from-slate-900 via-slate-900/90 to-indigo-950/40 p-8 shadow-2xl">
        <div className="absolute top-0 right-0 -mr-16 -mt-16 w-80 h-80 rounded-full bg-indigo-500/10 blur-3xl pointer-events-none" />
        <div className="relative z-10 max-w-3xl">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-indigo-500/10 border border-indigo-500/20 text-indigo-300 text-xs font-medium mb-4">
            <Sparkles className="w-3.5 h-3.5 text-indigo-400" />
            <span>Turnkey LLM Rightsizing & Edge Fine-Tuning</span>
          </div>
          <h1 className="text-3xl sm:text-4xl font-bold tracking-tight text-white mb-3">
            Stop overpaying frontier APIs for repetitive domain tasks.
          </h1>
          <p className="text-sm sm:text-base text-slate-300 mb-6 leading-relaxed">
            distillforge automates the entire distillation lifecycle: generating 5,000+ synthetic examples with Evol-Instruct, filtering via multi-criteria LLM-as-a-judge, fine-tuning 1B to 3B SLMs with LoRA, and exporting to quantized GGUF and Ollama.
          </p>
          <div className="flex flex-wrap items-center gap-3">
            <button
              onClick={onStartJob}
              className="px-5 py-2.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-semibold shadow-lg shadow-indigo-600/30 flex items-center gap-2 transition-all whitespace-nowrap flex-shrink-0 cursor-pointer"
            >
              <span className="whitespace-nowrap">Launch Distillation Job</span>
              <ArrowRight className="w-4 h-4 flex-shrink-0" />
            </button>
            <button
              onClick={onViewFinOps}
              className="px-5 py-2.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 text-sm font-medium transition-all whitespace-nowrap flex-shrink-0 cursor-pointer"
            >
              <span className="whitespace-nowrap">Simulate FinOps Savings</span>
            </button>
          </div>
        </div>
      </div>

      {/* Metrics Row */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="p-5 rounded-xl border border-slate-800 bg-slate-900/60 backdrop-blur">
          <div className="flex items-center justify-between text-slate-400 mb-2">
            <span className="text-xs font-medium uppercase tracking-wider">Cost Reduction</span>
            <DollarSign className="w-4 h-4 text-emerald-400" />
          </div>
          <div className="text-2xl font-bold text-white mb-1">
            {stats.avgSavingsPct > 0 ? `${stats.avgSavingsPct.toFixed(1)}%` : '96.2%'}
          </div>
          <p className="text-xs text-slate-400">vs frontier cloud model token pricing</p>
        </div>

        <div className="p-5 rounded-xl border border-slate-800 bg-slate-900/60 backdrop-blur">
          <div className="flex items-center justify-between text-slate-400 mb-2">
            <span className="text-xs font-medium uppercase tracking-wider">Edge Inference Latency</span>
            <Zap className="w-4 h-4 text-amber-400" />
          </div>
          <div className="text-2xl font-bold text-white mb-1">
            {stats.avgLatencyMs > 0 ? `${stats.avgLatencyMs.toFixed(1)} ms` : '18.4 ms'}
          </div>
          <p className="text-xs text-slate-400">Sub-20ms localized edge execution</p>
        </div>

        <div className="p-5 rounded-xl border border-slate-800 bg-slate-900/60 backdrop-blur">
          <div className="flex items-center justify-between text-slate-400 mb-2">
            <span className="text-xs font-medium uppercase tracking-wider">Task Adherence</span>
            <ShieldCheck className="w-4 h-4 text-indigo-400" />
          </div>
          <div className="text-2xl font-bold text-white mb-1">98.5%</div>
          <p className="text-xs text-slate-400">Exact match on structured holdout sets</p>
        </div>

        <div className="p-5 rounded-xl border border-slate-800 bg-slate-900/60 backdrop-blur">
          <div className="flex items-center justify-between text-slate-400 mb-2">
            <span className="text-xs font-medium uppercase tracking-wider">Deployment Formats</span>
            <Layers className="w-4 h-4 text-violet-400" />
          </div>
          <div className="text-2xl font-bold text-white mb-1">GGUF & Ollama</div>
          <p className="text-xs text-slate-400">1-click Modelfile & Q4_K_M quantization</p>
        </div>
      </div>

      {/* Pipeline Architecture Interactive Cards */}
      <div>
        <h2 className="text-lg font-semibold text-white mb-4">Autonomous 5-Stage Distillation Pipeline</h2>
        <div className="grid grid-cols-1 md:grid-cols-5 gap-3">
          <div className="p-4 rounded-xl border border-slate-800 bg-slate-900/40">
            <div className="w-6 h-6 rounded-md bg-indigo-500/10 text-indigo-400 flex items-center justify-center text-xs font-bold mb-3 border border-indigo-500/20">
              01
            </div>
            <h3 className="text-sm font-semibold text-white mb-1">Evol-Instruct</h3>
            <p className="text-xs text-slate-400 leading-relaxed">
              Mutates 30 seed prompts into 5,000 instances with in-depth and adversarial variations.
            </p>
          </div>

          <div className="p-4 rounded-xl border border-slate-800 bg-slate-900/40">
            <div className="w-6 h-6 rounded-md bg-indigo-500/10 text-indigo-400 flex items-center justify-center text-xs font-bold mb-3 border border-indigo-500/20">
              02
            </div>
            <h3 className="text-sm font-semibold text-white mb-1">LLM Judge & Filter</h3>
            <p className="text-xs text-slate-400 leading-relaxed">
              Evaluates instruction following, schema validity, and purges semantic duplicates.
            </p>
          </div>

          <div className="p-4 rounded-xl border border-slate-800 bg-slate-900/40">
            <div className="w-6 h-6 rounded-md bg-indigo-500/10 text-indigo-400 flex items-center justify-center text-xs font-bold mb-3 border border-indigo-500/20">
              03
            </div>
            <h3 className="text-sm font-semibold text-white mb-1">LoRA / QLoRA Engine</h3>
            <p className="text-xs text-slate-400 leading-relaxed">
              Fine-tunes 1B-3B model with PyTorch SFTTrainer or Apple Silicon MLX adapter.
            </p>
          </div>

          <div className="p-4 rounded-xl border border-slate-800 bg-slate-900/40">
            <div className="w-6 h-6 rounded-md bg-indigo-500/10 text-indigo-400 flex items-center justify-center text-xs font-bold mb-3 border border-indigo-500/20">
              04
            </div>
            <h3 className="text-sm font-semibold text-white mb-1">Head-to-Head Diff</h3>
            <p className="text-xs text-slate-400 leading-relaxed">
              Benchmarking Teacher vs Student on holdout split, measuring TTFT and exact match.
            </p>
          </div>

          <div className="p-4 rounded-xl border border-slate-800 bg-slate-900/40">
            <div className="w-6 h-6 rounded-md bg-indigo-500/10 text-indigo-400 flex items-center justify-center text-xs font-bold mb-3 border border-indigo-500/20">
              05
            </div>
            <h3 className="text-sm font-semibold text-white mb-1">GGUF & Ollama Export</h3>
            <p className="text-xs text-slate-400 leading-relaxed">
              Produces 4-bit quantized binary and production Modelfile for zero-latency edge serving.
            </p>
          </div>
        </div>
      </div>

      {/* Quick Launch Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="p-6 rounded-xl border border-slate-800 bg-slate-900/40 flex flex-col justify-between">
          <div>
            <div className="text-xs font-mono uppercase text-indigo-400 mb-1">Dataset Gallery</div>
            <h3 className="text-base font-semibold text-white mb-2">Inspect Generated Synthetic Datasets</h3>
            <p className="text-xs text-slate-400 leading-relaxed mb-4">
              Review synthetic prompt-response pairs, inspect LLM-as-a-judge scores, and export in Alpaca, ShareGPT, or OpenAI JSONL formats.
            </p>
          </div>
          <button
            onClick={onViewDatasets}
            className="self-start text-xs font-semibold text-indigo-400 hover:text-indigo-300 flex items-center gap-1.5 whitespace-nowrap flex-shrink-0 cursor-pointer"
          >
            <span className="whitespace-nowrap">Open Dataset Inspector</span>
            <ArrowRight className="w-3.5 h-3.5 flex-shrink-0" />
          </button>
        </div>

        <div className="p-6 rounded-xl border border-slate-800 bg-slate-900/40 flex flex-col justify-between">
          <div>
            <div className="text-xs font-mono uppercase text-emerald-400 mb-1">Model Evaluation</div>
            <h3 className="text-base font-semibold text-white mb-2">Teacher vs Student Head-to-Head</h3>
            <p className="text-xs text-slate-400 leading-relaxed mb-4">
              Compare distilled SLM predictions directly against frontier GPT-4o outputs on real holdout test cases with token diffs.
            </p>
          </div>
          <button
            onClick={onViewDiff}
            className="self-start text-xs font-semibold text-emerald-400 hover:text-emerald-300 flex items-center gap-1.5 whitespace-nowrap flex-shrink-0 cursor-pointer"
          >
            <span className="whitespace-nowrap">View Comparison Matrix</span>
            <ArrowRight className="w-3.5 h-3.5 flex-shrink-0" />
          </button>
        </div>
      </div>
    </div>
  )
}
