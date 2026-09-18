"use client";

// Small open/closed registration-status pill used on each event card.
export function Badge({ open }: { open: boolean }) {
  return (
    <span
      className={[
        "inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[0.65rem] font-semibold tracking-wide",
        open
          ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
          : "bg-white/5 text-[var(--color-text-muted)] border border-white/10",
      ].join(" ")}
    >
      <span
        className={[
          "w-1.5 h-1.5 rounded-full flex-shrink-0",
          open ? "bg-emerald-400" : "bg-white/25",
        ].join(" ")}
      />
      {open ? "Open" : "Closed"}
    </span>
  );
}
