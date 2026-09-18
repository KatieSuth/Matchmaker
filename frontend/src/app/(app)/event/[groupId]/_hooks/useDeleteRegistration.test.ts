// Tests for the delete-registration confirmation flow (single game vs. all games in a series).
import { act, renderHook } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { server } from "@/test/msw/server";
import { TEST_API_URL } from "@/test/constants";
import { buildEventGroupDetail, buildEventGroupEvent, buildEventRegistration, buildUser } from "@/test/fixtures";
import { useDeleteRegistration } from "./useDeleteRegistration";

const viewer = buildUser({ id: "viewer-1" });

function setup(group = buildEventGroupDetail()) {
  const loadGroup = vi.fn().mockResolvedValue(undefined);
  const setWorking = vi.fn();
  const setPageError = vi.fn();
  const view = renderHook(() => useDeleteRegistration(group, viewer, loadGroup, setWorking, setPageError));
  return { ...view, loadGroup, setWorking, setPageError, group };
}

describe("useDeleteRegistration", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("openDeleteConfirmation counts how many events in the group the target user is registered for", () => {
    const registration = buildEventRegistration({ user_id: "target-1", display_name: "Target", discord_name: "target#0001" });
    const group = buildEventGroupDetail({
      events: [
        buildEventGroupEvent({ registrations: [registration] }),
        buildEventGroupEvent({ registrations: [buildEventRegistration({ user_id: "target-1" })] }),
        buildEventGroupEvent({ registrations: [buildEventRegistration({ user_id: "someone-else" })] }),
      ],
    });
    const { result } = setup(group);

    act(() => result.current.openDeleteConfirmation(registration, 1, "single"));

    expect(result.current.deleteWarningSheetOpen).toBe(true);
    expect(result.current.pendingDeleteAction).toMatchObject({
      mode: "single",
      userId: "target-1",
      userName: "Target @target#0001",
      registrationsInGroup: 2,
    });
  });

  it("closeDeleteWarningSheet clears the pending action", () => {
    const registration = buildEventRegistration({ user_id: "target-1" });
    const { result } = setup();
    act(() => result.current.openDeleteConfirmation(registration, 1, "single"));

    act(() => result.current.closeDeleteWarningSheet());

    expect(result.current.deleteWarningSheetOpen).toBe(false);
    expect(result.current.pendingDeleteAction).toBeNull();
  });

  it("openDeleteAllForCurrentUserConfirmation opens an 'all' action for the viewer's own registration", () => {
    const myRegistration = buildEventRegistration({ user_id: viewer.id });
    const group = buildEventGroupDetail({ events: [buildEventGroupEvent({ registrations: [myRegistration] })] });
    const { result } = setup(group);

    act(() => result.current.openDeleteAllForCurrentUserConfirmation());

    expect(result.current.pendingDeleteAction).toMatchObject({ mode: "all", userId: viewer.id });
  });

  it("openDeleteAllForCurrentUserConfirmation is a no-op when the viewer has no registration", () => {
    const group = buildEventGroupDetail({ events: [buildEventGroupEvent({ registrations: [] })] });
    const { result } = setup(group);

    act(() => result.current.openDeleteAllForCurrentUserConfirmation());

    expect(result.current.pendingDeleteAction).toBeNull();
    expect(result.current.deleteWarningSheetOpen).toBe(false);
  });

  it("deletingSelf is true only when the pending action targets the viewer", () => {
    const otherRegistration = buildEventRegistration({ user_id: "someone-else" });
    const { result } = setup();
    act(() => result.current.openDeleteConfirmation(otherRegistration, 1, "single"));
    expect(result.current.deletingSelf).toBe(false);

    const myRegistration = buildEventRegistration({ user_id: viewer.id });
    act(() => result.current.openDeleteConfirmation(myRegistration, 1, "single"));
    expect(result.current.deletingSelf).toBe(true);
  });

  it("handleDeleteRegistration deletes a single registration, closes the sheet, and reloads", async () => {
    const registration = buildEventRegistration({ event_id: "event-1", user_id: "target-1" });
    let deletedPath: string | null = null;
    server.use(
      http.delete(`${TEST_API_URL}/registrations/:eventId/:userId`, ({ params }) => {
        deletedPath = `${params.eventId}/${params.userId}`;
        return new HttpResponse(null, { status: 204 });
      }),
    );
    const { result, loadGroup } = setup();
    act(() => result.current.openDeleteConfirmation(registration, 1, "single"));

    await act(async () => {
      await result.current.handleDeleteRegistration();
    });

    expect(deletedPath).toBe("event-1/target-1");
    expect(loadGroup).toHaveBeenCalledTimes(1);
    expect(result.current.deleteWarningSheetOpen).toBe(false);
  });

  it("handleDeleteRegistration deletes across every event the user is registered for in 'all' mode", async () => {
    const registration = buildEventRegistration({ event_id: "event-1", user_id: "target-1" });
    const group = buildEventGroupDetail({
      events: [
        buildEventGroupEvent({ id: "event-1", registrations: [registration] }),
        buildEventGroupEvent({ id: "event-2", registrations: [buildEventRegistration({ user_id: "target-1" })] }),
        buildEventGroupEvent({ id: "event-3", registrations: [buildEventRegistration({ user_id: "someone-else" })] }),
      ],
    });
    const deletedPaths: string[] = [];
    server.use(
      http.delete(`${TEST_API_URL}/registrations/:eventId/:userId`, ({ params }) => {
        deletedPaths.push(`${params.eventId}/${params.userId}`);
        return new HttpResponse(null, { status: 204 });
      }),
    );
    const { result } = setup(group);
    act(() => result.current.openDeleteConfirmation(registration, 1, "all"));

    await act(async () => {
      await result.current.handleDeleteRegistration();
    });

    expect(deletedPaths.sort()).toEqual(["event-1/target-1", "event-2/target-1"]);
  });

  it("handleDeleteRegistration surfaces a page error and keeps working state consistent on failure", async () => {
    const registration = buildEventRegistration({ event_id: "event-1", user_id: "target-1" });
    server.use(
      http.delete(`${TEST_API_URL}/registrations/:eventId/:userId`, () =>
        HttpResponse.json({ message: "boom" }, { status: 500 }),
      ),
    );
    const { result, setPageError, setWorking } = setup();
    act(() => result.current.openDeleteConfirmation(registration, 1, "single"));

    await act(async () => {
      await result.current.handleDeleteRegistration();
    });

    expect(setPageError).toHaveBeenCalledWith("Could not delete registration.");
    expect(setWorking).toHaveBeenLastCalledWith(false);
    // The sheet stays open on failure so the user can see the error and retry/cancel.
    expect(result.current.deleteWarningSheetOpen).toBe(true);
  });

  it("handleDeleteRegistration is a no-op when there is no pending action", async () => {
    const { result, loadGroup } = setup();
    await act(async () => {
      await result.current.handleDeleteRegistration();
    });
    expect(loadGroup).not.toHaveBeenCalled();
  });
});
