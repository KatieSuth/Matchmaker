"use client";

// Shared "copy to clipboard" status lifecycle, previously duplicated across every copy button
// on the event group page (share link, Discord pings, join-lobby info).
import { useCallback, useEffect, useRef, useState } from "react";

export type CopyStatus = "idle" | "success" | "error";

const DEFAULT_RESET_DELAY_MS = 1600;

/**
 * Tracks the idle/success/error state of a clipboard-copy action and auto-resets to idle after
 * `resetDelayMs`. Returns `copy(value)` to perform the write (no-op for empty/nullish values) and
 * `reset()` to clear status immediately (e.g. when closing the sheet that owns the button).
 */
export function useCopyStatus(resetDelayMs: number = DEFAULT_RESET_DELAY_MS) {
  const [status, setStatus] = useState<CopyStatus>("idle");
  const timerRef = useRef<number | null>(null);

  const clearPendingReset = useCallback(() => {
    if (timerRef.current !== null) {
      window.clearTimeout(timerRef.current);
      timerRef.current = null;
    }
  }, []);

  useEffect(() => clearPendingReset, [clearPendingReset]);

  const copy = useCallback(
    async (value: string | undefined | null) => {
      if (!value) return;
      try {
        await navigator.clipboard.writeText(value);
        setStatus("success");
      } catch {
        setStatus("error");
      }
      clearPendingReset();
      timerRef.current = window.setTimeout(() => setStatus("idle"), resetDelayMs);
    },
    [clearPendingReset, resetDelayMs]
  );

  const reset = useCallback(() => {
    clearPendingReset();
    setStatus("idle");
  }, [clearPendingReset]);

  return { status, copy, reset };
}
