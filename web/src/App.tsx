import { useState, useEffect } from 'react'
import { Header } from './components/Header'
import { Dashboard } from './components/Dashboard'
import { NewDistillation } from './components/NewDistillation'
import { TrainingCockpit } from './components/TrainingCockpit'
import { DatasetInspector } from './components/DatasetInspector'
import { HeadToHeadDiff } from './components/HeadToHeadDiff'
import { ROISummaryCard } from './components/ROISummaryCard'
import { Playground } from './components/Playground'

export default function App() {
  const [activeTab, setActiveTab] = useState('dashboard')
  const [pipelineRunning, setPipelineRunning] = useState(false)
  const [currentJob, setCurrentJob] = useState<any>(null)
  const [metrics, setMetrics] = useState<any[]>([])
  const [logs, setLogs] = useState<string[]>([])
  const [samples] = useState<any[]>([
    {
      id: 'sample-1',
      prompt: 'Customer reports fiber optic ONT box has red LOS light after thunderstorm.',
      response: '{"category": "Network Outage", "confidence": 0.98, "requires_escalation": false, "sla_hours": 4}',
      strategy: 'in_depth',
      judge_score: 0.96,
      is_accepted: true,
    },
    {
      id: 'sample-2',
      prompt: 'Cancel subscription renewal and refund the $49 charge from yesterday.',
      response: '{"category": "Billing", "confidence": 0.99, "requires_escalation": true, "sla_hours": 2}',
      strategy: 'adversarial',
      judge_score: 0.95,
      is_accepted: true,
    },
    {
      id: 'sample-3',
      prompt: 'Forgot 2FA password and switched to a new phone number.',
      response: '{"category": "Authentication", "confidence": 0.97, "requires_escalation": true, "sla_hours": 4}',
      strategy: 'in_breadth',
      judge_score: 0.94,
      is_accepted: true,
    },
    {
      id: 'sample-4',
      prompt: 'App crashes immediately on startup on Android 14 after latest update.',
      response: '{"category": "Technical Defect", "confidence": 0.96, "requires_escalation": false, "sla_hours": 8}',
      strategy: 'cot',
      judge_score: 0.98,
      is_accepted: true,
    },
  ])

  // Fetch initial system status
  useEffect(() => {
    fetch('/api/jobs')
      .then((res) => (res.ok ? res.json() : []))
      .then((jobs) => {
        if (jobs && jobs.length > 0) {
          const latest = jobs[0]
          setCurrentJob(latest)
          if (latest.status === 'running') {
            setPipelineRunning(true)
            subscribeToJobEvents(latest.id)
          }
        }
      })
      .catch(() => {})
  }, [])

  const subscribeToJobEvents = (jobId: string) => {
    const sse = new EventSource(`/api/jobs/${jobId}/events`)

    sse.addEventListener('progress', (e: any) => {
      try {
        const data = JSON.parse(e.data)
        setCurrentJob((prev: any) => ({ ...prev, progress: data.progress, status: 'running' }))
        setLogs((prev) => [...prev, `[PROGRESS] Stage: ${data.stage} (${data.done}/${data.total})`])
      } catch (_) {}
    })

    sse.addEventListener('training_metric', (e: any) => {
      try {
        const metric = JSON.parse(e.data)
        setMetrics((prev) => [...prev, metric])
        setLogs((prev) => [
          ...prev,
          `[TRAIN] Step ${metric.step} | Loss: ${metric.loss} | PPL: ${metric.perplexity} | VRAM: ${metric.vram_mb} MB`,
        ])
      } catch (_) {}
    })

    sse.addEventListener('log', (e: any) => {
      setLogs((prev) => [...prev, e.data])
    })

    sse.addEventListener('completed', (e: any) => {
      try {
        const data = JSON.parse(e.data)
        setCurrentJob((prev: any) => ({ ...prev, progress: 100, status: 'completed' }))
        setPipelineRunning(false)
        setLogs((prev) => [...prev, `[SUCCESS] Distillation complete! Model exported to ${data.gguf_path}`])
      } catch (_) {}
      sse.close()
    })

    sse.addEventListener('error', (e: any) => {
      try {
        const data = JSON.parse(e.data)
        setCurrentJob((prev: any) => ({ ...prev, status: 'failed', error_message: data.error_message }))
        setLogs((prev) => [...prev, `[ERROR] ${data.error_message}`])
      } catch (_) {}
      setPipelineRunning(false)
      sse.close()
    })
  }

  const handleLaunchJob = async (configPayload: any) => {
    setPipelineRunning(true)
    setMetrics([])
    setLogs(['[INIT] Submitting declarative distillation spec to orchestrator...'])
    setActiveTab('training')

    try {
      const res = await fetch('/api/jobs/run', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ config: configPayload }),
      })

      if (!res.ok) {
        throw new Error(`Failed to create job (${res.status})`)
      }

      const job = await res.json()
      setCurrentJob(job)
      setLogs((prev) => [...prev, `[STARTED] Distillation Job ${job.id} dispatched.`])
      subscribeToJobEvents(job.id)
    } catch (err: any) {
      setPipelineRunning(false)
      setLogs((prev) => [...prev, `[FAILED] ${err.message}`])
    }
  }

  return (
    <div className="min-h-screen bg-[#090d16] text-slate-100 flex flex-col font-sans">
      <Header
        activeTab={activeTab}
        setActiveTab={setActiveTab}
        pipelineRunning={pipelineRunning}
      />

      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {activeTab === 'dashboard' && (
          <Dashboard
            stats={{
              activeJobs: pipelineRunning ? 1 : 0,
              totalDatasets: 3,
              avgSavingsPct: 96.2,
              avgLatencyMs: 18.4,
            }}
            onStartJob={() => setActiveTab('new')}
            onViewDatasets={() => setActiveTab('datasets')}
            onViewFinOps={() => setActiveTab('finops')}
            onViewDiff={() => setActiveTab('diff')}
          />
        )}

        {activeTab === 'new' && (
          <NewDistillation
            onLaunchJob={handleLaunchJob}
            isLaunching={pipelineRunning}
          />
        )}

        {activeTab === 'training' && (
          <TrainingCockpit
            currentJob={currentJob}
            metrics={metrics}
            logs={logs}
          />
        )}

        {activeTab === 'datasets' && (
          <DatasetInspector samples={samples} />
        )}

        {activeTab === 'diff' && (
          <HeadToHeadDiff comparisons={[]} />
        )}

        {activeTab === 'finops' && (
          <ROISummaryCard />
        )}

        {activeTab === 'playground' && (
          <Playground />
        )}
      </main>

      <footer className="border-t border-slate-800/80 bg-[#0c1220]/60 py-6 text-center text-xs text-slate-500">
        <div className="max-w-7xl mx-auto px-4 flex flex-wrap items-center justify-between gap-4">
          <div className="flex items-center gap-2">
            <span className="font-semibold text-slate-400">distillforge</span>
            <span>&bull;</span>
            <span>Enterprise LLM-to-SLM Distillation & Edge Fine-Tuning Studio</span>
          </div>
          <div className="font-mono text-slate-500">
            Target: Sub-20ms On-Prem Inference & 90%+ Cost Reduction
          </div>
        </div>
      </footer>
    </div>
  )
}
