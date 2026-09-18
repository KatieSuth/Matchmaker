// Tests for the real AuthProvider (session bootstrap + logout). Other suites stub
// `<AuthContext.Provider>` via `renderWithProviders` so they don't hit `/auth/refresh`.
import { http, HttpResponse } from "msw";
import { afterEach, describe, expect, it, vi } from "vitest";
import { buildUser } from "@/test/fixtures";
import { server } from "@/test/msw/server";
import { TEST_API_URL } from "@/test/constants";
import { render, screen, userEvent, waitFor } from "@/test/render";
import { getAccessToken } from "@/app/_lib/auth";
import { AuthProvider, useAuth } from "./AuthContext";

/** Tiny consumer so tests can assert AuthProvider's hydrated state without stubbing the context. */
function AuthProbe() {
  const { isAuthenticated, isLoading, user, logout } = useAuth();
  if (isLoading) {
    return <p>auth-loading</p>;
  }
  return (
    <div>
      <p>{isAuthenticated ? `signed-in:${user?.display_name ?? "unknown"}` : "signed-out"}</p>
      <button type="button" onClick={() => void logout()}>
        Sign out
      </button>
    </div>
  );
}

/** Mounts AuthProvider (not the test stub) around AuthProbe. */
function renderProvider() {
  return render(
    <AuthProvider>
      <AuthProbe />
    </AuthProvider>,
  );
}

describe("AuthProvider", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.unstubAllEnvs();
  });

  it("hydrates the viewer after a successful silent refresh", async () => {
    const me = buildUser({ display_name: "Hydrated User" });
    vi.stubEnv("NEXT_PUBLIC_FRONTEND_DOMAIN", "example.com");
    vi.stubEnv("NEXT_PUBLIC_COOKIE_AUTH_EXPIRE_LIMIT", "3600");
    server.use(
      http.post(`${TEST_API_URL}/auth/refresh`, () => HttpResponse.json({ access_token: "session-token" })),
      http.get(`${TEST_API_URL}/users/me`, () => HttpResponse.json(me)),
    );

    renderProvider();

    expect(screen.getByText("auth-loading")).toBeInTheDocument();
    expect(await screen.findByText("signed-in:Hydrated User")).toBeInTheDocument();
    expect(getAccessToken()).toBe("session-token");
  });

  it("stays signed out when the silent refresh fails", async () => {
    server.use(http.post(`${TEST_API_URL}/auth/refresh`, () => HttpResponse.json({ message: "no session" }, { status: 401 })));

    renderProvider();

    expect(await screen.findByText("signed-out")).toBeInTheDocument();
    expect(getAccessToken()).toBeNull();
  });

  it("skips bootstrap on the Discord OAuth callback route", async () => {
    const refresh = vi.fn();
    vi.stubGlobal("location", { ...window.location, pathname: "/auth/callback" });
    server.use(http.post(`${TEST_API_URL}/auth/refresh`, () => {
      refresh();
      return HttpResponse.json({ access_token: "should-not-run" });
    }));

    renderProvider();

    expect(screen.getByText("signed-out")).toBeInTheDocument();
    expect(screen.queryByText("auth-loading")).not.toBeInTheDocument();
    await waitFor(() => expect(refresh).not.toHaveBeenCalled());
    expect(getAccessToken()).toBeNull();
  });

  it("does not apply bootstrap state after unmount", async () => {
    let release!: () => void;
    const held = new Promise<void>((resolve) => {
      release = resolve;
    });
    server.use(
      http.post(`${TEST_API_URL}/auth/refresh`, async () => {
        await held;
        return HttpResponse.json({ access_token: "late-token" });
      }),
    );

    const { unmount } = renderProvider();
    expect(screen.getByText("auth-loading")).toBeInTheDocument();
    unmount();
    release();

    await waitFor(() => expect(getAccessToken()).toBe("late-token"));
  });

  it("clears the access token and signed-in flag on logout", async () => {
    const me = buildUser({ display_name: "Hydrated User" });
    server.use(
      http.post(`${TEST_API_URL}/auth/refresh`, () => HttpResponse.json({ access_token: "session-token" })),
      http.get(`${TEST_API_URL}/users/me`, () => HttpResponse.json(me)),
    );
    const user = userEvent.setup();
    renderProvider();
    expect(await screen.findByText("signed-in:Hydrated User")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Sign out" }));

    expect(await screen.findByText("signed-out")).toBeInTheDocument();
    expect(getAccessToken()).toBeNull();
  });
});

describe("useAuth", () => {
  it("throws when used outside AuthProvider", () => {
    // Catch inside the component so React's render does not treat this as an uncaught error
    // (which jsdom would print as a failing-looking stack even though the assertion passes).
    let caught: unknown;
    function Probe() {
      try {
        useAuth();
      } catch (err) {
        caught = err;
      }
      return null;
    }

    render(<Probe />);

    expect(caught).toBeInstanceOf(Error);
    expect((caught as Error).message).toMatch(/useAuth must be used within AuthProvider/);
  });
});
