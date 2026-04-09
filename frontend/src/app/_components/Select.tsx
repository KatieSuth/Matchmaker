"use client";

import ReactSelect, { GroupBase, StylesConfig } from "react-select";

export interface SelectOption {
  value: string;
  label: string;
}

interface SelectProps {
  value: string;
  onChange: (value: string) => void;
  options: SelectOption[];
  placeholder?: string;
  disabled?: boolean;
}

// Matches the site's design tokens — dark surface, blue accent, drop-in animation.
// react-select renders the menu in a portal by default (menuPortalTarget),
// so it escapes any overflow:hidden containers cleanly.
const buildStyles = (): StylesConfig<SelectOption, false, GroupBase<SelectOption>> => ({
  control: (base, state) => ({
    ...base,
    minHeight: "36px",
    background: "rgba(255,255,255,0.05)",
    border: state.isFocused
      ? "1px solid var(--color-accent-blue)"
      : "1px solid rgba(255,255,255,0.10)",
    borderRadius: "0.5rem",
    boxShadow: state.isFocused
      ? "0 0 0 3px rgba(30,120,255,0.18)"
      : "none",
    cursor: "pointer",
    transition: "border-color 150ms, box-shadow 150ms",
    "&:hover": {
      borderColor: state.isFocused
        ? "var(--color-accent-blue)"
        : "rgba(255,255,255,0.20)",
    },
  }),
  valueContainer: (base) => ({
    ...base,
    padding: "0 12px",
  }),
  singleValue: (base) => ({
    ...base,
    color: "var(--color-text)",
    fontSize: "0.875rem",
    margin: 0,
  }),
  placeholder: (base) => ({
    ...base,
    color: "var(--color-text-muted)",
    fontSize: "0.875rem",
    margin: 0,
  }),
  input: (base) => ({
    ...base,
    color: "var(--color-text)",
    fontSize: "0.875rem",
    margin: 0,
    padding: 0,
  }),
  indicatorSeparator: () => ({
    display: "none",
  }),
  dropdownIndicator: (base, state) => ({
    ...base,
    color: "rgba(180,200,235,0.55)",
    padding: "0 10px",
    transition: "transform 150ms, color 150ms",
    transform: state.selectProps.menuIsOpen ? "rotate(180deg)" : "rotate(0deg)",
    "&:hover": {
      color: "var(--color-text-soft)",
    },
  }),
  menu: (base) => ({
    ...base,
    background: "linear-gradient(160deg, rgba(14,20,38,0.98) 0%, rgba(8,11,20,0.99) 100%)",
    border: "1px solid rgba(255,255,255,0.09)",
    borderRadius: "0.5rem",
    boxShadow:
      "0 0 0 1px rgba(255,255,255,0.03), 0 20px 60px rgba(0,0,0,0.65), 0 0 40px rgba(30,80,200,0.08)",
    overflow: "hidden",
    animation: "var(--animate-drop-in)",
    marginTop: "4px",
  }),
  menuPortal: (base) => ({
    ...base,
    zIndex: 9999,
  }),
  menuList: (base) => ({
    ...base,
    padding: "4px",
  }),
  option: (base, state) => ({
    ...base,
    background: state.isSelected
      ? "rgba(30,120,255,0.12)"
      : state.isFocused
      ? "rgba(255,255,255,0.05)"
      : "transparent",
    color: state.isSelected
      ? "var(--color-accent-blue)"
      : "var(--color-text-soft)",
    fontSize: "0.875rem",
    padding: "8px 12px",
    borderRadius: "0.375rem",
    cursor: "pointer",
    transition: "background 100ms, color 100ms",
    "&:active": {
      background: "rgba(30,120,255,0.18)",
    },
  }),
  noOptionsMessage: (base) => ({
    ...base,
    color: "var(--color-text-muted)",
    fontSize: "0.875rem",
  }),
});

const styles = buildStyles();

export function Select({
  value,
  onChange,
  options,
  placeholder = "— Select —",
  disabled = false,
}: SelectProps) {
  const selected = options.find((o) => o.value === value) ?? null;

  return (
    <ReactSelect<SelectOption>
      value={selected}
      onChange={(opt) => onChange(opt?.value ?? "")}
      options={options}
      placeholder={placeholder}
      isDisabled={disabled}
      styles={styles}
      // Render menu in document.body so it escapes overflow:hidden containers
      menuPortalTarget={typeof document !== "undefined" ? document.body : null}
      menuPosition="fixed"
      isSearchable={false}
    />
  );
}
