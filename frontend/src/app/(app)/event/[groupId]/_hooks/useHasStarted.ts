"use client";

// Reports whether an ISO start time has passed, flipping to true on its own when it does,
// without reading the clock during render.
import { useEffect, useState } from "react";

/** setTimeout's maximum delay (~24.8 days); later start times are re-checked on the next render. */
const MAX_TIMEOUT_MS = 2 ** 31 - 1;

export function useHasStarted(startTimeIso: string): boolean {
  // Remember which start time has passed so a changed start time resets the answer.
  const [startedFor, setStartedFor] = useState<string | null>(null);

  useEffect(() => {
    const startMs = Date.parse(startTimeIso);
    if (Number.isNaN(startMs)) return;
    const delay = Math.max(0, startMs - Date.now());
    if (delay > MAX_TIMEOUT_MS) return;
    const timer = setTimeout(() => setStartedFor(startTimeIso), delay);
    return () => clearTimeout(timer);
  }, [startTimeIso]);

  return startedFor === startTimeIso;
}
