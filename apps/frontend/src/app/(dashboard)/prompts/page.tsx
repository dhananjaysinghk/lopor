"use client";

import React, { useState } from "react";
import { Sparkles, Plus, Code, Copy, Check, Play, BookOpen, Terminal, Variable } from "lucide-react";
import { useAuth } from "@/context/AuthContext";

interface PromptTemplate {
  id: string;
  title: string;
  category: string;
  template: string;
  variables: string[];
}

export default function PromptsPage() {
  const { activeWorkspace } = useAuth();
  const [templates, setTemplates] = useState<PromptTemplate[]>([
    {
      id: "1",
      title: "Go Security & Performance Refactor",
      category: "Engineering",
      template: "Review the following Go code in {{language}}:\n\n```go\n{{code}}\n```\n\nProvide performance optimizations and security checks for {{target}}.",
      variables: ["language", "code", "target"],
    },
    {
      id: "2",
      title: "RAG Context Summarizer",
      category: "Research",
      template: "Summarize the vector search results below for workspace {{workspace}}:\n\n{{context}}\n\nTarget audience: {{audience}}.",
      variables: ["workspace", "context", "audience"],
    },
  ]);

  const [activeTemplate, setActiveTemplate] = useState<PromptTemplate>(templates[0]);
  const [variableValues, setVariableValues] = useState<Record<string, string>>({
    language: "Go 1.22",
    code: "func Process() {\n    // DB query\n}",
    target: "PostgreSQL pgvector",
  });
  const [hydratedText, setHydratedText] = useState("");

  const handleTestSubstitution = () => {
    let result = activeTemplate.template;
    activeTemplate.variables.forEach((v) => {
      const val = variableValues[v] || `{{${v}}}`;
      result = result.replaceAll(`{{${v}}}`, val);
    });
    setHydratedText(result);
  };

  return (
    <div className="max-w-6xl mx-auto space-y-6">
      {/* Top Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-white tracking-tight flex items-center gap-2">
            <Sparkles size={20} className="text-amber-400" />
            AI Prompt Engineering Studio & Shared Templates
          </h1>
          <p className="text-xs text-zinc-400 mt-1">
            Build reusable prompt templates with variable substitution ({"`{{var}}`"}), team sharing, and instant model execution.
          </p>
        </div>
        <button className="flex items-center gap-2 px-3 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-medium shadow-md shadow-indigo-600/20 transition-all">
          <Plus size={14} /> New Prompt Template
        </button>
      </div>

      {/* Main Studio Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {/* Template Selector Library */}
        <div className="p-4 rounded-xl border border-zinc-800/80 bg-zinc-900/40 space-y-3">
          <span className="text-xs font-semibold text-zinc-300 flex items-center gap-1.5">
            <BookOpen size={14} className="text-indigo-400" /> Shared Template Store
          </span>

          <div className="space-y-2">
            {templates.map((tpl) => (
              <div
                key={tpl.id}
                onClick={() => {
                  setActiveTemplate(tpl);
                  const initialVars: Record<string, string> = {};
                  tpl.variables.forEach((v) => (initialVars[v] = ""));
                  setVariableValues(initialVars);
                  setHydratedText("");
                }}
                className={`p-3 rounded-lg text-xs cursor-pointer border transition-all ${
                  activeTemplate.id === tpl.id
                    ? "bg-zinc-800 border-indigo-500/50 text-white font-medium shadow-md"
                    : "bg-zinc-950/60 border-zinc-800/80 text-zinc-400 hover:text-zinc-200 hover:border-zinc-700"
                }`}
              >
                <div className="flex items-center justify-between">
                  <span className="truncate font-semibold">{tpl.title}</span>
                  <span className="px-1.5 py-0.5 text-[9px] font-mono uppercase bg-amber-500/10 text-amber-400 border border-amber-500/20 rounded">
                    {tpl.category}
                  </span>
                </div>
                <p className="text-[11px] text-zinc-500 font-mono mt-2 truncate">
                  Vars: {tpl.variables.map((v) => `{{${v}}}`).join(", ")}
                </p>
              </div>
            ))}
          </div>
        </div>

        {/* Interactive Template Playground */}
        <div className="md:col-span-2 p-6 rounded-xl border border-zinc-800/80 bg-zinc-900/40 space-y-6 shadow-2xl">
          <div>
            <h3 className="text-sm font-semibold text-white">{activeTemplate.title}</h3>
            <p className="text-xs text-zinc-400 mt-1 font-mono bg-zinc-950 p-3 rounded-lg border border-zinc-800">
              {activeTemplate.template}
            </p>
          </div>

          {/* Variable Inputs */}
          <div className="space-y-3">
            <h4 className="text-xs font-semibold text-zinc-300 flex items-center gap-1.5">
              <Variable size={14} className="text-indigo-400" /> Template Variables ({activeTemplate.variables.length})
            </h4>

            <div className="grid grid-cols-1 gap-3">
              {activeTemplate.variables.map((v) => (
                <div key={v}>
                  <label className="block text-[11px] font-mono text-indigo-300 mb-1">
                    {"{{"}
                    {v}
                    {"}}"}
                  </label>
                  <input
                    type="text"
                    value={variableValues[v] || ""}
                    onChange={(e) => setVariableValues({ ...variableValues, [v]: e.target.value })}
                    placeholder={`Enter value for ${v}...`}
                    className="w-full bg-zinc-950/80 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-zinc-100 placeholder-zinc-500 focus:outline-none focus:border-indigo-500"
                  />
                </div>
              ))}
            </div>

            <button
              onClick={handleTestSubstitution}
              className="flex items-center gap-2 px-4 py-2 rounded-lg bg-amber-600 hover:bg-amber-500 text-white text-xs font-medium shadow-md shadow-amber-600/20 transition-all mt-2"
            >
              <Play size={12} /> Hydrate Prompt Template
            </button>
          </div>

          {/* Hydrated Output Preview */}
          {hydratedText && (
            <div className="p-4 rounded-lg bg-zinc-950 border border-zinc-800 font-mono text-xs space-y-2">
              <span className="text-[10px] text-emerald-400 font-bold uppercase">Hydrated System Prompt Output:</span>
              <p className="whitespace-pre-wrap text-zinc-200">{hydratedText}</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
