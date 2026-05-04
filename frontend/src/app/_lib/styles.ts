// Shared Tailwind class strings for form inputs across the app.
// Import `inputCls` wherever a styled text input is needed so all inputs
// stay visually consistent with a single source of truth.

/** Single portal root for react-datepicker poppers (`document.body`). */
export const DATEPICKER_PORTAL_ID = "matchmaker-datepicker-portal";

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
  /* Library uses #f0f0f0 + :not([aria-disabled]) — must match specificity / !important or hover looks bright white */
  .react-datepicker__day:not([aria-disabled=true]):hover,
  .react-datepicker__month-text:not([aria-disabled=true]):hover,
  .react-datepicker__quarter-text:not([aria-disabled=true]):hover,
  .react-datepicker__year-text:not([aria-disabled=true]):hover {
    background-color: rgba(30, 120, 255, 0.14) !important;
    color: var(--color-text) !important;
  }
  .react-datepicker__day--selected,
  .react-datepicker__day--range-start,
  .react-datepicker__day--range-end { background: var(--color-accent-blue) !important; color: #fff !important; border-radius: 0.375rem; }
  .react-datepicker__day--selected:not([aria-disabled=true]):hover,
  .react-datepicker__day--range-start:not([aria-disabled=true]):hover,
  .react-datepicker__day--range-end:not([aria-disabled=true]):hover,
  .react-datepicker__month-text--selected:not([aria-disabled=true]):hover,
  .react-datepicker__month-text--in-range:not([aria-disabled=true]):hover {
    background-color: #3d89ff !important;
    color: #fff !important;
  }
  .react-datepicker__day--in-range { background: rgba(30,120,255,0.12); color: var(--color-text); border-radius: 0; }
  .react-datepicker__day--in-range:not([aria-disabled=true]):hover {
    background-color: rgba(30, 120, 255, 0.22) !important;
    color: var(--color-text) !important;
  }
  .react-datepicker__day--in-selecting-range { background: rgba(30,120,255,0.08); border-radius: 0; }
  .react-datepicker__day--in-selecting-range:not([aria-disabled=true]):hover {
    background-color: rgba(30, 120, 255, 0.18) !important;
    color: var(--color-text) !important;
  }
  .react-datepicker__day--keyboard-selected { background: rgba(30,120,255,0.18); color: var(--color-text); }
  .react-datepicker__day--keyboard-selected:not([aria-disabled=true]):hover {
    background-color: rgba(30, 120, 255, 0.26) !important;
    color: var(--color-text) !important;
  }
  .react-datepicker__day--outside-month { color: rgba(180,200,235,0.2); }
  .react-datepicker__day--disabled { color: rgba(180,200,235,0.2) !important; cursor: not-allowed; }
  .react-datepicker__triangle { display: none; }
  .react-datepicker__month { margin: 0.4rem; }

  /* Month/year native selects — surface + text for closed state; option list (where browsers allow) */
  .react-datepicker {
    color-scheme: dark;
  }
  .react-datepicker__month-select,
  .react-datepicker__year-select,
  .react-datepicker__month-year-select {
    background-color: var(--color-surface-dropdown) !important;
    color: var(--color-text) !important;
    border: 1px solid var(--color-border-subtle) !important;
    border-radius: 0.375rem;
    padding: 0.2rem 0.45rem;
    margin-top: 4px;
    font-size: 0.75rem;
  }
  .react-datepicker__month-select:focus-visible,
  .react-datepicker__year-select:focus-visible,
  .react-datepicker__month-year-select:focus-visible {
    outline: none;
    border-color: var(--color-accent-blue) !important;
    box-shadow: 0 0 0 2px rgba(30, 120, 255, 0.25);
  }
  .react-datepicker__month-select option,
  .react-datepicker__year-select option,
  .react-datepicker__month-year-select option {
    background-color: var(--color-surface-dropdown);
    color: var(--color-text);
  }
  .react-datepicker-year-header {
    color: var(--color-text-soft) !important;
  }

  /* Default library top is 2px (too high vs title); 10px was too low — split for alignment with "Month YYYY" row */
  .react-datepicker:has(.react-datepicker__month-select) .react-datepicker__navigation {
    top: 5px;
  }
  .react-datepicker:has(.react-datepicker__month-select) .react-datepicker__navigation-icon {
    top: 0;
  }

  /* Time list (showTimeSelect) — library uses white bg / gray hover */
  .react-datepicker__time-container {
    border-left: 1px solid rgba(255, 255, 255, 0.1) !important;
  }
  .react-datepicker__time-container .react-datepicker__time {
    background: rgba(8, 11, 20, 0.98) !important;
  }
  .react-datepicker-time__header {
    color: var(--color-text-soft) !important;
    font-weight: 600;
  }
  .react-datepicker__time-list-item {
    color: var(--color-text-soft) !important;
  }
  .react-datepicker__time-list-item:hover {
    background-color: rgba(30, 120, 255, 0.14) !important;
    color: var(--color-text) !important;
  }
  .react-datepicker__time-list-item--selected,
  .react-datepicker__time-list-item--selected:hover {
    background-color: var(--color-accent-blue) !important;
    color: #fff !important;
  }
  .react-datepicker__time-list-item--disabled,
  .react-datepicker__time-list-item--disabled:hover {
    color: rgba(180, 200, 235, 0.28) !important;
    background-color: transparent !important;
  }
`;
