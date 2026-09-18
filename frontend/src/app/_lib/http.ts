// Shared helpers for classifying axios/fetch request outcomes.

/**
 * True when a request failed because it was aborted via AbortController (React Strict Mode
 * double-effects, unmount, or a dependency change), rather than a genuine network/API failure.
 * Callers should treat this as a no-op: skip setting error state and skip clearing prior data.
 */
export function isCanceledError(err: unknown): boolean {
  const e = err as { code?: string; name?: string } | null | undefined;
  return e?.code === "ERR_CANCELED" || e?.name === "CanceledError";
}
