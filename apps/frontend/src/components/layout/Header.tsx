"use client";

import React, { useState } from "react";
import { Search, Bell, User, LogOut, Command, ChevronDown } from "lucide-react";
import { useAuth } from "@/context/AuthContext";

interface HeaderProps {
  onOpenCommandPalette?: () => void;
}

export function Header({ onOpenCommandPalette }: HeaderProps) {
  const { user, logout, activeWorkspace, workspaces, setActiveWorkspace } = useAuth();
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const [workspaceDropdown, setWorkspaceDropdown] = useState(false);

  return (
    <header className="h-14 border-b border-zinc-800/60 bg-zinc-950/60 backdrop-blur-md px-6 flex items-center justify-between z-20">
      {/* Search Bar / Cmd+K Launcher */}
      <button
        onClick={onOpenCommandPalette}
        className="flex items-center gap-3 px-3 py-1.5 rounded-lg bg-zinc-900/80 border border-zinc-800 text-zinc-400 hover:text-zinc-200 hover:border-zinc-700 transition-all text-xs w-64 md:w-80 shadow-inner"
      >
        <Search size={14} className="text-zinc-500" />
        <span className="flex-1 text-left">Search workspace or ask AI...</span>
        <kbd className="hidden sm:flex items-center gap-0.5 px-1.5 py-0.5 text-[10px] font-mono bg-zinc-800 text-zinc-400 border border-zinc-700 rounded">
          <Command size={10} /> K
        </kbd>
      </button>

      {/* Right Controls */}
      <div className="flex items-center gap-3">
        {/* Notifications */}
        <button className="p-2 rounded-lg text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/60 transition-colors relative">
          <Bell size={16} />
          <span className="absolute top-1.5 right-1.5 w-2 h-2 rounded-full bg-indigo-500" />
        </button>

        {/* User Dropdown */}
        <div className="relative">
          <button
            onClick={() => setDropdownOpen(!dropdownOpen)}
            className="flex items-center gap-2 p-1.5 rounded-lg hover:bg-zinc-900 transition-colors text-xs text-zinc-300"
          >
            <div className="w-7 h-7 rounded-full bg-zinc-800 border border-zinc-700 flex items-center justify-center font-medium text-white text-xs">
              {user?.full_name ? user.full_name[0].toUpperCase() : "U"}
            </div>
            <span className="hidden md:inline font-medium">{user?.full_name || "User Account"}</span>
            <ChevronDown size={14} className="text-zinc-500" />
          </button>

          {dropdownOpen && (
            <div className="absolute right-0 mt-2 w-48 rounded-lg bg-zinc-900 border border-zinc-800 shadow-2xl py-1 z-50 text-xs">
              <div className="px-3 py-2 border-b border-zinc-800">
                <p className="font-medium text-zinc-200">{user?.full_name}</p>
                <p className="text-[11px] text-zinc-500 truncate">{user?.email}</p>
              </div>
              <button
                onClick={() => logout()}
                className="w-full text-left flex items-center gap-2 px-3 py-2 text-rose-400 hover:bg-zinc-800/60 transition-colors"
              >
                <LogOut size={14} />
                <span>Log Out</span>
              </button>
            </div>
          )}
        </div>
      </div>
    </header>
  );
}
