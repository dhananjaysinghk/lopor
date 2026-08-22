"use client";

import React, { useState } from "react";
import { User, Key, Shield, Moon, Globe, Copy, Check, Plus, Trash2, Sliders } from "lucide-react";
import { AIProvidersVault } from "@/components/settings/AIProvidersVault";
import { useAuth } from "@/context/AuthContext";

export default function SettingsPage() {
  const { user } = useAuth();
  const [activeTab, setActiveTab] = useState<"profile" | "api-keys" | "security">("profile");
  const [apiKey, setApiKey] = useState("lopor_live_9f823a104e7c99a215b3");
  const [copied, setCopied] = useState(false);

  const copyApiKey = () => {
    navigator.clipboard.writeText(apiKey);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <div>
        <h1 className="text-xl font-bold text-white tracking-tight flex items-center gap-2">
          <Sliders size={20} className="text-indigo-400" />
          User & Workspace Settings
        </h1>
        <p className="text-xs text-zinc-400 mt-1">
          Manage your profile, API keys, security sessions, and workspace preferences.
        </p>
      </div>

      {/* Settings Navigation Tabs */}
      <div className="flex border-b border-zinc-800/80 gap-6 text-xs font-medium">
        <button
          onClick={() => setActiveTab("profile")}
          className={`pb-3 transition-colors flex items-center gap-2 border-b-2 ${
            activeTab === "profile"
              ? "border-indigo-500 text-white font-semibold"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <User size={14} /> Profile & Theme
        </button>
        <button
          onClick={() => setActiveTab("api-keys")}
          className={`pb-3 transition-colors flex items-center gap-2 border-b-2 ${
            activeTab === "api-keys"
              ? "border-indigo-500 text-white font-semibold"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Key size={14} /> API Keys
        </button>
        <button
          onClick={() => setActiveTab("security")}
          className={`pb-3 transition-colors flex items-center gap-2 border-b-2 ${
            activeTab === "security"
              ? "border-indigo-500 text-white font-semibold"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Shield size={14} /> Security & Sessions
        </button>
      </div>

      {/* Tab Content Panels */}
      {activeTab === "profile" && (
        <div className="p-6 rounded-xl border border-zinc-800/80 bg-zinc-900/40 space-y-4 shadow-xl">
          <div>
            <label className="block text-xs font-medium text-zinc-300 mb-1">Full Name</label>
            <input
              type="text"
              defaultValue={user?.full_name || "Enterprise User"}
              className="w-full max-w-md bg-zinc-950/80 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-zinc-100 focus:outline-none focus:border-indigo-500"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-zinc-300 mb-1">Email Address</label>
            <input
              type="email"
              disabled
              defaultValue={user?.email || "user@lopor.ai"}
              className="w-full max-w-md bg-zinc-950/40 border border-zinc-800/60 rounded-lg px-3 py-2 text-xs text-zinc-500 cursor-not-allowed"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-zinc-300 mb-1">Role Permission</label>
            <span className="inline-flex px-2.5 py-1 rounded text-xs font-mono bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 uppercase font-bold">
              {user?.role || "workspace_owner"}
            </span>
          </div>
        </div>
      )}

      {activeTab === "api-keys" && (
        <div className="space-y-6">
          <AIProvidersVault />

          <div className="p-6 rounded-xl border border-zinc-800/80 bg-zinc-900/40 space-y-4 shadow-xl">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-sm font-semibold text-white">Active Secret API Keys</h3>
                <p className="text-xs text-zinc-400 mt-0.5">Use keys to authenticate REST and RAG endpoints programmatically.</p>
              </div>
              <button className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-medium shadow-md shadow-indigo-600/20 transition-all">
                <Plus size={14} /> Create Secret Key
              </button>
            </div>

            <div className="p-4 rounded-lg bg-zinc-950 border border-zinc-800 flex items-center justify-between font-mono text-xs">
              <span className="text-indigo-300">{apiKey}</span>
              <button
                onClick={copyApiKey}
                className="flex items-center gap-1 px-2.5 py-1 rounded bg-zinc-800 text-zinc-400 hover:text-white text-[11px]"
              >
                {copied ? <Check size={12} className="text-emerald-400" /> : <Copy size={12} />}
                <span>{copied ? "Copied" : "Copy Key"}</span>
              </button>
            </div>
          </div>
        </div>
      )}

      {activeTab === "security" && (
        <div className="p-6 rounded-xl border border-zinc-800/80 bg-zinc-900/40 space-y-4 shadow-xl">
          <h3 className="text-sm font-semibold text-white">Active Browser Sessions</h3>
          <div className="p-4 rounded-lg bg-zinc-950 border border-zinc-800 flex items-center justify-between text-xs">
            <div>
              <p className="font-semibold text-zinc-200">Chrome on Windows (Current Session)</p>
              <p className="text-[11px] text-zinc-500 mt-0.5 font-mono">127.0.0.1 • Argon2id Hashed Token</p>
            </div>
            <span className="text-[10px] font-mono text-emerald-400 bg-emerald-500/10 border border-emerald-500/20 px-2 py-0.5 rounded">
              Active
            </span>
          </div>
        </div>
      )}
    </div>
  );
}
