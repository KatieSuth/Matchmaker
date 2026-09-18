import { http, HttpResponse } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { waitFor } from "@testing-library/react";
import { buildMyEvent, buildUser } from "@/test/fixtures";
import { server } from "@/test/msw/server";
import { TEST_API_URL } from "@/test/constants";
import { resetNextNavigationMock } from "@/test/nextNavigationMock";
import { renderWithProviders, screen, userEvent } from "@/test/render";
import MyEventsPage from "./page";

vi.mock("next/navigation", () => import("@/test/nextNavigationMock"));

const USER = buildUser({ id: "user-1" });

describe("MyEventsPage", () => {
  beforeEach(() => {
    resetNextNavigationMock();
    server.use(
      http.get(`${TEST_API_URL}/games`, () => HttpResponse.json([])),
      http.get(`${TEST_API_URL}/users/me/events`, () =>
        HttpResponse.json({
          event_groups: [buildMyEvent({ id: "group-1", name: "Friday Customs", host_id: USER.id })],
          next_cursor: null,
          has_more: false,
        }),
      ),
    );
  });

  it("renders the dashboard heading and create CTA", async () => {
    renderWithProviders(<MyEventsPage />, { user: USER, isAuthenticated: true });
    expect(screen.getByRole("heading", { name: "My Events" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Host an event/ })).toBeInTheDocument();
  });

  it("lists events returned by the API", async () => {
    renderWithProviders(<MyEventsPage />, { user: USER, isAuthenticated: true });
    expect(await screen.findByText("Friday Customs")).toBeInTheDocument();
  });

  it("opens the create-event sheet from the host CTA", async () => {
    const user = userEvent.setup();
    renderWithProviders(<MyEventsPage />, { user: USER, isAuthenticated: true });

    await user.click(screen.getByRole("button", { name: /Host an event/ }));

    expect(screen.getByRole("heading", { name: "Host an event" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Create Event" })).toBeInTheDocument();
  });
});
