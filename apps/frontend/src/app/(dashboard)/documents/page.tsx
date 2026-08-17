"use client";

import React, { useState } from "react";
import { FileText, FolderPlus, Plus, Save, Clock, Share2, Sparkles, Folder, Check } from "lucide-react";
import { CollaborativePresence } from "@/components/documents/CollaborativePresence";
import { useAuth } from "@/context/AuthContext";

interface Doc {
  id: string;
  title: string;
  content: string;
  updated_at: string;
}

export default function DocumentsPage() {
  const { activeWorkspace } = useAuth();
  const [docs, setDocs] = useState<Doc[]>([
    {
      id: "1",
      title: "Product Requirements Document (PRD) - Lopor v1",
      content:
        "# Product Requirements Document (PRD)\n\n## Goal\nBuild a commercial-grade AI Workspace with Next.js and Go Fiber.\n\n## Key Modules\n1. AI Streaming Chat\n2. pgvector RAG Engine\n3. Document Editor & Knowledge Graph",
      updated_at: "Just now",
    },
    {
      id: "2",
      title: "Database Schema Migration Guide",
      content: "# Database Schema Guide\n\nAll 18 tables defined with UUID primary keys and pgvector HNSW index.",
      updated_at: "2 hours ago",
    },
  ]);
  const [activeDoc, setActiveDoc] = useState<Doc>(docs[0]);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);

  const handleSave = () => {
    setSaving(true);
    setTimeout(() => {
      setSaving(false);
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    }, 500);
  };

  const createNewDocument = () => {
    const newDoc: Doc = {
      id: Date.now().toString(),
      title: "Untitled Document",
      content: "# New Document\n\nStart writing markdown or ask AI...",
      updated_at: "Just now",
    };
    setDocs([newDoc, ...docs]);
    setActiveDoc(newDoc);
  };

  return (
    <div className="flex h-[calc(100vh-5rem)] gap-4 overflow-hidden">
      {/* Folder & Document Navigator Sidebar */}
      <div className="w-64 hidden md:flex flex-col border border-zinc-800/80 bg-zinc-900/40 rounded-xl p-3">
        <div className="flex items-center justify-between mb-3 px-1">
          <span className="text-xs font-semibold text-zinc-300">Documents</span>
          <button
            onClick={createNewDocument}
            className="p-1 rounded bg-indigo-600 hover:bg-indigo-500 text-white transition-all text-xs flex items-center gap-1"
          >
            <Plus size={14} />
          </button>
        </div>

        <div className="flex-1 space-y-1 overflow-y-auto pr-1">
          <div className="flex items-center gap-2 p-2 rounded-lg text-xs text-zinc-400 font-medium">
            <Folder size={14} className="text-amber-400" /> Engineering Specs
          </div>
          {docs.map((doc) => (
            <div
              key={doc.id}
              onClick={() => setActiveDoc(doc)}
              className={`p-2.5 rounded-lg text-xs cursor-pointer transition-all flex items-center gap-2 truncate ${
                activeDoc.id === doc.id
                  ? "bg-zinc-800 text-white font-medium shadow-sm border border-zinc-700/60"
                  : "text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/40"
              }`}
            >
              <FileText size={14} className="text-indigo-400 shrink-0" />
              <span className="truncate">{doc.title}</span>
            </div>
          ))}
        </div>
      </div>

      {/* Document Workspace Main Editor Area */}
      <div className="flex-1 flex flex-col border border-zinc-800/80 bg-zinc-900/40 rounded-xl overflow-hidden shadow-2xl">
        {/* Editor Action Header Bar */}
        <div className="h-12 border-b border-zinc-800/80 px-6 flex items-center justify-between bg-zinc-950/60">
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2 text-xs text-zinc-400">
              <Clock size={14} /> Updated {activeDoc.updated_at}
            </div>
            <CollaborativePresence documentId={activeDoc.id} />
          </div>

          <div className="flex items-center gap-3">
            <button
              onClick={handleSave}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-xs font-medium text-zinc-200 transition-colors"
            >
              {saved ? (
                <>
                  <Check size={14} className="text-emerald-400" /> Saved
                </>
              ) : (
                <>
                  <Save size={14} /> {saving ? "Saving..." : "Autosave"}
                </>
              )}
            </button>
            <button className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-xs font-medium text-white shadow-md shadow-indigo-600/20 transition-all">
              <Sparkles size={14} /> AI Rewrite
            </button>
          </div>
        </div>

        {/* Title & Rich Content Input */}
        <div className="flex-1 p-6 md:p-8 overflow-y-auto space-y-4">
          <input
            type="text"
            value={activeDoc.title}
            onChange={(e) => {
              const updated = { ...activeDoc, title: e.target.value };
              setActiveDoc(updated);
              setDocs(docs.map((d) => (d.id === updated.id ? updated : d)));
            }}
            placeholder="Untitled Document"
            className="w-full text-2xl md:text-3xl font-bold text-white bg-transparent focus:outline-none placeholder-zinc-600 tracking-tight"
          />

          <textarea
            value={activeDoc.content}
            onChange={(e) => {
              const updated = { ...activeDoc, content: e.target.value };
              setActiveDoc(updated);
              setDocs(docs.map((d) => (d.id === updated.id ? updated : d)));
            }}
            placeholder="Write markdown content or use / to invoke AI actions..."
            className="w-full h-[calc(100%-4rem)] bg-transparent text-xs md:text-sm text-zinc-200 focus:outline-none resize-none font-mono leading-relaxed placeholder-zinc-600"
          />
        </div>
      </div>
    </div>
  );
}
