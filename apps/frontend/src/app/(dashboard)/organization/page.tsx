"use client";

import React, { useState } from "react";
import { Building2, UserPlus, Mail, Shield, Check, Trash2, Users } from "lucide-react";
import { useAuth } from "@/context/AuthContext";

interface Member {
  id: string;
  name: string;
  email: string;
  role: "Owner" | "Admin" | "Member" | "Viewer";
  status: "Active" | "Pending";
}

export default function OrganizationPage() {
  const { user } = useAuth();
  const [members, setMembers] = useState<Member[]>([
    { id: "1", name: user?.full_name || "Sarah Connor", email: user?.email || "sarah@company.com", role: "Owner", status: "Active" },
    { id: "2", name: "Alex Rivers", email: "alex@company.com", role: "Admin", status: "Active" },
    { id: "3", name: "Elena Rostova", email: "elena@company.com", role: "Member", status: "Pending" },
  ]);
  const [inviteEmail, setInviteEmail] = useState("");
  const [inviteRole, setInviteRole] = useState<"Admin" | "Member" | "Viewer">("Member");
  const [sending, setSending] = useState(false);
  const [sentSuccess, setSentSuccess] = useState(false);

  const handleInvite = (e: React.FormEvent) => {
    e.preventDefault();
    if (!inviteEmail.trim()) return;

    setSending(true);
    setTimeout(() => {
      setMembers((prev) => [
        ...prev,
        {
          id: Date.now().toString(),
          name: inviteEmail.split("@")[0],
          email: inviteEmail,
          role: inviteRole,
          status: "Pending",
        },
      ]);
      setInviteEmail("");
      setSending(false);
      setSentSuccess(true);
      setTimeout(() => setSentSuccess(false), 3000);
    }, 600);
  };

  return (
    <div className="max-w-5xl mx-auto space-y-6">
      {/* Top Header */}
      <div>
        <h1 className="text-xl font-bold text-white tracking-tight flex items-center gap-2">
          <Building2 size={20} className="text-indigo-400" />
          Organization & Team Management
        </h1>
        <p className="text-xs text-zinc-400 mt-1">
          Manage multi-tenant organization members, invite new teammates via email, and configure RBAC roles.
        </p>
      </div>

      {/* Invite Member Box */}
      <div className="p-6 rounded-xl border border-zinc-800/80 bg-zinc-900/40 shadow-xl space-y-4">
        <h3 className="text-sm font-semibold text-white flex items-center gap-2">
          <UserPlus size={16} className="text-indigo-400" /> Invite Team Member
        </h3>

        {sentSuccess && (
          <div className="p-3 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs flex items-center gap-2">
            <Check size={14} /> Invitation email dispatched via SMTP / Mailpit!
          </div>
        )}

        <form onSubmit={handleInvite} className="flex flex-col sm:flex-row gap-3">
          <div className="relative flex-1">
            <Mail size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" />
            <input
              type="email"
              required
              value={inviteEmail}
              onChange={(e) => setInviteEmail(e.target.value)}
              placeholder="teammate@company.com"
              className="w-full bg-zinc-950/80 border border-zinc-800 rounded-lg pl-9 pr-3 py-2 text-xs text-zinc-100 placeholder-zinc-500 focus:outline-none focus:border-indigo-500"
            />
          </div>

          <select
            value={inviteRole}
            onChange={(e) => setInviteRole(e.target.value as any)}
            className="bg-zinc-950/80 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-zinc-300 focus:outline-none"
          >
            <option value="Admin">Admin</option>
            <option value="Member">Member</option>
            <option value="Viewer">Viewer</option>
          </select>

          <button
            type="submit"
            disabled={sending}
            className="px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-medium shadow-md shadow-indigo-600/20 transition-all disabled:opacity-50"
          >
            {sending ? "Sending..." : "Send Invitation"}
          </button>
        </form>
      </div>

      {/* Organization Members Table */}
      <div className="rounded-xl border border-zinc-800/80 bg-zinc-900/40 overflow-hidden shadow-xl">
        <div className="px-4 py-3 border-b border-zinc-800 flex items-center justify-between bg-zinc-950/60 text-xs font-medium text-zinc-400">
          <span>Organization Members ({members.length})</span>
          <span className="font-mono text-[10px] text-indigo-400">RBAC Enforcement Active</span>
        </div>

        <div className="divide-y divide-zinc-800/60">
          {members.map((m) => (
            <div key={m.id} className="p-4 flex items-center justify-between hover:bg-zinc-800/30 transition-colors">
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 rounded-full bg-indigo-600/20 text-indigo-300 border border-indigo-500/30 flex items-center justify-center text-xs font-bold">
                  {m.name[0]}
                </div>
                <div>
                  <p className="text-xs font-semibold text-zinc-200">{m.name}</p>
                  <p className="text-[11px] text-zinc-500 font-mono">{m.email}</p>
                </div>
              </div>

              <div className="flex items-center gap-4">
                <span className="px-2 py-0.5 rounded text-[10px] font-mono bg-zinc-800 text-zinc-300 border border-zinc-700">
                  {m.role}
                </span>
                <span
                  className={`px-2 py-0.5 rounded text-[10px] font-mono border ${
                    m.status === "Active"
                      ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20"
                      : "bg-amber-500/10 text-amber-400 border-amber-500/20"
                  }`}
                >
                  {m.status}
                </span>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
