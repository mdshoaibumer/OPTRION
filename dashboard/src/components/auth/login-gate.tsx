"use client";

import { useState } from "react";
import { motion } from "framer-motion";
import { useAuth } from "@/lib/auth-context";
import { Key, ArrowRight, AlertCircle, Shield } from "lucide-react";

export function LoginGate({ children }: { children: React.ReactNode }) {
  const { isConfigured, setCredentials } = useAuth();
  const [apiKey, setApiKey] = useState("");
  const [apiUrl, setApiUrl] = useState("http://localhost:8080");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  if (isConfigured) {
    return <>{children}</>;
  }

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!apiKey.trim()) {
      setError("API key is required");
      return;
    }

    setLoading(true);

    try {
      // Validate the API key by calling the healthz endpoint
      const res = await fetch(`${apiUrl}/readyz`, {
        headers: {
          Authorization: `Bearer ${apiKey}`,
        },
      });

      if (res.ok || res.status === 200) {
        setCredentials(apiKey, apiUrl);
      } else if (res.status === 401) {
        setError("Invalid API key");
      } else {
        // readyz doesn't require auth, so if we reach here, just save
        setCredentials(apiKey, apiUrl);
      }
    } catch {
      // Server might not be running — save credentials anyway
      // Dashboard will show errors on data fetch
      setCredentials(apiKey, apiUrl);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-background p-4">
      <motion.div
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        className="w-full max-w-md"
      >
        {/* Logo / Header */}
        <div className="text-center mb-8">
          <motion.div
            initial={{ opacity: 0, y: -20 }}
            animate={{ opacity: 1, y: 0 }}
            className="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-primary/10 mb-4"
          >
            <Shield className="w-8 h-8 text-primary" />
          </motion.div>
          <h1 className="text-2xl font-bold text-foreground">OPTRION</h1>
          <p className="text-sm text-muted mt-2">
            Engineering Intelligence Dashboard
          </p>
        </div>

        {/* Login Form */}
        <motion.form
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1 }}
          onSubmit={handleLogin}
          className="rounded-xl border border-card-border bg-card p-6 space-y-4"
        >
          <div>
            <label
              htmlFor="api-url"
              className="block text-sm font-medium text-foreground mb-1"
            >
              Server URL
            </label>
            <input
              id="api-url"
              type="url"
              value={apiUrl}
              onChange={(e) => setApiUrl(e.target.value)}
              placeholder="http://localhost:8080"
              className="w-full px-3 py-2 rounded-lg border border-card-border bg-background text-foreground text-sm focus:outline-none focus:ring-2 focus:ring-primary/50"
            />
          </div>

          <div>
            <label
              htmlFor="api-key"
              className="block text-sm font-medium text-foreground mb-1"
            >
              API Key
            </label>
            <div className="relative">
              <Key className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted" />
              <input
                id="api-key"
                type="password"
                value={apiKey}
                onChange={(e) => setApiKey(e.target.value)}
                placeholder="opk_..."
                autoComplete="off"
                className="w-full pl-10 pr-3 py-2 rounded-lg border border-card-border bg-background text-foreground text-sm focus:outline-none focus:ring-2 focus:ring-primary/50"
              />
            </div>
            <p className="text-xs text-muted mt-1">
              Get your API key from{" "}
              <code className="text-primary">POST /api/v1/register</code> or
              your admin.
            </p>
          </div>

          {error && (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              className="flex items-center gap-2 text-sm text-red-400 bg-red-500/10 p-3 rounded-lg"
            >
              <AlertCircle className="w-4 h-4 flex-shrink-0" />
              {error}
            </motion.div>
          )}

          <button
            type="submit"
            disabled={loading}
            className="w-full flex items-center justify-center gap-2 px-4 py-2.5 rounded-lg bg-primary text-primary-foreground font-medium text-sm hover:bg-primary/90 disabled:opacity-50 transition-colors"
          >
            {loading ? (
              <span className="animate-pulse">Connecting...</span>
            ) : (
              <>
                Connect to Optrion
                <ArrowRight className="w-4 h-4" />
              </>
            )}
          </button>
        </motion.form>

        <p className="text-xs text-muted text-center mt-4">
          Your API key is stored locally in your browser and never sent to third parties.
        </p>
      </motion.div>
    </div>
  );
}
