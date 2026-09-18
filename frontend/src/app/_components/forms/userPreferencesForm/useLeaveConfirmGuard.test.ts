// Tests for the leave-confirmation guard that intercepts in-app link clicks while a new user
// has an unfinished profile and a pending post-login event redirect.
import { act, renderHook } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { mockRouter, resetNextNavigationMock } from "@/test/nextNavigationMock";
import { peekPostLoginRedirect, persistPostLoginRedirect } from "@/app/_lib/postLoginRedirect";
import { useLeaveConfirmGuard } from "./useLeaveConfirmGuard";

vi.mock("next/navigation", () => import("@/test/nextNavigationMock"));

/**
 * Simulates a real left-click on an anchor, bubbling to `document` like a user's click would.
 * Returns whether the guard cancelled the click (capture), then always preventDefault so jsdom
 * does not attempt a full document navigation it cannot implement.
 */
function clickAnchor(href: string, mouseInit: MouseEventInit = {}) {
  const anchor = document.createElement("a");
  anchor.setAttribute("href", href);
  document.body.appendChild(anchor);

  let intercepted = false;
  // Capture, registered after the guard's capture listener, so we observe whether the guard
  // already cancelled and then always cancel — the guard uses stopPropagation, which would skip a
  // bubble listener.
  const stopJsdomNavigation = (e: Event) => {
    intercepted = e.defaultPrevented;
    e.preventDefault();
  };
  document.addEventListener("click", stopJsdomNavigation, true);

  const event = new MouseEvent("click", { bubbles: true, cancelable: true, button: 0, ...mouseInit });
  anchor.dispatchEvent(event);

  document.removeEventListener("click", stopJsdomNavigation, true);
  document.body.removeChild(anchor);
  return { defaultPrevented: intercepted };
}

describe("useLeaveConfirmGuard", () => {
  beforeEach(() => {
    resetNextNavigationMock();
    sessionStorage.clear();
  });

  afterEach(() => {
    sessionStorage.clear();
  });

  it("does not intercept clicks when the guard is inactive (not a new user)", () => {
    renderHook(() => useLeaveConfirmGuard(false, true));
    const event = clickAnchor("/my_events");
    expect(event.defaultPrevented).toBe(false);
  });

  it("does not intercept clicks when there is no pending event redirect", () => {
    renderHook(() => useLeaveConfirmGuard(true, false));
    const event = clickAnchor("/my_events");
    expect(event.defaultPrevented).toBe(false);
  });

  it("intercepts an in-app link click and opens the leave-confirm sheet", () => {
    const { result } = renderHook(() => useLeaveConfirmGuard(true, true));

    let event!: { defaultPrevented: boolean };
    act(() => {
      event = clickAnchor("/my_events");
    });

    expect(event.defaultPrevented).toBe(true);
    expect(result.current.leaveHref).toBe("/my_events");
  });

  it("does not intercept a click on the current page (/my_account)", () => {
    renderHook(() => useLeaveConfirmGuard(true, true));
    const event = clickAnchor("/my_account");
    expect(event.defaultPrevented).toBe(false);
  });

  it("does not intercept a hash-only link", () => {
    renderHook(() => useLeaveConfirmGuard(true, true));
    const event = clickAnchor("#section");
    expect(event.defaultPrevented).toBe(false);
  });

  it("does not intercept a cross-origin link", () => {
    renderHook(() => useLeaveConfirmGuard(true, true));
    const event = clickAnchor("https://evil.com/phish");
    expect(event.defaultPrevented).toBe(false);
  });

  it("does not intercept a modified click (ctrl/cmd/shift/alt — open in new tab)", () => {
    renderHook(() => useLeaveConfirmGuard(true, true));
    const event = clickAnchor("/my_events", { ctrlKey: true });
    expect(event.defaultPrevented).toBe(false);
  });

  it("closeLeaveSheet clears leaveHref without navigating", () => {
    const { result } = renderHook(() => useLeaveConfirmGuard(true, true));
    act(() => {
      clickAnchor("/my_events");
    });
    expect(result.current.leaveHref).toBe("/my_events");

    act(() => result.current.closeLeaveSheet());

    expect(result.current.leaveHref).toBeNull();
    expect(mockRouter.push).not.toHaveBeenCalled();
  });

  it("confirmLeave navigates to the blocked href and consumes the post-login redirect", () => {
    persistPostLoginRedirect("/event/abc123");
    const { result } = renderHook(() => useLeaveConfirmGuard(true, true));
    act(() => {
      clickAnchor("/my_events");
    });

    act(() => result.current.confirmLeave());

    expect(mockRouter.push).toHaveBeenCalledWith("/my_events");
    expect(result.current.leaveHref).toBeNull();
    // The pending post-login redirect is discarded once the user confirms leaving on purpose.
    expect(peekPostLoginRedirect()).toBeNull();
  });

  it("confirmLeave is a no-op when there is no pending leaveHref", () => {
    const { result } = renderHook(() => useLeaveConfirmGuard(true, true));
    act(() => result.current.confirmLeave());
    expect(mockRouter.push).not.toHaveBeenCalled();
  });

  it("removes the click listener on unmount", () => {
    const { unmount } = renderHook(() => useLeaveConfirmGuard(true, true));
    unmount();
    const event = clickAnchor("/my_events");
    expect(event.defaultPrevented).toBe(false);
  });
});
