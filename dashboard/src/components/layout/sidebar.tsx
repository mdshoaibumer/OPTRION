"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { motion } from "framer-motion";
import {
  Activity,
  AlertTriangle,
  Brain,
  LayoutDashboard,
  Network,
  Settings,
  Shield,
} from "lucide-react";
import { cn } from "@/lib/utils";

const navigation = [
  { name: "Dashboard", href: "/", icon: LayoutDashboard },
  { name: "Topology", href: "/topology", icon: Network },
  { name: "Incidents", href: "/incidents", icon: AlertTriangle },
  { name: "Health", href: "/health", icon: Activity },
  { name: "AI Insights", href: "/ai", icon: Brain },
  { name: "Alerts", href: "/alerts", icon: Shield },
  { name: "Settings", href: "/settings", icon: Settings },
];

export function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="fixed inset-y-0 left-0 z-50 w-16 lg:w-64 flex flex-col border-r border-(--glass-border) bg-(--background-elevated)/80 backdrop-blur-xl">
      {/* Logo */}
      <div className="flex h-16 items-center px-4 border-b border-(--glass-border)">
        <div className="flex items-center gap-3">
          <motion.div
            className="h-8 w-8 rounded-xl flex items-center justify-center relative overflow-hidden"
            style={{ background: "var(--accent-gradient)" }}
            whileHover={{ scale: 1.05, rotate: 2 }}
            whileTap={{ scale: 0.95 }}
          >
            <Activity className="h-4 w-4 text-white relative z-10" />
            <div className="absolute inset-0 bg-white/20 opacity-0 hover:opacity-100 transition-opacity" />
          </motion.div>
          <span className="hidden lg:block font-semibold text-lg tracking-tight">
            OPTRION
          </span>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 px-2 py-4 space-y-1">
        {navigation.map((item) => {
          const isActive = pathname === item.href;
          return (
            <Link
              key={item.name}
              href={item.href}
              className={cn(
                "relative flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium transition-all duration-200",
                isActive
                  ? "text-foreground"
                  : "text-muted hover:text-foreground hover:bg-(--glass)"
              )}
            >
              {/* Active indicator — tubelight glow */}
              {isActive && (
                <motion.div
                  layoutId="sidebar-active"
                  className="absolute inset-0 rounded-xl"
                  style={{
                    background: "var(--accent-glow)",
                    border: "1px solid rgba(99, 102, 241, 0.2)",
                    boxShadow: "0 0 20px var(--accent-glow), inset 0 0 20px var(--accent-glow)",
                  }}
                  transition={{ type: "spring", stiffness: 350, damping: 30 }}
                />
              )}

              {/* Left edge glow for active */}
              {isActive && (
                <motion.div
                  layoutId="sidebar-edge"
                  className="absolute left-0 top-1/2 -translate-y-1/2 w-0.5 h-5 rounded-full"
                  style={{
                    background: "var(--accent)",
                    boxShadow: "0 0 8px var(--accent), 0 0 16px var(--accent-glow)",
                  }}
                  transition={{ type: "spring", stiffness: 350, damping: 30 }}
                />
              )}

              <item.icon className={cn("h-5 w-5 shrink-0 relative z-10", isActive && "text-accent")} />
              <span className="hidden lg:block relative z-10">{item.name}</span>
            </Link>
          );
        })}
      </nav>

      {/* Status footer */}
      <div className="p-4 border-t border-(--glass-border)">
        <div className="hidden lg:flex items-center gap-2 text-xs text-muted">
          <motion.div
            className="h-2 w-2 rounded-full bg-success"
            animate={{
              boxShadow: [
                "0 0 0 0 rgba(34, 197, 94, 0.4)",
                "0 0 0 4px rgba(34, 197, 94, 0)",
              ],
            }}
            transition={{ duration: 2, repeat: Infinity, ease: "easeInOut" }}
          />
          <span>System Operational</span>
        </div>
        <div className="hidden lg:block mt-2">
          <div className="text-[10px] text-muted/60 font-mono">v0.1.0</div>
        </div>
      </div>
    </aside>
  );
}
