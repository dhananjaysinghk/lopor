"use client";

import React, { useState } from "react";
import { Database, Network, Share2, Info, FileText, Bot, User, Sparkles, Filter } from "lucide-react";
import { useAuth } from "@/context/AuthContext";

interface Node {
  id: string;
  label: string;
  type: "document" | "agent" | "vector_chunk" | "user";
  color: string;
}

interface Edge {
  source: string;
  target: string;
  label: string;
  weight: number;
}

export default function GraphPage() {
  const { activeWorkspace } = useAuth();
  const [nodes, setNodes] = useState<Node[]>([
    { id: "doc_1", label: "System Architecture PRD", type: "document", color: "#6366f1" },
    { id: "doc_2", label: "PostgreSQL pgvector Migration Guide", type: "document", color: "#10b981" },
    { id: "agent_1", label: "Code Security Reviewer Agent", type: "agent", color: "#a855f7" },
    { id: "vec_1", label: "pgvector HNSW Cosine Vector", type: "vector_chunk", color: "#f59e0b" },
    { id: "usr_1", label: "Sarah Connor (Workspace Owner)", type: "user", color: "#ec4899" },
  ]);

  const [selectedNode, setSelectedNode] = useState<Node | null>(nodes[0]);

  return (
    <div className="max-w-6xl mx-auto space-y-6">
      {/* Top Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-white tracking-tight flex items-center gap-2">
            <Network size={20} className="text-indigo-400" />
            Interactive Workspace Knowledge Graph Visualizer
          </h1>
          <p className="text-xs text-zinc-400 mt-1">
            Visual connection map displaying relationships between Documents, AI Agents, Vector Embeddings, and Users.
          </p>
        </div>
      </div>

      {/* Main Canvas & Inspector Layout */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 h-[calc(100vh-14rem)]">
        {/* Graph Canvas Container */}
        <div className="md:col-span-2 rounded-xl border border-zinc-800/80 bg-zinc-950 p-6 relative overflow-hidden flex items-center justify-center shadow-2xl">
          <div className="absolute top-4 left-4 flex items-center gap-2 bg-zinc-900/80 border border-zinc-800 px-3 py-1.5 rounded-lg text-[11px] font-mono text-zinc-400">
            <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" /> 5 Connected Relational Nodes
          </div>

          {/* SVG Force-Graph Representation */}
          <svg className="w-full h-full">
            <line x1="200" y1="120" x2="380" y2="240" stroke="#334155" strokeWidth="2" strokeDasharray="4" />
            <line x1="380" y1="240" x2="160" y2="340" stroke="#334155" strokeWidth="2" />
            <line x1="380" y1="240" x2="520" y2="300" stroke="#334155" strokeWidth="2" />

            {/* Nodes */}
            <g
              onClick={() => setSelectedNode(nodes[0])}
              className="cursor-pointer group"
              transform="translate(200, 120)"
            >
              <circle r="22" fill="#6366f1" opacity="0.9" className="group-hover:scale-110 transition-transform" />
              <text y="36" textAnchor="middle" fill="#cbd5e1" fontSize="10" fontFamily="monospace">
                PRD Spec
              </text>
            </g>

            <g
              onClick={() => setSelectedNode(nodes[1])}
              className="cursor-pointer group"
              transform="translate(380, 240)"
            >
              <circle r="22" fill="#10b981" opacity="0.9" className="group-hover:scale-110 transition-transform" />
              <text y="36" textAnchor="middle" fill="#cbd5e1" fontSize="10" fontFamily="monospace">
                pgvector Guide
              </text>
            </g>

            <g
              onClick={() => setSelectedNode(nodes[2])}
              className="cursor-pointer group"
              transform="translate(160, 340)"
            >
              <circle r="22" fill="#a855f7" opacity="0.9" className="group-hover:scale-110 transition-transform" />
              <text y="36" textAnchor="middle" fill="#cbd5e1" fontSize="10" fontFamily="monospace">
                Security Agent
              </text>
            </g>

            <g
              onClick={() => setSelectedNode(nodes[3])}
              className="cursor-pointer group"
              transform="translate(520, 300)"
            >
              <circle r="22" fill="#f59e0b" opacity="0.9" className="group-hover:scale-110 transition-transform" />
              <text y="36" textAnchor="middle" fill="#cbd5e1" fontSize="10" fontFamily="monospace">
                Vector Embedding
              </text>
            </g>
          </svg>
        </div>

        {/* Selected Node Details Inspector */}
        <div className="p-6 rounded-xl border border-zinc-800/80 bg-zinc-900/40 space-y-4 shadow-xl">
          <div className="flex items-center gap-2 text-xs font-semibold text-zinc-300 border-b border-zinc-800 pb-3">
            <Info size={16} className="text-indigo-400" /> Relational Node Inspector
          </div>

          {selectedNode ? (
            <div className="space-y-4 text-xs">
              <div>
                <span className="text-[10px] text-zinc-500 font-mono uppercase">Node Title</span>
                <p className="text-sm font-bold text-white mt-0.5">{selectedNode.label}</p>
              </div>

              <div>
                <span className="text-[10px] text-zinc-500 font-mono uppercase">Entity Type</span>
                <p className="mt-1">
                  <span
                    className="px-2 py-0.5 rounded text-[10px] font-mono uppercase font-bold"
                    style={{ backgroundColor: `${selectedNode.color}20`, color: selectedNode.color }}
                  >
                    {selectedNode.type}
                  </span>
                </p>
              </div>

              <div className="p-3.5 rounded-lg bg-zinc-950 border border-zinc-800 space-y-2">
                <span className="text-[10px] text-indigo-400 font-mono font-bold">Connections (Graph Edges):</span>
                <ul className="space-y-1 text-[11px] text-zinc-400 font-mono">
                  <li>➔ created_by: Sarah Connor</li>
                  <li>➔ vector_similar_to: pgvector Guide</li>
                  <li>➔ audited_by: Security Agent</li>
                </ul>
              </div>
            </div>
          ) : (
            <p className="text-xs text-zinc-500">Click any graph node to inspect relationships</p>
          )}
        </div>
      </div>
    </div>
  );
}
