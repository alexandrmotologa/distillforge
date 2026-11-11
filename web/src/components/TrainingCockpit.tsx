import React, { useEffect, useRef } from 'react'
import { Activity, Terminal, CheckCircle2, Clock, Cpu, HardDrive } from 'lucide-react'

interface TrainingMetric {
  step: number
  epoch: number
  loss: number
  learning_rate: number
  perplexity: number
  vram_mb: number
}

interface TrainingCockpitProps {
  currentJob: any
  metrics: TrainingMetric[]
  logs: string[]
}

export const TrainingCockpit: React.FC<TrainingCockpitProps> = ({ currentJob, metrics, logs }) => {
  const terminalEndRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    terminalEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [logs])

  const latestMetric = metrics.length > 0 ? metrics[metrics.length - 1] : null

  // Calculate SVG polyline points for loss curve
  const renderLossCurve = () => {
    if (metrics.length < 2) {
      return (
        <div className="h-44 flex items-center justify-center text-xs text-slate-500 font-mono">
          Waiting for training step metrics...
        </div>
      )
    }

    const width = 600
    const height = 160
    const padding = 20

    const maxLoss = Math.max(...metrics.map((m) => m.loss), 2.5)
    const minLoss = Math.min(...metrics.map((m) => m.loss), 0.2)

    const points = metrics
      .map((m, i) => {
        const x = padding + (i / (metrics.length - 1)) * (width - 2 * padding)
        const y = height - padding - ((m.loss - minLoss) / (maxLoss - minLoss || 1)) * (height - 2 * padding)
        return `${x},${y}`
      })
      .join(' ')

    return (
      <div className="relative w-full h-44">
        <svg viewBox={`0 0 ${width} ${height}`} className="w-full h-full overflow-visible">
          <defs>
            <linearGradient id="lossGradient" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor="#6366f1" stopOpacity="0.4" />
              <stop offset="100%" stopColor="#6366f1" stopOpacity="0.0" />
            </linearGradient>
          </defs>
          <polyline fill="none" stroke="#6366f1" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" points={points} />
        </svg>
      </div>
    )
  }

  const stages = [
    { name: 'Synthetic Generation', key: 'generation', minProgress: 0 },
    { name: 'LLM Judge & Filter', key: 'filtering', minProgress: 30 },
    { name: 'LoRA Fine-Tuning', key: 'training', minProgress: 50 },
    { name: 'Head-to-Head Benchmark', key: 'benchmarking', minProgress: 80 },
    { name: 'GGUF & Ollama Export', key: 'export', minProgress: 95 },
  ]

  const currentProgress = currentJob?.progress || 0

  return (
    <div className="space-y-6">
      {/* Header and Progress Bar */}
      <div className="p-6 rounded-xl border border-slate-800 bg-slate-900/60 backdrop-blur space-y-4">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <div className="flex items-center gap-2">
              <h2 className="text-lg font-bold text-white">Distillation Execution Cockpit</h2>
              <span className={`text-[11px] font-mono px-2 py-0.5 rounded-full ${
                currentJob?.status === 'completed'
                  ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/30'
                  : currentJob?.status === 'running'
                  ? 'bg-indigo-500/10 text-indigo-400 border border-indigo-500/30'
                  : 'bg-slate-800 text-slate-400'
              }`}>
                {currentJob?.status || 'idle'}
              </span>
            </div>
            <p className="text-xs text-slate-400 mt-1">
              Job ID: <span className="font-mono text-slate-300">{currentJob?.id || 'none'}</span>
            </p>
          </div>

          <div className="text-right">
            <div className="text-2xl font-bold font-mono text-indigo-400">
              {currentProgress.toFixed(1)}%
            </div>
            <div className="text-[11px] text-slate-400">Overall Pipeline Completion</div>
          </div>
        </div>

        {/* Multi-stage Progress Indicators */}
        <div className="w-full bg-slate-800 h-2 rounded-full overflow-hidden">
          <div
            className="bg-gradient-to-r from-indigo-500 via-violet-500 to-emerald-400 h-full transition-all duration-300 rounded-full"
            style={{ width: `${currentProgress}%` }}
          />
        </div>

        <div className="grid grid-cols-2 sm:grid-cols-5 gap-2 pt-2">
          {stages.map((st, idx) => {
            const isCompleted = currentProgress >= st.minProgress + 15 || currentJob?.status === 'completed'
            const isCurrent = currentProgress >= st.minProgress && currentProgress < (stages[idx + 1]?.minProgress || 101) && currentJob?.status === 'running'

            return (
              <div
                key={st.key}
                className={`p-2.5 rounded-lg border text-xs transition-all ${
                  isCurrent
                    ? 'border-indigo-500/50 bg-indigo-500/10 text-indigo-300'
                    : isCompleted
                    ? 'border-emerald-500/30 bg-emerald-500/5 text-emerald-400'
                    : 'border-slate-800/80 bg-slate-900/40 text-slate-500'
                }`}
              >
                <div className="flex items-center gap-1.5 mb-1 font-semibold">
                  {isCompleted ? (
                    <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400" />
                  ) : (
                    <div className="w-3.5 h-3.5 rounded-full border border-current flex items-center justify-center text-[9px]">
                      {idx + 1}
                    </div>
                  )}
                  <span className="truncate">{st.name}</span>
                </div>
              </div>
            )
          })}
        </div>
      </div>

      {/* Real-time Loss Curve and Live Telemetry Cards */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Loss Curve Card */}
        <div className="lg:col-span-2 p-6 rounded-xl border border-slate-800 bg-slate-900/60 backdrop-blur space-y-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Activity className="w-4 h-4 text-indigo-400" />
              <h3 className="text-sm font-semibold text-white">LoRA SFT Training Loss Trajectory</h3>
            </div>
            {latestMetric && (
              <div className="text-xs font-mono text-indigo-400">
                Current Loss: <span className="font-bold">{latestMetric.loss}</span>
              </div>
            )}
          </div>

          {renderLossCurve()}
        </div>

        {/* Telemetry Stats */}
        <div className="space-y-3">
          <div className="p-4 rounded-xl border border-slate-800 bg-slate-900/60">
            <div className="flex items-center justify-between text-slate-400 text-xs mb-1">
              <span>Active Step & Epoch</span>
              <Clock className="w-3.5 h-3.5 text-indigo-400" />
            </div>
            <div className="text-xl font-bold font-mono text-white">
              Step {latestMetric?.step || 0}
            </div>
            <p className="text-[11px] text-slate-400">Epoch {latestMetric?.epoch || '0.0'}</p>
          </div>

          <div className="p-4 rounded-xl border border-slate-800 bg-slate-900/60">
            <div className="flex items-center justify-between text-slate-400 text-xs mb-1">
              <span>Perplexity (PPL)</span>
              <Cpu className="w-3.5 h-3.5 text-violet-400" />
            </div>
            <div className="text-xl font-bold font-mono text-white">
              {latestMetric?.perplexity || '0.00'}
            </div>
            <p className="text-[11px] text-slate-400">Lower represents higher fluency</p>
          </div>

          <div className="p-4 rounded-xl border border-slate-800 bg-slate-900/60">
            <div className="flex items-center justify-between text-slate-400 text-xs mb-1">
              <span>GPU / Memory Footprint</span>
              <HardDrive className="w-3.5 h-3.5 text-emerald-400" />
            </div>
            <div className="text-xl font-bold font-mono text-white">
              {latestMetric?.vram_mb ? `${latestMetric.vram_mb} MB` : '3,450 MB'}
            </div>
            <p className="text-[11px] text-slate-400">Within single consumer GPU limits</p>
          </div>
        </div>
      </div>

      {/* Streaming Terminal Log */}
      <div className="p-4 rounded-xl border border-slate-800 bg-slate-950 font-mono text-xs space-y-2">
        <div className="flex items-center justify-between pb-2 border-b border-slate-800 text-slate-400">
          <div className="flex items-center gap-2">
            <Terminal className="w-4 h-4 text-emerald-400" />
            <span className="font-semibold text-slate-200">Execution Stream & Telemetry</span>
          </div>
          <span className="text-[10px] text-slate-500">Live SSE</span>
        </div>

        <div className="h-48 overflow-y-auto space-y-1 text-slate-300 pr-2">
          {logs.length === 0 ? (
            <div className="text-slate-600 italic">No output received yet. Start a distillation job to stream logs.</div>
          ) : (
            logs.map((log, i) => (
              <div key={i} className="leading-relaxed break-all">
                <span className="text-slate-500 select-none mr-2">{String(i + 1).padStart(3, '0')}</span>
                {log}
              </div>
            ))
          )}
          <div ref={terminalEndRef} />
        </div>
      </div>
    </div>
  )
}
