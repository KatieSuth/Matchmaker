"use client";

// Floating "Copied"/"Copy failed" bubble shown under a copy button, driven by useCopyStatus.
// Previously duplicated near-verbatim 4x on the event group page.
import { CopyStatus } from "@/app/_hooks/useCopyStatus";

interface CopyStatusTooltipProps {
  status: CopyStatus;
  successLabel?: string;
  errorLabel?: string;
  /** Matches the drop shadow used by the share/ping buttons; the join-lobby copy button omits it. */
  shadow?: boolean;
}

/** Renders nothing when status is idle. */
export function CopyStatusTooltip({
  status,
  successLabel = "Copied",
  errorLabel = "Copy failed",
  shadow = true,
}: CopyStatusTooltipProps) {
  if (status === "idle") return null;

  return (
    <div
      className={[
        "pointer-events-none absolute left-1/2 top-full mt-1.5 -translate-x-1/2 whitespace-nowrap rounded-md border px-2 py-1 text-[11px]",
        shadow ? "shadow-[0_10px_24px_rgba(0,0,0,0.45)]" : "",
        status === "success"
          ? "border-emerald-500/30 bg-[var(--color-bg)] text-emerald-300"
          : "border-[var(--color-text-danger)]/30 bg-[var(--color-bg)] text-[var(--color-text-danger)]",
      ].join(" ")}
    >
      {status === "success" ? successLabel : errorLabel}
    </div>
  );
}
