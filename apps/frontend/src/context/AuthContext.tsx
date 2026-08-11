"use client";

import React, { createContext, useContext, useEffect, useState } from "react";
import { apiFetch } from "@/lib/api";

interface User {
  id: string;
  email: string;
  full_name: string;
  role: string;
  avatar_url?: string;
}

interface Workspace {
  id: string;
  name: string;
  slug: string;
  icon?: string;
}

interface AuthContextType {
  user: User | null;
  workspaces: Workspace[];
  activeWorkspace: Workspace | null;
  isLoading: boolean;
  login: (token: string, user: User) => void;
  logout: () => void;
  setActiveWorkspace: (ws: Workspace) => void;
  refreshWorkspaces: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [workspaces, setWorkspaces] = useState<Workspace[]>([]);
  const [activeWorkspace, setActiveWorkspace] = useState<Workspace | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const initAuth = async () => {
      const token = localStorage.getItem("lopor_access_token");
      if (token) {
        try {
          const userData = await apiFetch<User>("/auth/me");
          setUser(userData);
          await loadWorkspaces();
        } catch (error) {
          console.error("Auth init error:", error);
          localStorage.removeItem("lopor_access_token");
        }
      }
      setIsLoading(false);
    };

    initAuth();
  }, []);

  const loadWorkspaces = async () => {
    try {
      const wsData = await apiFetch<Workspace[]>("/workspaces");
      setWorkspaces(wsData || []);
      if (wsData && wsData.length > 0) {
        setActiveWorkspace(wsData[0]);
      }
    } catch (err) {
      console.error("Failed to load workspaces:", err);
    }
  };

  const login = (token: string, userData: User) => {
    localStorage.setItem("lopor_access_token", token);
    setUser(userData);
    loadWorkspaces();
  };

  const logout = () => {
    localStorage.removeItem("lopor_access_token");
    setUser(null);
    setWorkspaces([]);
    setActiveWorkspace(null);
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        workspaces,
        activeWorkspace,
        isLoading,
        login,
        logout,
        setActiveWorkspace,
        refreshWorkspaces: loadWorkspaces,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
