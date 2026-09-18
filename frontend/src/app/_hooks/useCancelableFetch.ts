"use client";

// Shared cancel-safe "fetch on mount / dependency change" effect, matching the abort-in-flight-reads
// pattern required for effect-driven GETs (see frontend.mdc). Consolidates a shape that was
// previously hand-rolled in EventForm, UserPreferencesForm, and my_events.
import { useEffect } from "react";
import { isCanceledError } from "@/app/_lib/http";

interface UseCancelableFetchOptions<T> {
  /** Performs the request; must accept and respect the AbortSignal. */
  fetcher: (signal: AbortSignal) => Promise<T>;
  /** Skips the fetch entirely when false (e.g. waiting on auth or a required id). */
  enabled?: boolean;
  /** Called synchronously before the fetch starts (e.g. to flip a loading flag on). */
  onStart?: () => void;
  onSuccess: (data: T) => void;
  /** Called only for genuine failures; cancellations are swallowed automatically. */
  onError: (err: unknown) => void;
  /** Called after success or a genuine error, but not after a cancellation. */
  onSettled?: () => void;
  /** Re-runs the fetch when any of these change, same as a normal effect dependency array. */
  deps: React.DependencyList;
}

/**
 * Runs `fetcher` inside a cancel-safe `useEffect`: creates an AbortController, forwards its
 * signal, ignores results/errors after abort, and always cleans up by aborting on unmount or
 * dependency change. Intended for simple "load this on mount/when X changes" effects; effects
 * with multi-step/conditional fetching (e.g. chained requests) should stay bespoke.
 */
export function useCancelableFetch<T>({
  fetcher,
  enabled = true,
  onStart,
  onSuccess,
  onError,
  onSettled,
  deps,
}: UseCancelableFetchOptions<T>) {
  useEffect(() => {
    if (!enabled) return;
    const ac = new AbortController();
    onStart?.();

    (async () => {
      try {
        const data = await fetcher(ac.signal);
        if (ac.signal.aborted) return;
        onSuccess(data);
      } catch (err) {
        if (ac.signal.aborted || isCanceledError(err)) return;
        onError(err);
      } finally {
        if (!ac.signal.aborted) {
          onSettled?.();
        }
      }
    })();

    return () => ac.abort();
    // Callers pass an explicit `deps` array (mirroring the effect's own semantics); onSuccess/
    // onError/onSettled are intentionally excluded so callers don't need to memoize them.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [enabled, ...deps]);
}
