// Tests for the event group detail page's primary data source: Discord-guild access gating,
// group loading, owner-only rank loading, and Discord-lock error handling.
import { act, renderHook, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { server } from "@/test/msw/server";
import { TEST_API_URL } from "@/test/constants";
import { buildEventGroupDetail, buildEventGroupEvent, buildGameRank, buildUser } from "@/test/fixtures";
import { useEventGroupData } from "./useEventGroupData";

const owner = buildUser({ id: "owner-1" });
const viewer = buildUser({ id: "viewer-1" });

describe("useEventGroupData", () => {
  beforeEach(() => {
    server.use(http.get(`${TEST_API_URL}/events/:groupId/access`, () => new HttpResponse(null, { status: 204 })));
  });

  it("does not load anything while auth is still loading", async () => {
    let requested = false;
    server.use(http.get(`${TEST_API_URL}/events/:groupId/access`, () => { requested = true; return new HttpResponse(null, { status: 204 }); }));
    renderHook(() => useEventGroupData("group-1", viewer, true, false));
    await new Promise((resolve) => setTimeout(resolve, 20));
    expect(requested).toBe(false);
  });

  it("does not load when unauthenticated", async () => {
    let requested = false;
    server.use(http.get(`${TEST_API_URL}/events/:groupId/access`, () => { requested = true; return new HttpResponse(null, { status: 204 }); }));
    renderHook(() => useEventGroupData("group-1", null, false, false));
    await new Promise((resolve) => setTimeout(resolve, 20));
    expect(requested).toBe(false);
  });

  it("checks Discord-guild access, then loads the group and sets the first event active", async () => {
    const event = buildEventGroupEvent({ id: "event-1" });
    const group = buildEventGroupDetail({ id: "group-1", owner_id: owner.id, events: [event] });
    server.use(http.get(`${TEST_API_URL}/events/group-1`, () => HttpResponse.json(group)));

    const { result } = renderHook(() => useEventGroupData("group-1", viewer, false, true));

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.group).toEqual(group);
    expect(result.current.activeEventId).toBe("event-1");
    expect(result.current.pageError).toBeNull();
  });

  it("loads the group's rank list only when the viewer is the owner", async () => {
    const group = buildEventGroupDetail({ id: "group-1", owner_id: owner.id, game_id: "game-1" });
    server.use(
      http.get(`${TEST_API_URL}/events/group-1`, () => HttpResponse.json(group)),
      http.get(`${TEST_API_URL}/games/game-1/ranks`, () => HttpResponse.json([buildGameRank({ name: "Gold" })])),
    );

    const { result } = renderHook(() => useEventGroupData("group-1", owner, false, true));

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.gameRanks.map((r) => r.name)).toEqual(["Gold"]);
  });

  it("does not fetch ranks for a non-owner viewer", async () => {
    const group = buildEventGroupDetail({ id: "group-1", owner_id: owner.id, game_id: "game-1" });
    let ranksRequested = false;
    server.use(
      http.get(`${TEST_API_URL}/events/group-1`, () => HttpResponse.json(group)),
      http.get(`${TEST_API_URL}/games/game-1/ranks`, () => {
        ranksRequested = true;
        return HttpResponse.json([]);
      }),
    );

    const { result } = renderHook(() => useEventGroupData("group-1", viewer, false, true));

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(ranksRequested).toBe(false);
    expect(result.current.gameRanks).toEqual([]);
  });

  it("sets a Discord access-denial when the access check is guild-restricted", async () => {
    server.use(
      http.get(`${TEST_API_URL}/events/:groupId/access`, () =>
        HttpResponse.json(
          {
            message: "restricted",
            details: {
              code: "discord_guild_restricted",
              event_title: "Friday Customs",
              event_named: true,
              discord_guilds: [{ id: "g1", name: "Server A" }],
            },
          },
          { status: 403 },
        ),
      ),
    );

    const { result } = renderHook(() => useEventGroupData("group-1", viewer, false, true));

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.accessDenial).toMatchObject({ code: "discord_guild_restricted", event_title: "Friday Customs" });
    expect(result.current.group).toBeNull();
  });

  it("sets a generic page error when the access check fails for another reason", async () => {
    server.use(http.get(`${TEST_API_URL}/events/:groupId/access`, () => HttpResponse.json({ message: "boom" }, { status: 500 })));

    const { result } = renderHook(() => useEventGroupData("group-1", viewer, false, true));

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.pageError).toBe("Could not load this event group.");
  });

  it("sets a generic page error when loading the group itself fails", async () => {
    server.use(http.get(`${TEST_API_URL}/events/group-1`, () => HttpResponse.json({ message: "boom" }, { status: 500 })));

    const { result } = renderHook(() => useEventGroupData("group-1", viewer, false, true));

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.pageError).toBe("Could not load this event group.");
  });

  it("loadGroup does not show the full-page spinner on a refresh (group already loaded)", async () => {
    const group = buildEventGroupDetail({ id: "group-1", events: [buildEventGroupEvent({ id: "event-1" })] });
    server.use(http.get(`${TEST_API_URL}/events/group-1`, () => HttpResponse.json(group)));
    const { result } = renderHook(() => useEventGroupData("group-1", viewer, false, true));
    await waitFor(() => expect(result.current.loading).toBe(false));

    await act(async () => {
      await result.current.loadGroup();
    });

    // loading never flips back to true for a background refresh once data is present.
    expect(result.current.loading).toBe(false);
  });

  it("loadGroup keeps the current activeEventId when it still exists in the refreshed group", async () => {
    const group = buildEventGroupDetail({
      id: "group-1",
      events: [buildEventGroupEvent({ id: "event-1" }), buildEventGroupEvent({ id: "event-2" })],
    });
    server.use(http.get(`${TEST_API_URL}/events/group-1`, () => HttpResponse.json(group)));
    const { result } = renderHook(() => useEventGroupData("group-1", viewer, false, true));
    await waitFor(() => expect(result.current.loading).toBe(false));

    act(() => result.current.setActiveEventId("event-2"));
    await act(async () => {
      await result.current.loadGroup();
    });

    expect(result.current.activeEventId).toBe("event-2");
  });

  it("loadGroup resets activeEventId when it no longer exists in the refreshed group", async () => {
    let callCount = 0;
    server.use(
      http.get(`${TEST_API_URL}/events/group-1`, () => {
        callCount += 1;
        const eventId = callCount === 1 ? "event-1" : "event-2";
        return HttpResponse.json(buildEventGroupDetail({ id: "group-1", events: [buildEventGroupEvent({ id: eventId })] }));
      }),
    );
    const { result } = renderHook(() => useEventGroupData("group-1", viewer, false, true));
    await waitFor(() => expect(result.current.activeEventId).toBe("event-1"));

    await act(async () => {
      await result.current.loadGroup();
    });

    expect(result.current.activeEventId).toBe("event-2");
  });

  it("does nothing when groupId is undefined", async () => {
    const { result } = renderHook(() => useEventGroupData(undefined, viewer, false, true));
    await new Promise((resolve) => setTimeout(resolve, 20));
    expect(result.current.group).toBeNull();
  });
});
