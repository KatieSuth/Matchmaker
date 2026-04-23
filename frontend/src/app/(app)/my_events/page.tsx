"use client";

// My Events dashboard: list of hosted and joined event groups, filters, and create flow.
import { useState, useEffect, useCallback, useRef } from "react";
import Link from "next/link";
import ReactSelect, { GroupBase } from "react-select";
import DatePicker from "react-datepicker";
import "react-datepicker/dist/react-datepicker.css";
import api from "@/app/_lib/axios";
import { Game } from "@/app/_types/types";
import { useAuth } from "@/app/_context/AuthContext";
import { SelectOption, selectStyles } from "@/app/_components/Select";
import { SectionDivider } from "@/app/_components/SectionDivider";
import { ResponsiveSheet } from "@/app/_components/ResponsiveSheet";
import { fetchGames } from "@/app/_services/games";
import { inputCls, datepickerStyles } from "@/app/_lib/styles";
import { EventForm } from "@/app/_components/forms/EventForm";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface Event {
  id: string;
  game_name: string;
  game_mode: string | null;
  event_date: string; // ISO 8601 UTC
  host_name: string;
  host_id: string;
  registered_count: number;
  registration_open: boolean;
}

interface EventsPage {
  event_groups: Event[];
  next_cursor: string | null;
  has_more: boolean;
}

interface BucketState {
  events: Event[];
  nextCursor: string | null;
  hasMore: boolean;
  isLoadingMore: boolean;
  loaded: boolean;
}

type Tab = "hosting" | "registered";
type TimeFilter = "upcoming" | "past";

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const USER_TZ =
  typeof Intl !== "undefined"
    ? Intl.DateTimeFormat().resolvedOptions().timeZone
    : "UTC";

const DATE_FMT = new Intl.DateTimeFormat(undefined, {
  dateStyle: "medium",
  timeZone: USER_TZ,
});

const EMPTY_BUCKET: BucketState = {
  events: [],
  nextCursor: null,
  hasMore: false,
  isLoadingMore: false,
  loaded: false,
};

// ---------------------------------------------------------------------------
// API helpers
// ---------------------------------------------------------------------------

async function fetchEvents(
  endpoint: string,
  params: Record<string, string | undefined>
): Promise<EventsPage> {
  const query = new URLSearchParams();
  for (const [k, v] of Object.entries(params)) {
    if (v !== undefined && v !== "") query.set(k, v);
  }
  const res = await api.get<EventsPage>(`${endpoint}?${query.toString()}`);
  return res.data;
}

// ---------------------------------------------------------------------------
// DateInput — custom trigger for react-datepicker that matches inputCls
// ---------------------------------------------------------------------------

interface DateInputProps {
  value?: string;
  onClick?: () => void;
  placeholder: string;
  isClearable?: boolean;
  onClear?: () => void;
}

function DateInput({ value, onClick, placeholder, isClearable, onClear }: DateInputProps) {
  return (
    <div className="relative flex items-center w-full">
      <button
        type="button"
        onClick={onClick}
        className={[
          // Re-use inputCls as a base but override padding for the icon
          "h-9 w-full pl-9 pr-3 rounded-lg text-sm text-left",
          "bg-white/5 border transition-all duration-150",
          "focus:outline-none focus:border-[var(--color-accent-blue)] focus:ring-1 focus:ring-[var(--color-accent-blue)]/30",
          value
            ? "border-white/20 text-[var(--color-text)]"
            : "border-white/10 text-[var(--color-text-muted)]",
        ].join(" ")}
      >
        {value || placeholder}
      </button>
      {/* Calendar icon */}
      <svg
        className="absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none"
        width="13"
        height="13"
        viewBox="0 0 16 16"
        fill="none"
      >
        <rect x="1" y="3" width="14" height="12" rx="2" stroke="rgba(180,200,235,0.45)" strokeWidth="1.5" />
        <path d="M1 7h14" stroke="rgba(180,200,235,0.45)" strokeWidth="1.5" />
        <path d="M5 1v4M11 1v4" stroke="rgba(180,200,235,0.45)" strokeWidth="1.5" strokeLinecap="round" />
      </svg>
      {/* Clear button */}
      {isClearable && value && (
        <button
          type="button"
          onClick={(e) => { e.stopPropagation(); onClear?.(); }}
          className="absolute right-2.5 top-1/2 -translate-y-1/2 text-[rgba(180,200,235,0.45)] hover:text-[var(--color-text-soft)] transition-colors"
        >
          <svg width="10" height="10" viewBox="0 0 10 10" fill="currentColor">
            <path d="M1 1l8 8M9 1L1 9" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
          </svg>
        </button>
      )}
    </div>
  );
}

// ---------------------------------------------------------------------------
// Badge
// ---------------------------------------------------------------------------

function Badge({ open }: { open: boolean }) {
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

// ---------------------------------------------------------------------------
// SkeletonCard
// ---------------------------------------------------------------------------

function SkeletonCard() {
  return (
    <div className="card rounded-xl p-4 flex flex-col gap-3 animate-pulse">
      <div className="flex items-start justify-between gap-2">
        <div className="flex flex-col gap-1.5 flex-1">
          <div className="h-4 w-2/5 rounded bg-white/8" />
          <div className="h-3 w-1/4 rounded bg-white/5" />
        </div>
        <div className="h-5 w-14 rounded-full bg-white/5" />
      </div>
      <div className="h-px bg-white/5" />
      <div className="grid grid-cols-3 gap-3">
        <div className="h-3 rounded bg-white/5" />
        <div className="h-3 rounded bg-white/5" />
        <div className="h-3 rounded bg-white/5" />
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// EventCard
// ---------------------------------------------------------------------------

interface EventCardProps {
  event: Event;
  currentUserId?: string;
  isHostingList: boolean;
  hostingIds: Set<string>;
}

function EventCard({ event, currentUserId, isHostingList, hostingIds }: EventCardProps) {
  const isAlsoHosting = !isHostingList && hostingIds.has(event.id);
  const isCurrentUserHost = event.host_id === currentUserId;

  return (
    <Link href={`/event/${event.id}`} className="block group focus:outline-none">
      <div
        className={[
          "card rounded-xl p-4 flex flex-col gap-3 relative overflow-hidden",
          "transition-all duration-200",
          "group-hover:border-white/15 group-hover:shadow-[0_0_0_1px_rgba(255,255,255,0.04),0_30px_90px_rgba(0,0,0,0.7),0_0_40px_rgba(30,80,200,0.12)]",
          "group-focus-visible:ring-2 group-focus-visible:ring-[var(--color-accent-blue)]/50",
        ].join(" ")}
      >
        <div className="absolute top-0 left-4 right-4 h-px bg-top-edge opacity-0 group-hover:opacity-20 transition-opacity duration-200 rounded-full" />

        {/* Header row */}
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0 flex-1">
            <p className="font-semibold text-[var(--color-text)] text-sm leading-snug truncate group-hover:text-white transition-colors">
              {event.game_name}
            </p>
            {event.game_mode && (
              <p className="text-xs text-[var(--color-text-muted)] mt-0.5 truncate">
                {event.game_mode}
              </p>
            )}
          </div>
          <div className="flex items-center gap-1.5 flex-shrink-0">
            {isAlsoHosting && (
              <span className="inline-flex items-center px-2 py-0.5 rounded-full text-[0.65rem] font-semibold tracking-wide bg-[var(--color-accent-blue)]/10 text-[var(--color-accent-blue)] border border-[var(--color-accent-blue)]/20">
                Hosting
              </span>
            )}
            <Badge open={event.registration_open} />
          </div>
        </div>

        {/* Divider */}
        <div className="h-px bg-white/[0.06]" />

        {/* Details grid */}
        <div className="grid grid-cols-3 gap-x-3 gap-y-2">
          <EventDetail
            icon={
              <svg width="11" height="11" viewBox="0 0 16 16" fill="none">
                <rect x="1" y="3" width="14" height="12" rx="2" stroke="currentColor" strokeWidth="1.5" />
                <path d="M1 7h14" stroke="currentColor" strokeWidth="1.5" />
                <path d="M5 1v4M11 1v4" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
              </svg>
            }
            label="Date"
            value={DATE_FMT.format(new Date(event.event_date))}
          />
          <EventDetail
            icon={
              <svg width="11" height="11" viewBox="0 0 16 16" fill="none">
                <circle cx="8" cy="5" r="3" stroke="currentColor" strokeWidth="1.5" />
                <path d="M2 14c0-3.314 2.686-6 6-6s6 2.686 6 6" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
              </svg>
            }
            label="Host"
            value={isCurrentUserHost ? "You" : event.host_name}
          />
          <EventDetail
            icon={
              <svg width="11" height="11" viewBox="0 0 16 16" fill="none">
                <circle cx="5" cy="5" r="2.5" stroke="currentColor" strokeWidth="1.5" />
                <circle cx="11" cy="5" r="2.5" stroke="currentColor" strokeWidth="1.5" />
                <path d="M1 14c0-2.761 1.79-5 4-5M15 14c0-2.761-1.79-5-4-5M8 14c0-2.761 1.343-5 3-5s3 2.239 3 5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
              </svg>
            }
            label="Players"
            value={`${event.registered_count}`}
          />
        </div>

        {/* Chevron */}
        <div className="absolute right-4 bottom-4 text-[var(--color-text-faint)] group-hover:text-[var(--color-text-muted)] group-hover:translate-x-0.5 transition-all duration-150">
          <svg width="12" height="12" viewBox="0 0 12 12" fill="none">
            <path d="M4.5 2.5L8 6l-3.5 3.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
          </svg>
        </div>
      </div>
    </Link>
  );
}

function EventDetail({
  icon,
  label,
  value,
}: {
  icon: React.ReactNode;
  label: string;
  value: string;
}) {
  return (
    <div className="flex flex-col gap-0.5 min-w-0">
      <div className="flex items-center gap-1 text-[var(--color-text-faint)]">
        {icon}
        <span className="text-[0.6rem] font-medium tracking-wider uppercase">{label}</span>
      </div>
      <span className="text-xs text-[var(--color-text-soft)] truncate">{value}</span>
    </div>
  );
}

// ---------------------------------------------------------------------------
// EmptyState
// ---------------------------------------------------------------------------

function EmptyState({
  tab,
  time,
  hasFilters,
}: {
  tab: Tab;
  time: TimeFilter;
  hasFilters: boolean;
}) {
  return (
    <div className="flex flex-col items-center justify-center py-14 rounded-xl border border-dashed border-white/[0.08] text-center gap-3">
      <div className="w-10 h-10 rounded-full bg-white/[0.04] border border-white/[0.08] flex items-center justify-center">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" className="text-[var(--color-text-faint)]">
          <rect x="3" y="4" width="18" height="18" rx="3" stroke="currentColor" strokeWidth="1.5" />
          <path d="M3 10h18M8 2v4M16 2v4" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
        </svg>
      </div>
      <div>
        <p className="text-sm font-medium text-[var(--color-text-soft)]">
          {hasFilters
            ? "No events match your filters"
            : time === "past"
            ? "No past events"
            : tab === "hosting"
            ? "You're not hosting any upcoming events"
            : "You're not registered for any upcoming events"}
        </p>
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// LoadingSpinner
// ---------------------------------------------------------------------------

function LoadingSpinner() {
  return (
    <svg className="animate-spin" width="14" height="14" viewBox="0 0 24 24" fill="none">
      <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2" opacity="0.25" />
      <path d="M12 2a10 10 0 0 1 10 10" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
    </svg>
  );
}

// ---------------------------------------------------------------------------
// Main Page
// ---------------------------------------------------------------------------

export default function MyEventsPage() {
  const { user, isAuthenticated, isLoading: authLoading } = useAuth();
  const [isEventSheetOpen, setIsEventSheetOpen] = useState(false);

  const [activeTab, setActiveTab] = useState<Tab>("hosting");
  const [timeFilter, setTimeFilter] = useState<TimeFilter>("upcoming");

  // Filters: pending = staged in UI, applied = sent to API
  const [pendingGame, setPendingGame] = useState<SelectOption | null>(null);
  const [pendingFrom, setPendingFrom] = useState<Date | null>(null);
  const [pendingTo, setPendingTo] = useState<Date | null>(null);
  const [appliedGame, setAppliedGame] = useState<SelectOption | null>(null);
  const [appliedFrom, setAppliedFrom] = useState<Date | null>(null);
  const [appliedTo, setAppliedTo] = useState<Date | null>(null);
  const dateRangeActive = !!(appliedFrom || appliedTo);

  // Games for filter dropdown
  const [games, setGames] = useState<Game[]>([]);
  const [gamesLoading, setGamesLoading] = useState(true);
  const [gamesError, setGamesError] = useState(false);
  // gamesLoading is only meaningful once we know the user is authenticated; until then,
  // we derive a loading state from auth bootstrap without extra setState calls in effects.
  const gamesSelectLoading = isAuthenticated ? authLoading || gamesLoading : authLoading;

  // Four event buckets keyed by "tab:time"
  type BucketKey = `${Tab}:${TimeFilter}`;
  const [buckets, setBuckets] = useState<Record<BucketKey, BucketState>>({
    "hosting:upcoming": EMPTY_BUCKET,
    "hosting:past": EMPTY_BUCKET,
    "registered:upcoming": EMPTY_BUCKET,
    "registered:past": EMPTY_BUCKET,
  });

  const bucketFilters = useRef<Record<BucketKey, string>>({
    "hosting:upcoming": "",
    "hosting:past": "",
    "registered:upcoming": "",
    "registered:past": "",
  });

  const activeBucketKey: BucketKey = `${activeTab}:${timeFilter}`;
  const activeBucket = buckets[activeBucketKey];

  const hostingIds = new Set([
    ...buckets["hosting:upcoming"].events.map((e) => e.id),
    ...buckets["hosting:past"].events.map((e) => e.id),
  ]);

  // ---------------------------------------------------------------------------
  // Filter summary
  // ---------------------------------------------------------------------------

  const activeFilterLabels: string[] = [];
  if (appliedGame) activeFilterLabels.push(appliedGame.label);
  if (appliedFrom && appliedTo) {
    activeFilterLabels.push(`${DATE_FMT.format(appliedFrom)} – ${DATE_FMT.format(appliedTo)}`);
  } else if (appliedFrom) {
    activeFilterLabels.push(`From ${DATE_FMT.format(appliedFrom)}`);
  } else if (appliedTo) {
    activeFilterLabels.push(`Until ${DATE_FMT.format(appliedTo)}`);
  }

  const dateRangeValid = !pendingFrom || !pendingTo || pendingFrom <= pendingTo;

  const filterKey = `${appliedGame?.value ?? ""}|${appliedFrom?.toISOString() ?? ""}|${appliedTo?.toISOString() ?? ""}`;

  // ---------------------------------------------------------------------------
  // Load bucket
  // ---------------------------------------------------------------------------

  const loadBucket = useCallback(
    async (tab: Tab, time: TimeFilter, cursor?: string) => {
      const key: BucketKey = `${tab}:${time}`;
      const endpoint = "/users/me/events";

      const params: Record<string, string | undefined> = {};
      params.tz = USER_TZ;
      if (tab === "hosting") params.hosting = "true";
      if (appliedFrom || appliedTo) {
        if (appliedFrom) params.from = appliedFrom.toISOString().split("T")[0];
        if (appliedTo) params.to = appliedTo.toISOString().split("T")[0];
      } else {
        if (time === "past") params.past = "true";
      }
      if (appliedGame) params.game_id = appliedGame.value;
      if (cursor) params.cursor = cursor;

      setBuckets((prev) => ({
        ...prev,
        [key]: { ...prev[key], isLoadingMore: true },
      }));

      try {
        const page = await fetchEvents(endpoint, params);
        setBuckets((prev) => ({
          ...prev,
          [key]: {
            events: cursor ? [...prev[key].events, ...page.event_groups] : page.event_groups,
            nextCursor: page.next_cursor,
            hasMore: page.has_more,
            isLoadingMore: false,
            loaded: true,
          },
        }));
        bucketFilters.current[key] = filterKey;
      } catch {
        setBuckets((prev) => ({
          ...prev,
          [key]: { ...prev[key], isLoadingMore: false, loaded: true },
        }));
      }
    },
    [appliedFrom, appliedTo, appliedGame, filterKey]
  );

  // ---------------------------------------------------------------------------
  // Load active bucket on tab/time/filter change
  // ---------------------------------------------------------------------------

  useEffect(() => {
    if (authLoading || !isAuthenticated) {
      return;
    }

    const key = activeBucketKey;
    const needsReset = bucketFilters.current[key] !== filterKey;
    if (!activeBucket.loaded || needsReset) {
      if (needsReset) setBuckets((prev) => ({ ...prev, [key]: EMPTY_BUCKET }));
      loadBucket(activeTab, timeFilter);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [activeTab, timeFilter, filterKey, authLoading, isAuthenticated]);

  // ---------------------------------------------------------------------------
  // Load games
  // ---------------------------------------------------------------------------

  useEffect(() => {
    if (authLoading) {
      return;
    }
    if (!isAuthenticated) {
      return;
    }

    fetchGames()
      .then(setGames)
      .catch(() => setGamesError(true))
      .finally(() => setGamesLoading(false));
  }, [authLoading, isAuthenticated]);

  // ---------------------------------------------------------------------------
  // Game select options (grouped)
  // ---------------------------------------------------------------------------

  const systemGames = games.filter((g) => g.owner_id === null);
  const userGames = games.filter((g) => g.owner_id !== null);

  const gameSelectOptions: GroupBase<SelectOption>[] = [
    {
      label: "Standard",
      options: systemGames.map((g) => ({ value: g.id, label: g.name })),
    },
    ...(userGames.length > 0
      ? [{ label: "Custom", options: [{ value: "other", label: "Other" }] }]
      : []),
  ];

  // ---------------------------------------------------------------------------
  // Apply / clear filters
  // ---------------------------------------------------------------------------

  const resetAllBuckets = () =>
    setBuckets({
      "hosting:upcoming": EMPTY_BUCKET,
      "hosting:past": EMPTY_BUCKET,
      "registered:upcoming": EMPTY_BUCKET,
      "registered:past": EMPTY_BUCKET,
    });

  const applyFilters = () => {
    setAppliedGame(pendingGame);
    setAppliedFrom(pendingFrom);
    setAppliedTo(pendingTo);
    resetAllBuckets();
  };

  const clearFilters = () => {
    setPendingGame(null);
    setPendingFrom(null);
    setPendingTo(null);
    setAppliedGame(null);
    setAppliedFrom(null);
    setAppliedTo(null);
    resetAllBuckets();
  };

  const filtersAreDirty =
    pendingGame?.value !== appliedGame?.value ||
    pendingFrom?.toISOString() !== appliedFrom?.toISOString() ||
    pendingTo?.toISOString() !== appliedTo?.toISOString();

  // ---------------------------------------------------------------------------
  // Render
  // ---------------------------------------------------------------------------

  const isInitialLoading = !activeBucket.loaded && !activeBucket.isLoadingMore;

  return (
    <div className="flex-1 flex flex-col items-center py-8 px-4">
      <div
        className="w-full max-w-2xl flex flex-col gap-6"
        style={{ animation: "var(--animate-rise)" }}
      >
        {/* Page header */}
        <div style={{ animation: "var(--animate-rise-1)" }}>
          <h1 className="text-2xl font-semibold text-[var(--color-text)] tracking-tight">
            My Events
          </h1>
          <p className="mt-1 text-sm text-[var(--color-text-muted)]">
            {"Events you're hosting or registered in."}
          </p>
        </div>
        
        {/* Create event CTA */}
        <div className="flex justify-center" style={{ animation: "var(--animate-rise-1)" }}>
          <button
            type="button"
            onClick={() => setIsEventSheetOpen(true)}
            className="relative overflow-hidden flex items-center gap-2 px-5 py-2.5 rounded-lg
                       text-sm font-medium border border-white/10 bg-white/[0.04]
                       text-[var(--color-text)] hover:bg-white/[0.09]
                       hover:border-[var(--color-accent-blue)]/40
                       focus-visible:outline-none focus-visible:ring-2
                       focus-visible:ring-[var(--color-accent-blue)]/40
                       transition-all duration-150"
          >
            <span className="absolute top-0 left-0 right-0 h-px bg-top-edge opacity-30 rounded-full" />
            <span className="text-[var(--color-accent-blue)] text-base leading-none">+</span>
            Host an event
          </button>
        </div>

        {/* Tabs */}
        <div
          className="sticky top-0 z-20 -mx-4 px-4 py-3"
          style={{
            background: "linear-gradient(180deg, rgba(6,8,15,0.97) 80%, transparent 100%)",
            animation: "var(--animate-rise-1)",
          }}
        >
          <div className="flex items-center gap-1 p-1 rounded-xl bg-white/[0.04] border border-white/[0.07] w-full">
            {(["hosting", "registered"] as Tab[]).map((tab) => (
              <button
                key={tab}
                type="button"
                onClick={() => setActiveTab(tab)}
                className={[
                  "flex-1 py-2 px-4 rounded-lg text-sm font-medium transition-all duration-200",
                  activeTab === tab
                    ? "bg-white/[0.09] text-[var(--color-text)] shadow-sm border border-white/[0.10]"
                    : "text-[var(--color-text-muted)] hover:text-[var(--color-text-soft)] hover:bg-white/[0.03]",
                ].join(" ")}
              >
                {tab === "hosting" ? "Hosted by Me" : "Registered"}
              </button>
            ))}
          </div>
        </div>

        {/* Filters card */}
        <div
          className="card rounded-xl p-4 flex flex-col gap-3 relative overflow-hidden"
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
          <div className="flex flex-col sm:flex-row gap-2 sm:items-center">
            <div className="flex-1 flex flex-col gap-1.5">
              <label className="text-xs font-medium tracking-wide text-[var(--color-text-soft)]">From</label>
              <DatePicker
                wrapperClassName="w-full"
                selected={pendingFrom}
                onChange={(d: Date | null) => setPendingFrom(d)}
                selectsStart
                startDate={pendingFrom}
                endDate={pendingTo}
                maxDate={pendingTo ?? undefined}
                dateFormat="MMM d, yyyy"
                popperPlacement="bottom-end"
                popperProps={{
                  strategy: "fixed"
                }}
                customInput={
                  <DateInput
                    placeholder="From"
                    isClearable
                    onClear={() => setPendingFrom(null)}
                  />
                }
              />
            </div>
            <div className="flex-1 flex flex-col gap-1.5">
              <label className="text-xs font-medium tracking-wide text-[var(--color-text-soft)]">To</label>
              <DatePicker
                wrapperClassName="w-full"
                selected={pendingTo}
                onChange={(d: Date | null) => setPendingTo(d)}
                selectsEnd
                startDate={pendingFrom}
                endDate={pendingTo}
                minDate={pendingFrom ?? undefined}
                dateFormat="MMM d, yyyy"
                popperPlacement="bottom-end"
                popperProps={{
                  strategy: "fixed"
                }}
                customInput={
                  <DateInput
                    placeholder="To"
                    isClearable
                    onClear={() => setPendingTo(null)}
                  />
                }
              />
            </div>
            <button
              type="button"
              onClick={applyFilters}
              disabled={!dateRangeValid || !filtersAreDirty}
              title={!dateRangeValid ? "'From' must be before 'To'" : undefined}
              className={[
                "px-4 py-2 rounded-lg text-sm font-medium transition-all duration-150",
                "border border-[var(--color-accent-blue)]/30 bg-[var(--color-accent-blue)]/10 text-[var(--color-accent-blue)]",
                "hover:bg-[var(--color-accent-blue)]/20 hover:border-[var(--color-accent-blue)]/50",
                "disabled:opacity-35 disabled:cursor-not-allowed",
                "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--color-accent-blue)]/40",
              ].join(" ")}
            >
              Apply
            </button>
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
                onClick={clearFilters}
                className="ml-auto text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text-soft)] transition-colors underline underline-offset-2"
              >
                Clear
              </button>
            </div>
          )}
        </div>

        {/* Event list */}
        <div className="flex flex-col gap-3" style={{ animation: "var(--animate-rise-3)" }}>
          <SectionDivider
            title={activeTab === "hosting" ? "Hosted by Me" : "Registered"}
          />

          {isInitialLoading ? (
            <>
              <SkeletonCard />
              <SkeletonCard />
              <SkeletonCard />
            </>
          ) : activeBucket.events.length === 0 ? (
            <EmptyState
              tab={activeTab}
              time={timeFilter}
              hasFilters={activeFilterLabels.length > 0}
            />
          ) : (
            <>
              {activeBucket.events.map((event) => (
                <EventCard
                  key={event.id}
                  event={event}
                  currentUserId={user?.id}
                  isHostingList={activeTab === "hosting"}
                  hostingIds={hostingIds}
                />
              ))}

              {activeBucket.hasMore && (
                <button
                  type="button"
                  onClick={() =>
                    loadBucket(activeTab, timeFilter, activeBucket.nextCursor ?? undefined)
                  }
                  disabled={activeBucket.isLoadingMore}
                  className={[
                    "w-full py-3 rounded-xl text-sm font-medium transition-all duration-150",
                    "border border-white/[0.08] bg-white/[0.03] text-[var(--color-text-muted)]",
                    "hover:bg-white/[0.06] hover:border-white/[0.14] hover:text-[var(--color-text-soft)]",
                    "disabled:opacity-40 disabled:cursor-not-allowed",
                  ].join(" ")}
                >
                  {activeBucket.isLoadingMore ? (
                    <span className="flex items-center justify-center gap-2">
                      <LoadingSpinner />
                      Loading…
                    </span>
                  ) : (
                    "Load more"
                  )}
                </button>
              )}
            </>
          )}
        </div>
      </div>

      <ResponsiveSheet
        isOpen={isEventSheetOpen}
        onClose={() => setIsEventSheetOpen(false)}
        title="Host an event"
      >
        <EventForm mode="create" onCancel={() => setIsEventSheetOpen(false)} />
      </ResponsiveSheet>

      <style>{datepickerStyles}</style>
    </div>
  );
}
