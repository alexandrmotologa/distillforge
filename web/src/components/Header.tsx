import React, { useState } from 'react'
import { Cpu, Activity, Database, GitCompare, DollarSign, Terminal, Play, Flame, Menu, X } from 'lucide-react'

interface HeaderProps {
  activeTab: string
  setActiveTab: (tab: string) => void
  pipelineRunning: boolean
}

export const Header: React.FC<HeaderProps> = ({ activeTab, setActiveTab, pipelineRunning }) => {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false)

  const tabs = [
    { id: 'dashboard', label: 'Dashboard', icon: Activity },
    { id: 'new', label: 'New Job', icon: Flame },
    { id: 'training', label: 'Training Cockpit', icon: Cpu },
    { id: 'datasets', label: 'Datasets', icon: Database },
    { id: 'diff', label: 'Head-to-Head Diff', icon: GitCompare },
    { id: 'finops', label: 'FinOps ROI', icon: DollarSign },
    { id: 'playground', label: 'Playground', icon: Terminal },
  ]

  const handleSelectTab = (tabId: string) => {
    setActiveTab(tabId)
    setIsMobileMenuOpen(false)
  }

  return (
    <header className="border-b border-slate-800/80 bg-[#0c1220]/90 backdrop-blur-md sticky top-0 z-50">
      <div className="max-w-[1440px] mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between gap-4">
        {/* Brand */}
        <div className="flex items-center gap-3 flex-shrink-0">
          <div className="w-9 h-9 rounded-xl bg-slate-900 border border-slate-700/60 p-1 flex items-center justify-center shadow-lg shadow-indigo-500/10 flex-shrink-0">
            <img
              src="/distillforge_logo.svg"
              alt="distillforge emblem"
              className="w-full h-full object-contain"
            />
          </div>
          <div className="flex flex-col flex-shrink-0">
            <div className="flex items-center gap-2">
              <span className="font-bold text-lg tracking-tight bg-gradient-to-r from-white via-slate-100 to-slate-300 bg-clip-text text-transparent whitespace-nowrap">
                distillforge
              </span>
              <span className="text-[10px] font-mono uppercase bg-indigo-500/15 text-indigo-400 border border-indigo-500/30 px-1.5 py-0.5 rounded font-semibold whitespace-nowrap">
                v1.0 Studio
              </span>
            </div>
            <p className="text-[11px] text-slate-400 hidden xl:block whitespace-nowrap leading-none mt-0.5">
              Turnkey LLM-to-SLM Distillation &amp; FinOps
            </p>
          </div>
        </div>

        {/* Desktop Navigation Tabs */}
        <nav className="hidden lg:flex items-center gap-1 flex-shrink-0">
          {tabs.map((tab) => {
            const Icon = tab.icon
            const isActive = activeTab === tab.id
            return (
              <button
                key={tab.id}
                onClick={() => handleSelectTab(tab.id)}
                className={`flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs font-medium whitespace-nowrap flex-shrink-0 transition-all cursor-pointer ${
                  isActive
                    ? 'bg-indigo-600/20 text-indigo-300 border border-indigo-500/40 shadow-sm shadow-indigo-500/10'
                    : 'text-slate-400 hover:text-slate-100 hover:bg-slate-800/60 border border-transparent'
                }`}
              >
                <Icon className="w-3.5 h-3.5 flex-shrink-0" />
                <span className="whitespace-nowrap">{tab.label}</span>
              </button>
            )
          })}
        </nav>

        {/* Status / Quick Action / Mobile Toggle */}
        <div className="flex items-center gap-2.5 flex-shrink-0">
          {pipelineRunning ? (
            <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 text-xs font-medium whitespace-nowrap flex-shrink-0">
              <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse flex-shrink-0" />
              <span>Job Active</span>
            </div>
          ) : (
            <button
              onClick={() => handleSelectTab('new')}
              className="hidden sm:inline-flex items-center gap-1.5 px-3.5 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold shadow-md shadow-indigo-600/25 transition-all whitespace-nowrap flex-shrink-0 cursor-pointer"
            >
              <Play className="w-3.5 h-3.5 fill-current flex-shrink-0" />
              <span className="whitespace-nowrap">Start Distillation</span>
            </button>
          )}

          {/* Mobile hamburger button */}
          <button
            onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
            className="lg:hidden flex items-center justify-center p-2 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 border border-slate-800 transition-colors"
            aria-label="Toggle navigation menu"
          >
            {isMobileMenuOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
          </button>
        </div>
      </div>

      {/* Mobile Drawer */}
      {isMobileMenuOpen && (
        <div className="lg:hidden border-t border-slate-800 bg-[#0c1220] px-4 pt-3 pb-4 space-y-1 shadow-2xl animate-fade-in">
          {tabs.map((tab) => {
            const Icon = tab.icon
            const isActive = activeTab === tab.id
            return (
              <button
                key={tab.id}
                onClick={() => handleSelectTab(tab.id)}
                className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-xs font-medium transition-all ${
                  isActive
                    ? 'bg-indigo-600/20 text-indigo-300 border border-indigo-500/40'
                    : 'text-slate-300 hover:bg-slate-800/80 hover:text-white'
                }`}
              >
                <Icon className="w-4 h-4 text-indigo-400 flex-shrink-0" />
                <span className="whitespace-nowrap font-medium">{tab.label}</span>
              </button>
            )
          })}
          <div className="pt-2">
            <button
              onClick={() => handleSelectTab('new')}
              className="w-full flex items-center justify-center gap-2 px-3 py-2.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold shadow-md shadow-indigo-600/25 transition-all whitespace-nowrap"
            >
              <Play className="w-3.5 h-3.5 fill-current" />
              <span>Start Distillation</span>
            </button>
          </div>
        </div>
      )}
    </header>
  )
}
