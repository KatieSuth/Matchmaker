// Tests for the register/edit-registration sheet: draft population, auto-open behavior for new
// guests, validation flags, and the two-step save flow (game profile, then registrations).
import { act, renderHook, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { server } from "@/test/msw/server";
import { TEST_API_URL } from "@/test/constants";
import {
  buildEventGroupDetail,
  buildEventGroupEvent,
  buildEventRegistration,
  buildGameRank,
  buildUser,
  buildUserGame,
} from "@/test/fixtures";
import { EventGroupDetail, EventRegistration, User } from "@/app/_types/types";
import { useRegistrationEditor, UseRegistrationEditorOptions } from "./useRegistrationEditor";

const user: User = buildUser({ id: "user-1", region: "AMER" });

function baseOptions(overrides: Partial<UseRegistrationEditorOptions> = {}): UseRegistrationEditorOptions {
  return {
    groupId: "group-1",
    group: buildEventGroupDetail({ id: "group-1", game_id: "game-1", region: "AMER" }),
    user,
    authLoading: false,
    isAuthenticated: true,
    pageLoading: false,
    isHost: false,
    myRegistrationsByEvent: new Map<string, EventRegistration>(),
    hasAnyLobbies: false,
    working: false,
    setWorking: vi.fn(),
    loadGroup: vi.fn().mockResolvedValue(undefined),
    setAccessDenial: vi.fn(),
    setGroup: vi.fn(),
    openDeleteAllForCurrentUserConfirmation: vi.fn(),
    ...overrides,
  };
}

function setup(overrides: Partial<UseRegistrationEditorOptions> = {}) {
  const options = baseOptions(overrides);
  const view = renderHook((props: UseRegistrationEditorOptions) => useRegistrationEditor(props), {
    initialProps: options,
  });
  return { ...view, options };
}

describe("useRegistrationEditor", () => {
  beforeEach(() => {
    server.use(
      http.get(`${TEST_API_URL}/users/me/games`, () => HttpResponse.json([])),
      http.get(`${TEST_API_URL}/games/:gameId/ranks`, () =>
        HttpResponse.json([buildGameRank({ name: "Gold" }), buildGameRank({ name: "Platinum" })]),
      ),
    );
  });

  describe("handleOpenRegistrationSheet", () => {
    it("selects every event with defaults when the user isn't registered for anything yet", async () => {
      const event1 = buildEventGroupEvent({ id: "event-1" });
      const event2 = buildEventGroupEvent({ id: "event-2" });
      const group = buildEventGroupDetail({ id: "group-1", game_id: "game-1", events: [event1, event2] });
      const { result } = setup({ group });

      await act(async () => {
        await result.current.handleOpenRegistrationSheet();
      });

      expect(result.current.registrationEditorOpen).toBe(true);
      expect(result.current.registrationDraft.selected_event_ids).toEqual(["event-1", "event-2"]);
      expect(result.current.registrationDraft.per_event["event-1"]).toEqual({ can_substitute: true, can_lobby_host: true });
    });

    it("pre-fills only the events the user is already registered for, using their saved preferences", async () => {
      const event1 = buildEventGroupEvent({ id: "event-1" });
      const event2 = buildEventGroupEvent({ id: "event-2" });
      const group = buildEventGroupDetail({ id: "group-1", game_id: "game-1", events: [event1, event2] });
      const registration = buildEventRegistration({
        event_id: "event-1",
        user_id: "user-1",
        can_substitute: false,
        can_lobby_host: true,
        duo_request: "buddy#0001",
      });
      const myRegistrationsByEvent = new Map([["event-1", registration]]);
      const { result } = setup({ group, myRegistrationsByEvent });

      await act(async () => {
        await result.current.handleOpenRegistrationSheet();
      });

      expect(result.current.registrationDraft.selected_event_ids).toEqual(["event-1"]);
      expect(result.current.registrationDraft.per_event["event-1"]).toEqual({ can_substitute: false, can_lobby_host: true });
      expect(result.current.registrationDraft.duo_request).toBe("buddy#0001");
    });

    it("adds the passed-in registration's event even if not otherwise selected", async () => {
      const event1 = buildEventGroupEvent({ id: "event-1" });
      const event2 = buildEventGroupEvent({ id: "event-2" });
      const group = buildEventGroupDetail({ id: "group-1", game_id: "game-1", events: [event1, event2] });
      const existingReg = buildEventRegistration({ event_id: "event-1", user_id: "user-1" });
      const myRegistrationsByEvent = new Map([["event-1", existingReg]]);
      const extraRegistration = buildEventRegistration({ event_id: "event-2", user_id: "user-1", can_substitute: false });
      const { result } = setup({ group, myRegistrationsByEvent });

      await act(async () => {
        await result.current.handleOpenRegistrationSheet(extraRegistration);
      });

      expect(result.current.registrationDraft.selected_event_ids).toEqual(["event-1", "event-2"]);
      expect(result.current.registrationDraft.per_event["event-2"]).toEqual({ can_substitute: false, can_lobby_host: false });
    });

    it("pre-fills the game profile draft from the user's existing game, when one exists", async () => {
      server.use(
        http.get(`${TEST_API_URL}/users/me/games`, () =>
          HttpResponse.json([buildUserGame({ game_id: "game-1", in_game_name: "MyIGN", current_rank: "gold", peak_rank: "platinum", show_rank: true })]),
        ),
      );
      const { result } = setup();

      await act(async () => {
        await result.current.handleOpenRegistrationSheet();
      });

      expect(result.current.userGameDraft).toEqual({
        game_id: "game-1",
        in_game_name: "MyIGN",
        current_rank: "gold",
        peak_rank: "platinum",
        show_rank: true,
      });
      expect(result.current.userGameRanks.map((r) => r.name)).toEqual(["Gold", "Platinum"]);
    });

    it("falls back to an empty game profile draft when the user has no existing game entry", async () => {
      const { result } = setup();

      await act(async () => {
        await result.current.handleOpenRegistrationSheet();
      });

      expect(result.current.userGameDraft).toEqual({
        game_id: "game-1",
        in_game_name: "",
        current_rank: "",
        peak_rank: "",
        show_rank: false,
      });
    });

    it("surfaces an error and resets the game profile draft if loading game data fails", async () => {
      server.use(http.get(`${TEST_API_URL}/users/me/games`, () => HttpResponse.json({ message: "boom" }, { status: 500 })));
      const { result } = setup();

      await act(async () => {
        await result.current.handleOpenRegistrationSheet();
      });

      expect(result.current.registrationError).toBe("boom");
      expect(result.current.userGameDraft.in_game_name).toBe("");
      // The sheet still opens so the user can see the error and retry.
      expect(result.current.registrationEditorOpen).toBe(true);
    });

    it("is a no-op when there is no group", async () => {
      const { result } = setup({ group: null });
      await act(async () => {
        await result.current.handleOpenRegistrationSheet();
      });
      expect(result.current.registrationEditorOpen).toBe(false);
    });
  });

  describe("handleCloseRegistrationEditor", () => {
    it("clears the error, closes the sheet, and reloads the group", async () => {
      const { result, options } = setup();
      await act(async () => {
        await result.current.handleOpenRegistrationSheet();
      });

      await act(async () => {
        await result.current.handleCloseRegistrationEditor();
      });

      expect(result.current.registrationEditorOpen).toBe(false);
      expect(result.current.registrationError).toBeNull();
      expect(options.loadGroup).toHaveBeenCalledTimes(1);
    });
  });

  describe("auto-open behavior", () => {
    it("auto-opens the registration sheet for an unregistered guest on an open group", async () => {
      const { result } = setup();
      await waitFor(() => expect(result.current.registrationEditorOpen).toBe(true));
    });

    it("does not auto-open for the host", async () => {
      const { result } = setup({ isHost: true });
      await new Promise((resolve) => setTimeout(resolve, 20));
      expect(result.current.registrationEditorOpen).toBe(false);
    });

    it("does not auto-open when the user is already registered", async () => {
      const registration = buildEventRegistration({ event_id: "event-1", user_id: "user-1" });
      const { result } = setup({ myRegistrationsByEvent: new Map([["event-1", registration]]) });
      await new Promise((resolve) => setTimeout(resolve, 20));
      expect(result.current.registrationEditorOpen).toBe(false);
    });

    it("does not auto-open when registration is closed", async () => {
      const group = buildEventGroupDetail({ id: "group-1", game_id: "game-1", registration_open: false });
      const { result } = setup({ group });
      await new Promise((resolve) => setTimeout(resolve, 20));
      expect(result.current.registrationEditorOpen).toBe(false);
    });

    it("does not auto-open while the page or auth is still loading, or the user is unauthenticated", async () => {
      const { result: r1 } = setup({ pageLoading: true });
      const { result: r2 } = setup({ authLoading: true });
      const { result: r3 } = setup({ isAuthenticated: false, user: null });
      await new Promise((resolve) => setTimeout(resolve, 20));
      expect(r1.current.registrationEditorOpen).toBe(false);
      expect(r2.current.registrationEditorOpen).toBe(false);
      expect(r3.current.registrationEditorOpen).toBe(false);
    });

    it("does not re-auto-open after the user manually closes it (once per group visit)", async () => {
      const { result } = setup();
      await waitFor(() => expect(result.current.registrationEditorOpen).toBe(true));

      await act(async () => {
        await result.current.handleCloseRegistrationEditor();
      });
      await new Promise((resolve) => setTimeout(resolve, 20));

      expect(result.current.registrationEditorOpen).toBe(false);
    });
  });

  describe("validation flags", () => {
    it("hasUserGameErrors and canSaveRegistration reflect an incomplete game profile", async () => {
      const event1 = buildEventGroupEvent({ id: "event-1" });
      const group = buildEventGroupDetail({ id: "group-1", game_id: "game-1", events: [event1] });
      const { result } = setup({ group });
      await act(async () => {
        await result.current.handleOpenRegistrationSheet();
      });

      expect(result.current.hasUserGameErrors).toBe(true);
      expect(result.current.canSaveRegistration).toBe(false);

      act(() =>
        result.current.setUserGameDraft({ game_id: "game-1", in_game_name: "IGN", current_rank: "gold", peak_rank: "plat", show_rank: false }),
      );

      expect(result.current.hasUserGameErrors).toBe(false);
      expect(result.current.canSaveRegistration).toBe(true);
    });

    it("canDeleteAllViaSave is true only in edit mode with zero selected events, no lobbies, and not busy", async () => {
      const event1 = buildEventGroupEvent({ id: "event-1" });
      const group = buildEventGroupDetail({ id: "group-1", game_id: "game-1", events: [event1] });
      const registration = buildEventRegistration({ event_id: "event-1", user_id: "user-1" });
      const { result } = setup({ group, myRegistrationsByEvent: new Map([["event-1", registration]]) });
      await act(async () => {
        await result.current.handleOpenRegistrationSheet();
      });

      act(() => result.current.setRegistrationDraft((prev) => ({ ...prev, selected_event_ids: [] })));

      expect(result.current.canDeleteAllViaSave).toBe(true);
      expect(result.current.canSubmitRegistration).toBe(true);
    });

    it("selectedValidEventIds de-dupes and drops ids no longer present on the group", async () => {
      const event1 = buildEventGroupEvent({ id: "event-1" });
      const group = buildEventGroupDetail({ id: "group-1", game_id: "game-1", events: [event1] });
      const { result } = setup({ group });

      act(() =>
        result.current.setRegistrationDraft((prev) => ({
          ...prev,
          selected_event_ids: ["event-1", "event-1", "stale-event"],
        })),
      );

      expect(result.current.selectedValidEventIds).toEqual(["event-1"]);
    });

    it("regionMismatchWarning appears only when the user's region differs from the event's, with events selected", async () => {
      const event1 = buildEventGroupEvent({ id: "event-1" });
      const group = buildEventGroupDetail({ id: "group-1", game_id: "game-1", events: [event1], region: "EU" });
      const { result } = setup({ group, user: buildUser({ id: "user-1", region: "AMER" }) });

      expect(result.current.regionMismatchWarning).toBeNull();

      act(() => result.current.setRegistrationDraft((prev) => ({ ...prev, selected_event_ids: ["event-1"] })));

      expect(result.current.regionMismatchWarning).toContain("AMER");
      expect(result.current.regionMismatchWarning).toContain("EU");
    });
  });

  describe("handleSaveRegistration", () => {
    function groupWithOneEvent(overrides: Partial<EventGroupDetail> = {}): EventGroupDetail {
      return buildEventGroupDetail({
        id: "group-1",
        game_id: "game-1",
        events: [buildEventGroupEvent({ id: "event-1" })],
        ...overrides,
      });
    }

    // `isHost: true` suppresses the unrelated auto-open-for-guests effect (see "auto-open
    // behavior" above) so it can't race with these tests by re-opening the sheet mid-assertion.
    function setupForSave(overrides: Partial<UseRegistrationEditorOptions> = {}) {
      return setup({ isHost: true, ...overrides });
    }

    it("is a no-op when there is no group", async () => {
      const { result, options } = setup({ group: null });
      await act(async () => {
        await result.current.handleSaveRegistration();
      });
      expect(options.setWorking).not.toHaveBeenCalled();
    });

    it("triggers the delete-all confirmation instead of saving when zero events are selected in edit mode", async () => {
      const group = groupWithOneEvent();
      const registration = buildEventRegistration({ event_id: "event-1", user_id: "user-1" });
      const { result, options } = setupForSave({ group, myRegistrationsByEvent: new Map([["event-1", registration]]) });
      act(() => result.current.setRegistrationDraft((prev) => ({ ...prev, selected_event_ids: [] })));

      await act(async () => {
        await result.current.handleSaveRegistration();
      });

      expect(options.openDeleteAllForCurrentUserConfirmation).toHaveBeenCalledTimes(1);
      expect(options.setWorking).not.toHaveBeenCalled();
    });

    it("sets an error when zero events are selected and not in edit mode", async () => {
      const group = groupWithOneEvent();
      const { result } = setupForSave({ group });
      act(() => result.current.setRegistrationDraft((prev) => ({ ...prev, selected_event_ids: [] })));

      await act(async () => {
        await result.current.handleSaveRegistration();
      });

      expect(result.current.registrationError).toMatch(/select at least one event/i);
    });

    it("sets an error and does not call the API when the game profile has validation errors", async () => {
      const group = groupWithOneEvent();
      let called = false;
      server.use(http.put(`${TEST_API_URL}/users/me/games/:gameId`, () => { called = true; return new HttpResponse(null, { status: 204 }); }));
      const { result } = setupForSave({ group });
      act(() => result.current.setRegistrationDraft((prev) => ({ ...prev, selected_event_ids: ["event-1"] })));

      await act(async () => {
        await result.current.handleSaveRegistration();
      });

      expect(result.current.registrationError).toMatch(/complete your game profile/i);
      expect(called).toBe(false);
    });

    it("saves the game profile then the registrations, closes the sheet, and reloads on success", async () => {
      const group = groupWithOneEvent();
      let gameProfileBody: unknown;
      let registrationsBody: unknown;
      server.use(
        http.put(`${TEST_API_URL}/users/me/games/:gameId`, async ({ request }) => {
          gameProfileBody = await request.json();
          return new HttpResponse(null, { status: 204 });
        }),
        http.put(`${TEST_API_URL}/registrations/group/:groupId/me`, async ({ request }) => {
          registrationsBody = await request.json();
          return new HttpResponse(null, { status: 204 });
        }),
      );
      const { result, options } = setupForSave({ group });
      act(() => {
        result.current.setRegistrationDraft((prev) => ({
          ...prev,
          selected_event_ids: ["event-1"],
          duo_request: "buddy#0001",
        }));
        result.current.setUserGameDraft({ game_id: "game-1", in_game_name: " IGN ", current_rank: "gold", peak_rank: "plat", show_rank: true });
      });

      await act(async () => {
        await result.current.handleSaveRegistration();
      });

      expect(gameProfileBody).toEqual({ in_game_name: "IGN", current_rank: "gold", peak_rank: "plat", show_rank: true });
      expect(registrationsBody).toEqual({
        duo_request: "buddy#0001",
        events: [{ event_id: "event-1", can_substitute: true, can_lobby_host: true }],
      });
      expect(result.current.registrationEditorOpen).toBe(false);
      expect(options.loadGroup).toHaveBeenCalledTimes(1);
      expect(options.setWorking).toHaveBeenLastCalledWith(false);
    });

    it("stops before saving registrations if the game-profile save fails", async () => {
      const group = groupWithOneEvent();
      let registrationsCalled = false;
      server.use(
        http.put(`${TEST_API_URL}/users/me/games/:gameId`, () => HttpResponse.json({ message: "Invalid rank" }, { status: 400 })),
        http.put(`${TEST_API_URL}/registrations/group/:groupId/me`, () => {
          registrationsCalled = true;
          return new HttpResponse(null, { status: 204 });
        }),
      );
      const { result, options } = setupForSave({ group });
      act(() => {
        result.current.setRegistrationDraft((prev) => ({ ...prev, selected_event_ids: ["event-1"] }));
        result.current.setUserGameDraft({ game_id: "game-1", in_game_name: "IGN", current_rank: "gold", peak_rank: "plat", show_rank: false });
      });

      await act(async () => {
        await result.current.handleSaveRegistration();
      });

      expect(registrationsCalled).toBe(false);
      expect(result.current.registrationError).toMatch(/Invalid rank/);
      expect(options.setWorking).toHaveBeenLastCalledWith(false);
    });

    it("handles a Discord-guild restriction on the registration save by denying access and closing the sheet", async () => {
      const group = groupWithOneEvent();
      server.use(
        http.put(`${TEST_API_URL}/users/me/games/:gameId`, () => new HttpResponse(null, { status: 204 })),
        http.put(`${TEST_API_URL}/registrations/group/:groupId/me`, () =>
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
      const { result, options } = setupForSave({ group });
      act(() => {
        result.current.setRegistrationDraft((prev) => ({ ...prev, selected_event_ids: ["event-1"] }));
        result.current.setUserGameDraft({ game_id: "game-1", in_game_name: "IGN", current_rank: "gold", peak_rank: "plat", show_rank: false });
      });

      await act(async () => {
        await result.current.handleSaveRegistration();
      });

      expect(options.setAccessDenial).toHaveBeenCalledWith(expect.objectContaining({ code: "discord_guild_restricted" }));
      expect(options.setGroup).toHaveBeenCalledWith(null);
      expect(result.current.registrationEditorOpen).toBe(false);
    });

    it("keeps the sheet open and surfaces a combined error when the registration save fails for another reason", async () => {
      const group = groupWithOneEvent();
      server.use(
        http.put(`${TEST_API_URL}/users/me/games/:gameId`, () => new HttpResponse(null, { status: 204 })),
        http.put(`${TEST_API_URL}/registrations/group/:groupId/me`, () => HttpResponse.json({ message: "Event full" }, { status: 409 })),
      );
      const { result, options } = setupForSave({ group });
      await act(async () => {
        await result.current.handleOpenRegistrationSheet();
      });
      act(() => {
        result.current.setRegistrationDraft((prev) => ({ ...prev, selected_event_ids: ["event-1"] }));
        result.current.setUserGameDraft({ game_id: "game-1", in_game_name: "IGN", current_rank: "gold", peak_rank: "plat", show_rank: false });
      });

      await act(async () => {
        await result.current.handleSaveRegistration();
      });

      expect(result.current.registrationError).toMatch(/Event full/);
      // The sheet stays open (unlike the discord-restriction case) so the user can retry saving.
      expect(result.current.registrationEditorOpen).toBe(true);
      expect(options.loadGroup).not.toHaveBeenCalled();
    });
  });
});
