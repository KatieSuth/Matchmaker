import { http, HttpResponse } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { waitFor } from "@testing-library/react";
import { buildUser } from "@/test/fixtures";
import { persistPostLoginRedirect } from "@/app/_lib/postLoginRedirect";
import { server } from "@/test/msw/server";
import { TEST_API_URL } from "@/test/constants";
import { mockRouter, resetNextNavigationMock } from "@/test/nextNavigationMock";
import { renderWithProviders } from "@/test/render";
import CallbackPage from "./page";

vi.mock("next/navigation", () => import("@/test/nextNavigationMock"));

const USER = buildUser({ id: "user-1", new_user: false });

describe("Auth callback page", () => {
  beforeEach(() => {
    resetNextNavigationMock();
    sessionStorage.clear();
  });

  it("redirects home when no otc is present", async () => {
    window.history.pushState({}, "", "/auth/callback");
    renderWithProviders(<CallbackPage />, { setUser: vi.fn(), setIsAuthenticated: vi.fn() });
    await waitFor(() => expect(mockRouter.replace).toHaveBeenCalledWith("/"));
  });

  it("exchanges a valid otc and routes an existing user to My Events", async () => {
    window.history.pushState({}, "", "/auth/callback?otc=abc123&new_user=false");
    server.use(
      http.post(`${TEST_API_URL}/auth/complete`, () => HttpResponse.json({ access_token: "tok" })),
      http.get(`${TEST_API_URL}/users/me`, () => HttpResponse.json(USER)),
    );

    renderWithProviders(<CallbackPage />, { setUser: vi.fn(), setIsAuthenticated: vi.fn() });

    await waitFor(() => expect(mockRouter.replace).toHaveBeenCalledWith("/my_events"));
  });

  it("routes a new user to /my_account", async () => {
    window.history.pushState({}, "", "/auth/callback?otc=abc123&new_user=true");
    server.use(
      http.post(`${TEST_API_URL}/auth/complete`, () => HttpResponse.json({ access_token: "tok" })),
      http.get(`${TEST_API_URL}/users/me`, () => HttpResponse.json({ ...USER, new_user: true })),
    );

    renderWithProviders(<CallbackPage />, { setUser: vi.fn(), setIsAuthenticated: vi.fn() });

    await waitFor(() => expect(mockRouter.replace).toHaveBeenCalledWith("/my_account"));
  });

  it("honors a stored post-login event redirect for an existing user", async () => {
    persistPostLoginRedirect("/event/group-1");
    window.history.pushState({}, "", "/auth/callback?otc=abc123&new_user=false");
    server.use(
      http.post(`${TEST_API_URL}/auth/complete`, () => HttpResponse.json({ access_token: "tok" })),
      http.get(`${TEST_API_URL}/users/me`, () => HttpResponse.json(USER)),
    );

    renderWithProviders(<CallbackPage />, { setUser: vi.fn(), setIsAuthenticated: vi.fn() });

    await waitFor(() => expect(mockRouter.replace).toHaveBeenCalledWith("/event/group-1"));
  });

  it("redirects home when complete-auth fails", async () => {
    window.history.pushState({}, "", "/auth/callback?otc=abc123");
    server.use(http.post(`${TEST_API_URL}/auth/complete`, () => new HttpResponse(null, { status: 401 })));

    renderWithProviders(<CallbackPage />, { setUser: vi.fn(), setIsAuthenticated: vi.fn() });

    await waitFor(() => expect(mockRouter.replace).toHaveBeenCalledWith("/"));
  });
});
