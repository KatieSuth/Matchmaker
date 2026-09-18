// Tests for the "Join Lobby" sheet: edit permissions, draft validation, copy, and save flow.
import { act, renderHook } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { server } from "@/test/msw/server";
import { TEST_API_URL } from "@/test/constants";
import { EventGroupDetail } from "@/app/_types/types";
import { buildEventGroupDetail, buildEventLobby, buildUser } from "@/test/fixtures";
import { useJoinLobbySheet } from "./useJoinLobbySheet";

const host = buildUser({ id: "host-1" });
const otherUser = buildUser({ id: "user-2" });

function setup(group: EventGroupDetail = buildEventGroupDetail({ join_link_base: "https://play.example.com" })) {
  const loadGroup = vi.fn().mockResolvedValue(undefined);
  const setWorking = vi.fn();
  const setToast = vi.fn();
  const view = renderHook(() => useJoinLobbySheet(group, host, true, loadGroup, setWorking, setToast));
  return { ...view, loadGroup, setWorking, setToast, group };
}

describe("useJoinLobbySheet", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("opens with the draft pre-filled from the lobby's current join code", () => {
    const lobby = buildEventLobby({ join_code: "ABC-123" });
    const { result } = setup();

    act(() => result.current.openJoinLobbySheet(lobby, 0, 1, "2026-09-20T12:00:00Z"));

    expect(result.current.joinLobbySheetOpen).toBe(true);
    expect(result.current.joinLobbyDraft).toBe("ABC-123");
    expect(result.current.pendingJoinLobby?.lobby.id).toBe(lobby.id);
  });

  it("pre-fills the rebuilt full URL when the join code is a stored path", () => {
    const lobby = buildEventLobby({ join_code: "/LOL?code=abc" });
    const { result } = setup();

    act(() => result.current.openJoinLobbySheet(lobby, 0, 1, "2026-09-20T12:00:00Z"));

    expect(result.current.joinLobbyDraft).toBe("https://play.example.com/LOL?code=abc");
  });

  it("closeJoinLobbySheet clears all sheet state", () => {
    const lobby = buildEventLobby({ join_code: "ABC-123" });
    const { result } = setup();
    act(() => result.current.openJoinLobbySheet(lobby, 0, 1, "2026-09-20T12:00:00Z"));

    act(() => result.current.closeJoinLobbySheet());

    expect(result.current.joinLobbySheetOpen).toBe(false);
    expect(result.current.pendingJoinLobby).toBeNull();
    expect(result.current.joinLobbyDraft).toBe("");
  });

  it("handleJoinLobbyDraftChange updates the draft and clears any error", () => {
    const { result } = setup();
    act(() => result.current.handleJoinLobbyDraftChange("NEW-CODE"));
    expect(result.current.joinLobbyDraft).toBe("NEW-CODE");
  });

  it("canEditPendingJoinLobby is true for the event host", () => {
    const lobby = buildEventLobby({ host_id: "someone-else" });
    const { result } = renderHook(() =>
      useJoinLobbySheet(buildEventGroupDetail(), host, true, vi.fn(), vi.fn(), vi.fn()),
    );
    act(() => result.current.openJoinLobbySheet(lobby, 0, 1, "2026-09-20T12:00:00Z"));
    expect(result.current.canEditPendingJoinLobby).toBe(true);
  });

  it("canEditPendingJoinLobby is true for the lobby's own host even if not the event host", () => {
    const lobby = buildEventLobby({ host_id: otherUser.id });
    const { result } = renderHook(() =>
      useJoinLobbySheet(buildEventGroupDetail(), otherUser, false, vi.fn(), vi.fn(), vi.fn()),
    );
    act(() => result.current.openJoinLobbySheet(lobby, 0, 1, "2026-09-20T12:00:00Z"));
    expect(result.current.canEditPendingJoinLobby).toBe(true);
  });

  it("canEditPendingJoinLobby is false for a non-host, non-lobby-host viewer", () => {
    const lobby = buildEventLobby({ host_id: "someone-else" });
    const { result } = renderHook(() =>
      useJoinLobbySheet(buildEventGroupDetail(), otherUser, false, vi.fn(), vi.fn(), vi.fn()),
    );
    act(() => result.current.openJoinLobbySheet(lobby, 0, 1, "2026-09-20T12:00:00Z"));
    expect(result.current.canEditPendingJoinLobby).toBe(false);
  });

  it("handleSaveJoinLobby sets a validation error and does not call the API for invalid input", async () => {
    const lobby = buildEventLobby({ join_code: null });
    const { result, loadGroup } = setup();
    act(() => result.current.openJoinLobbySheet(lobby, 0, 1, "2026-09-20T12:00:00Z"));
    act(() => result.current.handleJoinLobbyDraftChange("bad code!"));

    await act(async () => {
      await result.current.handleSaveJoinLobby();
    });

    expect(result.current.joinLobbyError).toBeTruthy();
    expect(loadGroup).not.toHaveBeenCalled();
  });

  it("handleSaveJoinLobby saves, reloads, closes the sheet, and toasts on success", async () => {
    const lobby = buildEventLobby({ join_code: null });
    let receivedBody: unknown;
    server.use(
      http.patch(`${TEST_API_URL}/lobbies/${lobby.id}/join-code`, async ({ request }) => {
        receivedBody = await request.json();
        return new HttpResponse(null, { status: 204 });
      }),
    );
    const { result, loadGroup, setToast } = setup();
    act(() => result.current.openJoinLobbySheet(lobby, 0, 1, "2026-09-20T12:00:00Z"));
    act(() => result.current.handleJoinLobbyDraftChange("ABC-123"));

    await act(async () => {
      await result.current.handleSaveJoinLobby();
    });

    expect(receivedBody).toEqual({ join_code: "ABC-123" });
    expect(loadGroup).toHaveBeenCalledTimes(1);
    expect(result.current.joinLobbySheetOpen).toBe(false);
    expect(setToast).toHaveBeenCalledWith("Lobby join info saved.");
  });

  it("handleSaveJoinLobby clears the join code (sends null) when the draft is emptied", async () => {
    const lobby = buildEventLobby({ join_code: "ABC-123" });
    let receivedBody: unknown;
    server.use(
      http.patch(`${TEST_API_URL}/lobbies/${lobby.id}/join-code`, async ({ request }) => {
        receivedBody = await request.json();
        return new HttpResponse(null, { status: 204 });
      }),
    );
    const { result } = setup();
    act(() => result.current.openJoinLobbySheet(lobby, 0, 1, "2026-09-20T12:00:00Z"));
    act(() => result.current.handleJoinLobbyDraftChange(""));

    await act(async () => {
      await result.current.handleSaveJoinLobby();
    });

    expect(receivedBody).toEqual({ join_code: null });
  });

  it("handleSaveJoinLobby surfaces the API error on failure and keeps the sheet open", async () => {
    const lobby = buildEventLobby({ join_code: null });
    server.use(
      http.patch(`${TEST_API_URL}/lobbies/${lobby.id}/join-code`, () =>
        HttpResponse.json({ message: "Lobby not found" }, { status: 404 }),
      ),
    );
    const { result } = setup();
    act(() => result.current.openJoinLobbySheet(lobby, 0, 1, "2026-09-20T12:00:00Z"));
    act(() => result.current.handleJoinLobbyDraftChange("ABC-123"));

    await act(async () => {
      await result.current.handleSaveJoinLobby();
    });

    expect(result.current.joinLobbyError).toBe("Lobby not found");
    expect(result.current.joinLobbySheetOpen).toBe(true);
  });

  it("handleCopyJoinLobby copies the draft value when editable", async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.assign(navigator, { clipboard: { writeText } });
    const lobby = buildEventLobby({ join_code: "ABC-123" });
    const { result } = setup();
    act(() => result.current.openJoinLobbySheet(lobby, 0, 1, "2026-09-20T12:00:00Z"));
    act(() => result.current.handleJoinLobbyDraftChange("EDITED-CODE"));

    await act(async () => {
      await result.current.handleCopyJoinLobby();
    });

    expect(writeText).toHaveBeenCalledWith("EDITED-CODE");
  });

  it("handleCopyJoinLobby copies the read-only display value for a non-editing viewer", async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.assign(navigator, { clipboard: { writeText } });
    const lobby = buildEventLobby({ join_code: "ABC-123", host_id: "someone-else" });
    const { result } = renderHook(() =>
      useJoinLobbySheet(buildEventGroupDetail(), otherUser, false, vi.fn(), vi.fn(), vi.fn()),
    );
    act(() => result.current.openJoinLobbySheet(lobby, 0, 1, "2026-09-20T12:00:00Z"));

    await act(async () => {
      await result.current.handleCopyJoinLobby();
    });

    expect(writeText).toHaveBeenCalledWith("ABC-123");
  });
});
