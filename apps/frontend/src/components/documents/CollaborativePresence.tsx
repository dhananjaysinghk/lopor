"use client";

import React, { useEffect, useState } from "react";
import { Users, Wifi } from "lucide-react";

interface UserPresence {
  id: string;
  name: string;
  color: string;
}

interface CollaborativePresenceProps {
  documentId: string;
}

export function CollaborativePresence({ documentId }: CollaborativePresenceProps) {
  const [onlineUsers, setOnlineUsers] = useState<UserPresence[]>([
    { id: "usr_1", name: "Sarah Connor (You)", color: "bg-indigo-500" },
    { id: "usr_2", name: "Alex Rivers", color: "bg-emerald-500" },
    { id: "usr_3", name: "David Miller", color: "bg-purple-500" },
  ]);

  return (
    <div className="flex items-center gap-3">
      <div className="flex items-center -space-x-2 overflow-hidden">
        {onlineUsers.map((u) => (
          <div
            key={u.id}
            title={u.name}
            className={`w-7 h-7 rounded-full border-2 border-zinc-950 flex items-center justify-center text-[10px] font-bold text-white shadow-md ${u.color}`}
          >
            {u.name[0]}
          </div>
        ))}
      </div>
      <span className="flex items-center gap-1.5 text-[11px] font-mono text-emerald-400 bg-emerald-500/10 border border-emerald-500/20 px-2 py-0.5 rounded">
        <Wifi size={10} className="animate-pulse" /> Live WebSockets
      </span>
    </div>
  );
}
