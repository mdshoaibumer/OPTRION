"use client";

import { useEffect, useState } from "react";

function formatRelativeTime(isoString: string): string {
  const now = Date.now();
  const then = new Date(isoString).getTime();
  const diffMs = now - then;
  const diffMin = Math.floor(diffMs / 60000);
  const diffHr = Math.floor(diffMin / 60);
  const diffDay = Math.floor(diffHr / 24);

  if (diffMin < 1) return "just now";
  if (diffMin < 60) return `${diffMin}m ago`;
  if (diffHr < 24) return `${diffHr}h ago`;
  return `${diffDay}d ago`;
}

export function RelativeTime({ date }: { date: string }) {
  const [text, setText] = useState("");

  useEffect(() => {
    setText(formatRelativeTime(date));
    const interval = setInterval(() => {
      setText(formatRelativeTime(date));
    }, 60000);
    return () => clearInterval(interval);
  }, [date]);

  return <span suppressHydrationWarning>{text}</span>;
}
