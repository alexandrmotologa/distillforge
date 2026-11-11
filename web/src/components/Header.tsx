import React from 'react'
import { Cpu, Activity, Database, GitCompare, DollarSign, Terminal, Play, Flame } from 'lucide-react'

interface HeaderProps {
  activeTab: string
  setActiveTab: (tab: string) => void
  pipelineRunning: boolean
}

export const Header: React.FC<HeaderProps> = ({ activeTab, setActiveTab, pipelineRunning }) => {
  const tabs = [
    { id: 'dashboard', label: 'Dashboard', icon: Activity },
    { id: 'new', label: 'New Job', icon: Flame },
    { id: 'training', label: 'Training Cockpit', icon: Cpu },
    { id: 'datasets', label: 'Datasets', icon: Database },
    { id: 'diff', label: 'Head-to-Head Diff', icon: GitCompare },
    { id: 'finops', label: 'FinOps ROI', icon: DollarSign },
    { id: 'playground', label: 'Playground', icon: Terminal },
  ]

  return (
    <header className="border-b border-slate-800 bg-[#0c1220]/80 backdrop-blur sticky top-0 z-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
        {/* Brand */}
        <div className="flex items-center gap-3">
          <div className="w-9 h-9 rounded-lg bg-gradient-to-br from-indigo-500 to-violet-600 flex items-center justify-center shadow-lg shadow-indigo-500/20">
            <Cpu className="w-5 h-5 text-white" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <span className="font-bold text-lg tracking-tight bg-gradient-to-r from-white via-slate-200 to-slate-400 bg-clip-text text-transparent">
                distillforge
              </span>
              <span className="text-[10px] font-mono uppercase bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 px-1.5 py-0.5 rounded">
                v1.0 Studio
              </span>
            </div>
            <p className="text-xs text-slate-400 hidden sm:block">Turnkey LLM-to-SLM Distillation & FinOps</p>
          </div>
        </div>

        {/* Navigation Tabs */}
        <nav className="flex items-center gap-1 overflow-x-auto py-1">
          {tabs.map((tab) => {
            const Icon = tab.icon
            const isActive = activeTab === tab.id
            return (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`flex items-center gap-2 px-3 py-1.5 rounded-md text-xs font-medium transition-all ${
                  isActive
                    ? 'bg-indigo-600/15 text-indigo-400 border border-indigo-500/30'
                    : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/50'
                }`}
              >
                <Icon className="w-3.5 h-3.5" />
                <span>{tab.label}</span>
              </button>
            )
          })}
        </nav>

        {/* Status / Quick Action */}
        <div className="flex items-center gap-3">
          {pipelineRunning ? (
            <div className="flex items-center gap-2 px-2.5 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 text-xs">
              <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
              <span>Job Active</span>
            </div>
          ) : (
            <button
              onClick={() => setActiveTab('new')}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold shadow-md shadow-indigo-600/20 transition-all"
            >
              <Play className="w-3.5 h-3.5 fill-current" />
              <span>Start Distillation</span>
            </button>
          )}
        </div>
      </div>
    </header>
  )
}
