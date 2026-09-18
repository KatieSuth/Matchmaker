"use client";

// My Events dashboard: list of hosted and joined event groups, filters, and create flow.
import { useState } from "react";
import "react-datepicker/dist/react-datepicker.css";
import { Game } from "@/app/_types/types";
import { useAuth } from "@/app/_context/AuthContext";
import { SelectOption } from "@/app/_components/Select";
import { SectionDivider } from "@/app/_components/SectionDivider";
import { ResponsiveSheet } from "@/app/_components/ResponsiveSheet";
import { fetchGames } from "@/app/_services/games";
import { datepickerStyles } from "@/app/_lib/styles";
import { EventForm } from "@/app/_components/forms/EventForm";
import { useCancelableFetch } from "@/app/_hooks/useCancelableFetch";
import { Tab, TimeFilter } from "./_types";
import { useMyEventsBuckets } from "./_hooks/useMyEventsBuckets";
import { SkeletonCard } from "./_components/SkeletonCard";
import { EventCard } from "./_components/EventCard";
import { EmptyState } from "./_components/EmptyState";
import { LoadingSpinner } from "./_components/LoadingSpinner";
import { EventFilters } from "./_components/EventFilters";

export default function MyEventsPage() {
  const { user, isAuthenticated, isLoading: authLoading } = useAuth();
  const [isEventSheetOpen, setIsEventSheetOpen] = useState(false);

  const [activeTab, setActiveTab] = useState<Tab>("registered");
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

  useCancelableFetch({
    fetcher: (signal) => fetchGames(signal),
    enabled: !authLoading && isAuthenticated,
    onSuccess: setGames,
    onError: () => setGamesError(true),
    onSettled: () => setGamesLoading(false),
    deps: [authLoading, isAuthenticated],
  });

  const { activeBucket, hostingIds, loadBucket, resetAllBuckets } = useMyEventsBuckets({
    authLoading,
    isAuthenticated,
    activeTab,
    timeFilter,
    appliedGame,
    appliedFrom,
    appliedTo,
  });

  const systemGames = games.filter((g) => g.owner_id === null);
  const userGames = games.filter((g) => g.owner_id !== null);

  const gameSelectOptions = [
    {
      options: systemGames.map((g) => ({ value: g.id, label: g.name })),
    },
    ...(userGames.length > 0
      ? [{ label: "Custom", options: [{ value: "other", label: "Other" }] }]
      : []),
  ];

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

  const hasActiveFilters = !!(appliedGame || appliedFrom || appliedTo);
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
          className="sticky top-0 z-20 -mx-4 px-4 py-2"
          style={{ animation: "var(--animate-rise-1)" }}
        >
          <div className="flex items-center gap-1 p-1 rounded-xl bg-white/[0.04] border border-white/[0.07] w-full">
            {(["registered", "hosting"] as Tab[]).map((tab) => (
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
        <EventFilters
          timeFilter={timeFilter}
          setTimeFilter={setTimeFilter}
          dateRangeActive={dateRangeActive}
          pendingGame={pendingGame}
          setPendingGame={setPendingGame}
          gameSelectOptions={gameSelectOptions}
          gamesSelectLoading={gamesSelectLoading}
          gamesError={gamesError}
          pendingFrom={pendingFrom}
          setPendingFrom={setPendingFrom}
          pendingTo={pendingTo}
          setPendingTo={setPendingTo}
          appliedGame={appliedGame}
          appliedFrom={appliedFrom}
          appliedTo={appliedTo}
          onApply={applyFilters}
          onClear={clearFilters}
        />

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
              hasFilters={hasActiveFilters}
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
