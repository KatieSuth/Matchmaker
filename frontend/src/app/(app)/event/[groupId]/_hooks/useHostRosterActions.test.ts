// Tests for host-only roster management: lock-in/delete teams, swap players, move to
// subs/unplaced, and lobby-host reassignment.
import { act, renderHook, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { server } from "@/test/msw/server";
import { TEST_API_URL } from "@/test/constants";
import {
  buildEventGroupDetail,
  buildEventGroupEvent,
  buildEventLobby,
  buildEventTeam,
  buildLobbyPlayer,
} from "@/test/fixtures";
import { PlayerPlacement } from "../_types";
import { useHostRosterActions } from "./useHostRosterActions";

function setup(group = buildEventGroupDetail()) {
  const loadGroup = vi.fn().mockResolvedValue(undefined);
  const setWorking = vi.fn();
  const setPageError = vi.fn();
  const view = renderHook(() => useHostRosterActions(group, loadGroup, setWorking, setPageError));
  return { ...view, loadGroup, setWorking, setPageError, group };
}

function teamPlacement(overrides: Partial<PlayerPlacement> = {}): PlayerPlacement {
  return {
    eventId: "event-1",
    userId: "user-1",
    discordName: "TestUser#0001",
    lobbyId: "lobby-1",
    sourceLobbyIndex: 0,
    teamNumber: 1,
    ...overrides,
  };
}

describe("useHostRosterActions", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("lockInTeams", () => {
    it("creates teams, reloads the group, and does not open the sub-capacity sheet by default", async () => {
      server.use(http.post(`${TEST_API_URL}/events/:groupId/teams`, () => HttpResponse.json({ sub_capacity_adjusted: false })));
      const { result, loadGroup } = setup();

      await act(async () => {
        await result.current.lockInTeams();
      });

      expect(loadGroup).toHaveBeenCalledTimes(1);
      expect(result.current.subCapacitySheetOpen).toBe(false);
    });

    it("opens the sub-capacity sheet when the backend had to adjust sub capacity", async () => {
      server.use(http.post(`${TEST_API_URL}/events/:groupId/teams`, () => HttpResponse.json({ sub_capacity_adjusted: true })));
      const { result } = setup();

      await act(async () => {
        await result.current.lockInTeams();
      });

      expect(result.current.subCapacitySheetOpen).toBe(true);
    });

    it("surfaces an API error on failure", async () => {
      server.use(http.post(`${TEST_API_URL}/events/:groupId/teams`, () => HttpResponse.json({ message: "Not enough players" }, { status: 400 })));
      const { result, setPageError, setWorking } = setup();

      await act(async () => {
        await result.current.lockInTeams();
      });

      expect(setPageError).toHaveBeenCalledWith("Not enough players");
      expect(setWorking).toHaveBeenLastCalledWith(false);
    });
  });

  describe("confirmDeleteTeams", () => {
    it("deletes teams and reloads the group", async () => {
      let deleted = false;
      server.use(http.delete(`${TEST_API_URL}/events/:groupId/teams`, () => { deleted = true; return new HttpResponse(null, { status: 204 }); }));
      const { result, loadGroup } = setup();

      act(() => result.current.confirmDeleteTeams());
      await waitFor(() => expect(loadGroup).toHaveBeenCalledTimes(1));

      expect(deleted).toBe(true);
      expect(result.current.warningSheetOpen).toBe(false);
    });
  });

  describe("swap sheet", () => {
    it("openSwapSheet/closeSwapSheet manage sheet state", () => {
      const { result } = setup();
      const placement = teamPlacement();

      act(() => result.current.openSwapSheet(placement));
      expect(result.current.swapSheetOpen).toBe(true);
      expect(result.current.pendingSwap).toEqual(placement);

      act(() => result.current.setSwapTargetUserId("target-1"));
      act(() => result.current.closeSwapSheet());

      expect(result.current.swapSheetOpen).toBe(false);
      expect(result.current.pendingSwap).toBeNull();
      expect(result.current.swapTargetUserId).toBe("");
    });

    it("swapCandidateOptions lists eligible players for the pending swap's event", () => {
      const targetPlayer = buildLobbyPlayer({ user_id: "target-1", display_name: "Target Player" });
      const lobby = buildEventLobby({ id: "lobby-1", teams: [buildEventTeam({ team_number: 2, players: [targetPlayer] })] });
      const event = buildEventGroupEvent({ id: "event-1", lobbies: [lobby] });
      const group = buildEventGroupDetail({ events: [event] });
      const { result } = setup(group);

      act(() => result.current.openSwapSheet(teamPlacement({ eventId: "event-1", lobbyId: "other-lobby" })));

      expect(result.current.swapCandidateOptions.some((opt) => opt.value === "target-1")).toBe(true);
    });

    it("handleSwapSubmit swaps players, reloads, and closes the sheet on success", async () => {
      let receivedBody: unknown;
      server.use(
        http.post(`${TEST_API_URL}/registrations/:eventId/player-swap`, async ({ request }) => {
          receivedBody = await request.json();
          return new HttpResponse(null, { status: 204 });
        }),
      );
      const { result, loadGroup } = setup();
      act(() => result.current.openSwapSheet(teamPlacement({ userId: "user-1" })));
      act(() => result.current.setSwapTargetUserId("user-2"));

      await act(async () => {
        await result.current.handleSwapSubmit();
      });

      expect(receivedBody).toEqual({ user_id_a: "user-1", user_id_b: "user-2" });
      expect(loadGroup).toHaveBeenCalledTimes(1);
      expect(result.current.swapSheetOpen).toBe(false);
    });

    it("handleSwapSubmit surfaces a swap-scoped error and keeps the sheet open on failure", async () => {
      server.use(http.post(`${TEST_API_URL}/registrations/:eventId/player-swap`, () => HttpResponse.json({ message: "Cannot swap" }, { status: 400 })));
      const { result } = setup();
      act(() => result.current.openSwapSheet(teamPlacement()));
      act(() => result.current.setSwapTargetUserId("user-2"));

      await act(async () => {
        await result.current.handleSwapSubmit();
      });

      expect(result.current.swapError).toBe("Cannot swap");
      expect(result.current.swapSheetOpen).toBe(true);
    });

    it("handleSwapSubmit is a no-op without a pending swap or target", async () => {
      const { result, loadGroup } = setup();
      await act(async () => {
        await result.current.handleSwapSubmit();
      });
      expect(loadGroup).not.toHaveBeenCalled();
    });
  });

  describe("move to unplaced / subs", () => {
    it("handleMoveToUnplaced calls the API and reloads via withHostAction", async () => {
      let receivedBody: unknown;
      server.use(
        http.post(`${TEST_API_URL}/registrations/:eventId/sub-to-unplaced`, async ({ request }) => {
          receivedBody = await request.json();
          return new HttpResponse(null, { status: 204 });
        }),
      );
      const { result, loadGroup } = setup();

      act(() => result.current.handleMoveToUnplaced(teamPlacement({ userId: "sub-1" })));
      await waitFor(() => expect(loadGroup).toHaveBeenCalledTimes(1));

      expect(receivedBody).toEqual({ user_id: "sub-1" });
    });

    it("handleMoveToSubs auto-assigns to the only lobby when there's just one", async () => {
      let receivedBody: unknown;
      server.use(
        http.post(`${TEST_API_URL}/registrations/:eventId/unplaced-to-subs`, async ({ request }) => {
          receivedBody = await request.json();
          return new HttpResponse(null, { status: 204 });
        }),
      );
      const lobby = buildEventLobby({ id: "lobby-only" });
      const event = buildEventGroupEvent({ id: "event-1", lobbies: [lobby] });
      const group = buildEventGroupDetail({ events: [event] });
      const { result, loadGroup } = setup(group);

      act(() => result.current.handleMoveToSubs(teamPlacement({ eventId: "event-1", userId: "unplaced-1" })));
      await waitFor(() => expect(loadGroup).toHaveBeenCalledTimes(1));

      expect(receivedBody).toEqual({ user_id: "unplaced-1", lobby_id: "lobby-only" });
      expect(result.current.moveToSubsSheetOpen).toBe(false);
    });

    it("handleMoveToSubs opens the lobby-picker sheet when there are multiple lobbies", () => {
      const event = buildEventGroupEvent({
        id: "event-1",
        lobbies: [buildEventLobby({ id: "lobby-1" }), buildEventLobby({ id: "lobby-2" })],
      });
      const group = buildEventGroupDetail({ events: [event] });
      const { result } = setup(group);

      act(() => result.current.handleMoveToSubs(teamPlacement({ eventId: "event-1", userId: "unplaced-1" })));

      expect(result.current.moveToSubsSheetOpen).toBe(true);
      expect(result.current.pendingMoveToSubs?.userId).toBe("unplaced-1");
      expect(result.current.moveToSubsLobbyOptions).toEqual([
        { value: "lobby-1", label: "Lobby 1" },
        { value: "lobby-2", label: "Lobby 2" },
      ]);
    });

    it("handleMoveToSubsSubmit saves the picked lobby, reloads, and closes the sheet", async () => {
      let receivedBody: unknown;
      server.use(
        http.post(`${TEST_API_URL}/registrations/:eventId/unplaced-to-subs`, async ({ request }) => {
          receivedBody = await request.json();
          return new HttpResponse(null, { status: 204 });
        }),
      );
      const event = buildEventGroupEvent({
        id: "event-1",
        lobbies: [buildEventLobby({ id: "lobby-1" }), buildEventLobby({ id: "lobby-2" })],
      });
      const group = buildEventGroupDetail({ events: [event] });
      const { result, loadGroup } = setup(group);
      act(() => result.current.handleMoveToSubs(teamPlacement({ eventId: "event-1", userId: "unplaced-1" })));
      act(() => result.current.setMoveToSubsLobbyId("lobby-2"));

      await act(async () => {
        await result.current.handleMoveToSubsSubmit();
      });

      expect(receivedBody).toEqual({ user_id: "unplaced-1", lobby_id: "lobby-2" });
      expect(loadGroup).toHaveBeenCalledTimes(1);
      expect(result.current.moveToSubsSheetOpen).toBe(false);
    });

    it("handleMoveToSubsSubmit surfaces an error and keeps the sheet open on failure", async () => {
      server.use(http.post(`${TEST_API_URL}/registrations/:eventId/unplaced-to-subs`, () => HttpResponse.json({ message: "Lobby full" }, { status: 400 })));
      const event = buildEventGroupEvent({
        id: "event-1",
        lobbies: [buildEventLobby({ id: "lobby-1" }), buildEventLobby({ id: "lobby-2" })],
      });
      const group = buildEventGroupDetail({ events: [event] });
      const { result } = setup(group);
      act(() => result.current.handleMoveToSubs(teamPlacement({ eventId: "event-1", userId: "unplaced-1" })));
      act(() => result.current.setMoveToSubsLobbyId("lobby-2"));

      await act(async () => {
        await result.current.handleMoveToSubsSubmit();
      });

      expect(result.current.moveToSubsError).toBe("Lobby full");
      expect(result.current.moveToSubsSheetOpen).toBe(true);
    });

    it("closeMoveToSubsSheet clears all move-to-subs state", () => {
      const event = buildEventGroupEvent({
        id: "event-1",
        lobbies: [buildEventLobby({ id: "lobby-1" }), buildEventLobby({ id: "lobby-2" })],
      });
      const group = buildEventGroupDetail({ events: [event] });
      const { result } = setup(group);
      act(() => result.current.handleMoveToSubs(teamPlacement({ eventId: "event-1" })));

      act(() => result.current.closeMoveToSubsSheet());

      expect(result.current.moveToSubsSheetOpen).toBe(false);
      expect(result.current.pendingMoveToSubs).toBeNull();
      expect(result.current.moveToSubsLobbyId).toBe("");
    });
  });

  describe("lobby host reassignment", () => {
    it("handleMakeLobbyHost submits immediately when the target player can lobby-host", async () => {
      const player = buildLobbyPlayer({ user_id: "user-1", can_lobby_host: true });
      const lobby = buildEventLobby({ id: "lobby-1", teams: [buildEventTeam({ team_number: 1, players: [player] })] });
      const event = buildEventGroupEvent({ id: "event-1", lobbies: [lobby] });
      const group = buildEventGroupDetail({ events: [event] });
      let receivedBody: unknown;
      server.use(
        http.post(`${TEST_API_URL}/registrations/:eventId/lobby-host`, async ({ request }) => {
          receivedBody = await request.json();
          return new HttpResponse(null, { status: 204 });
        }),
      );
      const { result, loadGroup } = setup(group);

      act(() => result.current.handleMakeLobbyHost(teamPlacement({ eventId: "event-1", userId: "user-1", lobbyId: "lobby-1" })));
      await waitFor(() => expect(loadGroup).toHaveBeenCalledTimes(1));

      expect(receivedBody).toEqual({ user_id: "user-1" });
      expect(result.current.lobbyHostConfirmOpen).toBe(false);
    });

    it("handleMakeLobbyHost opens a confirmation with volunteer options when the target cannot lobby-host", () => {
      const target = buildLobbyPlayer({ user_id: "user-1", can_lobby_host: false });
      const volunteer = buildLobbyPlayer({ user_id: "user-2", can_lobby_host: true, display_name: "Volunteer" });
      const lobby = buildEventLobby({
        id: "lobby-1",
        host_id: "user-2",
        teams: [buildEventTeam({ team_number: 1, players: [target, volunteer] })],
      });
      const event = buildEventGroupEvent({ id: "event-1", lobbies: [lobby] });
      const group = buildEventGroupDetail({ events: [event] });
      const { result } = setup(group);

      act(() => result.current.handleMakeLobbyHost(teamPlacement({ eventId: "event-1", userId: "user-1", lobbyId: "lobby-1" })));

      expect(result.current.lobbyHostConfirmOpen).toBe(true);
      expect(result.current.pendingLobbyHostChange?.volunteerOptions).toEqual([
        { userId: "user-2", discordName: "Volunteer @TestUser#0001", teamNumber: 1, isCurrentHost: true },
      ]);
    });

    it("handleMakeLobbyHost is a no-op for a sub/unplaced placement (not team-assigned)", () => {
      const event = buildEventGroupEvent({ id: "event-1", lobbies: [buildEventLobby({ id: "lobby-1" })] });
      const group = buildEventGroupDetail({ events: [event] });
      const { result } = setup(group);

      act(() => result.current.handleMakeLobbyHost(teamPlacement({ eventId: "event-1", lobbyId: "lobby-1", teamNumber: null })));

      expect(result.current.lobbyHostConfirmOpen).toBe(false);
    });

    it("submitLobbyHostChange sets the host, reloads, and closes the confirm sheet", async () => {
      server.use(http.post(`${TEST_API_URL}/registrations/:eventId/lobby-host`, () => new HttpResponse(null, { status: 204 })));
      const { result, loadGroup } = setup();

      await act(async () => {
        await result.current.submitLobbyHostChange(teamPlacement());
      });

      expect(loadGroup).toHaveBeenCalledTimes(1);
      expect(result.current.lobbyHostConfirmOpen).toBe(false);
    });

    it("submitLobbyHostChange surfaces a page error on failure", async () => {
      server.use(http.post(`${TEST_API_URL}/registrations/:eventId/lobby-host`, () => HttpResponse.json({ message: "boom" }, { status: 500 })));
      const { result, setPageError } = setup();

      await act(async () => {
        await result.current.submitLobbyHostChange(teamPlacement());
      });

      expect(setPageError).toHaveBeenCalledWith("boom");
    });

    it("closeLobbyHostConfirm clears the pending change", () => {
      const target = buildLobbyPlayer({ user_id: "user-1", can_lobby_host: false });
      const lobby = buildEventLobby({ id: "lobby-1", teams: [buildEventTeam({ team_number: 1, players: [target] })] });
      const event = buildEventGroupEvent({ id: "event-1", lobbies: [lobby] });
      const group = buildEventGroupDetail({ events: [event] });
      const { result } = setup(group);
      act(() => result.current.handleMakeLobbyHost(teamPlacement({ eventId: "event-1", userId: "user-1", lobbyId: "lobby-1" })));

      act(() => result.current.closeLobbyHostConfirm());

      expect(result.current.lobbyHostConfirmOpen).toBe(false);
      expect(result.current.pendingLobbyHostChange).toBeNull();
    });
  });
});
