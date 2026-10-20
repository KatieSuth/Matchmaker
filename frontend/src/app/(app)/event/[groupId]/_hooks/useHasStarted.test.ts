// Tests for the start-time hook that gates "Mark Absent".
import { act, renderHook } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { useHasStarted } from "./useHasStarted";

describe("useHasStarted", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-10-19T12:00:00Z"));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("is true right after mount for a start time in the past", () => {
    const { result } = renderHook(() => useHasStarted("2026-10-19T11:00:00Z"));
    act(() => vi.advanceTimersByTime(0));
    expect(result.current).toBe(true);
  });

  it("flips to true when a future start time arrives", () => {
    const { result } = renderHook(() => useHasStarted("2026-10-19T12:30:00Z"));
    act(() => vi.advanceTimersByTime(29 * 60 * 1000));
    expect(result.current).toBe(false);
    act(() => vi.advanceTimersByTime(60 * 1000));
    expect(result.current).toBe(true);
  });

  it("resets when the start time moves into the future", () => {
    const { result, rerender } = renderHook(({ start }) => useHasStarted(start), {
      initialProps: { start: "2026-10-19T11:00:00Z" },
    });
    act(() => vi.advanceTimersByTime(0));
    expect(result.current).toBe(true);

    rerender({ start: "2026-10-20T11:00:00Z" });
    expect(result.current).toBe(false);
  });

  it("stays false for an invalid or very distant start time", () => {
    const { result: invalid } = renderHook(() => useHasStarted("not-a-date"));
    const { result: distant } = renderHook(() => useHasStarted("2099-01-01T00:00:00Z"));
    act(() => vi.advanceTimersByTime(60_000));
    expect(invalid.current).toBe(false);
    expect(distant.current).toBe(false);
  });
});
