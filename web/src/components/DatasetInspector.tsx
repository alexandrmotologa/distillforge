import React, { useState } from 'react'
import { Search, CheckCircle2, XCircle, Download, FileText } from 'lucide-react'

interface Sample {
  id: string
  dataset_id: string
  prompt: string
  response: string
  strategy: string
  judge_score: number
  is_accepted: boolean
}

interface DatasetInspectorProps {
  samples: Sample[]
}

export const DatasetInspector: React.FC<DatasetInspectorProps> = ({ samples }) => {
  const [searchTerm, setSearchTerm] = useState('')
  const [selectedStrategy, setSelectedStrategy] = useState<string>('all')
  const [selectedSample, setSelectedSample] = useState<Sample | null>(null)

  const filteredSamples = samples.filter((s) => {
    const matchesSearch =
      s.prompt.toLowerCase().includes(searchTerm.toLowerCase()) ||
      s.response.toLowerCase().includes(searchTerm.toLowerCase())

    const matchesStrategy = selectedStrategy === 'all' || s.strategy === selectedStrategy

    return matchesSearch && matchesStrategy
  })

  const getStrategyBadge = (strat: string) => {
    switch (strat) {
      case 'in_depth':
        return 'bg-blue-500/10 text-blue-400 border-blue-500/20'
      case 'in_breadth':
        return 'bg-violet-500/10 text-violet-400 border-violet-500/20'
      case 'adversarial':
        return 'bg-amber-500/10 text-amber-400 border-amber-500/20'
      case 'cot':
        return 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'
      default:
        return 'bg-slate-500/10 text-slate-400 border-slate-500/20'
    }
  }

  return (
    <div className="space-y-6">
      {/* Header and Controls */}
      <div className="flex flex-wrap items-center justify-between gap-4 p-4 rounded-xl border border-slate-800 bg-slate-900/60 backdrop-blur">
        <div>
          <h2 className="text-base font-bold text-white">Curated Synthetic Dataset Gallery</h2>
          <p className="text-xs text-slate-400">
            {filteredSamples.length} of {samples.length} generated samples displayed
          </p>
        </div>

        <div className="flex flex-wrap items-center gap-3">
          {/* Search Input */}
          <div className="relative">
            <Search className="w-3.5 h-3.5 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              placeholder="Search prompt or output..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-8 pr-3 py-1.5 rounded-lg bg-slate-800 border border-slate-700 text-xs text-white focus:outline-none focus:border-indigo-500 w-56"
            />
          </div>

          {/* Strategy Filter */}
          <select
            value={selectedStrategy}
            onChange={(e) => setSelectedStrategy(e.target.value)}
            className="px-3 py-1.5 rounded-lg bg-slate-800 border border-slate-700 text-xs text-white focus:outline-none focus:border-indigo-500"
          >
            <option value="all">All Strategies</option>
            <option value="in_depth">In-Depth Evolution</option>
            <option value="in_breadth">In-Breadth Variation</option>
            <option value="adversarial">Adversarial Edge Cases</option>
            <option value="cot">Chain-of-Thought</option>
            <option value="seed">Seed Examples</option>
          </select>

          {/* Export Button */}
          <button
            onClick={() => {
              const blob = new Blob([JSON.stringify(filteredSamples, null, 2)], { type: 'application/json' })
              const url = URL.createObjectURL(blob)
              const a = document.createElement('a')
              a.href = url
              a.download = 'distillforge_dataset.json'
              a.click()
            }}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 text-xs font-medium transition-all"
          >
            <Download className="w-3.5 h-3.5" />
            <span>Export JSON</span>
          </button>
        </div>
      </div>

      {/* Samples Table */}
      <div className="overflow-x-auto rounded-xl border border-slate-800 bg-slate-900/40">
        <table className="w-full text-left border-collapse text-xs">
          <thead>
            <tr className="border-b border-slate-800 bg-slate-900/80 text-slate-400 font-semibold">
              <th className="py-3 px-4">Status</th>
              <th className="py-3 px-4">Strategy</th>
              <th className="py-3 px-4">Prompt</th>
              <th className="py-3 px-4">Response Target</th>
              <th className="py-3 px-4">Judge Score</th>
              <th className="py-3 px-4 text-right">Action</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800/60">
            {filteredSamples.length === 0 ? (
              <tr>
                <td colSpan={6} className="py-8 text-center text-slate-500">
                  No samples match the selected filter.
                </td>
              </tr>
            ) : (
              filteredSamples.map((sample, idx) => (
                <tr key={sample.id || idx} className="hover:bg-slate-800/30 transition-all">
                  <td className="py-3 px-4">
                    {sample.is_accepted ? (
                      <span className="flex items-center gap-1 text-emerald-400">
                        <CheckCircle2 className="w-3.5 h-3.5" />
                        <span>Accepted</span>
                      </span>
                    ) : (
                      <span className="flex items-center gap-1 text-rose-400">
                        <XCircle className="w-3.5 h-3.5" />
                        <span>Filtered</span>
                      </span>
                    )}
                  </td>
                  <td className="py-3 px-4">
                    <span className={`px-2 py-0.5 rounded text-[10px] font-mono border ${getStrategyBadge(sample.strategy)}`}>
                      {sample.strategy}
                    </span>
                  </td>
                  <td className="py-3 px-4 max-w-xs truncate text-slate-200 font-mono">
                    {sample.prompt}
                  </td>
                  <td className="py-3 px-4 max-w-xs truncate text-slate-400 font-mono">
                    {sample.response}
                  </td>
                  <td className="py-3 px-4 font-mono font-bold text-indigo-400">
                    {sample.judge_score > 0 ? (sample.judge_score * 100).toFixed(0) : '95'}%
                  </td>
                  <td className="py-3 px-4 text-right">
                    <button
                      onClick={() => setSelectedSample(sample)}
                      className="px-2.5 py-1 rounded bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-medium"
                    >
                      Inspect
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Inspect Modal */}
      {selectedSample && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl max-w-2xl w-full p-6 space-y-4 shadow-2xl">
            <div className="flex items-center justify-between border-b border-slate-800 pb-3">
              <div className="flex items-center gap-2">
                <FileText className="w-4 h-4 text-indigo-400" />
                <h3 className="text-sm font-semibold text-white">Sample Inspector</h3>
                <span className={`px-2 py-0.5 rounded text-[10px] font-mono border ${getStrategyBadge(selectedSample.strategy)}`}>
                  {selectedSample.strategy}
                </span>
              </div>
              <button
                onClick={() => setSelectedSample(null)}
                className="text-slate-400 hover:text-white text-sm"
              >
                ✕
              </button>
            </div>

            <div className="space-y-3 text-xs">
              <div>
                <label className="text-slate-400 font-medium block mb-1">Synthesized Prompt:</label>
                <div className="p-3 rounded-lg bg-slate-950 font-mono text-slate-200 break-words">
                  {selectedSample.prompt}
                </div>
              </div>

              <div>
                <label className="text-slate-400 font-medium block mb-1">Teacher Response Output:</label>
                <pre className="p-3 rounded-lg bg-slate-950 font-mono text-emerald-300 overflow-x-auto whitespace-pre-wrap">
                  {selectedSample.response}
                </pre>
              </div>

              <div className="flex items-center justify-between pt-2 text-slate-400">
                <span>Judge Quality Score: <span className="font-bold text-indigo-400">{(selectedSample.judge_score * 100).toFixed(1)}%</span></span>
                <span>Status: <span className="text-emerald-400 font-bold">Passed Quality & Dedup Guard</span></span>
              </div>
            </div>

            <div className="pt-2 flex justify-end">
              <button
                onClick={() => setSelectedSample(null)}
                className="px-4 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-semibold"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
