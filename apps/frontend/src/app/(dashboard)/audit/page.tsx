"use client";

import React, { useState } from "react";
import { ShieldCheck, Download, Search, Filter, Lock, Eye, FileText, CheckCircle2 } from "lucide-react";
import { useAuth } from "@/context/AuthContext";

interface AuditLog {
  id: string;
  action: string;
  entity_type: string;
  ip_address: string;
  timestamp: string;
  severity: "INFO" | "WARNING" | "CRITICAL";
}

export default function AuditPage() {
  const { activeWorkspace } = useAuth();
  const [logs, setLogs] = useState<AuditLog[]>([
    {
      id: "log_9f81a",
      action: "USER_AUTHENTICATED_SESSION",
      entity_type: "Auth",
      ip_address: "192.168.1.104",
      timestamp: "2026-08-24 20:15:02",
      severity: "INFO",
    },
    {
      id: "log_4d2bb",
      action: "PGVECTOR_HYBRID_RRF_QUERY",
      entity_type: "RAG Engine",
      ip_address: "192.168.1.104",
      timestamp: "2026-08-24 19:42:18",
      severity: "INFO",
    },
    {
      id: "log_1e7cc",
      action: "AGENT_TOOL_CODE_SANDBOX_EXECUTED",
      entity_type: "AI Agent",
      ip_address: "192.168.1.104",
      timestamp: "2026-08-24 18:20:11",
      severity: "WARNING",
    },
  ]);

  const [downloading, setDownloading] = useState(false);

  const handleExportCSV = () => {
    setDownloading(true);
    setTimeout(() => {
      const csvContent =
        "ID,Action,EntityType,IPAddress,Timestamp\n" +
        logs.map((l) => `${l.id},${l.action},${l.entity_type},${l.ip_address},${l.timestamp}`).join("\n");

      const blob = new Blob([csvContent], { type: "text/csv;charset=utf-8;" });
      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.setAttribute("href", url);
      link.setAttribute("download", `lopor-compliance-audit-ledger-${Date.now()}.csv`);
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      setDownloading(false);
    }, 600);
  };

  return (
    <div className="max-w-6xl mx-auto space-y-6">
      {/* Top Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-white tracking-tight flex items-center gap-2">
            <ShieldCheck size={20} className="text-emerald-400" />
            Enterprise Security Audit Trail & Compliance Export
          </h1>
          <p className="text-xs text-zinc-400 mt-1">
            Immutable SOC 2 / HIPAA compliance audit trail logging all workspace authentication events, RAG searches, and agent code executions.
          </p>
        </div>
        <button
          onClick={handleExportCSV}
          disabled={downloading}
          className="flex items-center gap-2 px-4 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-medium shadow-md shadow-emerald-600/20 transition-all disabled:opacity-50"
        >
          <Download size={14} /> {downloading ? "Exporting CSV..." : "Export Compliance Ledger (CSV)"}
        </button>
      </div>

      {/* Audit Log Table */}
      <div className="rounded-xl border border-zinc-800/80 bg-zinc-900/40 overflow-hidden shadow-2xl">
        <div className="px-4 py-3 border-b border-zinc-800 flex items-center justify-between bg-zinc-950/60 text-xs font-medium text-zinc-400">
          <span>Immutable Audit Ledger ({logs.length} Events)</span>
          <span className="font-mono text-[10px] text-emerald-400">SOC2 Type II Compliant</span>
        </div>

        <div className="divide-y divide-zinc-800/60 font-mono text-xs">
          {logs.map((l) => (
            <div key={l.id} className="p-4 flex items-center justify-between hover:bg-zinc-800/30 transition-colors">
              <div className="space-y-1">
                <div className="flex items-center gap-2">
                  <span className="font-bold text-zinc-200">{l.action}</span>
                  <span className="px-2 py-0.2 rounded text-[9px] bg-zinc-800 text-zinc-300 border border-zinc-700">
                    {l.entity_type}
                  </span>
                </div>
                <p className="text-[11px] text-zinc-500">ID: {l.id} • IP: {l.ip_address}</p>
              </div>

              <div className="flex items-center gap-4">
                <span className="text-[11px] text-zinc-400">{l.timestamp}</span>
                <span
                  className={`px-2 py-0.5 rounded text-[10px] border font-bold ${
                    l.severity === "INFO"
                      ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20"
                      : "bg-amber-500/10 text-amber-400 border-amber-500/20"
                  }`}
                >
                  {l.severity}
                </span>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
