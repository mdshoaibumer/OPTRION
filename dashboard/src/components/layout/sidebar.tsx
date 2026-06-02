"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
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
    <aside className="fixed inset-y-0 left-0 z-50 w-16 lg:w-64 flex flex-col border-r border-card-border bg-card">
      {/* Logo */}
      <div className="flex h-16 items-center px-4 border-b border-card-border">
        <div className="flex items-center gap-3">
          <div className="h-8 w-8 rounded-lg bg-accent flex items-center justify-center">
            <Activity className="h-4 w-4 text-white" />
          </div>
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
                "flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-all duration-200",
                isActive
                  ? "bg-(--accent-glow) text-accent border border-accent/20"
                  : "text-muted hover:text-foreground hover:bg-card-border"
              )}
            >
              <item.icon className="h-5 w-5 shrink-0" />
              <span className="hidden lg:block">{item.name}</span>
            </Link>
          );
        })}
      </nav>

      {/* Status */}
      <div className="p-4 border-t border-card-border">
        <div className="hidden lg:flex items-center gap-2 text-xs text-muted">
          <div className="h-2 w-2 rounded-full bg-success pulse-healthy" />
          <span>System Operational</span>
        </div>
      </div>
    </aside>
  );
}
