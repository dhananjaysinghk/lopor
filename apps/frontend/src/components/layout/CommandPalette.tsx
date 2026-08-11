"use client";

import React, { useEffect, useState } from "react";
import { Search, Sparkles, FileText, MessageSquare, Database, X } from "lucide-react";
import { useRouter } from "next/navigation";

interface CommandPaletteProps {
  isOpen: boolean;
  onClose: () => void;
}

export function CommandPalette({ isOpen, onClose }: CommandPaletteProps) {
  const [query, setQuery] = useState("");
  const router = useRouter();

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === "k") {
        e.preventDefault();
        if (isOpen) onClose();
        else setQuery("");
      }
      if (e.key === "Escape" && isOpen) {
        onClose();
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  const actions = [
    { label: "New AI Chat Session", icon: Sparkles, path: "/chat" },
    { label: "Create Document", icon: FileText, path: "/documents" },
    { label: "Open Knowledge Base", icon: Database, path: "/knowledge" },
    { label: "Search Documents & Vectors", icon: Search, path: "/search" },
  ];

  const filtered = actions.filter((a) =>
    a.label.toLowerCase().includes(query.toLowerCase())
  );

  return (
    <div className="fixed inset-0 bg-black/70 backdrop-blur-md flex items-start justify-center pt-20 z-50 p-4">
      <div className="w-full max-w-xl bg-zinc-900 border border-zinc-800 rounded-xl shadow-2xl overflow-hidden animate-in fade-in zoom-in-95 duration-150">
        {/* Search Bar Input */}
        <div className="flex items-center px-4 py-3 border-b border-zinc-800 gap-3">
          <Search size={18} className="text-zinc-500" />
          <input
            type="text"
            placeholder="Type a command or search workspace..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            className="w-full bg-transparent text-sm text-zinc-100 placeholder-zinc-500 focus:outline-none"
            autoFocus
          />
          <button onClick={onClose} className="text-zinc-500 hover:text-zinc-300">
            <X size={16} />
          </button>
        </div>

        {/* Action List */}
        <div className="p-2 space-y-1 max-h-80 overflow-y-auto">
          {filtered.length === 0 ? (
            <p className="p-4 text-center text-xs text-zinc-500">No actions found</p>
          ) : (
            filtered.map((action, idx) => {
              const Icon = action.icon;
              return (
                <button
                  key={idx}
                  onClick={() => {
                    router.push(action.path);
                    onClose();
                  }}
                  className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-xs font-medium text-zinc-300 hover:text-white hover:bg-zinc-800/80 transition-colors"
                >
                  <Icon size={16} className="text-indigo-400" />
                  <span>{action.label}</span>
                </button>
              );
            })
          )}
        </div>
      </div>
    </div>
  );
}
