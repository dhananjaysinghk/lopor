"use client";

import React from "react";
import { ShieldAlert, Users, HardDrive, Database, Activity, ToggleLeft, ToggleRight, CheckCircle2 } from "lucide-react";

export default function AdminDashboardPage() {
  const systemMetrics = [
    { label: "Registered Users", value: "1,248", change: "+12% this week", color: "text-indigo-400" },
    { label: "Active Workspaces", value: "312", change: "98% online", color: "text-emerald-400" },
    { label: "RAG Vectors Stored", value: "48,590", change: "pgvector HNSW", color: "text-purple-400" },
    { label: "Storage Consumed", value: "42.8 GB", change: "Cloudflare R2", color: "text-amber-400" },
  ];

  const auditLogs = [
    { action: "USER_LOGIN", user: "sarah@lopor.ai", ip: "192.168.1.45", time: "2 mins ago" },
    { action: "VECTOR_INDEX_CREATED", user: "system_worker", ip: "127.0.0.1", time: "14 mins ago" },
    { action: "WORKSPACE_CREATED", user: "alex@company.com", ip: "10.0.4.12", time: "1 hour ago" },
  ];

  return (
    <div className="max-w-6xl mx-auto space-y-8">
      <div>
        <h1 className="text-xl font-bold text-white tracking-tight flex items-center gap-2">
          <ShieldAlert size={20} className="text-indigo-400" />
          Enterprise Superadmin Control Panel
        </h1>
        <p className="text-xs text-zinc-400 mt-1">
          Monitor system metrics, tenant workspaces, feature flag toggles, and security audit trails.
        </p>
      </div>

      {/* Metrics Cards Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {systemMetrics.map((m, idx) => (
          <div key={idx} className="p-5 rounded-xl bg-zinc-900/60 border border-zinc-800/80 shadow-md">
            <span className="text-xs text-zinc-400 font-medium">{m.label}</span>
            <p className={`text-xl font-bold mt-2 ${m.color}`}>{m.value}</p>
            <span className="text-[10px] text-zinc-500 font-mono mt-1 block">{m.change}</span>
          </div>
        ))}
      </div>

      {/* Feature Flags Control Box */}
      <div className="p-6 rounded-xl border border-zinc-800/80 bg-zinc-900/40 space-y-4 shadow-xl">
        <h3 className="text-sm font-semibold text-white">Production Feature Flags</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-xs">
          <div className="p-3.5 rounded-lg bg-zinc-950 border border-zinc-800 flex items-center justify-between">
            <div>
              <p className="font-semibold text-zinc-200">OpenAI SSE Real-Time Streaming</p>
              <p className="text-[10px] text-zinc-500 font-mono">Enables low-latency chat deltas</p>
            </div>
            <ToggleRight size={24} className="text-indigo-400 cursor-pointer" />
          </div>

          <div className="p-3.5 rounded-lg bg-zinc-950 border border-zinc-800 flex items-center justify-between">
            <div>
              <p className="font-semibold text-zinc-200">pgvector HNSW Cosine Indexing</p>
              <p className="text-[10px] text-zinc-500 font-mono">1536-dimensional similarity vector store</p>
            </div>
            <ToggleRight size={24} className="text-emerald-400 cursor-pointer" />
          </div>
        </div>
      </div>

      {/* Security Audit Trail Log Table */}
      <div className="rounded-xl border border-zinc-800/80 bg-zinc-900/40 overflow-hidden shadow-xl">
        <div className="px-4 py-3 border-b border-zinc-800 flex items-center justify-between bg-zinc-950/60 text-xs font-medium text-zinc-400">
          <span>Security Audit Trail Logs</span>
          <span className="font-mono text-[10px] text-indigo-400">Immutable Ledger</span>
        </div>

        <div className="divide-y divide-zinc-800/60">
          {auditLogs.map((log, idx) => (
            <div key={idx} className="p-3.5 flex items-center justify-between text-xs hover:bg-zinc-800/30 transition-colors">
              <div className="flex items-center gap-3">
                <span className="px-2 py-0.5 rounded text-[10px] font-mono bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 font-bold">
                  {log.action}
                </span>
                <span className="text-zinc-200 font-medium">{log.user}</span>
              </div>
              <div className="flex items-center gap-4 text-zinc-500 font-mono text-[11px]">
                <span>{log.ip}</span>
                <span>{log.time}</span>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
