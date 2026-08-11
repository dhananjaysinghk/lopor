"use client";

import React, { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  MessageSquare,
  FileText,
  Database,
  Search,
  Settings,
  Bot,
  ChevronLeft,
  ChevronRight,
  Plus,
  Sparkles,
  LayoutDashboard,
  FolderKanban
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useAuth } from "@/context/AuthContext";

export function Sidebar() {
  const [collapsed, setCollapsed] = useState(false);
  const pathname = usePathname();
  const { activeWorkspace } = useAuth();

  const navItems = [
    { name: "Dashboard", href: "/dashboard", icon: LayoutDashboard },
    { name: "AI Chat", href: "/chat", icon: MessageSquare, badge: "GPT-4o" },
    { name: "Documents", href: "/documents", icon: FileText },
    { name: "Knowledge Base", href: "/knowledge", icon: Database },
    { name: "AI Agents", href: "/agents", icon: Bot },
    { name: "Files & RAG", href: "/files", icon: FolderKanban },
    { name: "Search", href: "/search", icon: Search },
    { name: "Settings", href: "/settings", icon: Settings },
  ];

  return (
    <aside
      className={cn(
        "relative flex flex-col h-screen border-r border-zinc-800/60 bg-zinc-950/80 backdrop-blur-xl transition-all duration-300 z-30 select-none",
        collapsed ? "w-16" : "w-64"
      )}
    >
      {/* Workspace Branding Header */}
      <div className="flex items-center justify-between h-14 px-4 border-b border-zinc-800/60">
        <div className="flex items-center gap-3 overflow-hidden">
          <div className="flex items-center justify-center w-8 h-8 rounded-lg bg-gradient-to-br from-indigo-500 to-purple-600 text-white font-bold text-sm shadow-lg shadow-indigo-500/20">
            L
          </div>
          {!collapsed && (
            <div className="flex flex-col truncate">
              <span className="font-semibold text-sm text-zinc-100 truncate">
                {activeWorkspace?.name || "Lopor AI Workspace"}
              </span>
              <span className="text-[10px] text-zinc-400 font-mono tracking-wider uppercase">
                Enterprise
              </span>
            </div>
          )}
        </div>
        <button
          onClick={() => setCollapsed(!collapsed)}
          className="p-1 rounded-md text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800/60 transition-colors"
        >
          {collapsed ? <ChevronRight size={16} /> : <ChevronLeft size={16} />}
        </button>
      </div>

      {/* Quick Action Button */}
      <div className="p-3">
        <button
          className={cn(
            "w-full flex items-center justify-center gap-2 py-2 px-3 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-medium shadow-md shadow-indigo-600/20 transition-all",
            collapsed && "p-2"
          )}
        >
          <Sparkles size={14} />
          {!collapsed && <span>New AI Chat</span>}
        </button>
      </div>

      {/* Navigation Menu */}
      <nav className="flex-1 px-2 py-2 space-y-1 overflow-y-auto">
        {navItems.map((item) => {
          const Icon = item.icon;
          const isActive = pathname === item.href;
          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                "flex items-center gap-3 px-3 py-2 rounded-md text-xs font-medium transition-all group relative",
                isActive
                  ? "bg-zinc-800/80 text-white shadow-sm"
                  : "text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900/60"
              )}
            >
              <Icon
                size={16}
                className={cn(
                  "transition-colors",
                  isActive ? "text-indigo-400" : "text-zinc-400 group-hover:text-zinc-200"
                )}
              />
              {!collapsed && (
                <div className="flex items-center justify-between w-full truncate">
                  <span className="truncate">{item.name}</span>
                  {item.badge && (
                    <span className="px-1.5 py-0.5 text-[9px] font-mono uppercase bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 rounded">
                      {item.badge}
                    </span>
                  )}
                </div>
              )}
            </Link>
          );
        })}
      </nav>

      {/* Footer Profile Status */}
      <div className="p-3 border-t border-zinc-800/60">
        {!collapsed && (
          <div className="flex items-center justify-between text-xs text-zinc-500 px-1">
            <span className="flex items-center gap-1.5">
              <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
              Engine Online
            </span>
            <span className="font-mono text-[10px]">v1.0.0</span>
          </div>
        )}
      </div>
    </aside>
  );
}
