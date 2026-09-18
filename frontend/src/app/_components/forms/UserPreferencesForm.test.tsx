// Tests for the /my_account profile form: loading, hydration, save, and new-user redirect.
import { http, HttpResponse } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { waitFor } from "@testing-library/react";
import { axe } from "@/test/axe";
import { buildGame, buildUser, buildUserGame } from "@/test/fixtures";
import { server } from "@/test/msw/server";
import { TEST_API_URL } from "@/test/constants";
import { mockRouter, resetNextNavigationMock } from "@/test/nextNavigationMock";
import { renderWithProviders, screen, userEvent } from "@/test/render";
import UserPreferencesForm from "./UserPreferencesForm";

vi.mock("next/navigation", () => import("@/test/nextNavigationMock"));

const USER = buildUser({
  id: "user-1",
  discord_id: "111",
  discord_name: "Ash#0001",
  display_name: "Ash",
  new_user: false,
});
const GAME = buildGame({ id: "123e4567-e89b-12d3-a456-426614174000", name: "Valorant" });

function stubProfileLoads() {
  server.use(
    http.get(`${TEST_API_URL}/games`, () => HttpResponse.json([GAME])),
    http.get(`${TEST_API_URL}/users/me/games`, () => HttpResponse.json([])),
    http.get(`${TEST_API_URL}/games/:gameId/ranks`, () => HttpResponse.json([])),
  );
}

describe("UserPreferencesForm", () => {
  beforeEach(() => {
    resetNextNavigationMock();
    stubProfileLoads();
  });

  it("shows a loading state until games and user games have loaded", () => {
    renderWithProviders(<UserPreferencesForm />, { user: USER, isAuthenticated: true, isLoading: false });
    expect(screen.getByText("Loading…")).toBeInTheDocument();
  });

  it("hydrates the settings heading and Discord identity once data arrives", async () => {
    renderWithProviders(<UserPreferencesForm />, { user: USER, isAuthenticated: true, isLoading: false });

    expect(await screen.findByRole("heading", { name: "Settings" })).toBeInTheDocument();
    expect(screen.getByText("Ash#0001")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Save changes" })).toBeDisabled();
  });

  it("saves preference changes and shows a success message", async () => {
    const setUser = vi.fn();
    const user = userEvent.setup();
    server.use(
      http.put(`${TEST_API_URL}/users/me`, () => new HttpResponse(null, { status: 204 })),
      http.get(`${TEST_API_URL}/users/me`, () => HttpResponse.json({ ...USER, display_name: "Ash Updated" })),
    );

    renderWithProviders(<UserPreferencesForm />, {
      user: USER,
      isAuthenticated: true,
      isLoading: false,
      setUser,
    });

    expect(await screen.findByLabelText(/Display Name/)).toBeInTheDocument();
    const displayName = screen.getByLabelText(/Display Name/);
    await user.clear(displayName);
    await user.type(displayName, "Ash Updated");
    await user.click(screen.getByRole("button", { name: "Save changes" }));

    expect(await screen.findByText("Changes saved.")).toBeInTheDocument();
    expect(setUser).toHaveBeenCalled();
  });

  it("redirects a finishing new user to My Events after save", async () => {
    const newUser = { ...USER, new_user: true };
    const user = userEvent.setup();
    server.use(
      http.put(`${TEST_API_URL}/users/me`, () => new HttpResponse(null, { status: 204 })),
      http.get(`${TEST_API_URL}/users/me`, () => HttpResponse.json({ ...newUser, new_user: false })),
    );

    renderWithProviders(<UserPreferencesForm />, {
      user: newUser,
      isAuthenticated: true,
      isLoading: false,
      setUser: vi.fn(),
    });

    expect(await screen.findByRole("heading", { name: "Welcome to Matchmaker!" })).toBeInTheDocument();
    await user.type(screen.getByLabelText(/Display Name/), "Ash");
    await user.click(screen.getByRole("button", { name: "Save Profile" }));

    await waitFor(() => expect(mockRouter.push).toHaveBeenCalledWith("/my_events"));
  });

  it("shows an API error message when save fails", async () => {
    const user = userEvent.setup();
    server.use(
      http.put(`${TEST_API_URL}/users/me`, () =>
        HttpResponse.json({ message: "Display name is invalid." }, { status: 400 }),
      ),
    );

    renderWithProviders(<UserPreferencesForm />, { user: USER, isAuthenticated: true, isLoading: false });

    expect(await screen.findByLabelText(/Display Name/)).toBeInTheDocument();
    await user.type(screen.getByLabelText(/Display Name/), "X");
    await user.click(screen.getByRole("button", { name: "Save changes" }));

    expect(await screen.findByText("Display name is invalid.")).toBeInTheDocument();
  });

  it("has no accessibility violations once hydrated", async () => {
    const { container } = renderWithProviders(<UserPreferencesForm />, {
      user: USER,
      isAuthenticated: true,
      isLoading: false,
    });
    expect(await screen.findByRole("heading", { name: "Settings" })).toBeInTheDocument();
    expect(await axe(container)).toHaveNoViolations();
  });
});
