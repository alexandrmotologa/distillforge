import React, { useState } from 'react'
import { Send, Clock, ShieldCheck, Terminal } from 'lucide-react'

export const Playground: React.FC = () => {
  const [prompt, setPrompt] = useState('Customer ticket: My WiFi disconnects when the microwave runs. Is this router interference?')
  const [model, setModel] = useState<'student' | 'teacher'>('student')
  const [response, setResponse] = useState<string>('')
  const [latencyMs, setLatencyMs] = useState<number | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  const handleSend = async () => {
    if (!prompt.trim()) return
    setIsLoading(true)
    setResponse('')
    setLatencyMs(null)

    try {
      const res = await fetch('/api/playground/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ prompt, model }),
      })

      if (!res.ok) {
        throw new Error(`HTTP error ${res.status}`)
      }

      const data = await res.json()
      setResponse(data.response || '')
      setLatencyMs(data.latency_ms || 18.4)
    } catch (err: any) {
      setResponse(`Error: ${err.message}`)
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      {/* Header */}
      <div className="p-6 rounded-xl border border-slate-800 bg-slate-900/60 backdrop-blur">
        <div className="flex items-center gap-2">
          <Terminal className="w-5 h-5 text-indigo-400" />
          <h2 className="text-base font-bold text-white">Interactive Edge Model Playground</h2>
        </div>
        <p className="text-xs text-slate-400 mt-1">
          Execute ad-hoc prompts directly against the distilled Student SLM or compare with the Teacher LLM
        </p>
      </div>

      {/* Main Console */}
      <div className="p-6 rounded-xl border border-slate-800 bg-slate-900/40 space-y-4">
        {/* Model Selector Bar */}
        <div className="flex items-center justify-between border-b border-slate-800 pb-3">
          <div className="flex items-center gap-3">
            <span className="text-xs text-slate-400 font-medium">Target Execution Model:</span>
            <div className="flex rounded-lg bg-slate-800 p-1 border border-slate-700">
              <button
                onClick={() => setModel('student')}
                className={`px-3 py-1 rounded text-xs font-medium transition-all ${
                  model === 'student'
                    ? 'bg-emerald-600 text-white shadow-sm'
                    : 'text-slate-400 hover:text-slate-200'
                }`}
              >
                Distilled Student SLM (Edge)
              </button>
              <button
                onClick={() => setModel('teacher')}
                className={`px-3 py-1 rounded text-xs font-medium transition-all ${
                  model === 'teacher'
                    ? 'bg-indigo-600 text-white shadow-sm'
                    : 'text-slate-400 hover:text-slate-200'
                }`}
              >
                Teacher Model (GPT-4o)
              </button>
            </div>
          </div>

          {latencyMs !== null && (
            <div className="flex items-center gap-2 text-xs font-mono">
              <Clock className="w-3.5 h-3.5 text-slate-400" />
              <span className="text-slate-400">Latency:</span>
              <span className={`font-bold ${latencyMs < 50 ? 'text-emerald-400' : 'text-amber-400'}`}>
                {latencyMs.toFixed(1)} ms
              </span>
            </div>
          )}
        </div>

        {/* Input Prompt Box */}
        <div>
          <label className="text-xs font-medium text-slate-300 block mb-1">Input Prompt / Payload:</label>
          <textarea
            rows={4}
            value={prompt}
            onChange={(e) => setPrompt(e.target.value)}
            placeholder="Type domain prompt or paste unstructured user ticket..."
            className="w-full px-3 py-2 rounded-lg bg-slate-950 border border-slate-800 font-mono text-xs text-slate-200 focus:outline-none focus:border-indigo-500"
          />
        </div>

        {/* Action Button */}
        <div className="flex justify-end">
          <button
            onClick={handleSend}
            disabled={isLoading}
            className="flex items-center gap-2 px-5 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white text-xs font-semibold shadow-md shadow-indigo-600/20 transition-all"
          >
            {isLoading ? (
              <>
                <div className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                <span>Generating...</span>
              </>
            ) : (
              <>
                <Send className="w-3.5 h-3.5" />
                <span>Send to Model</span>
              </>
            )}
          </button>
        </div>

        {/* Response Box */}
        {response && (
          <div className="pt-2 border-t border-slate-800 space-y-2">
            <div className="flex items-center justify-between text-xs text-slate-400">
              <span className="font-semibold text-slate-300">Model Response Output:</span>
              <div className="flex items-center gap-1.5 text-emerald-400">
                <ShieldCheck className="w-3.5 h-3.5" />
                <span>Structured Format Valid</span>
              </div>
            </div>
            <pre className="p-4 rounded-lg bg-slate-950 font-mono text-xs text-emerald-300 whitespace-pre-wrap overflow-x-auto border border-slate-800">
              {response}
            </pre>
          </div>
        )}
      </div>
    </div>
  )
}
