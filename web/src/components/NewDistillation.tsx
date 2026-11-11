import React, { useState } from 'react'
import { Play, Sparkles, Sliders, Code2, AlertCircle } from 'lucide-react'

interface NewDistillationProps {
  onLaunchJob: (config: any) => Promise<void>
  isLaunching: boolean
}

export const NewDistillation: React.FC<NewDistillationProps> = ({ onLaunchJob, isLaunching }) => {
  const [projectName, setProjectName] = useState('ticket-classifier')
  const [description, setDescription] = useState('Distill GPT-4o into Llama-3.2-1B for real-time customer support routing')
  const [teacherProvider, setTeacherProvider] = useState('mock') // default mock for immediate testing
  const [teacherModel, setTeacherModel] = useState('gpt-4o')
  const [studentModel, setStudentModel] = useState('meta-llama/Llama-3.2-1B-Instruct')
  const [trainingMethod, setTrainingMethod] = useState('qlora')
  const [datasetSize, setDatasetSize] = useState(100)
  const [epochs, setEpochs] = useState(3)
  const [inDepthWeight, setInDepthWeight] = useState(0.4)
  const [inBreadthWeight, setInBreadthWeight] = useState(0.3)
  const [adversarialWeight, setAdversarialWeight] = useState(0.3)
  const [activeTab, setActiveTab] = useState<'form' | 'yaml'>('form')

  const yamlPreview = `version: "1"
project:
  name: "${projectName}"
  description: "${description}"

teacher:
  provider: "${teacherProvider}"
  model: "${teacherModel}"
  temperature: 0.7
  max_tokens: 1024

student:
  base_model: "${studentModel}"
  quantization_target: "q4_k_m"

seed_data:
  file: "./examples/ticket_classifier/data/raw_tickets.jsonl"
  input_key: "prompt"
  target_key: "response"

generation:
  target_dataset_size: ${datasetSize}
  concurrency: 4
  evolution_strategies:
    - name: "in_depth"
      weight: ${inDepthWeight}
    - name: "in_breadth"
      weight: ${inBreadthWeight}
    - name: "adversarial"
      weight: ${adversarialWeight}
  quality_threshold: 0.80

filter:
  dedup_threshold: 0.85
  min_response_chars: 10
  max_response_chars: 4000

training:
  method: "${trainingMethod}"
  r: 16
  lora_alpha: 32
  learning_rate: 0.0002
  epochs: ${epochs}
  batch_size: 4
  gradient_accumulation_steps: 4

export:
  formats: ["gguf", "ollama"]
  ollama_modelfile_path: "./dist/Modelfile"
`

  const handleLaunch = async () => {
    const configPayload = {
      version: '1',
      project: { name: projectName, description },
      teacher: { provider: teacherProvider, model: teacherModel, temperature: 0.7, max_tokens: 1024 },
      student: { base_model: studentModel, quantization_target: 'q4_k_m' },
      seed_data: {
        file: './examples/ticket_classifier/data/raw_tickets.jsonl',
        input_key: 'prompt',
        target_key: 'response',
      },
      generation: {
        target_dataset_size: datasetSize,
        concurrency: 4,
        evolution_strategies: [
          { name: 'in_depth', weight: inDepthWeight },
          { name: 'in_breadth', weight: inBreadthWeight },
          { name: 'adversarial', weight: adversarialWeight },
        ],
        quality_threshold: 0.80,
      },
      filter: { dedup_threshold: 0.85, min_response_chars: 10, max_response_chars: 4000 },
      training: {
        method: trainingMethod,
        r: 16,
        lora_alpha: 32,
        learning_rate: 0.0002,
        epochs: epochs,
        batch_size: 4,
        gradient_accumulation_steps: 4,
        max_seq_length: 2048,
      },
      export: { formats: ['gguf', 'ollama'], ollama_modelfile_path: './dist/Modelfile' },
    }

    await onLaunchJob(configPayload)
  }

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-bold text-white">Create New Distillation Pipeline</h2>
          <p className="text-xs text-slate-400">Configure synthetic data evolution, quality filtering, and LoRA fine-tuning</p>
        </div>

        <div className="flex items-center rounded-lg bg-slate-900 border border-slate-800 p-1">
          <button
            onClick={() => setActiveTab('form')}
            className={`px-3 py-1 rounded text-xs font-medium ${
              activeTab === 'form' ? 'bg-indigo-600 text-white' : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            Form Wizard
          </button>
          <button
            onClick={() => setActiveTab('yaml')}
            className={`flex items-center gap-1.5 px-3 py-1 rounded text-xs font-medium ${
              activeTab === 'yaml' ? 'bg-indigo-600 text-white' : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            <Code2 className="w-3.5 h-3.5" />
            <span>YAML Preview</span>
          </button>
        </div>
      </div>

      {activeTab === 'form' ? (
        <div className="space-y-6">
          {/* Project Details */}
          <div className="p-6 rounded-xl border border-slate-800 bg-slate-900/60 backdrop-blur space-y-4">
            <h3 className="text-sm font-semibold text-white flex items-center gap-2">
              <Sparkles className="w-4 h-4 text-indigo-400" />
              <span>Project Identity</span>
            </h3>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <label className="block text-xs font-medium text-slate-300 mb-1">Project Name</label>
                <input
                  type="text"
                  value={projectName}
                  onChange={(e) => setProjectName(e.target.value)}
                  className="w-full px-3 py-2 rounded-lg bg-slate-800 border border-slate-700 text-xs text-white focus:outline-none focus:border-indigo-500"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-slate-300 mb-1">Domain Description</label>
                <input
                  type="text"
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  className="w-full px-3 py-2 rounded-lg bg-slate-800 border border-slate-700 text-xs text-white focus:outline-none focus:border-indigo-500"
                />
              </div>
            </div>
          </div>

          {/* Model Selection */}
          <div className="p-6 rounded-xl border border-slate-800 bg-slate-900/60 backdrop-blur space-y-4">
            <h3 className="text-sm font-semibold text-white flex items-center gap-2">
              <Sliders className="w-4 h-4 text-violet-400" />
              <span>Model Architecture</span>
            </h3>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <label className="block text-xs font-medium text-slate-300 mb-1">Teacher Provider & Model</label>
                <div className="flex gap-2">
                  <select
                    value={teacherProvider}
                    onChange={(e) => setTeacherProvider(e.target.value)}
                    className="w-1/3 px-3 py-2 rounded-lg bg-slate-800 border border-slate-700 text-xs text-white focus:outline-none focus:border-indigo-500"
                  >
                    <option value="mock">Mock / Simulation</option>
                    <option value="openai">OpenAI</option>
                    <option value="anthropic">Anthropic</option>
                    <option value="ollama">Ollama</option>
                  </select>
                  <input
                    type="text"
                    value={teacherModel}
                    onChange={(e) => setTeacherModel(e.target.value)}
                    className="w-2/3 px-3 py-2 rounded-lg bg-slate-800 border border-slate-700 text-xs text-white focus:outline-none focus:border-indigo-500"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-medium text-slate-300 mb-1">Student Base SLM</label>
                <select
                  value={studentModel}
                  onChange={(e) => setStudentModel(e.target.value)}
                  className="w-full px-3 py-2 rounded-lg bg-slate-800 border border-slate-700 text-xs text-white focus:outline-none focus:border-indigo-500"
                >
                  <option value="meta-llama/Llama-3.2-1B-Instruct">meta-llama/Llama-3.2-1B-Instruct (Edge default)</option>
                  <option value="meta-llama/Llama-3.2-3B-Instruct">meta-llama/Llama-3.2-3B-Instruct</option>
                  <option value="Qwen/Qwen2.5-1.5B-Instruct">Qwen/Qwen2.5-1.5B-Instruct</option>
                  <option value="HuggingFaceTB/SmolLM2-1.7B-Instruct">HuggingFaceTB/SmolLM2-1.7B-Instruct</option>
                </select>
              </div>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 pt-2">
              <div>
                <label className="block text-xs font-medium text-slate-300 mb-1">Training Method</label>
                <select
                  value={trainingMethod}
                  onChange={(e) => setTrainingMethod(e.target.value)}
                  className="w-full px-3 py-2 rounded-lg bg-slate-800 border border-slate-700 text-xs text-white focus:outline-none focus:border-indigo-500"
                >
                  <option value="qlora">QLoRA (4-bit bitsandbytes)</option>
                  <option value="lora">LoRA (FP16 / BF16)</option>
                  <option value="mlx">Apple Silicon MLX</option>
                  <option value="cpu">CPU Fallback</option>
                </select>
              </div>

              <div>
                <label className="block text-xs font-medium text-slate-300 mb-1">Synthetic Dataset Size</label>
                <input
                  type="number"
                  value={datasetSize}
                  onChange={(e) => setDatasetSize(Number(e.target.value))}
                  className="w-full px-3 py-2 rounded-lg bg-slate-800 border border-slate-700 text-xs text-white focus:outline-none focus:border-indigo-500"
                />
              </div>

              <div>
                <label className="block text-xs font-medium text-slate-300 mb-1">Training Epochs</label>
                <input
                  type="number"
                  value={epochs}
                  onChange={(e) => setEpochs(Number(e.target.value))}
                  className="w-full px-3 py-2 rounded-lg bg-slate-800 border border-slate-700 text-xs text-white focus:outline-none focus:border-indigo-500"
                />
              </div>
            </div>
          </div>

          {/* Evol-Instruct Strategies */}
          <div className="p-6 rounded-xl border border-slate-800 bg-slate-900/60 backdrop-blur space-y-4">
            <h3 className="text-sm font-semibold text-white">Evol-Instruct Mutation Ratios</h3>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
              <div>
                <div className="flex justify-between text-xs text-slate-300 mb-1">
                  <span>In-Depth Evolution</span>
                  <span className="font-mono text-indigo-400">{(inDepthWeight * 100).toFixed(0)}%</span>
                </div>
                <input
                  type="range"
                  min="0"
                  max="1"
                  step="0.05"
                  value={inDepthWeight}
                  onChange={(e) => setInDepthWeight(parseFloat(e.target.value))}
                  className="w-full accent-indigo-500"
                />
                <p className="text-[11px] text-slate-400 mt-1">Deepens operational and domain constraints.</p>
              </div>

              <div>
                <div className="flex justify-between text-xs text-slate-300 mb-1">
                  <span>In-Breadth Evolution</span>
                  <span className="font-mono text-indigo-400">{(inBreadthWeight * 100).toFixed(0)}%</span>
                </div>
                <input
                  type="range"
                  min="0"
                  max="1"
                  step="0.05"
                  value={inBreadthWeight}
                  onChange={(e) => setInBreadthWeight(parseFloat(e.target.value))}
                  className="w-full accent-indigo-500"
                />
                <p className="text-[11px] text-slate-400 mt-1">Expands scenario diversity and tone.</p>
              </div>

              <div>
                <div className="flex justify-between text-xs text-slate-300 mb-1">
                  <span>Adversarial Edge Cases</span>
                  <span className="font-mono text-indigo-400">{(adversarialWeight * 100).toFixed(0)}%</span>
                </div>
                <input
                  type="range"
                  min="0"
                  max="1"
                  step="0.05"
                  value={adversarialWeight}
                  onChange={(e) => setAdversarialWeight(parseFloat(e.target.value))}
                  className="w-full accent-indigo-500"
                />
                <p className="text-[11px] text-slate-400 mt-1">Injects boundary triggers and input noise.</p>
              </div>
            </div>
          </div>
        </div>
      ) : (
        <div className="p-4 rounded-xl border border-slate-800 bg-slate-950 font-mono text-xs text-indigo-300 overflow-x-auto">
          <pre>{yamlPreview}</pre>
        </div>
      )}

      {/* Action Bar */}
      <div className="flex items-center justify-between p-4 rounded-xl border border-slate-800 bg-slate-900/80">
        <div className="flex items-center gap-2 text-xs text-slate-400">
          <AlertCircle className="w-4 h-4 text-indigo-400" />
          <span>Execution runs: Generation → Judge Filter → LoRA Train → Benchmark → GGUF Export</span>
        </div>

        <button
          onClick={handleLaunch}
          disabled={isLaunching}
          className="flex items-center gap-2 px-6 py-2.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white text-xs font-semibold shadow-lg shadow-indigo-600/30 transition-all"
        >
          {isLaunching ? (
            <>
              <div className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
              <span>Starting Pipeline...</span>
            </>
          ) : (
            <>
              <Play className="w-3.5 h-3.5 fill-current" />
              <span>Launch Distillation Pipeline</span>
            </>
          )}
        </button>
      </div>
    </div>
  )
}
