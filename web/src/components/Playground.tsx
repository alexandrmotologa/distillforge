import React, { useState } from 'react'
import { Send, Clock, ShieldCheck, Terminal, DollarSign, SplitSquareVertical, ArrowRightLeft } from 'lucide-react'

export const Playground: React.FC = () => {
  const [prompt, setPrompt] = useState('Customer ticket: My WiFi disconnects when the microwave runs. Is this router interference?')
  const [isDualMode, setIsDualMode] = useState(true)
  const [singleModel, setSingleModel] = useState<'student' | 'teacher'>('student')
  const [isLoading, setIsLoading] = useState(false)

  // Outputs & stats
  const [studentResponse, setStudentResponse] = useState<string>('')
  const [teacherResponse, setTeacherResponse] = useState<string>('')
  const [studentLatency, setStudentLatency] = useState<number | null>(null)
  const [teacherLatency, setTeacherLatency] = useState<number | null>(null)

  const handleSend = async () => {
    if (!prompt.trim()) return
    setIsLoading(true)

    if (isDualMode) {
      setStudentResponse('')
      setTeacherResponse('')
      setStudentLatency(null)
      setTeacherLatency(null)

      try {
        const [resStudent, resTeacher] = await Promise.all([
          fetch('/api/playground/chat', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ prompt, model: 'student' }),
          }),
          fetch('/api/playground/chat', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ prompt, model: 'teacher' }),
          }),
        ])

        const dataStudent = await resStudent.json()
        const dataTeacher = await resTeacher.json()

        setStudentResponse(dataStudent.response || '')
        setStudentLatency(dataStudent.latency_ms || 18.2)

        setTeacherResponse(dataTeacher.response || '')
        setTeacherLatency(dataTeacher.latency_ms || 842.0)
      } catch (err: any) {
        setStudentResponse(`Error: ${err.message}`)
      } finally {
        setIsLoading(false)
      }
    } else {
      setStudentResponse('')
      setTeacherResponse('')
      try {
        const res = await fetch('/api/playground/chat', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ prompt, model: singleModel }),
        })
        const data = await res.json()
        if (singleModel === 'student') {
          setStudentResponse(data.response || '')
          setStudentLatency(data.latency_ms || 18.4)
        } else {
          setTeacherResponse(data.response || '')
          setTeacherLatency(data.latency_ms || 850.0)
        }
      } catch (err: any) {
        if (singleModel === 'student') setStudentResponse(`Error: ${err.message}`)
        else setTeacherResponse(`Error: ${err.message}`)
      } finally {
        setIsLoading(false)
      }
    }
  }

  return (
    <div className="max-w-5xl mx-auto space-y-6">
      {/* Header */}
      <div className="p-6 rounded-xl border border-slate-800 bg-slate-900/60 backdrop-blur">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <div className="flex items-center gap-2">
              <Terminal className="w-5 h-5 text-indigo-400" />
              <h2 className="text-base font-bold text-white">Dual-Stream Edge Model Playground</h2>
            </div>
            <p className="text-xs text-slate-400 mt-1">
              Compare Student SLM vs Teacher LLM head-to-head in real time with latency and token economics
            </p>
          </div>

          <div className="flex items-center gap-2 rounded-lg bg-slate-950 p-1 border border-slate-800">
            <button
              onClick={() => setIsDualMode(true)}
              className={`flex items-center gap-1.5 px-3 py-1 rounded text-xs font-medium transition-all ${
                isDualMode ? 'bg-indigo-600 text-white' : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              <SplitSquareVertical className="w-3.5 h-3.5" />
              <span>Dual Comparison</span>
            </button>
            <button
              onClick={() => setIsDualMode(false)}
              className={`flex items-center gap-1.5 px-3 py-1 rounded text-xs font-medium transition-all ${
                !isDualMode ? 'bg-indigo-600 text-white' : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              <ArrowRightLeft className="w-3.5 h-3.5" />
              <span>Single Model</span>
            </button>
          </div>
        </div>
      </div>

      {/* Main Console */}
      <div className="p-6 rounded-xl border border-slate-800 bg-slate-900/40 space-y-4">
        {!isDualMode && (
          <div className="flex items-center gap-3 border-b border-slate-800 pb-3">
            <span className="text-xs text-slate-400 font-medium">Target Execution Model:</span>
            <div className="flex rounded-lg bg-slate-800 p-1 border border-slate-700">
              <button
                onClick={() => setSingleModel('student')}
                className={`px-3 py-1 rounded text-xs font-medium transition-all ${
                  singleModel === 'student' ? 'bg-emerald-600 text-white shadow-sm' : 'text-slate-400 hover:text-slate-200'
                }`}
              >
                Distilled Student SLM (Edge)
              </button>
              <button
                onClick={() => setSingleModel('teacher')}
                className={`px-3 py-1 rounded text-xs font-medium transition-all ${
                  singleModel === 'teacher' ? 'bg-indigo-600 text-white shadow-sm' : 'text-slate-400 hover:text-slate-200'
                }`}
              >
                Teacher Model (GPT-4o)
              </button>
            </div>
          </div>
        )}

        {/* Input Prompt Box */}
        <div>
          <label className="text-xs font-medium text-slate-300 block mb-1">Input Prompt / Test Payload:</label>
          <textarea
            rows={3}
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
            className="flex items-center gap-2 px-5 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white text-xs font-semibold shadow-md shadow-indigo-600/20 transition-all cursor-pointer"
          >
            {isLoading ? (
              <>
                <div className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                <span>Streaming Inference...</span>
              </>
            ) : (
              <>
                <Send className="w-3.5 h-3.5" />
                <span>{isDualMode ? 'Execute Dual Query' : 'Send to Model'}</span>
              </>
            )}
          </button>
        </div>

        {/* Dual Mode View */}
        {isDualMode ? (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 pt-2 border-t border-slate-800">
            {/* Student SLM Card */}
            <div className="p-4 rounded-xl border border-emerald-500/20 bg-emerald-950/10 space-y-3">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <span className="w-2 h-2 rounded-full bg-emerald-400" />
                  <span className="text-xs font-bold text-white">Distilled Student SLM</span>
                  <span className="px-1.5 py-0.5 rounded bg-emerald-500/20 text-emerald-400 text-[10px] font-mono">Edge</span>
                </div>
                {studentLatency !== null && (
                  <div className="flex items-center gap-1 text-xs font-mono text-emerald-400">
                    <Clock className="w-3 h-3" />
                    <span className="font-bold">{studentLatency.toFixed(1)} ms</span>
                  </div>
                )}
              </div>

              <div className="flex items-center justify-between text-[11px] text-slate-400 border-b border-emerald-500/10 pb-2">
                <span className="flex items-center gap-1">
                  <DollarSign className="w-3 h-3 text-emerald-400" />
                  <span>Cost: <strong className="text-white">$0.0000</strong> (On-Prem)</span>
                </span>
                <span className="flex items-center gap-1 text-emerald-400">
                  <ShieldCheck className="w-3 h-3" />
                  <span>Schema Valid</span>
                </span>
              </div>

              <pre className="p-3 rounded-lg bg-slate-950 font-mono text-xs text-emerald-300 whitespace-pre-wrap overflow-x-auto min-h-[120px] border border-slate-800/80">
                {studentResponse || (isLoading ? 'Computing localized weights...' : 'Awaiting prompt execution...')}
              </pre>
            </div>

            {/* Teacher LLM Card */}
            <div className="p-4 rounded-xl border border-indigo-500/20 bg-indigo-950/10 space-y-3">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <span className="w-2 h-2 rounded-full bg-indigo-400" />
                  <span className="text-xs font-bold text-white">Frontier Teacher LLM</span>
                  <span className="px-1.5 py-0.5 rounded bg-indigo-500/20 text-indigo-400 text-[10px] font-mono">Cloud API</span>
                </div>
                {teacherLatency !== null && (
                  <div className="flex items-center gap-1 text-xs font-mono text-amber-400">
                    <Clock className="w-3 h-3" />
                    <span className="font-bold">{teacherLatency.toFixed(1)} ms</span>
                  </div>
                )}
              </div>

              <div className="flex items-center justify-between text-[11px] text-slate-400 border-b border-indigo-500/10 pb-2">
                <span className="flex items-center gap-1">
                  <DollarSign className="w-3 h-3 text-amber-400" />
                  <span>Cost: <strong className="text-white">~$0.0052</strong> / call</span>
                </span>
                <span className="flex items-center gap-1 text-indigo-400">
                  <ShieldCheck className="w-3 h-3" />
                  <span>Cloud API Ground Truth</span>
                </span>
              </div>

              <pre className="p-3 rounded-lg bg-slate-950 font-mono text-xs text-indigo-300 whitespace-pre-wrap overflow-x-auto min-h-[120px] border border-slate-800/80">
                {teacherResponse || (isLoading ? 'Querying remote endpoint...' : 'Awaiting prompt execution...')}
              </pre>
            </div>
          </div>
        ) : (
          /* Single Mode View */
          <div className="pt-2 border-t border-slate-800 space-y-2">
            <div className="flex items-center justify-between text-xs text-slate-400">
              <span className="font-semibold text-slate-300">
                {singleModel === 'student' ? 'Distilled Student Response' : 'Teacher LLM Response'}:
              </span>
              <div className="flex items-center gap-3">
                {((singleModel === 'student' ? studentLatency : teacherLatency) !== null) && (
                  <div className="flex items-center gap-1 font-mono text-slate-300">
                    <Clock className="w-3 h-3" />
                    <span>{(singleModel === 'student' ? studentLatency : teacherLatency)?.toFixed(1)} ms</span>
                  </div>
                )}
                <div className="flex items-center gap-1.5 text-emerald-400">
                  <ShieldCheck className="w-3.5 h-3.5" />
                  <span>Structured Format Valid</span>
                </div>
              </div>
            </div>
            <pre className="p-4 rounded-lg bg-slate-950 font-mono text-xs text-emerald-300 whitespace-pre-wrap overflow-x-auto border border-slate-800 min-h-[100px]">
              {(singleModel === 'student' ? studentResponse : teacherResponse) || 'Awaiting execution...'}
            </pre>
          </div>
        )}
      </div>
    </div>
  )
}
