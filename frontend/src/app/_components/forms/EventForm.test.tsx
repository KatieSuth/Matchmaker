// Integration-style tests for the create/edit event form: validation, submit, read-only, delete.
import { http, HttpResponse } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { waitFor } from "@testing-library/react";
import { buildGame, buildGameMode, buildUser } from "@/test/fixtures";
import { server } from "@/test/msw/server";
import { TEST_API_URL } from "@/test/constants";
import { mockRouter, resetNextNavigationMock } from "@/test/nextNavigationMock";
import { renderWithProviders, screen, userEvent } from "@/test/render";
import { EventForm } from "./EventForm";

vi.mock("next/navigation", () => import("@/test/nextNavigationMock"));

const GAME = buildGame({ id: "game-1", name: "Valorant" });
const MODE = buildGameMode({ id: "mode-1", game_id: "game-1", name: "5v5" });
const USER = buildUser({ id: "user-1" });

function stubCatalog() {
  server.use(
    http.get(`${TEST_API_URL}/games/users/:ownerId`, () => HttpResponse.json([GAME])),
    http.get(`${TEST_API_URL}/games/:gameId/modes`, () => HttpResponse.json([MODE])),
  );
}

async function pickOption(user: ReturnType<typeof userEvent.setup>, placeholder: string, option: string) {
  await user.click(screen.getByText(placeholder));
  await user.click(await screen.findByText(option));
}

describe("EventForm (create)", () => {
  beforeEach(() => {
    resetNextNavigationMock();
    stubCatalog();
  });

  it("calls onCancel when Cancel is clicked", async () => {
    const onCancel = vi.fn();
    const user = userEvent.setup();
    renderWithProviders(<EventForm mode="create" onCancel={onCancel} />, {
      user: USER,
      isAuthenticated: true,
    });

    await user.click(screen.getByRole("button", { name: "Cancel" }));
    expect(onCancel).toHaveBeenCalledTimes(1);
  });

  it("shows validation errors when required fields are empty", async () => {
    const user = userEvent.setup();
    renderWithProviders(<EventForm mode="create" onCancel={vi.fn()} />, {
      user: USER,
      isAuthenticated: true,
    });

    await user.click(screen.getByRole("button", { name: "Create Event" }));

    await waitFor(() => {
      expect(screen.getByText("Game is required.")).toBeInTheDocument();
      expect(screen.getByText("Region is required.")).toBeInTheDocument();
    });
  });

  it("creates an event and navigates to the new group on success", async () => {
    server.use(http.post(`${TEST_API_URL}/events`, () => HttpResponse.json({ group_id: "group-created" })));
    const onCancel = vi.fn();
    const user = userEvent.setup();
    renderWithProviders(<EventForm mode="create" onCancel={onCancel} />, {
      user: USER,
      isAuthenticated: true,
    });

    expect(await screen.findByText("Select game")).toBeInTheDocument();
    await pickOption(user, "Select game", "Valorant");
    expect(await screen.findByText("Select game mode")).toBeInTheDocument();
    await pickOption(user, "Select game mode", "5v5");
    await pickOption(user, "Select region", "AMER");

    await user.click(screen.getByRole("button", { name: "Create Event" }));

    await waitFor(() => expect(onCancel).toHaveBeenCalled());
    expect(mockRouter.push).toHaveBeenCalledWith("/event/group-created");
  });

  it("surfaces an API error message when create fails", async () => {
    server.use(
      http.post(`${TEST_API_URL}/events`, () => HttpResponse.json({ message: "Too many events." }, { status: 400 })),
    );
    const user = userEvent.setup();
    renderWithProviders(<EventForm mode="create" onCancel={vi.fn()} />, {
      user: USER,
      isAuthenticated: true,
    });

    expect(await screen.findByText("Select game")).toBeInTheDocument();
    await pickOption(user, "Select game", "Valorant");
    expect(await screen.findByText("Select game mode")).toBeInTheDocument();
    await pickOption(user, "Select game mode", "5v5");
    await pickOption(user, "Select region", "AMER");
    await user.click(screen.getByRole("button", { name: "Create Event" }));

    expect(await screen.findByText("Too many events.")).toBeInTheDocument();
  });
});

describe("EventForm (edit)", () => {
  beforeEach(() => {
    resetNextNavigationMock();
    stubCatalog();
  });

  it("hides mutation controls in read-only mode", async () => {
    const onCancel = vi.fn();
    const user = userEvent.setup();
    renderWithProviders(
      <EventForm
        mode="edit"
        eventGroupId="group-1"
        onCancel={onCancel}
        readOnly
        initialValues={{
          name: "Friday Customs",
          game_id: GAME.id,
          region: "AMER",
          sub_min: 0,
          games_to_run: 1,
          registration_open: true,
          sort_logic: "balanced",
          discord_lock: false,
          discord_guild_ids: [],
        }}
        editSchedule={[{ id: "event-1", start_time: "2026-12-01T18:00:00Z", game_mode_id: MODE.id }]}
      />,
      { user: USER, isAuthenticated: true },
    );

    expect(screen.queryByRole("button", { name: "Save Settings" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Delete Event" })).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Close" }));
    expect(onCancel).toHaveBeenCalledTimes(1);
  });

  it("saves settings and calls onSubmitted", async () => {
    const onSubmitted = vi.fn();
    const onCancel = vi.fn();
    const user = userEvent.setup();
    renderWithProviders(
      <EventForm
        mode="edit"
        eventGroupId="group-1"
        onCancel={onCancel}
        onSubmitted={onSubmitted}
        initialValues={{
          name: "Friday Customs",
          game_id: GAME.id,
          region: "AMER",
          sub_min: 0,
          games_to_run: 1,
          registration_open: true,
          sort_logic: "balanced",
          discord_lock: false,
          discord_guild_ids: [],
        }}
        editSchedule={[{ id: "event-1", start_time: "2026-12-01T18:00:00Z", game_mode_id: MODE.id }]}
      />,
      { user: USER, isAuthenticated: true },
    );

    await user.click(screen.getByRole("button", { name: "Save Settings" }));

    await waitFor(() => expect(onSubmitted).toHaveBeenCalled());
    expect(onCancel).toHaveBeenCalled();
  });

  it("deletes the event group after confirmation and navigates to My Events", async () => {
    const onCancel = vi.fn();
    const user = userEvent.setup();
    renderWithProviders(
      <EventForm
        mode="edit"
        eventGroupId="group-1"
        onCancel={onCancel}
        initialValues={{
          name: "Friday Customs",
          game_id: GAME.id,
          region: "AMER",
          sub_min: 0,
          games_to_run: 1,
          registration_open: true,
          sort_logic: "balanced",
          discord_lock: false,
          discord_guild_ids: [],
        }}
        editSchedule={[{ id: "event-1", start_time: "2026-12-01T18:00:00Z", game_mode_id: MODE.id }]}
      />,
      { user: USER, isAuthenticated: true },
    );

    await user.click(screen.getByRole("button", { name: "Delete Event" }));
    await user.click(screen.getByRole("button", { name: "Delete Permanently" }));

    await waitFor(() => expect(onCancel).toHaveBeenCalled());
    expect(mockRouter.push).toHaveBeenCalledWith("/my_events");
  });
});
