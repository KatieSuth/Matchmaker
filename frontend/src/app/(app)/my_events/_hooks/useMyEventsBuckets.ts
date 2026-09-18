"use client";

// Manages the four hosting/registered x upcoming/past event buckets: fetching, pagination,
// and re-fetching when the active tab/time/filter combination changes.
import { useCallback, useEffect, useRef, useState } from "react";
import { getUserTimeZone } from "@/app/_lib/dateTime";
import { isCanceledError } from "@/app/_lib/http";
import { fetchMyEvents } from "@/app/_services/events";
import { SelectOption } from "@/app/_components/Select";
import { BucketKey, BucketState, EMPTY_BUCKET, Tab, TimeFilter } from "../_types";

const INITIAL_BUCKETS: Record<BucketKey, BucketState> = {
  "hosting:upcoming": EMPTY_BUCKET,
  "hosting:past": EMPTY_BUCKET,
  "registered:upcoming": EMPTY_BUCKET,
  "registered:past": EMPTY_BUCKET,
};

interface UseMyEventsBucketsOptions {
  authLoading: boolean;
  isAuthenticated: boolean;
  activeTab: Tab;
  timeFilter: TimeFilter;
  appliedGame: SelectOption | null;
  appliedFrom: Date | null;
  appliedTo: Date | null;
}

/**
 * Owns bucket state/pagination for the My Events dashboard. Callers own the tab/time/filter
 * selection state; this hook reacts to it, loads the active bucket, and exposes `loadBucket`
 * for "Load more" and `resetAllBuckets` for filter apply/clear.
 */
export function useMyEventsBuckets({
  authLoading,
  isAuthenticated,
  activeTab,
  timeFilter,
  appliedGame,
  appliedFrom,
  appliedTo,
}: UseMyEventsBucketsOptions) {
  const [buckets, setBuckets] = useState<Record<BucketKey, BucketState>>(INITIAL_BUCKETS);

  // Tracks which applied-filter combination each bucket was last loaded with, so we can tell
  // a filter change apart from simply revisiting an already-loaded tab/time combination.
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

  const filterKey = `${appliedGame?.value ?? ""}|${appliedFrom?.toISOString() ?? ""}|${appliedTo?.toISOString() ?? ""}`;

  const loadBucket = useCallback(
    async (tab: Tab, time: TimeFilter, cursor?: string, signal?: AbortSignal) => {
      const key: BucketKey = `${tab}:${time}`;
      const params: Record<string, string | undefined> = {};
      // Read fresh (rather than a module-scope constant) so the value isn't baked in at import
      // time — keeps this overridable/deterministic in tests without `vi.resetModules()`.
      params.tz = getUserTimeZone();
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
        const page = await fetchMyEvents(params, signal);
        if (signal?.aborted) return;
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
      } catch (err) {
        if (signal?.aborted || isCanceledError(err)) {
          setBuckets((prev) => ({
            ...prev,
            [key]: { ...prev[key], isLoadingMore: false },
          }));
          return;
        }
        setBuckets((prev) => ({
          ...prev,
          [key]: { ...prev[key], isLoadingMore: false, loaded: true },
        }));
      }
    },
    [appliedFrom, appliedTo, appliedGame, filterKey]
  );

  // Load (or reload) the active bucket whenever the tab/time selection or applied filters change.
  useEffect(() => {
    if (authLoading || !isAuthenticated) {
      return;
    }

    const key = activeBucketKey;
    const needsReset = bucketFilters.current[key] !== filterKey;
    if (!activeBucket.loaded || needsReset) {
      if (needsReset) setBuckets((prev) => ({ ...prev, [key]: EMPTY_BUCKET }));
      const ac = new AbortController();
      loadBucket(activeTab, timeFilter, undefined, ac.signal);
      return () => {
        ac.abort();
      };
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [activeTab, timeFilter, filterKey, authLoading, isAuthenticated]);

  const resetAllBuckets = useCallback(() => setBuckets(INITIAL_BUCKETS), []);

  return {
    activeBucket,
    hostingIds,
    loadBucket,
    resetAllBuckets,
  };
}
