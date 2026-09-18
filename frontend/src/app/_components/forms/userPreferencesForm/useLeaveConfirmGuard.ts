"use client";

// Guards in-app navigation away from /my_account while a new user has an unfinished profile and
// a pending post-login event redirect: intercepts same-origin link clicks and asks for confirmation
// instead of letting the user lose their way back to the event they came here for.
import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { consumePostLoginRedirect } from "@/app/_lib/postLoginRedirect";

export function useLeaveConfirmGuard(newUser: boolean | undefined, hasPendingEventRedirect: boolean) {
  const router = useRouter();
  /** Non-null while the leave-confirm sheet is open; value is the blocked in-app href. */
  const [leaveHref, setLeaveHref] = useState<string | null>(null);

  const closeLeaveSheet = useCallback(() => setLeaveHref(null), []);

  const confirmLeave = useCallback(() => {
    if (!leaveHref) return;
    const href = leaveHref;
    setLeaveHref(null);
    consumePostLoginRedirect();
    router.push(href);
  }, [leaveHref, router]);

  useEffect(() => {
    if (!newUser || !hasPendingEventRedirect) return;

    const onClick = (e: MouseEvent) => {
      if (e.defaultPrevented || e.button !== 0) return;
      if (e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return;

      const target = e.target;
      if (!(target instanceof Element)) return;
      const anchor = target.closest("a[href]");
      if (!(anchor instanceof HTMLAnchorElement)) return;

      const hrefAttr = anchor.getAttribute("href");
      if (!hrefAttr || hrefAttr.startsWith("#")) return;

      let url: URL;
      try {
        url = new URL(hrefAttr, window.location.origin);
      } catch {
        return;
      }
      if (url.origin !== window.location.origin) return;
      if (url.pathname === "/my_account") return;

      e.preventDefault();
      e.stopPropagation();
      setLeaveHref(url.pathname + url.search);
    };

    document.addEventListener("click", onClick, true);
    return () => document.removeEventListener("click", onClick, true);
  }, [newUser, hasPendingEventRedirect]);

  return { leaveHref, closeLeaveSheet, confirmLeave };
}
