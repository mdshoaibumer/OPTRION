"use client";

import { motion } from "framer-motion";
import { useState } from "react";
import { Sidebar } from "@/components/layout/sidebar";
import { Header } from "@/components/layout/header";
import { Key, Save } from "lucide-react";

export default function SettingsPage() {
  const [apiKey, setApiKey] = useState("");
  const [apiUrl, setApiUrl] = useState("http://localhost:8080");
  const [saved, setSaved] = useState(false);

  const handleSave = () => {
    if (apiKey) {
      localStorage.setItem("optrion_api_key", apiKey);
    }
    if (apiUrl) {
      localStorage.setItem("optrion_api_url", apiUrl);
    }
    setSaved(true);
    setTimeout(() => setSaved(false), 2000);
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

              <button
                onClick={handleSave}
                className="flex items-center gap-2 px-4 py-2 rounded-lg bg-accent text-white text-sm font-medium hover:opacity-90 transition-opacity"
              >
                <Save className="h-4 w-4" />
                {saved ? "Saved!" : "Save Settings"}
              </button>
            </div>
          </motion.div>
        </main>
      </div>
    </div>
  );
}
