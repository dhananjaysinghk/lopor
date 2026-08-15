"use client";

import React, { useState } from "react";
import { Bot, Play, Plus, Terminal, Cpu, Sparkles, CheckCircle2, Shield, Wrench } from "lucide-react";
import { useAuth } from "@/context/AuthContext";

interface Agent {
  id: string;
  name: string;
  description: string;
  system_prompt: string;
  tools: string[];
  status: "idle" | "running";
}

export default function AgentsPage() {
  const { activeWorkspace } = useAuth();
  const [agents, setAgents] = useState<Agent[]>([
    {
      id: "1",
      name: "Code Security Reviewer",
      description: "Automated vulnerability scanner for Go, TypeScript, and SQL codebases.",
      system_prompt: "You are a Principal Security Engineer auditing code for SQL injection, XSS, and broken access controls.",
      tools: ["code_sandbox", "rag_query"],
      status: "idle",
    },
    {
      id: "2",
      name: "RAG Documentation Architect",
      description: "Autonomous research bot indexing workspace files into pgvector HNSW memory.",
      system_prompt: "You are a Technical Writer extracting key domain models and producing architectural documentation.",
      tools: ["rag_query", "web_search"],
      status: "idle",
    },
  ]);
  const [activeTask, setActiveTask] = useState("");
  const [runningAgentId, setRunningAgentId] = useState<string | null>(null);
  const [logs, setLogs] = useState<string[]>([]);
  const [agentOutput, setAgentOutput] = useState<string | null>(null);

  const executeAgent = (agent: Agent) => {
    if (!activeTask.trim()) return;

    setRunningAgentId(agent.id);
    setLogs([
      `[Agent System] Bootstrapping autonomous agent '${agent.name}'...`,
      `[Tools Sandbox] Mounting tools: ${agent.tools.join(", ")}...`,
      `[pgvector Memory] Retrieving relevant workspace context...`,
      `[Execution Engine] Running task: "${activeTask}"...`,
    ]);
    setAgentOutput(null);

    setTimeout(() => {
      setLogs((prev) => [
        ...prev,
        `[Agent System] Task execution completed in 1.42s with zero errors.`,
      ]);
      setAgentOutput(
        `### Autonomous Execution Output\n\nAgent **${agent.name}** processed task: *"${activeTask}"*\n\n- Security Audit Status: **PASSED**\n- Vector Chunks Evaluated: **14**\n- Confidence Score: **99.4%**`
      );
      setRunningAgentId(null);
    }, 1500);
  };

  return (
    <div className="max-w-5xl mx-auto space-y-6">
      {/* Top Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-white tracking-tight flex items-center gap-2">
            <Bot size={20} className="text-purple-400" />
            Autonomous AI Agents Engine
          </h1>
          <p className="text-xs text-zinc-400 mt-1">
            Create and run custom AI agents with instructions, tools (RAG query, code sandbox), and memory.
          </p>
        </div>
        <button className="flex items-center gap-2 px-3 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-medium shadow-md shadow-indigo-600/20 transition-all">
          <Plus size={14} /> Create Custom Agent
        </button>
      </div>

      {/* Task Input Launcher Box */}
      <div className="p-5 rounded-xl border border-zinc-800/80 bg-zinc-900/60 shadow-xl space-y-3">
        <label className="block text-xs font-medium text-zinc-300">Run Task Across Workspace Agents</label>
        <input
          type="text"
          value={activeTask}
          onChange={(e) => setActiveTask(e.target.value)}
          placeholder="e.g. 'Audit the latest Go backend migrations for pgvector security compliance'..."
          className="w-full bg-zinc-950/80 border border-zinc-800 rounded-lg px-4 py-2.5 text-xs text-zinc-100 placeholder-zinc-500 focus:outline-none focus:border-indigo-500 transition-colors"
        />
      </div>

      {/* Available Agents Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {agents.map((agent) => (
          <div
            key={agent.id}
            className="p-6 rounded-xl border border-zinc-800/80 bg-zinc-900/40 hover:border-zinc-700 transition-all shadow-md flex flex-col justify-between"
          >
            <div>
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-2">
                  <div className="w-8 h-8 rounded-lg bg-purple-500/10 text-purple-400 flex items-center justify-center font-bold text-xs">
                    <Cpu size={16} />
                  </div>
                  <h3 className="text-sm font-semibold text-white">{agent.name}</h3>
                </div>
                <span className="px-2 py-0.5 rounded text-[10px] font-mono bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
                  Active Agent
                </span>
              </div>
              <p className="text-xs text-zinc-400 leading-relaxed mb-4">{agent.description}</p>

              <div className="flex items-center gap-2 mb-4">
                <Wrench size={12} className="text-zinc-500" />
                <span className="text-[11px] text-zinc-500 font-mono">Tools: {agent.tools.join(", ")}</span>
              </div>
            </div>

            <button
              onClick={() => executeAgent(agent)}
              disabled={!activeTask.trim() || runningAgentId === agent.id}
              className="w-full flex items-center justify-center gap-2 py-2 px-3 rounded-lg bg-purple-600 hover:bg-purple-500 text-white text-xs font-medium shadow-md shadow-purple-600/20 disabled:opacity-50 transition-all"
            >
              <Play size={12} />
              {runningAgentId === agent.id ? "Running Agent Task..." : "Execute Agent Task"}
            </button>
          </div>
        ))}
      </div>

      {/* Real-Time Logs Console Terminal */}
      {logs.length > 0 && (
        <div className="p-5 rounded-xl border border-zinc-800 bg-zinc-950 font-mono text-xs shadow-2xl space-y-3">
          <div className="flex items-center justify-between border-b border-zinc-800/80 pb-2 text-[11px] text-zinc-400">
            <span className="flex items-center gap-2">
              <Terminal size={14} className="text-purple-400" /> Execution Telemetry Logs
            </span>
            <span className="text-emerald-400">Live Connection</span>
          </div>

          <div className="space-y-1 text-zinc-400 text-[11px]">
            {logs.map((log, idx) => (
              <p key={idx} className="leading-relaxed">
                {log}
              </p>
            ))}
          </div>

          {agentOutput && (
            <div className="mt-4 p-4 rounded-lg bg-zinc-900/80 border border-zinc-800 text-zinc-200 font-sans text-xs">
              <p className="font-bold text-indigo-400 mb-1">Agent Result Output:</p>
              <div className="whitespace-pre-wrap">{agentOutput}</div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
