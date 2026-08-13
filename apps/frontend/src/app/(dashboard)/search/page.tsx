"use client";

import React, { useState } from "react";
import { Search, Sparkles, Database, FileText, ArrowUpRight, Zap, Check } from "lucide-react";
import { apiFetch } from "@/lib/api";
import { useAuth } from "@/context/AuthContext";

interface VectorResult {
  id: string;
  chunk_index: number;
  chunk_text: string;
  similarity_score: number;
}

export default function SearchPage() {
  const { activeWorkspace } = useAuth();
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<VectorResult[]>([
    {
      id: "vec_1",
      chunk_index: 0,
      chunk_text:
        "Lopor uses Modular Monolith architecture in Go Fiber paired with PostgreSQL 16 + pgvector cosine similarity index HNSW with m=16, ef_construction=64.",
      similarity_score: 0.942,
    },
    {
      id: "vec_2",
      chunk_index: 1,
      chunk_text:
        "Authentication uses JWT Access Tokens (15 min expiration) and Refresh Tokens stored in HTTP-Only cookies with Argon2id password hashing.",
      similarity_score: 0.887,
    },
  ]);
  const [isSearching, setIsSearching] = useState(false);

  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!query.trim()) return;

    setIsSearching(true);
    setTimeout(() => {
      setIsSearching(false);
    }, 600);
  };

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      {/* Top Search Title */}
      <div className="text-center space-y-2 py-4">
        <h1 className="text-2xl font-bold text-white tracking-tight flex items-center justify-center gap-2">
          <Database size={24} className="text-indigo-400" />
          pgvector Hybrid Semantic Search
        </h1>
        <p className="text-xs text-zinc-400">
          Query your entire workspace vector embeddings instantly using dense cosine similarity matching.
        </p>
      </div>

      {/* Main Search Input Form */}
      <form onSubmit={handleSearch} className="relative">
        <Search size={18} className="absolute left-4 top-1/2 -translate-y-1/2 text-zinc-500" />
        <input
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Ask anything (e.g. 'How does the Go JWT token refresh strategy work?')..."
          className="w-full bg-zinc-900/80 border border-zinc-800 rounded-xl pl-11 pr-24 py-3.5 text-xs md:text-sm text-zinc-100 placeholder-zinc-500 focus:outline-none focus:border-indigo-500 transition-colors shadow-2xl"
        />
        <button
          type="submit"
          disabled={isSearching || !query.trim()}
          className="absolute right-2 top-1/2 -translate-y-1/2 px-4 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-medium transition-all shadow-md shadow-indigo-600/20 disabled:opacity-50"
        >
          {isSearching ? "Searching..." : "Search"}
        </button>
      </form>

      {/* Vector Results List */}
      <div className="space-y-4 pt-4">
        <div className="flex items-center justify-between text-xs font-medium text-zinc-400 px-1">
          <span>Vector Match Candidates ({results.length})</span>
          <span className="font-mono text-[10px] text-indigo-400">Similarity Metric: Cosine Distance</span>
        </div>

        {results.map((res) => (
          <div
            key={res.id}
            className="p-5 rounded-xl bg-zinc-900/50 border border-zinc-800/80 hover:border-indigo-500/40 transition-all space-y-3 shadow-md"
          >
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <FileText size={14} className="text-indigo-400" />
                <span className="text-xs font-semibold text-zinc-200">Chunk #{res.chunk_index + 1}</span>
              </div>
              <span className="px-2 py-0.5 rounded text-[10px] font-mono bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 font-bold">
                {(res.similarity_score * 100).toFixed(1)}% Match
              </span>
            </div>

            <p className="text-xs text-zinc-300 leading-relaxed font-mono bg-zinc-950 p-3 rounded-lg border border-zinc-800/60">
              "{res.chunk_text}"
            </p>
          </div>
        ))}
      </div>
    </div>
  );
}
