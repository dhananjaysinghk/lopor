"use client";

import React, { useState } from "react";
import { Sparkles, Send, Bot, User, RefreshCw, Pin, Trash2, Plus, Code, Lightbulb, FileText, Check } from "lucide-react";
import { MarkdownRenderer } from "@/components/chat/MarkdownRenderer";
import { useAuth } from "@/context/AuthContext";

interface Message {
  id: string;
  sender_role: "user" | "assistant";
  content: string;
}

export default function ChatPage() {
  const { activeWorkspace } = useAuth();
  const [messages, setMessages] = useState<Message[]>([
    {
      id: "1",
      sender_role: "assistant",
      content:
        "Hello! I am your **Lopor AI Assistant**. How can I assist with your code, technical documentation, or RAG workspace files today?",
    },
  ]);
  const [input, setInput] = useState("");
  const [isStreaming, setIsStreaming] = useState(false);
  const [selectedModel, setSelectedModel] = useState("gemini-3.6-flash");

  const promptTemplates = [
    { title: "Code Review", prompt: "Perform a security & performance code review on this Go struct:", icon: Code },
    { title: "Brainstorm RAG", prompt: "Brainstorm architectural improvements for pgvector cosine indexing:", icon: Lightbulb },
    { title: "Summarize Docs", prompt: "Provide a concise executive summary of our workspace architecture:", icon: FileText },
  ];

  const handleSend = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    if (!input.trim() || isStreaming) return;

    const userMsg: Message = {
      id: Date.now().toString(),
      sender_role: "user",
      content: input,
    };

    setMessages((prev) => [...prev, userMsg]);
    const promptText = input;
    setInput("");
    setIsStreaming(true);

    const assistantMsgId = (Date.now() + 1).toString();
    const initialAssistantMsg: Message = {
      id: assistantMsgId,
      sender_role: "assistant",
      content: "",
    };
    setMessages((prev) => [...prev, initialAssistantMsg]);

    try {
      const wsId = activeWorkspace?.id || "00000000-0000-0000-0000-000000000001";
      const chatId = "00000000-0000-0000-0000-000000000002";
      const response = await fetch(
        `http://localhost:8080/api/v1/workspaces/${wsId}/chats/${chatId}/stream`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            prompt: promptText,
            model: selectedModel,
          }),
        }
      );

      if (!response.ok || !response.body) {
        throw new Error(`Backend error: ${response.statusText}`);
      }

      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      let accumulatedContent = "";

      while (true) {
        const { value, done } = await reader.read();
        if (done) break;

        const chunk = decoder.decode(value, { stream: true });
        const lines = chunk.split("\n");

        for (const line of lines) {
          if (line.startsWith("data: ")) {
            const dataStr = line.slice(6).trim();
            if (dataStr === "[DONE]") continue;
            try {
              const data = JSON.parse(dataStr);
              const deltaContent =
                data.delta ||
                data.content ||
                data.choices?.[0]?.delta?.content ||
                "";
              if (deltaContent) {
                accumulatedContent += deltaContent;
                setMessages((prev) =>
                  prev.map((msg) =>
                    msg.id === assistantMsgId
                      ? { ...msg, content: accumulatedContent }
                      : msg
                  )
                );
              }
            } catch {
              // Ignore non-JSON lines or partial buffers
            }
          }
        }
      }
    } catch (err: any) {
      console.error("Chat streaming error:", err);
      setMessages((prev) =>
        prev.map((msg) =>
          msg.id === assistantMsgId
            ? {
                ...msg,
                content: `⚠️ **Error**: ${err?.message || "Failed to connect to backend engine at http://localhost:8080."}`,
              }
            : msg
        )
      );
    } finally {
      setIsStreaming(false);
    }
  };

  return (
    <div className="flex h-[calc(100vh-5rem)] gap-4 overflow-hidden">
      {/* Left Chat History Sidebar */}
      <div className="w-64 hidden md:flex flex-col border border-zinc-800/80 bg-zinc-900/40 rounded-xl p-3">
        <button className="flex items-center justify-center gap-2 py-2 px-3 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-medium transition-all shadow-md shadow-indigo-600/20 mb-3">
          <Plus size={14} /> New Chat
        </button>

        <div className="flex-1 space-y-1 overflow-y-auto pr-1">
          <p className="px-2 py-1 text-[10px] uppercase font-mono text-zinc-500 font-bold">Pinned</p>
          <div className="flex items-center justify-between p-2 rounded-lg bg-zinc-800/60 text-xs text-zinc-200 border border-zinc-700/50 cursor-pointer">
            <span className="truncate">Go Fiber API Refactor</span>
            <Pin size={12} className="text-amber-400 shrink-0" />
          </div>

          <p className="px-2 py-1 text-[10px] uppercase font-mono text-zinc-500 font-bold mt-4">Recent</p>
          <div className="p-2 rounded-lg text-xs text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/40 cursor-pointer truncate">
            pgvector HNSW vs IVFFlat
          </div>
          <div className="p-2 rounded-lg text-xs text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/40 cursor-pointer truncate">
            RAG Chunking Strategy
          </div>
        </div>
      </div>

      {/* Main Chat Viewport */}
      <div className="flex-1 flex flex-col border border-zinc-800/80 bg-zinc-900/40 rounded-xl overflow-hidden shadow-2xl relative">
        {/* Model Selector Bar */}
        <div className="h-12 border-b border-zinc-800/80 px-4 flex items-center justify-between bg-zinc-950/60">
          <div className="flex items-center gap-2">
            <span className="text-xs font-medium text-zinc-300">Model:</span>
            <select
              value={selectedModel}
              onChange={(e) => setSelectedModel(e.target.value)}
              className="bg-zinc-900 border border-zinc-800 rounded-md text-xs text-indigo-300 px-2 py-1 focus:outline-none"
            >
              <option value="gemini-3.6-flash">Google Gemini 3.6 Flash (Recommended)</option>
              <option value="gpt-4o">OpenAI GPT-4o</option>
              <option value="claude-3-5">Claude 3.5 Sonnet</option>
              <option value="ollama-llama3">Ollama Llama 3 (Local)</option>
            </select>
          </div>
          <span className="text-[10px] font-mono bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 px-2 py-0.5 rounded">
            SSE Stream Ready
          </span>
        </div>

        {/* Messages Stream Container */}
        <div className="flex-1 overflow-y-auto p-4 md:p-6 space-y-6">
          {messages.map((msg) => (
            <div
              key={msg.id}
              className={`flex gap-3 max-w-3xl ${msg.sender_role === "user" ? "ml-auto flex-row-reverse" : ""
                }`}
            >
              <div
                className={`w-8 h-8 rounded-lg flex items-center justify-center shrink-0 text-xs font-bold ${msg.sender_role === "user"
                    ? "bg-indigo-600 text-white"
                    : "bg-gradient-to-br from-purple-600 to-indigo-600 text-white shadow-md shadow-indigo-500/20"
                  }`}
              >
                {msg.sender_role === "user" ? <User size={16} /> : <Bot size={16} />}
              </div>

              <div
                className={`p-4 rounded-xl text-xs md:text-sm border shadow-sm ${msg.sender_role === "user"
                    ? "bg-indigo-600/15 border-indigo-500/30 text-zinc-100"
                    : "bg-zinc-950/80 border-zinc-800 text-zinc-200"
                  }`}
              >
                <MarkdownRenderer content={msg.content || (isStreaming ? "Thinking..." : "")} />
              </div>
            </div>
          ))}
        </div>

        {/* Prompt Templates Drawer */}
        <div className="px-4 py-2 border-t border-zinc-800/40 bg-zinc-950/40 flex items-center gap-2 overflow-x-auto">
          {promptTemplates.map((tpl, idx) => {
            const Icon = tpl.icon;
            return (
              <button
                key={idx}
                onClick={() => setInput(tpl.prompt)}
                className="flex items-center gap-1.5 px-2.5 py-1 rounded-md bg-zinc-900 border border-zinc-800 text-[11px] text-zinc-400 hover:text-zinc-200 hover:border-zinc-700 transition-all shrink-0"
              >
                <Icon size={12} className="text-indigo-400" />
                <span>{tpl.title}</span>
              </button>
            );
          })}
        </div>

        {/* Input Bar Form */}
        <form onSubmit={handleSend} className="p-3 bg-zinc-950/80 border-t border-zinc-800/80 flex items-center gap-2">
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder="Ask Lopor AI or request code generation..."
            className="flex-1 bg-zinc-900 border border-zinc-800 rounded-lg px-4 py-2.5 text-xs md:text-sm text-zinc-100 placeholder-zinc-500 focus:outline-none focus:border-indigo-500 transition-colors"
          />
          <button
            type="submit"
            disabled={!input.trim() || isStreaming}
            className="p-2.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white shadow-md shadow-indigo-600/20 disabled:opacity-50 transition-all"
          >
            <Send size={16} />
          </button>
        </form>
      </div>
    </div>
  );
}
