"use client";

import React, { useState } from "react";
import { Code, Terminal, Download, Copy, Check, Key, BookOpen, ExternalLink, Zap } from "lucide-react";
import { useAuth } from "@/context/AuthContext";

export default function DeveloperPage() {
  const { activeWorkspace } = useAuth();
  const [copied, setCopied] = useState(false);
  const apiKey = "lopor_live_9f81a7b4c2d3e4f5a6b7c8d9e0f1a2b3";

  const handleCopyKey = () => {
    navigator.clipboard.writeText(apiKey);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const curlSample = `curl -X POST "http://localhost:8080/api/v1/workspaces/${activeWorkspace?.id || "ws_default"}/search/hybrid" \\
  -H "X-API-Key: ${apiKey}" \\
  -H "Content-Type: application/json" \\
  -d '{
    "query": "PostgreSQL pgvector HNSW cosine similarity search",
    "top_k": 5
  }'`;

  return (
    <div className="max-w-6xl mx-auto space-y-6">
      {/* Top Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-white tracking-tight flex items-center gap-2">
            <Code size={20} className="text-indigo-400" />
            Public Developer API Hub & OpenAPI 3.0 Documentation
          </h1>
          <p className="text-xs text-zinc-400 mt-1">
            Programmatic access endpoints for pgvector RAG semantic search, autonomous agent executions, and document CRUD.
          </p>
        </div>
        <a
          href="http://localhost:8080/api/v1/developer/openapi.json"
          target="_blank"
          rel="noreferrer"
          className="flex items-center gap-2 px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-medium shadow-md shadow-indigo-600/20 transition-all"
        >
          <Download size={14} /> Download OpenAPI Spec (.json)
        </a>
      </div>

      {/* Secret API Key Banner */}
      <div className="p-6 rounded-xl border border-zinc-800/80 bg-zinc-900/40 space-y-3 shadow-xl">
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-semibold text-white flex items-center gap-2">
            <Key size={16} className="text-amber-400" /> Public Developer Secret API Key
          </h3>
          <span className="px-2 py-0.5 rounded text-[10px] font-mono bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 font-bold">
            Live Production Key
          </span>
        </div>

        <div className="p-4 rounded-lg bg-zinc-950 border border-zinc-800 flex items-center justify-between font-mono text-xs">
          <span className="text-indigo-300">{apiKey}</span>
          <button
            onClick={handleCopyKey}
            className="flex items-center gap-1 px-3 py-1.5 rounded bg-zinc-800 text-zinc-300 hover:text-white text-xs"
          >
            {copied ? <Check size={14} className="text-emerald-400" /> : <Copy size={14} />}
            <span>{copied ? "Copied Key" : "Copy Header Key"}</span>
          </button>
        </div>
      </div>

      {/* cURL Interactive Playground */}
      <div className="p-6 rounded-xl border border-zinc-800/80 bg-zinc-900/40 space-y-4 shadow-xl">
        <h3 className="text-sm font-semibold text-white flex items-center gap-2">
          <Terminal size={16} className="text-indigo-400" /> cURL Request Example (RRF Hybrid Search)
        </h3>

        <div className="p-4 rounded-lg bg-zinc-950 border border-zinc-800 font-mono text-xs overflow-x-auto text-zinc-300 space-y-2">
          <pre className="text-indigo-300">{curlSample}</pre>
        </div>
      </div>
    </div>
  );
}
