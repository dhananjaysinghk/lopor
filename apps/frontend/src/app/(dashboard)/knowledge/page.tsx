"use client";

import React from "react";
import { Database, Folder, Bookmark, Tag, Sparkles, FileText, ArrowRight } from "lucide-react";
import Link from "next/link";

export default function KnowledgePage() {
  const collections = [
    { title: "System Architecture & Specs", count: "8 Documents", color: "from-indigo-500/20 to-purple-500/10", border: "border-indigo-500/30" },
    { title: "Database Schemas & RAG Vector", count: "14 Documents", color: "from-emerald-500/20 to-teal-500/10", border: "border-emerald-500/30" },
    { title: "Frontend Design System", count: "6 Documents", color: "from-amber-500/20 to-orange-500/10", border: "border-amber-500/30" },
  ];

  return (
    <div className="max-w-5xl mx-auto space-y-8">
      <div>
        <h1 className="text-xl font-bold text-white tracking-tight flex items-center gap-2">
          <Database size={20} className="text-indigo-400" />
          Workspace Knowledge Base
        </h1>
        <p className="text-xs text-zinc-400 mt-1">
          Organized knowledge collections, tags, and AI-indexed document hubs for team collaboration.
        </p>
      </div>

      {/* Collection Cards Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {collections.map((col, idx) => (
          <div
            key={idx}
            className={`p-6 rounded-xl bg-gradient-to-br ${col.color} border ${col.border} backdrop-blur-xl flex flex-col justify-between group hover:scale-[1.02] transition-all shadow-xl`}
          >
            <div>
              <div className="w-10 h-10 rounded-lg bg-zinc-900/80 flex items-center justify-center text-white mb-4 shadow-md">
                <Folder size={20} className="text-indigo-400" />
              </div>
              <h3 className="text-sm font-semibold text-white">{col.title}</h3>
              <p className="text-xs text-zinc-400 mt-1 font-mono">{col.count}</p>
            </div>
            <Link
              href="/documents"
              className="flex items-center gap-1 text-xs text-indigo-400 hover:text-indigo-300 font-medium mt-6 group-hover:translate-x-1 transition-transform"
            >
              Explore Collection <ArrowRight size={14} />
            </Link>
          </div>
        ))}
      </div>
    </div>
  );
}
