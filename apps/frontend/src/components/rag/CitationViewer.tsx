"use client";

import React, { useState } from "react";
import { FileText, ExternalLink, Bookmark, Check } from "lucide-react";

interface Citation {
  documentTitle: string;
  chunkIndex: number;
  snippetText: string;
  rrfScore: number;
}

interface CitationViewerProps {
  citations: Citation[];
}

export function CitationViewer({ citations }: CitationViewerProps) {
  const [bookmarked, setBookmarked] = useState<Record<number, boolean>>({});

  const toggleBookmark = (idx: number) => {
    setBookmarked((prev) => ({ ...prev, [idx]: !prev[idx] }));
  };

  return (
    <div className="space-y-3 pt-2">
      <div className="flex items-center justify-between text-xs font-medium text-zinc-400">
        <span>RAG Context Sources & Citations ({citations.length})</span>
        <span className="font-mono text-[10px] text-emerald-400">Fused BM25 + Vector RRF</span>
      </div>

      <div className="grid grid-cols-1 gap-2.5">
        {citations.map((c, idx) => (
          <div
            key={idx}
            className="p-3.5 rounded-lg bg-zinc-950/80 border border-zinc-800/80 hover:border-zinc-700 transition-all text-xs space-y-2"
          >
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <FileText size={14} className="text-indigo-400" />
                <span className="font-semibold text-zinc-200">{c.documentTitle}</span>
                <span className="text-[10px] text-zinc-500 font-mono">Chunk #{c.chunkIndex + 1}</span>
              </div>
              <div className="flex items-center gap-2">
                <span className="px-2 py-0.5 rounded text-[10px] font-mono bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 font-bold">
                  {(c.rrfScore * 100).toFixed(1)}% RRF Confidence
                </span>
                <button
                  onClick={() => toggleBookmark(idx)}
                  className="text-zinc-500 hover:text-amber-400 transition-colors p-1"
                >
                  <Bookmark size={12} className={bookmarked[idx] ? "fill-amber-400 text-amber-400" : ""} />
                </button>
              </div>
            </div>

            <p className="text-[11px] text-zinc-400 leading-relaxed font-mono bg-zinc-900/60 p-2.5 rounded border border-zinc-800/40">
              "{c.snippetText}"
            </p>
          </div>
        ))}
      </div>
    </div>
  );
}
