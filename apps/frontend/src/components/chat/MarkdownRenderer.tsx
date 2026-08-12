"use client";

import React from "react";
import { Copy, Check } from "lucide-react";

interface MarkdownRendererProps {
  content: string;
}

export function MarkdownRenderer({ content }: MarkdownRendererProps) {
  const [copied, setCopied] = React.useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText(content);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  // Process code blocks for syntax styling
  const parts = content.split(/(```[\s\S]*?```)/g);

  return (
    <div className="text-xs md:text-sm leading-relaxed text-zinc-200 space-y-3 relative group">
      <button
        onClick={handleCopy}
        className="absolute right-0 top-0 opacity-0 group-hover:opacity-100 p-1.5 rounded bg-zinc-800 text-zinc-400 hover:text-white transition-all text-xs flex items-center gap-1"
      >
        {copied ? <Check size={12} className="text-emerald-400" /> : <Copy size={12} />}
        <span>{copied ? "Copied" : "Copy"}</span>
      </button>

      {parts.map((part, index) => {
        if (part.startsWith("```")) {
          const match = part.match(/^```(\w+)?\n([\s\S]*?)```$/);
          const lang = match ? match[1] || "code" : "code";
          const code = match ? match[2] : part.slice(3, -3);

          return (
            <div key={index} className="my-3 rounded-lg overflow-hidden border border-zinc-800 bg-zinc-950 font-mono text-xs">
              <div className="flex items-center justify-between px-3 py-1.5 bg-zinc-900 border-b border-zinc-800 text-zinc-400 text-[10px] uppercase font-bold">
                <span>{lang}</span>
              </div>
              <pre className="p-3 overflow-x-auto text-indigo-300">
                <code>{code}</code>
              </pre>
            </div>
          );
        }

        return (
          <p key={index} className="whitespace-pre-wrap">
            {part}
          </p>
        );
      })}
    </div>
  );
}
