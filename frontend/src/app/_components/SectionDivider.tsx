interface SectionDividerProps {
  title: string;
}

export function SectionDivider({ title }: SectionDividerProps) {
  return (
    <div className="flex items-center gap-3 mb-1">
      <span className="divider-gem w-1.5 h-1.5 rounded-full flex-shrink-0" />
      <h2 className="text-xs font-semibold tracking-[0.14em] uppercase text-[var(--color-text-muted)]">
        {title}
      </h2>
      <div className="flex-1 h-px bg-divider-blue opacity-60" />
    </div>
  );
}