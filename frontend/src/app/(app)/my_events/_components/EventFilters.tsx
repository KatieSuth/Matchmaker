"use client";

// Filters card: upcoming/past pill, game select, date range pickers, apply/clear, and the
// resulting active-filter summary. Computes its own validation/dirty-state from the pending vs.
// applied filter values it's given.
import DatePicker from "react-datepicker";
import ReactSelect, { GroupBase } from "react-select";
import { SelectOption, selectStyles } from "@/app/_components/Select";
import { DATEPICKER_PORTAL_ID } from "@/app/_lib/styles";
import { formatMyEventsDate } from "../_lib/dateFmt";
import { TimeFilter } from "../_types";
import { DateInput } from "./DateInput";

export function EventFilters({
  timeFilter,
  setTimeFilter,
  dateRangeActive,
  pendingGame,
  setPendingGame,
  gameSelectOptions,
  gamesSelectLoading,
  gamesError,
  pendingFrom,
  setPendingFrom,
  pendingTo,
  setPendingTo,
  appliedGame,
  appliedFrom,
  appliedTo,
  onApply,
  onClear,
}: {
  timeFilter: TimeFilter;
  setTimeFilter: (t: TimeFilter) => void;
  dateRangeActive: boolean;
  pendingGame: SelectOption | null;
  setPendingGame: (opt: SelectOption | null) => void;
  gameSelectOptions: GroupBase<SelectOption>[];
  gamesSelectLoading: boolean;
  gamesError: boolean;
  pendingFrom: Date | null;
  setPendingFrom: (d: Date | null) => void;
  pendingTo: Date | null;
  setPendingTo: (d: Date | null) => void;
  appliedGame: SelectOption | null;
  appliedFrom: Date | null;
  appliedTo: Date | null;
  onApply: () => void;
  onClear: () => void;
}) {
  const dateRangeValid = !pendingFrom || !pendingTo || pendingFrom <= pendingTo;
  const filtersAreDirty =
    pendingGame?.value !== appliedGame?.value ||
    pendingFrom?.toISOString() !== appliedFrom?.toISOString() ||
    pendingTo?.toISOString() !== appliedTo?.toISOString();

  const activeFilterLabels: string[] = [];
  if (appliedGame) activeFilterLabels.push(appliedGame.label);
  if (appliedFrom && appliedTo) {
    activeFilterLabels.push(`${formatMyEventsDate(appliedFrom)} – ${formatMyEventsDate(appliedTo)}`);
  } else if (appliedFrom) {
    activeFilterLabels.push(`From ${formatMyEventsDate(appliedFrom)}`);
  } else if (appliedTo) {
    activeFilterLabels.push(`Until ${formatMyEventsDate(appliedTo)}`);
  }

  return (
    <div
      className="card rounded-xl p-4 flex flex-col gap-3 relative overflow-visible"
      style={{ animation: "var(--animate-rise-2)" }}
    >
      <div className="absolute top-0 left-4 right-4 h-px bg-top-edge opacity-15 rounded-full" />

      {/* Time toggle + game select row */}
      <div className="flex flex-col sm:flex-row sm:items-center gap-3">
        {/* Upcoming / Past pill */}
        <div className="flex items-center gap-1 p-0.5 rounded-lg bg-white/[0.04] border border-white/[0.07] flex-shrink-0">
          {(["upcoming", "past"] as TimeFilter[]).map((t) => {
            const disabled = dateRangeActive;
            return (
              <button
                key={t}
                type="button"
                title={disabled ? "Clear date filter to use this" : undefined}
                onClick={() => !disabled && setTimeFilter(t)}
                className={[
                  "px-3 py-1 rounded-md text-xs font-medium transition-all duration-150",
                  disabled ? "opacity-40 cursor-not-allowed" : "cursor-pointer",
                  timeFilter === t && !disabled
                    ? "bg-white/[0.09] text-[var(--color-text)] border border-white/10"
                    : "text-[var(--color-text-muted)] hover:text-[var(--color-text-soft)]",
                ].join(" ")}
              >
                {t.charAt(0).toUpperCase() + t.slice(1)}
              </button>
            );
          })}
        </div>

        {/* Game select */}
        <div className="flex-1 min-w-0">
          <ReactSelect<SelectOption, false, GroupBase<SelectOption>>
            instanceId="my-events-filter-game"
            inputId="my-events-filter-game"
            aria-label="Filter by game"
            value={pendingGame}
            onChange={(opt) => setPendingGame(opt)}
            options={gameSelectOptions}
            placeholder="Filter by game…"
            isLoading={gamesSelectLoading}
            isDisabled={gamesError}
            isClearable
            isSearchable
            styles={selectStyles}
            menuPortalTarget={typeof document !== "undefined" ? document.body : null}
            menuPosition="fixed"
            noOptionsMessage={() =>
              gamesError ? "Games unavailable" : "No games found"
            }
          />
        </div>
      </div>

      {/* Date range row */}
      <div className="flex flex-col gap-2 sm:flex-row sm:items-start">
        <div className="flex-1 flex flex-col gap-1.5">
          <label htmlFor="my-events-filter-from" className="text-xs font-medium tracking-wide text-[var(--color-text-soft)]">From</label>
          <DatePicker
            portalId={DATEPICKER_PORTAL_ID}
            wrapperClassName="w-full"
            selected={pendingFrom}
            onChange={(d: Date | null) => setPendingFrom(d)}
            selectsStart
            startDate={pendingFrom}
            endDate={pendingTo}
            maxDate={pendingTo ?? undefined}
            showMonthDropdown
            showYearDropdown
            dropdownMode="select"
            dateFormat="MMM d, yyyy"
            placeholderText="From"
            popperPlacement="bottom-end"
            popperProps={{
              strategy: "fixed"
            }}
            customInput={
              <DateInput
                id="my-events-filter-from"
                placeholder="From"
                emptyLabel="From"
                isClearable
                onClear={() => setPendingFrom(null)}
              />
            }
          />
        </div>
        <div className="flex-1 flex flex-col gap-1.5">
          <label htmlFor="my-events-filter-to" className="text-xs font-medium tracking-wide text-[var(--color-text-soft)]">To</label>
          <DatePicker
            portalId={DATEPICKER_PORTAL_ID}
            wrapperClassName="w-full"
            selected={pendingTo}
            onChange={(d: Date | null) => setPendingTo(d)}
            selectsEnd
            startDate={pendingFrom}
            endDate={pendingTo}
            minDate={pendingFrom ?? undefined}
            showMonthDropdown
            showYearDropdown
            dropdownMode="select"
            dateFormat="MMM d, yyyy"
            placeholderText="To"
            popperPlacement="bottom-end"
            popperProps={{
              strategy: "fixed"
            }}
            customInput={
              <DateInput
                id="my-events-filter-to"
                placeholder="To"
                emptyLabel="To"
                isClearable
                onClear={() => setPendingTo(null)}
              />
            }
          />
        </div>
        <div className="flex shrink-0 flex-col gap-1.5">
          <span
            className="text-xs font-medium tracking-wide text-transparent select-none"
            aria-hidden
          >
            From
          </span>
          <button
            type="button"
            onClick={onApply}
            disabled={!dateRangeValid || !filtersAreDirty}
            title={!dateRangeValid ? "'From' must be before 'To'" : undefined}
            className={[
              "inline-flex h-9 items-center justify-center rounded-lg px-4 text-sm font-medium transition-all duration-150",
              "border border-[var(--color-accent-blue)]/30 bg-[var(--color-accent-blue)]/10 text-[var(--color-accent-blue)]",
              "hover:bg-[var(--color-accent-blue)]/20 hover:border-[var(--color-accent-blue)]/50",
              "disabled:opacity-35 disabled:cursor-not-allowed",
              "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--color-accent-blue)]/40",
            ].join(" ")}
          >
            Apply
          </button>
        </div>
      </div>

      {/* Validation */}
      {!dateRangeValid && (
        <p className="text-xs text-[var(--color-text-danger)] -mt-1">
          {"\"From\" date must be on or before \"To\" date."}
        </p>
      )}

      {/* Active filter summary */}
      {activeFilterLabels.length > 0 && (
        <div className="flex items-center gap-2 pt-0.5">
          <span className="text-xs text-[var(--color-text-muted)]">Showing:</span>
          <div className="flex items-center gap-1.5 flex-wrap">
            {activeFilterLabels.map((label) => (
              <span
                key={label}
                className="px-2 py-0.5 rounded-full text-xs bg-[var(--color-accent-blue)]/10 text-[var(--color-accent-blue)] border border-[var(--color-accent-blue)]/20"
              >
                {label}
              </span>
            ))}
          </div>
          <button
            type="button"
            onClick={onClear}
            className="ml-auto text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text-soft)] transition-colors underline underline-offset-2"
          >
            Clear
          </button>
        </div>
      )}
    </div>
  );
}
