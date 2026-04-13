// Shared Tailwind class strings for form inputs across the app.
// Import `inputCls` wherever a styled text input is needed so all inputs
// stay visually consistent with a single source of truth.

export const inputCls =
  "w-full px-3 py-2 rounded-lg text-sm " +
  "bg-white/5 border border-white/10 " +
  "text-[var(--color-text)] placeholder:text-[var(--color-text-muted)] " +
  "focus:outline-none focus:border-[var(--color-accent-blue)] " +
  "focus:ring-1 focus:ring-[var(--color-accent-blue)]/30 " +
  "transition-colors duration-150 appearance-none";

// ---------------------------------------------------------------------------
// react-datepicker CSS overrides — keyed to site design tokens
// ---------------------------------------------------------------------------

export const datepickerStyles = `
  .react-datepicker-popper { z-index: 9999 !important; }
  .react-datepicker {
    font-family: inherit;
    background: linear-gradient(160deg, rgba(14,20,38,0.98) 0%, rgba(8,11,20,0.99) 100%);
    border: 1px solid rgba(255,255,255,0.09);
    border-radius: 0.5rem;
    box-shadow: 0 0 0 1px rgba(255,255,255,0.03), 0 20px 60px rgba(0,0,0,0.65), 0 0 40px rgba(30,80,200,0.08);
    color: var(--color-text);
    overflow: hidden;
  }
  .react-datepicker__header {
    background: rgba(255,255,255,0.03);
    border-bottom: 1px solid rgba(255,255,255,0.07);
    padding-top: 10px;
  }
  .react-datepicker__current-month { color: var(--color-text-soft); font-size: 0.8rem; }
  .react-datepicker__day-name { color: rgba(180,200,235,0.4); font-size: 0.7rem; font-weight: 600; letter-spacing: 0.05em; }
  .react-datepicker__navigation-icon::before { border-color: rgba(180,200,235,0.45); }
  .react-datepicker__navigation:hover .react-datepicker__navigation-icon::before { border-color: var(--color-text-soft); }
  .react-datepicker__day { color: var(--color-text-soft); border-radius: 0.375rem; font-size: 0.8rem; transition: background 100ms, color 100ms; }
  .react-datepicker__day:hover { background: rgba(255,255,255,0.07); color: var(--color-text); }
  .react-datepicker__day--selected,
  .react-datepicker__day--range-start,
  .react-datepicker__day--range-end { background: var(--color-accent-blue) !important; color: #fff !important; border-radius: 0.375rem; }
  .react-datepicker__day--in-range { background: rgba(30,120,255,0.12); color: var(--color-text); border-radius: 0; }
  .react-datepicker__day--in-selecting-range { background: rgba(30,120,255,0.08); border-radius: 0; }
  .react-datepicker__day--keyboard-selected { background: rgba(30,120,255,0.18); color: var(--color-text); }
  .react-datepicker__day--outside-month { color: rgba(180,200,235,0.2); }
  .react-datepicker__day--disabled { color: rgba(180,200,235,0.2) !important; cursor: not-allowed; }
  .react-datepicker__triangle { display: none; }
  .react-datepicker__month { margin: 0.4rem; }
`;
