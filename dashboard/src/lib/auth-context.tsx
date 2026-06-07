"use client";

import { createContext, useContext, useEffect, useState, useCallback } from "react";

interface AuthState {
  apiKey: string | null;
  apiUrl: string;
  isConfigured: boolean;
  setCredentials: (apiKey: string, apiUrl: string) => void;
  clearCredentials: () => void;
}

const AuthContext = createContext<AuthState>({
  apiKey: null,
  apiUrl: "http://localhost:8080",
  isConfigured: false,
  setCredentials: () => {},
  clearCredentials: () => {},
});

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [apiKey, setApiKey] = useState<string | null>(null);
  const [apiUrl, setApiUrl] = useState("http://localhost:8080");
  const [isConfigured, setIsConfigured] = useState(false);

  useEffect(() => {
    const storedKey = localStorage.getItem("optrion_api_key");
    const storedUrl = localStorage.getItem("optrion_api_url") || "http://localhost:8080";
    setApiKey(storedKey);
    setApiUrl(storedUrl);
    setIsConfigured(!!storedKey);
  }, []);

  const setCredentials = useCallback((key: string, url: string) => {
    localStorage.setItem("optrion_api_key", key);
    localStorage.setItem("optrion_api_url", url);
    setApiKey(key);
    setApiUrl(url);
    setIsConfigured(true);
  }, []);

  const clearCredentials = useCallback(() => {
    localStorage.removeItem("optrion_api_key");
    localStorage.removeItem("optrion_api_url");
    setApiKey(null);
    setApiUrl("http://localhost:8080");
    setIsConfigured(false);
  }, []);

  return (
    <AuthContext.Provider value={{ apiKey, apiUrl, isConfigured, setCredentials, clearCredentials }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  return useContext(AuthContext);
}
