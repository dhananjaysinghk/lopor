"use client";

import React, { useState } from "react";
import { Cpu, Play, CheckCircle2, Clock, AlertTriangle, RefreshCw, Layers, Database } from "lucide-react";
import { useAuth } from "@/context/AuthContext";

interface JobItem {
  id: string;
  type: string;
  status: "QUEUED" | "PROCESSING" | "COMPLETED" | "FAILED";
  result?: string;
  created_at: string;
}

export default function JobsPage() {
  const { activeWorkspace } = useAuth();
  const [jobs, setJobs] = useState<JobItem[]>([
    {
      id: "job_8f29a",
      type: "rag_reindex",
      status: "COMPLETED",
      result: "Re-indexed 48 vector chunks into pgvector HNSW index.",
      created_at: "2 mins ago",
    },
    {
      id: "job_312bc",
      type: "document_pdf_export",
      status: "COMPLETED",
      result: "Generated PDF export for 'System Architecture PRD'.",
      created_at: "18 mins ago",
    },
  ]);
  const [isEnqueueing, setIsEnqueueing] = useState(false);

  const triggerJob = (jobType: string) => {
    setIsEnqueueing(true);
    const newJob: JobItem = {
      id: "job_" + Math.random().toString(36).substr(2, 5),
      type: jobType,
      status: "PROCESSING",
      created_at: "Just now",
    };

    setJobs((prev) => [newJob, ...prev]);

    setTimeout(() => {
      setJobs((prev) =>
        prev.map((j) =>
          j.id === newJob.id
            ? { ...j, status: "COMPLETED", result: `Worker successfully completed async task '${jobType}'.` }
            : j
        )
      );
      setIsEnqueueing(false);
    }, 1200);
  };

  return (
    <div className="max-w-5xl mx-auto space-y-6">
      {/* Top Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-white tracking-tight flex items-center gap-2">
            <Cpu size={20} className="text-indigo-400" />
            Redis Async Job Queue & Task Automation
          </h1>
          <p className="text-xs text-zinc-400 mt-1">
            Monitor background worker tasks, scheduled cron automations, PDF exporters, and pgvector re-indexing.
          </p>
        </div>
      </div>

      {/* Manual Worker Trigger Buttons */}
      <div className="p-6 rounded-xl border border-zinc-800/80 bg-zinc-900/40 space-y-4 shadow-xl">
        <h3 className="text-sm font-semibold text-white flex items-center gap-2">
          <Layers size={16} className="text-indigo-400" /> Dispatch Background Worker Tasks
        </h3>

        <div className="flex flex-wrap gap-3">
          <button
            onClick={() => triggerJob("rag_reindex")}
            disabled={isEnqueueing}
            className="flex items-center gap-2 px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-medium shadow-md shadow-indigo-600/20 disabled:opacity-50 transition-all"
          >
            <Database size={14} /> Re-index pgvector Embeddings
          </button>
          <button
            onClick={() => triggerJob("document_pdf_export")}
            disabled={isEnqueueing}
            className="flex items-center gap-2 px-4 py-2 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-zinc-200 text-xs font-medium transition-all"
          >
            <Play size={14} /> Export Workspace Documents to PDF
          </button>
        </div>
      </div>

      {/* Real-time Jobs Feed Table */}
      <div className="rounded-xl border border-zinc-800/80 bg-zinc-900/40 overflow-hidden shadow-xl">
        <div className="px-4 py-3 border-b border-zinc-800 flex items-center justify-between bg-zinc-950/60 text-xs font-medium text-zinc-400">
          <span>Async Job Queue Telemetry ({jobs.length})</span>
          <span className="font-mono text-[10px] text-emerald-400">Redis 7 Worker Pool</span>
        </div>

        <div className="divide-y divide-zinc-800/60">
          {jobs.map((j) => (
            <div key={j.id} className="p-4 flex items-center justify-between hover:bg-zinc-800/30 transition-colors">
              <div className="space-y-1">
                <div className="flex items-center gap-2">
                  <span className="text-xs font-semibold text-zinc-200 font-mono">{j.id}</span>
                  <span className="px-2 py-0.5 rounded text-[10px] font-mono bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 uppercase font-bold">
                    {j.type}
                  </span>
                </div>
                {j.result && <p className="text-xs text-zinc-400 font-mono">{j.result}</p>}
              </div>

              <div className="flex items-center gap-3">
                <span className="text-[11px] text-zinc-500 font-mono">{j.created_at}</span>
                <span
                  className={`px-2 py-0.5 rounded text-[10px] font-mono border font-bold ${
                    j.status === "COMPLETED"
                      ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20"
                      : "bg-amber-500/10 text-amber-400 border-amber-500/20 animate-pulse"
                  }`}
                >
                  {j.status}
                </span>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
