"use client";

import { motion } from "framer-motion";
import { Bell, Command, Search, User, Sparkles } from "lucide-react";

export function Header() {
  return (
    <header className="sticky top-0 z-40 h-16 border-b border-(--glass-border) bg-(--background)/60 backdrop-blur-xl">
      <div className="flex h-full items-center justify-between px-6">
        {/* Search — command palette style */}
        <motion.div
          initial={{ opacity: 0, y: -5 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1 }}
          className="flex items-center gap-2 px-4 py-2.5 rounded-xl bg-(--glass) border border-(--glass-border) w-80 group hover:border-(--glass-border-hover) transition-all duration-200 cursor-pointer"
        >
          <Search className="h-4 w-4 text-muted group-hover:text-accent transition-colors" />
          <input
            type="text"
            placeholder="Search components, incidents..."
            className="bg-transparent text-sm outline-none w-full text-foreground placeholder:text-muted cursor-pointer"
            readOnly
          />
          <kbd className="hidden sm:inline-flex items-center gap-0.5 text-[10px] px-1.5 py-0.5 rounded-md bg-background-elevated border border-(--glass-border) text-muted font-mono">
            <Command className="h-2.5 w-2.5" />K
          </kbd>
        </motion.div>

        {/* Actions */}
        <motion.div
          initial={{ opacity: 0, y: -5 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2 }}
          className="flex items-center gap-2"
        >
          {/* AI button */}
          <motion.button
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            className="relative p-2.5 rounded-xl bg-(--glass) border border-(--glass-border) hover:border-purple-500/30 transition-all duration-200 group"
          >
            <Sparkles className="h-4 w-4 text-muted group-hover:text-purple-400 transition-colors" />
          </motion.button>

          {/* Notifications */}
          <motion.button
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            className="relative p-2.5 rounded-xl bg-(--glass) border border-(--glass-border) hover:border-(--glass-border-hover) transition-all duration-200 group"
          >
            <Bell className="h-4 w-4 text-muted group-hover:text-foreground transition-colors" />
            <motion.span
              className="absolute top-1.5 right-1.5 h-2 w-2 rounded-full bg-danger"
              animate={{
                scale: [1, 1.2, 1],
                opacity: [1, 0.8, 1],
              }}
              transition={{ duration: 2, repeat: Infinity }}
            />
          </motion.button>

          {/* User avatar */}
          <motion.button
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            className="flex items-center gap-2 p-1.5 rounded-xl bg-(--glass) border border-(--glass-border) hover:border-(--glass-border-hover) transition-all duration-200"
          >
            <div className="h-7 w-7 rounded-lg flex items-center justify-center" style={{ background: "var(--accent-gradient)" }}>
              <User className="h-3.5 w-3.5 text-white" />
            </div>
          </motion.button>
        </motion.div>
      </div>
    </header>
  );
}
