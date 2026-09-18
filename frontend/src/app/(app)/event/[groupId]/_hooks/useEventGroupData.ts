"use client";

// Loads the event group detail payload, gates it behind the Discord-guild access check, and
// tracks which game tab is active. This is the primary data source the rest of the page composes.
import { useCallback, useEffect, useRef, useState } from "react";
import { isCanceledError } from "@/app/_lib/http";
import {
  DiscordGuildRestrictionDetails,
  extractDiscordGuildRestriction,
  fetchGameRanks,
} from "@/app/_services/games";
import { fetchEventGroup, fetchEventGroupAccess } from "@/app/_services/events";
import { EventGroupDetail, GameRank, User } from "@/app/_types/types";

export function useEventGroupData(
  groupId: string | undefined,
  user: User | null,
  authLoading: boolean,
  isAuthenticated: boolean,
) {
  const [group, setGroupState] = useState<EventGroupDetail | null>(null);
  const [pageError, setPageError] = useState<string | null>(null);
  const [accessChecking, setAccessChecking] = useState(false);
  const [accessDenial, setAccessDenial] = useState<DiscordGuildRestrictionDetails | null>(null);
  const [loading, setLoading] = useState(true);
  const [activeEventId, setActiveEventIdState] = useState<string | null>(null);
  const [gameRanks, setGameRanks] = useState<GameRank[]>([]);

  // Mirrors of `group`/`activeEventId` that `loadGroup` reads synchronously. Reading React state
  // via a functional `setState` "peek" (the previous approach) only reflects the latest value
  // when React's eager-update-bailout optimization happens to fire, which it skips once another
  // update is already pending on the fiber (e.g. the `setAccessChecking(true)` just before this
  // runs) — silently leaving `loading` stuck at its initial value. Plain refs, updated in lockstep
  // with the state setters below, are always current regardless of React's internal scheduling.
  const groupRef = useRef<EventGroupDetail | null>(null);
  const activeEventIdRef = useRef<string | null>(null);

  /** Sets `group` state and its same-tick-readable ref together (see note above). */
  const setGroup = useCallback((next: EventGroupDetail | null) => {
    groupRef.current = next;
    setGroupState(next);
  }, []);

  /** Sets `activeEventId` state and its same-tick-readable ref together (see note above). */
  const setActiveEventId = useCallback((next: string | null) => {
    activeEventIdRef.current = next;
    setActiveEventIdState(next);
  }, []);

  /** Refetches the group (and, for the owner, the game's rank list) without dropping current UI state. */
  const loadGroup = useCallback(
    async (signal?: AbortSignal) => {
      if (!groupId) return;
      // Only show the full-page spinner for the very first load; background refreshes (e.g.
      // after a host action) keep the current content on screen while refetching.
      const showFullPageLoading = groupRef.current === null;
      if (showFullPageLoading) {
        setLoading(true);
      }
      setPageError(null);
      try {
        const data = await fetchEventGroup(groupId, signal);
        if (signal?.aborted) return;
        setGroup(data);
        if (user?.id === data.owner_id) {
          const ranks = await fetchGameRanks(data.game_id, signal);
          if (signal?.aborted) return;
          setGameRanks(ranks);
        } else {
          setGameRanks([]);
        }
        const currentActiveEventId = activeEventIdRef.current;
        if (!currentActiveEventId || !data.events.some((event) => event.id === currentActiveEventId)) {
          setActiveEventId(data.events[0]?.id ?? null);
        }
      } catch (err) {
        const canceled = isCanceledError(err) || signal?.aborted;
        if (canceled) return;
        const restriction = extractDiscordGuildRestriction(err);
        if (restriction) {
          setAccessDenial(restriction);
          setGroup(null);
          return;
        }
        setPageError("Could not load this event group.");
      } finally {
        if (!signal?.aborted && showFullPageLoading) {
          setLoading(false);
        }
      }
    },
    // Deliberately excludes `activeEventId`: it's read via `activeEventIdRef` instead so that
    // selecting a different active event tab doesn't change `loadGroup`'s identity and re-trigger
    // the access-check effect below (which depends on `loadGroup`).
    [groupId, user, setActiveEventId, setGroup],
  );

  // Discord-guild access check runs before the first load; a successful check (or none required)
  // is followed immediately by loadGroup.
  useEffect(() => {
    if (authLoading) return;
    if (!isAuthenticated || !groupId) return;
    const ac = new AbortController();
    const timer = window.setTimeout(() => {
      setAccessChecking(true);
      setAccessDenial(null);
      void (async () => {
        try {
          await fetchEventGroupAccess(groupId, ac.signal);
          if (ac.signal.aborted) return;
          await loadGroup(ac.signal);
        } catch (err) {
          const canceled = isCanceledError(err) || ac.signal.aborted;
          if (canceled) return;
          const restriction = extractDiscordGuildRestriction(err);
          if (restriction) {
            setAccessDenial(restriction);
            setGroup(null);
            setLoading(false);
            return;
          }
          setPageError("Could not load this event group.");
          setLoading(false);
        } finally {
          if (!ac.signal.aborted) {
            setAccessChecking(false);
          }
        }
      })();
    }, 0);
    return () => {
      window.clearTimeout(timer);
      ac.abort();
    };
  }, [authLoading, groupId, isAuthenticated, loadGroup, setGroup]);

  return {
    group,
    setGroup,
    pageError,
    setPageError,
    accessChecking,
    accessDenial,
    setAccessDenial,
    loading,
    gameRanks,
    loadGroup,
    activeEventId,
    setActiveEventId,
  };
}
