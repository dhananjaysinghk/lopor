"use client";

import React, { useState } from "react";
import { Upload, FileText, CheckCircle, Database, Trash2, Eye, Sparkles, HardDrive } from "lucide-react";
import { useAuth } from "@/context/AuthContext";

interface FileItem {
  id: string;
  filename: string;
  mime_type: string;
  file_size: number;
  created_at: string;
  status: "embedded" | "processing";
}

export default function FilesPage() {
  const { activeWorkspace } = useAuth();
  const [files, setFiles] = useState<FileItem[]>([
    {
      id: "1",
      filename: "System_Architecture_Blueprint_v1.pdf",
      mime_type: "application/pdf",
      file_size: 2450000,
      created_at: new Date().toISOString(),
      status: "embedded",
    },
    {
      id: "2",
      filename: "Database_Migration_pgvector.docx",
      mime_type: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
      file_size: 1120000,
      created_at: new Date().toISOString(),
      status: "embedded",
    },
  ]);
  const [isUploading, setIsUploading] = useState(false);

  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files || e.target.files.length === 0) return;
    setIsUploading(true);

    const uploaded = Array.from(e.target.files).map((f, i) => ({
      id: (Date.now() + i).toString(),
      filename: f.name,
      mime_type: f.type || "application/octet-stream",
      file_size: f.size,
      created_at: new Date().toISOString(),
      status: "embedded" as const,
    }));

    setTimeout(() => {
      setFiles((prev) => [...uploaded, ...prev]);
      setIsUploading(false);
    }, 1000);
  };

  return (
    <div className="max-w-5xl mx-auto space-y-6">
      {/* Top Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-white tracking-tight flex items-center gap-2">
            <HardDrive size={20} className="text-indigo-400" />
            File Storage & RAG Vector Ingestion
          </h1>
          <p className="text-xs text-zinc-400 mt-1">
            Upload PDF, DOCX, TXT, or Code files. Automatically chunked & embedded into PostgreSQL pgvector HNSW index.
          </p>
        </div>
      </div>

      {/* File Drag and Drop Box */}
      <div className="p-8 rounded-xl border border-dashed border-zinc-700/80 bg-zinc-900/40 flex flex-col items-center justify-center text-center relative hover:border-indigo-500/60 transition-all group">
        <input
          type="file"
          multiple
          onChange={handleFileUpload}
          className="absolute inset-0 opacity-0 cursor-pointer z-10"
        />
        <div className="w-12 h-12 rounded-xl bg-indigo-500/10 text-indigo-400 flex items-center justify-center mb-3 group-hover:scale-110 transition-transform">
          <Upload size={22} />
        </div>
        <p className="text-sm font-semibold text-zinc-200">
          {isUploading ? "Chunking & Embedding into pgvector..." : "Click or drag files here to upload"}
        </p>
        <p className="text-xs text-zinc-500 mt-1">Supports PDF, DOCX, Markdown, TXT, Code files (max 50MB)</p>
      </div>

      {/* Uploaded Files Table */}
      <div className="rounded-xl border border-zinc-800/80 bg-zinc-900/40 overflow-hidden shadow-xl">
        <div className="px-4 py-3 border-b border-zinc-800 flex items-center justify-between bg-zinc-950/60 text-xs font-medium text-zinc-400">
          <span>Uploaded Files ({files.length})</span>
          <span className="font-mono text-[10px] text-emerald-400">1536d pgvector HNSW</span>
        </div>

        <div className="divide-y divide-zinc-800/60">
          {files.map((file) => (
            <div key={file.id} className="p-4 flex items-center justify-between hover:bg-zinc-800/30 transition-colors">
              <div className="flex items-center gap-3">
                <div className="w-9 h-9 rounded-lg bg-zinc-800 flex items-center justify-center text-indigo-400">
                  <FileText size={18} />
                </div>
                <div>
                  <p className="text-xs font-semibold text-zinc-200">{file.filename}</p>
                  <p className="text-[10px] text-zinc-500 font-mono mt-0.5">
                    {(file.file_size / 1000000).toFixed(2)} MB • {file.mime_type}
                  </p>
                </div>
              </div>

              <div className="flex items-center gap-4">
                <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[10px] font-mono bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                  <CheckCircle size={10} /> Vector Indexed
                </span>
                <button className="text-zinc-500 hover:text-zinc-300 p-1">
                  <Eye size={14} />
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
