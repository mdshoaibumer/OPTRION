"use client";

import { Bell, Search, User } from "lucide-react";

export function Header() {
  return (
    <header className="sticky top-0 z-40 h-16 border-b border-[var(--card-border)] bg-[var(--background)]/80 backdrop-blur-sm">
      <div className="flex h-full items-center justify-between px-6">
        {/* Search */}
        <div className="flex items-center gap-2 px-3 py-2 rounded-lg bg-[var(--card)] border border-[var(--card-border)] w-80">
          <Search className="h-4 w-4 text-[var(--muted)]" />
          <input
            type="text"
            placeholder="Search components, incidents... (⌘K)"
            className="bg-transparent text-sm outline-none w-full text-[var(--foreground)] placeholder:text-[var(--muted)]"
          />
          <kbd className="hidden sm:inline-flex text-xs px-1.5 py-0.5 rounded bg-[var(--card-border)] text-[var(--muted)] font-mono">
            ⌘K
          </kbd>
        </div>

        {/* Actions */}
        <div className="flex items-center gap-4">
          <button className="relative p-2 rounded-lg hover:bg-[var(--card)] transition-colors">
            <Bell className="h-5 w-5 text-[var(--muted)]" />
            <span className="absolute top-1 right-1 h-2 w-2 rounded-full bg-[var(--danger)]" />
          </button>
          <button className="flex items-center gap-2 p-2 rounded-lg hover:bg-[var(--card)] transition-colors">
            <div className="h-7 w-7 rounded-full bg-[var(--accent)] flex items-center justify-center">
              <User className="h-4 w-4 text-white" />
            </div>
          </button>
        </div>
      </div>
    </header>
  );
}
