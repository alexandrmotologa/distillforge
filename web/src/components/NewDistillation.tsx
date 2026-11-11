import React, { useState, useRef } from 'react'
import { Play, Sparkles, Sliders, Code2, AlertCircle, UploadCloud, Check, Cpu, Laptop, Server, HardDrive } from 'lucide-react'

interface NewDistillationProps {
  onLaunchJob: (config: any) => Promise<void>
  isLaunching: boolean
}

export const NewDistillation: React.FC<NewDistillationProps> = ({ onLaunchJob, isLaunching }) => {
  const [projectName, setProjectName] = useState('ticket-classifier')
  const [description, setDescription] = useState('Distill GPT-4o into Llama-3.2-1B for real-time customer support routing')
  const [teacherProvider, setTeacherProvider] = useState('mock')
  const [teacherModel, setTeacherModel] = useState('gpt-4o')
  const [studentModel, setStudentModel] = useState('meta-llama/Llama-3.2-1B-Instruct')
  const [trainingMethod, setTrainingMethod] = useState('qlora')
  const [quantTarget, setQuantTarget] = useState('q4_k_m')
  const [datasetSize, setDatasetSize] = useState(100)
  const [epochs, setEpochs] = useState(3)
  const [inDepthWeight, setInDepthWeight] = useState(0.4)
  const [inBreadthWeight, setInBreadthWeight] = useState(0.3)
  const [adversarialWeight, setAdversarialWeight] = useState(0.3)
  const [activeTab, setActiveTab] = useState<'form' | 'yaml'>('form')

  // Seed data & upload state
  const [seedFilePath, setSeedFilePath] = useState('./examples/ticket_classifier/data/raw_tickets.jsonl')
  const [inputKey, setInputKey] = useState('prompt')
  const [targetKey, setTargetKey] = useState('response')
  const [isUploading, setIsUploading] = useState(false)
  const [uploadSuccess, setUploadSuccess] = useState<string | null>(null)
  const [detectedKeys, setDetectedKeys] = useState<string[]>([])
  const [activePreset, setActivePreset] = useState<string>('gpu_8gb')
  const fileInputRef = useRef<HTMLInputElement>(null)

  const applyHardwarePreset = (preset: string) => {
    setActivePreset(preset)
    switch (preset) {
      case 'apple_silicon':
        setStudentModel('meta-llama/Llama-3.2-1B-Instruct')
        setTrainingMethod('mlx')
        setQuantTarget('q4_k_m')
        break
      case 'gpu_8gb':
        setStudentModel('meta-llama/Llama-3.2-1B-Instruct')
        setTrainingMethod('qlora')
        setQuantTarget('q4_k_m')
        break
      case 'gpu_16gb_24gb':
        setStudentModel('meta-llama/Llama-3.2-3B-Instruct')
        setTrainingMethod('qlora')
        setQuantTarget('q8_0')
        break
      case 'cpu':
        setStudentModel('HuggingFaceTB/SmolLM2-1.7B-Instruct')
        setTrainingMethod('cpu')
        setQuantTarget('q4_k_m')
        break
    }
  }

  const handleFileUpload = async (file: File) => {
    setIsUploading(true)
    setUploadSuccess(null)

    const formData = new FormData()
    formData.append('file', file)

    try {
      const res = await fetch('/api/upload', {
        method: 'POST',
        body: formData,
      })

      if (!res.ok) {
        throw new Error(`Upload failed with status ${res.status}`)
      }

      const data = await res.json()
      setSeedFilePath(data.file_path)
      setUploadSuccess(`Uploaded ${file.name} (${data.detected_count} samples detected)`)

      if (data.detected_keys && data.detected_keys.length > 0) {
        setDetectedKeys(data.detected_keys)
        // Auto-select candidate keys
        const keys: string[] = data.detected_keys
        const foundInput = keys.find((k) => ['prompt', 'input', 'text', 'ticket', 'instruction', 'query'].includes(k.toLowerCase()))
        const foundTarget = keys.find((k) => ['response', 'output', 'target', 'label', 'completion', 'answer'].includes(k.toLowerCase()))
        if (foundInput) setInputKey(foundInput)
        if (foundTarget) setTargetKey(foundTarget)
      }
    } catch (err: any) {
      alert(`Upload error: ${err.message}`)
    } finally {
      setIsUploading(false)
    }
  }

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault()
    if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
      handleFileUpload(e.dataTransfer.files[0])
    }
  }

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
  quantization_target: "${quantTarget}"

seed_data:
  file: "${seedFilePath}"
  input_key: "${inputKey}"
  target_key: "${targetKey}"

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
  max_seq_length: 2048

export:
  formats: ["gguf", "ollama"]
  ollama_modelfile_path: "./dist/Modelfile"
`

  const handleLaunch = async () => {
    const configPayload = {
      version: '1',
      project: { name: projectName, description },
      teacher: { provider: teacherProvider, model: teacherModel, temperature: 0.7, max_tokens: 1024 },
      student: { base_model: studentModel, quantization_target: quantTarget },
      seed_data: {
        file: seedFilePath,
        input_key: inputKey,
        target_key: targetKey,
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
                  className="w-full px-3 py-2 rounded-lg bg-slate-800 border border-slate-700 text-xs text-white focus:outline-none focus:border-indigo-500 font-mono"
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

          {/* Seed Data Upload & Ingestion Dropzone */}
          <div className="p-6 rounded-xl border border-slate-800 bg-slate-900/60 backdrop-blur space-y-4">
            <h3 className="text-sm font-semibold text-white flex items-center gap-2">
              <UploadCloud className="w-4 h-4 text-cyan-400" />
              <span>Seed Dataset Ingestion & Schema Mapping</span>
            </h3>

            {/* Dropzone */}
            <div
              onDragOver={(e) => e.preventDefault()}
              onDrop={handleDrop}
              onClick={() => fileInputRef.current?.click()}
              className="border-2 border-dashed border-slate-700 hover:border-indigo-500/70 rounded-xl p-6 text-center cursor-pointer transition-colors bg-slate-950/40"
            >
              <input
                ref={fileInputRef}
                type="file"
                accept=".json,.jsonl,.csv"
                className="hidden"
                onChange={(e) => e.target.files && handleFileUpload(e.target.files[0])}
              />
              <UploadCloud className="w-8 h-8 text-indigo-400 mx-auto mb-2" />
              <div className="text-xs text-slate-200 font-medium">
                {isUploading ? 'Analyzing and uploading seed file...' : 'Drag and drop seed data (.json, .jsonl, .csv) or click to browse'}
              </div>
              <div className="text-[11px] text-slate-500 mt-1">
                DistillForge inspects field structures, counts raw examples, and identifies input/output candidates
              </div>
            </div>

            {uploadSuccess && (
              <div className="p-3 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs flex items-center gap-2">
                <Check className="w-4 h-4 shrink-0" />
                <span>{uploadSuccess}</span>
              </div>
            )}

            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 pt-1">
              <div>
                <label className="block text-xs font-medium text-slate-300 mb-1">Seed File Path</label>
                <input
                  type="text"
                  value={seedFilePath}
                  onChange={(e) => setSeedFilePath(e.target.value)}
                  className="w-full px-3 py-2 rounded-lg bg-slate-800 border border-slate-700 text-xs text-white focus:outline-none focus:border-indigo-500 font-mono"
                />
              </div>

              <div>
                <label className="block text-xs font-medium text-slate-300 mb-1">Prompt / Input Field Key</label>
                {detectedKeys.length > 0 ? (
                  <select
                    value={inputKey}
                    onChange={(e) => setInputKey(e.target.value)}
                    className="w-full px-3 py-2 rounded-lg bg-slate-800 border border-slate-700 text-xs text-white focus:outline-none focus:border-indigo-500 font-mono"
                  >
                    {detectedKeys.map((k) => (
                      <option key={k} value={k}>
                        {k}
                      </option>
                    ))}
                  </select>
                ) : (
                  <input
                    type="text"
                    value={inputKey}
                    onChange={(e) => setInputKey(e.target.value)}
                    className="w-full px-3 py-2 rounded-lg bg-slate-800 border border-slate-700 text-xs text-white focus:outline-none focus:border-indigo-500 font-mono"
                  />
                )}
              </div>

              <div>
                <label className="block text-xs font-medium text-slate-300 mb-1">Target / Response Field Key</label>
                {detectedKeys.length > 0 ? (
                  <select
                    value={targetKey}
                    onChange={(e) => setTargetKey(e.target.value)}
                    className="w-full px-3 py-2 rounded-lg bg-slate-800 border border-slate-700 text-xs text-white focus:outline-none focus:border-indigo-500 font-mono"
                  >
                    {detectedKeys.map((k) => (
                      <option key={k} value={k}>
                        {k}
                      </option>
                    ))}
                  </select>
                ) : (
                  <input
                    type="text"
                    value={targetKey}
                    onChange={(e) => setTargetKey(e.target.value)}
                    className="w-full px-3 py-2 rounded-lg bg-slate-800 border border-slate-700 text-xs text-white focus:outline-none focus:border-indigo-500 font-mono"
                  />
                )}
              </div>
            </div>
          </div>

          {/* 1-Click Hardware Target Presets */}
          <div className="p-6 rounded-xl border border-slate-800 bg-slate-900/60 backdrop-blur space-y-4">
            <h3 className="text-sm font-semibold text-white flex items-center justify-between">
              <span className="flex items-center gap-2">
                <Cpu className="w-4 h-4 text-emerald-400" />
                <span>Hardware Target & FinOps Recipe Presets</span>
              </span>
              <span className="text-[11px] text-slate-400 font-normal">Automates optimal rank, precision & quant matrix</span>
            </h3>

            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
              <button
                type="button"
                onClick={() => applyHardwarePreset('apple_silicon')}
                className={`p-3 rounded-xl border text-left transition-all ${
                  activePreset === 'apple_silicon'
                    ? 'border-indigo-500 bg-indigo-500/10 ring-1 ring-indigo-500'
                    : 'border-slate-800 bg-slate-950/40 hover:border-slate-700'
                }`}
              >
                <div className="flex items-center gap-2 text-xs font-semibold text-white mb-1">
                  <Laptop className="w-3.5 h-3.5 text-indigo-400" />
                  <span>Apple Silicon</span>
                </div>
                <div className="text-[11px] text-slate-400">M-series MLX unified memory, 85+ tok/s edge inference</div>
              </button>

              <button
                type="button"
                onClick={() => applyHardwarePreset('gpu_8gb')}
                className={`p-3 rounded-xl border text-left transition-all ${
                  activePreset === 'gpu_8gb'
                    ? 'border-indigo-500 bg-indigo-500/10 ring-1 ring-indigo-500'
                    : 'border-slate-800 bg-slate-950/40 hover:border-slate-700'
                }`}
              >
                <div className="flex items-center gap-2 text-xs font-semibold text-white mb-1">
                  <Cpu className="w-3.5 h-3.5 text-emerald-400" />
                  <span>Consumer GPU 8GB</span>
                </div>
                <div className="text-[11px] text-slate-400">RTX 3070/4060, QLoRA 4-bit, sub-20ms edge latency</div>
              </button>

              <button
                type="button"
                onClick={() => applyHardwarePreset('gpu_16gb_24gb')}
                className={`p-3 rounded-xl border text-left transition-all ${
                  activePreset === 'gpu_16gb_24gb'
                    ? 'border-indigo-500 bg-indigo-500/10 ring-1 ring-indigo-500'
                    : 'border-slate-800 bg-slate-950/40 hover:border-slate-700'
                }`}
              >
                <div className="flex items-center gap-2 text-xs font-semibold text-white mb-1">
                  <Server className="w-3.5 h-3.5 text-violet-400" />
                  <span>Enterprise GPU</span>
                </div>
                <div className="text-[11px] text-slate-400">A10G/RTX 4090 (16-24GB), 3B SLM, high rank r=32 Q8_0</div>
              </button>

              <button
                type="button"
                onClick={() => applyHardwarePreset('cpu')}
                className={`p-3 rounded-xl border text-left transition-all ${
                  activePreset === 'cpu'
                    ? 'border-indigo-500 bg-indigo-500/10 ring-1 ring-indigo-500'
                    : 'border-slate-800 bg-slate-950/40 hover:border-slate-700'
                }`}
              >
                <div className="flex items-center gap-2 text-xs font-semibold text-white mb-1">
                  <HardDrive className="w-3.5 h-3.5 text-amber-400" />
                  <span>Standard CPU</span>
                </div>
                <div className="text-[11px] text-slate-400">Edge gateway / CPU fallback, SmolLM2-1.7B, minimal memory</div>
              </button>
            </div>
          </div>

          {/* Model Selection */}
          <div className="p-6 rounded-xl border border-slate-800 bg-slate-900/60 backdrop-blur space-y-4">
            <h3 className="text-sm font-semibold text-white flex items-center gap-2">
              <Sliders className="w-4 h-4 text-violet-400" />
              <span>Model Architecture & Quantization Matrix</span>
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

            <div className="grid grid-cols-1 sm:grid-cols-4 gap-4 pt-2">
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
                <label className="block text-xs font-medium text-slate-300 mb-1">Quant Target</label>
                <select
                  value={quantTarget}
                  onChange={(e) => setQuantTarget(e.target.value)}
                  className="w-full px-3 py-2 rounded-lg bg-slate-800 border border-slate-700 text-xs text-white focus:outline-none focus:border-indigo-500 font-mono"
                >
                  <option value="q4_k_m">Q4_K_M (Balanced 4-bit)</option>
                  <option value="q5_k_m">Q5_K_M (Enhanced 5-bit)</option>
                  <option value="q8_0">Q8_0 (Near-FP16 8-bit)</option>
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
          className="flex items-center gap-2 px-6 py-2.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white text-xs font-semibold shadow-lg shadow-indigo-600/30 transition-all cursor-pointer"
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
