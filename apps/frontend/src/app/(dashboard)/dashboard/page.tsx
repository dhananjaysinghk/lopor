"use client";

import React from "react";
import { Sparkles, FileText, Bot, Database, ArrowUpRight, Zap, FolderKanban } from "lucide-react";
import Link from "next/link";
import { useAuth } from "@/context/AuthContext";

export default function DashboardPage() {
  const { user, activeWorkspace } = useAuth();

  const quickStats = [
    { title: "Active AI Models", value: "GPT-4o / pgvector", icon: Zap, color: "text-amber-400" },
    { title: "Workspace Documents", value: "12 Pages", icon: FileText, color: "text-indigo-400" },
    { title: "RAG Vector Store", value: "1,536 Dimensional", icon: Database, color: "text-emerald-400" },
    { title: "Autonomous Agents", value: "3 Ready", icon: Bot, color: "text-purple-400" },
  ];

  return (
    <div className="max-w-6xl mx-auto space-y-8">
      {/* Welcome Banner */}
      <div className="p-8 rounded-2xl bg-gradient-to-r from-indigo-900/40 via-purple-900/20 to-zinc-900 border border-indigo-500/20 relative overflow-hidden shadow-2xl">
        <div className="relative z-10 max-w-2xl">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-indigo-500/10 border border-indigo-500/20 text-indigo-400 text-xs font-mono mb-3">
            <Sparkles size={12} /> Enterprise AI Monolith Engine Active
          </div>
          <h1 className="text-2xl md:text-3xl font-bold text-white tracking-tight">
            Welcome, {user?.full_name || "Knowledge Worker"}
          </h1>
          <p className="text-xs md:text-sm text-zinc-400 mt-2 leading-relaxed">
            Your Lopor workspace is ready. Brainstorm with real-time AI streams, perform pgvector hybrid search over your documents, or configure autonomous background agents.
          </p>

          <div className="flex flex-wrap gap-3 mt-6">
            <Link
              href="/chat"
              className="flex items-center gap-2 px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-medium shadow-lg shadow-indigo-600/25 transition-all"
            >
              <Sparkles size={14} /> Start AI Chat Session
            </Link>
            <Link
              href="/documents"
              className="flex items-center gap-2 px-4 py-2 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-zinc-200 text-xs font-medium transition-all"
            >
              <FileText size={14} /> Create New Document
            </Link>
          </div>
        </div>
      </div>

      {/* Quick Stats Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {quickStats.map((stat, idx) => {
          const Icon = stat.icon;
          return (
            <div
              key={idx}
              className="p-5 rounded-xl bg-zinc-900/60 border border-zinc-800/80 hover:border-zinc-700/80 transition-all shadow-md"
            >
              <div className="flex items-center justify-between">
                <span className="text-xs text-zinc-400 font-medium">{stat.title}</span>
                <Icon size={16} className={stat.color} />
              </div>
              <p className="text-lg font-semibold text-white mt-2">{stat.value}</p>
            </div>
          );
        })}
      </div>

      {/* Feature Tiles Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="p-6 rounded-xl bg-zinc-900/40 border border-zinc-800/60 flex flex-col justify-between group hover:border-zinc-700 transition-all">
          <div>
            <div className="w-10 h-10 rounded-lg bg-indigo-500/10 text-indigo-400 flex items-center justify-center mb-4">
              <Sparkles size={20} />
            </div>
            <h3 className="text-sm font-semibold text-white">AI Chat & Streaming</h3>
            <p className="text-xs text-zinc-400 mt-1">
              Low-latency SSE streaming with prompt history, model switching, and code blocks.
            </p>
          </div>
          <Link
            href="/chat"
            className="flex items-center gap-1 text-xs text-indigo-400 hover:text-indigo-300 font-medium mt-6 group-hover:translate-x-0.5 transition-transform"
          >
            Launch Chat <ArrowUpRight size={14} />
          </Link>
        </div>

        <div className="p-6 rounded-xl bg-zinc-900/40 border border-zinc-800/60 flex flex-col justify-between group hover:border-zinc-700 transition-all">
          <div>
            <div className="w-10 h-10 rounded-lg bg-emerald-500/10 text-emerald-400 flex items-center justify-center mb-4">
              <Database size={20} />
            </div>
            <h3 className="text-sm font-semibold text-white">RAG Vector Search</h3>
            <p className="text-xs text-zinc-400 mt-1">
              Chunk PDF & DOCX files into 1536d pgvector HNSW embeddings with precise citations.
            </p>
          </div>
          <Link
            href="/search"
            className="flex items-center gap-1 text-xs text-emerald-400 hover:text-emerald-300 font-medium mt-6 group-hover:translate-x-0.5 transition-transform"
          >
            Explore Search <ArrowUpRight size={14} />
          </Link>
        </div>

        <div className="p-6 rounded-xl bg-zinc-900/40 border border-zinc-800/60 flex flex-col justify-between group hover:border-zinc-700 transition-all">
          <div>
            <div className="w-10 h-10 rounded-lg bg-purple-500/10 text-purple-400 flex items-center justify-center mb-4">
              <Bot size={20} />
            </div>
            <h3 className="text-sm font-semibold text-white">AI Agents Engine</h3>
            <p className="text-xs text-zinc-400 mt-1">
              Create autonomous task runners with custom system prompts, memory, and tools.
            </p>
          </div>
          <Link
            href="/agents"
            className="flex items-center gap-1 text-xs text-purple-400 hover:text-purple-300 font-medium mt-6 group-hover:translate-x-0.5 transition-transform"
          >
            Manage Agents <ArrowUpRight size={14} />
          </Link>
        </div>
      </div>
    </div>
  );
}
