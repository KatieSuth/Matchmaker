"use client";

// Question-mark hint next to “Can lobby host”; hover or tap opens a floating popover (portal).
import { useEffect, useLayoutEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";

function LobbyHostInfoContent() {
  return (
    <div className="text-left">
      <p className="mb-2 font-medium text-[var(--color-text-soft)]">The lobby host is responsible for:</p>
      <ul className="list-disc space-y-1 pl-4 text-[var(--color-text-muted)]">
        <li>Creating the lobby and sharing the lobby code or inviting players</li>
        <li>Choosing the most balanced server for the majority</li>
        <li>Alerting the host of any toxicity occurring during matches</li>
      </ul>
    </div>
  );
}

export function LobbyHostInfoHint() {
  const [pinned, setPinned] = useState(false);
  const [hover, setHover] = useState(false);
  const wrapRef = useRef<HTMLDivElement>(null);
  const btnRef = useRef<HTMLButtonElement>(null);
  const panelRef = useRef<HTMLDivElement>(null);
  const hideTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const pinnedRef = useRef(false);
  const [panelPos, setPanelPos] = useState<{ top: number; left: number } | null>(null);

  const visible = pinned || hover;

  useEffect(() => {
    pinnedRef.current = pinned;
  }, [pinned]);

  const cancelScheduledHide = () => {
    if (hideTimerRef.current != null) {
      clearTimeout(hideTimerRef.current);
      hideTimerRef.current = null;
    }
  };

  const scheduleHide = () => {
    cancelScheduledHide();
    hideTimerRef.current = setTimeout(() => {
      hideTimerRef.current = null;
      if (!pinnedRef.current) setHover(false);
    }, 220);
  };

  useLayoutEffect(() => {
    if (!visible) return;

    const el = btnRef.current;
    if (!el) return;

    const update = () => {
      const r = el.getBoundingClientRect();
      const maxW = Math.min(288, typeof window !== "undefined" ? window.innerWidth - 16 : 288);
      let left = r.left;
      if (typeof window !== "undefined") {
        left = Math.max(8, Math.min(left, window.innerWidth - maxW - 8));
      }
      setPanelPos({ top: r.bottom + 8, left });
    };

    update();
    window.addEventListener("scroll", update, true);
    window.addEventListener("resize", update);
    return () => {
      window.removeEventListener("scroll", update, true);
      window.removeEventListener("resize", update);
    };
  }, [visible]);

  useEffect(() => {
    if (!visible) return;
    const onDocPointerDown = (e: PointerEvent) => {
      const t = e.target as Node;
      if (wrapRef.current?.contains(t)) return;
      if (panelRef.current?.contains(t)) return;
      cancelScheduledHide();
      setPinned(false);
      setHover(false);
    };
    document.addEventListener("pointerdown", onDocPointerDown, true);
    return () => document.removeEventListener("pointerdown", onDocPointerDown, true);
  }, [visible]);

  const portalTarget = typeof document !== "undefined" ? document.body : null;

  return (
    <>
      <div ref={wrapRef} className="relative inline-flex shrink-0 items-center">
        <button
          ref={btnRef}
          type="button"
          className={[
            "inline-flex h-5 min-w-[1.25rem] px-0.5 items-center justify-center rounded-full border border-white/15 text-[var(--color-text-muted)]",
            "hover:border-white/25 hover:text-[var(--color-text-soft)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-[var(--color-accent-blue)]/50",
          ].join(" ")}
          aria-label="What is a lobby host?"
          aria-expanded={visible}
          aria-haspopup="dialog"
          onMouseEnter={() => {
            cancelScheduledHide();
            setHover(true);
          }}
          onMouseLeave={() => scheduleHide()}
          onClick={(e) => {
            e.preventDefault();
            e.stopPropagation();
            cancelScheduledHide();
            setPinned((wasPinned) => {
              const next = !wasPinned;
              if (!next) setHover(false);
              return next;
            });
          }}
        >
          <span className="text-[11px] font-semibold leading-none text-current select-none" aria-hidden="true">
            ?
          </span>
        </button>
      </div>
      {visible &&
        panelPos &&
        portalTarget &&
        createPortal(
          <div
            ref={panelRef}
            role="dialog"
            aria-label="Lobby host responsibilities"
            className="fixed z-[200] w-[min(18rem,calc(100vw-1rem))] rounded-lg border border-white/10 bg-[var(--color-bg)] p-3 text-xs shadow-[0_12px_36px_rgba(0,0,0,0.55)]"
            style={{ top: panelPos.top, left: panelPos.left }}
            onMouseEnter={() => {
              cancelScheduledHide();
              setHover(true);
            }}
            onMouseLeave={() => scheduleHide()}
          >
            <LobbyHostInfoContent />
          </div>,
          portalTarget
        )}
    </>
  );
}
