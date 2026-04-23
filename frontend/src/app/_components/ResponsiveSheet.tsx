"use client";

// Modal "sheet" that portals to document.body; used for registration details and event edit flows.
import { useEffect, useMemo } from "react";
import { createPortal } from "react-dom";

interface ResponsiveSheetProps {
  isOpen: boolean;
  onClose: () => void;
  title?: string;
  children: React.ReactNode;
  maxWidthClassName?: string;
}

export function ResponsiveSheet({
  isOpen,
  onClose,
  title,
  children,
  maxWidthClassName = "lg:max-w-xl",
}: ResponsiveSheetProps) {
  useEffect(() => {
    if (!isOpen) return;

    const prevBodyOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";

    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") onClose();
    };

    window.addEventListener("keydown", onKeyDown);
    return () => {
      document.body.style.overflow = prevBodyOverflow;
      window.removeEventListener("keydown", onKeyDown);
    };
  }, [isOpen, onClose]);

  const content = useMemo(() => {
    if (!isOpen) return null;

    return (
      <div
        className="fixed inset-0 z-50"
        role="dialog"
        aria-modal="true"
        aria-label={title ?? "Dialog"}
      >
        <button
          type="button"
          aria-label="Close"
          onClick={onClose}
          className="absolute inset-0 bg-black/70 backdrop-blur-[2px]"
        />

        {/* Below lg: bottom sheet (~90% height). lg+: centered modal. lg (1024px) avoids iPad portrait (768) and wide RDM still using the “desktop” modal. */}
        <div className="absolute inset-x-0 bottom-0 top-0 flex min-h-0 flex-col justify-end px-4 pb-[max(0.75rem,env(safe-area-inset-bottom))] pt-[max(0.75rem,env(safe-area-inset-top))] lg:inset-0 lg:flex-row lg:items-center lg:justify-center lg:p-6 lg:pb-6 lg:pt-6">
          <div
            className={[
              "responsive-sheet-panel relative flex w-full min-h-0 flex-col overflow-hidden rounded-t-2xl border border-white/10 bg-[var(--color-bg)] shadow-[0_30px_90px_rgba(0,0,0,0.7)]",
              "animate-[var(--animate-rise)] lg:rounded-2xl",
              maxWidthClassName,
            ].join(" ")}
          >
            <div className="absolute -top-px left-4 right-4 h-px bg-top-edge opacity-30 rounded-full" />

            <div className="flex shrink-0 items-start justify-between gap-3 border-b border-white/[0.06] px-5 pb-4 pt-5">
              {title ? (
                <h2 className="text-lg font-semibold text-[var(--color-text)] tracking-tight">{title}</h2>
              ) : (
                <span />
              )}
              <button
                type="button"
                onClick={onClose}
                className="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-full border border-white/10 bg-white/[0.03] text-[var(--color-text-muted)] transition-colors hover:text-[var(--color-text-soft)]"
                aria-label="Close sheet"
              >
                <svg width="12" height="12" viewBox="0 0 12 12" fill="none">
                  <path
                    d="M2 2l8 8M10 2L2 10"
                    stroke="currentColor"
                    strokeWidth="1.5"
                    strokeLinecap="round"
                  />
                </svg>
              </button>
            </div>

            <div className="min-h-0 flex-1 overflow-y-auto overscroll-y-contain px-5 py-4">
              {children}
            </div>
          </div>
        </div>
      </div>
    );
  }, [children, isOpen, maxWidthClassName, onClose, title]);

  if (typeof document === "undefined") return null;
  return createPortal(content, document.body);
}
