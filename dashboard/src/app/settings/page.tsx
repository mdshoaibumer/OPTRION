"use client";

import { motion } from "framer-motion";
import { useState, useEffect } from "react";
import { Sidebar } from "@/components/layout/sidebar";
import { Header } from "@/components/layout/header";
import { useAuth } from "@/lib/auth-context";
import { api } from "@/lib/api";
import { Key, Save, CheckCircle2, XCircle, LogOut } from "lucide-react";

export default function SettingsPage() {
  const { apiKey: storedKey, apiUrl: storedUrl, isConfigured, setCredentials, clearCredentials } = useAuth();
  const [apiKey, setApiKey] = useState("");
  const [apiUrl, setApiUrl] = useState("http://localhost:8080");
  const [saved, setSaved] = useState(false);
  const [connectionStatus, setConnectionStatus] = useState<"idle" | "checking" | "ok" | "error">("idle");

  useEffect(() => {
    if (storedKey) setApiKey(storedKey);
    if (storedUrl) setApiUrl(storedUrl);
  }, [storedKey, storedUrl]);

  const handleSave = () => {
    setCredentials(apiKey, apiUrl);
    setSaved(true);
    setTimeout(() => setSaved(false), 2000);
  };

  const handleTest = async () => {
    setConnectionStatus("checking");
    try {
      await api.healthz();
      setConnectionStatus("ok");
    } catch {
      setConnectionStatus("error");
    }
    setTimeout(() => setConnectionStatus("idle"), 3000);
  };

  return (
    <div className="flex min-h-screen">
      <Sidebar />
      <div className="flex-1 ml-16 lg:ml-64">
        <Header />
        <main className="p-6 space-y-6">
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
          >
            <h1 className="text-2xl font-bold">Settings</h1>
            <p className="text-sm text-muted mt-1">
              Configure your Optrion dashboard connection
            </p>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1 }}
            className="max-w-lg rounded-xl border border-card-border bg-card p-6"
          >
            <h2 className="text-sm font-semibold mb-6 flex items-center gap-2">
              <Key className="h-4 w-4 text-accent" />
              API Connection
            </h2>

            <div className="space-y-4">
              <div>
                <label className="text-xs text-muted block mb-1.5">
                  API URL
                </label>
                <input
                  type="text"
                  value={apiUrl}
                  onChange={(e) => setApiUrl(e.target.value)}
                  className="w-full px-3 py-2 rounded-lg bg-background border border-card-border text-sm outline-none focus:border-accent transition-colors"
                  placeholder="http://localhost:8080"
                />
              </div>

              <div>
                <label className="text-xs text-muted block mb-1.5">
                  API Key
                </label>
                <input
                  type="password"
                  value={apiKey}
                  onChange={(e) => setApiKey(e.target.value)}
                  className="w-full px-3 py-2 rounded-lg bg-background border border-card-border text-sm outline-none focus:border-accent transition-colors font-mono"
                  placeholder="opk_..."
                />
                <p className="text-[10px] text-muted mt-1">
                  Your API key is stored locally in the browser and never sent to any third-party service.
                </p>
              </div>

              {/* Connection status */}
              {connectionStatus === "ok" && (
                <div className="flex items-center gap-2 text-xs text-success">
                  <CheckCircle2 className="h-3.5 w-3.5" />
                  Connected to OPTRION backend
                </div>
              )}
              {connectionStatus === "error" && (
                <div className="flex items-center gap-2 text-xs text-danger">
                  <XCircle className="h-3.5 w-3.5" />
                  Connection failed. Check URL and API key.
                </div>
              )}

              <div className="flex items-center gap-3">
                <button
                  onClick={handleSave}
                  className="flex items-center gap-2 px-4 py-2 rounded-lg bg-accent text-white text-sm font-medium hover:opacity-90 transition-opacity"
                >
                  <Save className="h-4 w-4" />
                  {saved ? "Saved!" : "Save Settings"}
                </button>
                <button
                  onClick={handleTest}
                  disabled={connectionStatus === "checking"}
                  className="flex items-center gap-2 px-4 py-2 rounded-lg border border-card-border text-sm font-medium hover:border-accent transition-colors disabled:opacity-50"
                >
                  {connectionStatus === "checking" ? "Testing..." : "Test Connection"}
                </button>
                {isConfigured && (
                  <button
                    onClick={clearCredentials}
                    className="flex items-center gap-2 px-4 py-2 rounded-lg border border-danger/30 text-sm font-medium text-danger hover:bg-danger/10 transition-colors"
                  >
                    <LogOut className="h-4 w-4" />
                    Disconnect
                  </button>
                )}
              </div>
            </div>
          </motion.div>
        </main>
      </div>
    </div>
  );
}
