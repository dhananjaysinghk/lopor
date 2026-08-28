"use client";

import React, { useState } from "react";
import { Terminal, Play, CheckCircle2, AlertCircle, Code, Copy, Check } from "lucide-react";

interface CodeSandboxModalProps {
  initialCode?: string;
  initialLanguage?: string;
}

export function CodeSandboxModal({
  initialCode = 'package main\n\nimport "fmt"\n\nfunc main() {\n    fmt.Println("Hello from Lopor AI Code Sandbox!")\n}',
  initialLanguage = "go",
}: CodeSandboxModalProps) {
  const [code, setCode] = useState(initialCode);
  const [language, setLanguage] = useState(initialLanguage);
  const [output, setOutput] = useState<string | null>(null);
  const [isExecuting, setIsExecuting] = useState(false);
  const [duration, setDuration] = useState("");

  const handleRunCode = () => {
    setIsExecuting(true);
    setOutput(null);

    setTimeout(() => {
      setOutput(`[Lopor Sandbox Engine v1.0]\nExecuting ${language} snippet...\n----------------------------------------\nHello from Lopor AI Code Sandbox!\nProcess finished with exit code 0.`);
      setDuration("0.12s");
      setIsExecuting(false);
    }, 600);
  };

  return (
    <div className="p-6 rounded-xl border border-zinc-800/80 bg-zinc-900/40 space-y-4 shadow-2xl">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold text-white flex items-center gap-2">
          <Terminal size={16} className="text-emerald-400" /> Multi-Language Code Execution Sandbox
        </h3>

        <div className="flex items-center gap-3">
          <select
            value={language}
            onChange={(e) => setLanguage(e.target.value)}
            className="bg-zinc-950/80 border border-zinc-800 rounded-lg px-2.5 py-1 text-xs text-zinc-300 font-mono focus:outline-none"
          >
            <option value="go">Go 1.22</option>
            <option value="python">Python 3.11</option>
            <option value="javascript">Node.js ES2022</option>
            <option value="bash">Bash Shell</option>
          </select>

          <button
            onClick={handleRunCode}
            disabled={isExecuting}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-medium shadow-md shadow-emerald-600/20 transition-all disabled:opacity-50"
          >
            <Play size={14} /> {isExecuting ? "Executing..." : "Run Code"}
          </button>
        </div>
      </div>

      <textarea
        value={code}
        onChange={(e) => setCode(e.target.value)}
        rows={6}
        className="w-full bg-zinc-950 border border-zinc-800 rounded-lg p-3 font-mono text-xs text-indigo-300 focus:outline-none focus:border-indigo-500"
      />

      {output && (
        <div className="p-4 rounded-lg bg-zinc-950 border border-zinc-800 font-mono text-xs space-y-2">
          <div className="flex items-center justify-between text-[10px] text-zinc-500">
            <span className="text-emerald-400 font-bold uppercase">Terminal Output</span>
            <span>Duration: {duration}</span>
          </div>
          <pre className="text-zinc-200 whitespace-pre-wrap">{output}</pre>
        </div>
      )}
    </div>
  );
}
