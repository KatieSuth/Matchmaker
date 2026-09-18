// Form layout primitive: label, optional error/hint, and a slot for the control.
interface FieldProps {
  label: string;
  error?: string;
  hint?: string;
  /** Associates this label with a control via htmlFor (native id or react-select inputId). */
  htmlFor?: string;
  children: React.ReactNode;
}

export function Field({ label, error, hint, htmlFor, children }: FieldProps) {
  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={htmlFor} className="text-xs font-medium tracking-wide text-[var(--color-text-soft)]">
        {label}
      </label>
      {hint && (
        <p className="text-xs text-[var(--color-text-muted)] -mt-0.5">{hint}</p>
      )}
      {children}
      {error && (
        <p className="text-xs text-[var(--color-text-danger)]">{error}</p>
      )}
    </div>
  );
}