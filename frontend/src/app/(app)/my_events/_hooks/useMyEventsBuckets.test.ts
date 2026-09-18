// Tests for the My Events dashboard's four-bucket (hosting/registered x upcoming/past) fetch,
// pagination, and filter-driven reset/reload logic.
import { act, renderHook, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { server } from "@/test/msw/server";
import { TEST_API_URL } from "@/test/constants";
import { buildEventsPage, buildMyEvent } from "@/test/fixtures";
import { SelectOption } from "@/app/_components/Select";
import { useMyEventsBuckets } from "./useMyEventsBuckets";

interface Props {
  authLoading: boolean;
  isAuthenticated: boolean;
  activeTab: "hosting" | "registered";
  timeFilter: "upcoming" | "past";
  appliedGame: SelectOption | null;
  appliedFrom: Date | null;
  appliedTo: Date | null;
}

const baseProps: Props = {
  authLoading: false,
  isAuthenticated: true,
  activeTab: "hosting",
  timeFilter: "upcoming",
  appliedGame: null,
  appliedFrom: null,
  appliedTo: null,
};

function requestParams(url: string): URLSearchParams {
  return new URL(url).searchParams;
}

describe("useMyEventsBuckets", () => {
  beforeEach(() => {
    server.use(http.get(`${TEST_API_URL}/users/me/events`, () => HttpResponse.json(buildEventsPage())));
  });

  it("does not fetch while auth is still loading", async () => {
    let requested = false;
    server.use(
      http.get(`${TEST_API_URL}/users/me/events`, () => {
        requested = true;
        return HttpResponse.json(buildEventsPage());
      }),
    );
    renderHook(() => useMyEventsBuckets({ ...baseProps, authLoading: true }));

    await new Promise((resolve) => setTimeout(resolve, 0));
    expect(requested).toBe(false);
  });

  it("does not fetch when unauthenticated", async () => {
    let requested = false;
    server.use(
      http.get(`${TEST_API_URL}/users/me/events`, () => {
        requested = true;
        return HttpResponse.json(buildEventsPage());
      }),
    );
    renderHook(() => useMyEventsBuckets({ ...baseProps, isAuthenticated: false }));

    await new Promise((resolve) => setTimeout(resolve, 0));
    expect(requested).toBe(false);
  });

  it("loads the active bucket on mount when authenticated", async () => {
    const event = buildMyEvent({ id: "group-1" });
    server.use(http.get(`${TEST_API_URL}/users/me/events`, () => HttpResponse.json(buildEventsPage({ event_groups: [event] }))));

    const { result } = renderHook(() => useMyEventsBuckets(baseProps));

    await waitFor(() => expect(result.current.activeBucket.loaded).toBe(true));
    expect(result.current.activeBucket.events).toEqual([event]);
  });

  it("sends hosting=true only for the hosting tab", async () => {
    let seenParams: URLSearchParams | undefined;
    server.use(
      http.get(`${TEST_API_URL}/users/me/events`, ({ request }) => {
        seenParams = requestParams(request.url);
        return HttpResponse.json(buildEventsPage());
      }),
    );
    const { result } = renderHook(() => useMyEventsBuckets({ ...baseProps, activeTab: "registered" }));
    await waitFor(() => expect(result.current.activeBucket.loaded).toBe(true));

    expect(seenParams?.get("hosting")).toBeNull();
  });

  it("sends past=true for the past tab when no date range filter is applied", async () => {
    let seenParams: URLSearchParams | undefined;
    server.use(
      http.get(`${TEST_API_URL}/users/me/events`, ({ request }) => {
        seenParams = requestParams(request.url);
        return HttpResponse.json(buildEventsPage());
      }),
    );
    const { result } = renderHook(() => useMyEventsBuckets({ ...baseProps, timeFilter: "past" }));
    await waitFor(() => expect(result.current.activeBucket.loaded).toBe(true));

    expect(seenParams?.get("past")).toBe("true");
  });

  it("sends from/to instead of past when a date range filter is applied", async () => {
    let seenParams: URLSearchParams | undefined;
    server.use(
      http.get(`${TEST_API_URL}/users/me/events`, ({ request }) => {
        seenParams = requestParams(request.url);
        return HttpResponse.json(buildEventsPage());
      }),
    );
    const from = new Date("2026-09-01T00:00:00Z");
    const to = new Date("2026-09-30T00:00:00Z");
    const { result } = renderHook(() =>
      useMyEventsBuckets({ ...baseProps, timeFilter: "past", appliedFrom: from, appliedTo: to }),
    );
    await waitFor(() => expect(result.current.activeBucket.loaded).toBe(true));

    expect(seenParams?.get("from")).toBe("2026-09-01");
    expect(seenParams?.get("to")).toBe("2026-09-30");
    expect(seenParams?.get("past")).toBeNull();
  });

  it("sends game_id when a game filter is applied", async () => {
    let seenParams: URLSearchParams | undefined;
    server.use(
      http.get(`${TEST_API_URL}/users/me/events`, ({ request }) => {
        seenParams = requestParams(request.url);
        return HttpResponse.json(buildEventsPage());
      }),
    );
    const { result } = renderHook(() =>
      useMyEventsBuckets({ ...baseProps, appliedGame: { value: "game-1", label: "Game One" } }),
    );
    await waitFor(() => expect(result.current.activeBucket.loaded).toBe(true));

    expect(seenParams?.get("game_id")).toBe("game-1");
  });

  it("switching the active tab loads that tab's bucket independently", async () => {
    server.use(
      http.get(`${TEST_API_URL}/users/me/events`, ({ request }) => {
        const hosting = requestParams(request.url).get("hosting") === "true";
        return HttpResponse.json(
          buildEventsPage({ event_groups: [buildMyEvent({ id: hosting ? "hosted-event" : "registered-event" })] }),
        );
      }),
    );
    const { result, rerender } = renderHook((props: Props) => useMyEventsBuckets(props), { initialProps: baseProps });
    await waitFor(() => expect(result.current.activeBucket.loaded).toBe(true));
    expect(result.current.activeBucket.events[0].id).toBe("hosted-event");

    rerender({ ...baseProps, activeTab: "registered" });
    await waitFor(() => expect(result.current.activeBucket.loaded).toBe(true));
    expect(result.current.activeBucket.events[0].id).toBe("registered-event");
  });

  it("hostingIds aggregates ids from both hosting buckets, not the registered ones", async () => {
    server.use(
      http.get(`${TEST_API_URL}/users/me/events`, ({ request }) => {
        const hosting = requestParams(request.url).get("hosting") === "true";
        return HttpResponse.json(
          buildEventsPage({ event_groups: [buildMyEvent({ id: hosting ? "hosted-event" : "registered-event" })] }),
        );
      }),
    );
    const { result, rerender } = renderHook((props: Props) => useMyEventsBuckets(props), { initialProps: baseProps });
    await waitFor(() => expect(result.current.activeBucket.loaded).toBe(true));

    rerender({ ...baseProps, activeTab: "registered" });
    await waitFor(() => expect(result.current.activeBucket.loaded).toBe(true));

    expect(result.current.hostingIds.has("hosted-event")).toBe(true);
    expect(result.current.hostingIds.has("registered-event")).toBe(false);
  });

  it("loadBucket appends events to the existing list when given a cursor (pagination)", async () => {
    const firstPage = buildEventsPage({ event_groups: [buildMyEvent({ id: "e1" })], has_more: true, next_cursor: "cursor-1" });
    const secondPage = buildEventsPage({ event_groups: [buildMyEvent({ id: "e2" })], has_more: false, next_cursor: null });
    let callCount = 0;
    server.use(
      http.get(`${TEST_API_URL}/users/me/events`, () => {
        callCount += 1;
        return HttpResponse.json(callCount === 1 ? firstPage : secondPage);
      }),
    );
    const { result } = renderHook(() => useMyEventsBuckets(baseProps));
    await waitFor(() => expect(result.current.activeBucket.loaded).toBe(true));
    expect(result.current.activeBucket.events.map((e) => e.id)).toEqual(["e1"]);

    await act(async () => {
      await result.current.loadBucket("hosting", "upcoming", "cursor-1");
    });

    await waitFor(() => expect(result.current.activeBucket.events.map((e) => e.id)).toEqual(["e1", "e2"]));
  });

  it("resets and reloads the active bucket when the applied filters change", async () => {
    let callCount = 0;
    server.use(
      http.get(`${TEST_API_URL}/users/me/events`, () => {
        callCount += 1;
        return HttpResponse.json(buildEventsPage({ event_groups: [buildMyEvent({ id: `event-${callCount}` })] }));
      }),
    );
    const { result, rerender } = renderHook((props: Props) => useMyEventsBuckets(props), { initialProps: baseProps });
    await waitFor(() => expect(result.current.activeBucket.events[0]?.id).toBe("event-1"));

    rerender({ ...baseProps, appliedGame: { value: "game-1", label: "Game One" } });

    await waitFor(() => expect(result.current.activeBucket.events[0]?.id).toBe("event-2"));
  });

  it("does not refetch when re-rendered with the same tab/time/filters", async () => {
    let callCount = 0;
    server.use(
      http.get(`${TEST_API_URL}/users/me/events`, () => {
        callCount += 1;
        return HttpResponse.json(buildEventsPage());
      }),
    );
    const { result, rerender } = renderHook((props: Props) => useMyEventsBuckets(props), { initialProps: baseProps });
    await waitFor(() => expect(result.current.activeBucket.loaded).toBe(true));

    rerender({ ...baseProps });
    await new Promise((resolve) => setTimeout(resolve, 0));

    expect(callCount).toBe(1);
  });

  it("resetAllBuckets clears every bucket back to its empty/unloaded state", async () => {
    const { result } = renderHook(() => useMyEventsBuckets(baseProps));
    await waitFor(() => expect(result.current.activeBucket.loaded).toBe(true));

    act(() => result.current.resetAllBuckets());

    await waitFor(() => expect(result.current.activeBucket.loaded).toBe(false));
    expect(result.current.activeBucket.events).toEqual([]);
  });
});
