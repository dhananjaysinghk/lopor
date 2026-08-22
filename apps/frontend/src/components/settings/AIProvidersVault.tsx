"use client";

import React, { useState } from "react";
import { Cpu, CheckCircle2, RefreshCw, Key, Shield, Zap, Server } from "lucide-react";

interface ModelInfo {
  id: string;
  name: string;
  provider: string;
  context_size: number;
  is_local: boolean;
  is_healthy: boolean;
}

export function AIProvidersVault() {
  const [models, setModels] = useState<ModelInfo[]>([
    { id: "gpt-4o", name: "OpenAI GPT-4o", provider: "openai", context_size: 128000, is_local: false, is_healthy: true },
    { id: "claude-3-5-sonnet", name: "Anthropic Claude 3.5 Sonnet", provider: "anthropic", context_size: 200000, is_local: false, is_healthy: true },
    { id: "llama3-8b", name: "Ollama Llama 3 (Local GPU)", provider: "ollama", context_size: 8192, is_local: true, is_healthy: true },
    { id: "deepseek-coder", name: "DeepSeek Coder V2", provider: "deepseek", context_size: 64000, is_local: false, is_healthy: true },
  ]);

  return (
    <div className="p-6 rounded-xl border border-zinc-800/80 bg-zinc-900/40 space-y-4 shadow-xl">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-sm font-semibold text-white flex items-center gap-2">
            <Cpu size={16} className="text-indigo-400" /> Multi-Model Gateway & Provider Health
          </h3>
          <p className="text-xs text-zinc-400 mt-0.5">
            Automatic failover router across cloud API providers and local Ollama GPU instances.
          </p>
        </div>
        <span className="px-2.5 py-1 rounded text-xs font-mono bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 font-bold flex items-center gap-1.5">
          <CheckCircle2 size={12} /> Router Stack Active
        </span>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-3 pt-2">
        {models.map((m) => (
          <div key={m.id} className="p-4 rounded-lg bg-zinc-950 border border-zinc-800 flex items-center justify-between text-xs">
            <div className="space-y-1">
              <div className="flex items-center gap-2">
                <span className="font-semibold text-zinc-200">{m.name}</span>
                {m.is_local && (
                  <span className="px-1.5 py-0.2 rounded text-[9px] font-mono bg-purple-500/10 text-purple-400 border border-purple-500/20">
                    Local GPU
                  </span>
                )}
              </div>
              <p className="text-[10px] text-zinc-500 font-mono">Context: {(m.context_size / 1000).toFixed(0)}k Tokens • {m.provider}</p>
            </div>

            <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" title="Healthy" />
          </div>
        ))}
      </div>
    </div>
  );
}
