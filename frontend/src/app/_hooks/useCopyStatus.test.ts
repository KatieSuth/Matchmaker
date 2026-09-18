// Tests for the shared clipboard-copy status lifecycle (idle -> success/error -> auto-reset).
import { act, renderHook } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { useCopyStatus } from "./useCopyStatus";

describe("useCopyStatus", () => {
  let writeText: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    // jsdom doesn't implement the Clipboard API; stub it so `copy()` has something to call.
    writeText = vi.fn().mockResolvedValue(undefined);
    Object.assign(navigator, { clipboard: { writeText } });
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("starts idle", () => {
    const { result } = renderHook(() => useCopyStatus());
    expect(result.current.status).toBe("idle");
  });

  it("sets status to success after a successful copy, then auto-resets to idle", async () => {
    const { result } = renderHook(() => useCopyStatus(1600));

    await act(async () => {
      await result.current.copy("share-link-value");
    });

    expect(writeText).toHaveBeenCalledWith("share-link-value");
    expect(result.current.status).toBe("success");

    // `waitFor` polls via real timers, which never fire once `vi.useFakeTimers()` is active, so
    // assert synchronously — `act()` flushes the state update the fake timer's callback triggers.
    act(() => {
      vi.advanceTimersByTime(1600);
    });
    expect(result.current.status).toBe("idle");
  });

  it("sets status to error when the clipboard write rejects", async () => {
    writeText.mockRejectedValueOnce(new Error("denied"));
    const { result } = renderHook(() => useCopyStatus());

    await act(async () => {
      await result.current.copy("value");
    });

    expect(result.current.status).toBe("error");
  });

  it("is a no-op for empty/nullish values", async () => {
    const { result } = renderHook(() => useCopyStatus());

    await act(async () => {
      await result.current.copy("");
      await result.current.copy(null);
      await result.current.copy(undefined);
    });

    expect(writeText).not.toHaveBeenCalled();
    expect(result.current.status).toBe("idle");
  });

  it("reset() clears status immediately without waiting for the timer", async () => {
    const { result } = renderHook(() => useCopyStatus(5000));

    await act(async () => {
      await result.current.copy("value");
    });
    expect(result.current.status).toBe("success");

    act(() => result.current.reset());
    expect(result.current.status).toBe("idle");
  });
});
