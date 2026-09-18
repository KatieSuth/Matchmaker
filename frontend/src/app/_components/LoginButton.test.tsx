// Tests for the Discord OAuth login link (forces a full document navigation, not client routing).
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { POST_LOGIN_REDIRECT_STORAGE_KEY } from "@/app/_lib/constants";
import { TEST_API_URL } from "@/test/constants";
import { axe } from "@/test/axe";
import { render, screen, userEvent } from "@/test/render";
import DiscordLoginButton from "./LoginButton";

const loginUrl = `${TEST_API_URL}/auth/login`;

describe("DiscordLoginButton", () => {
  let assignSpy: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    assignSpy = vi.fn();
    vi.stubGlobal("location", { ...window.location, assign: assignSpy, pathname: "/", search: "" });
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    sessionStorage.clear();
  });

  it("points its href at the API's Discord login endpoint", () => {
    render(<DiscordLoginButton />);
    expect(screen.getByRole("link", { name: "Log in with Discord" })).toHaveAttribute(
      "href",
      loginUrl,
    );
  });

  it("forces a full document navigation instead of a client-side transition", async () => {
    const user = userEvent.setup();
    render(<DiscordLoginButton />);

    await user.click(screen.getByRole("link", { name: "Log in with Discord" }));

    expect(assignSpy).toHaveBeenCalledWith(loginUrl);
  });

  it("clears a stale post-login redirect when clicked from the bare landing page", async () => {
    sessionStorage.setItem(POST_LOGIN_REDIRECT_STORAGE_KEY, "/event/stale");
    const user = userEvent.setup();
    render(<DiscordLoginButton />);

    await user.click(screen.getByRole("link", { name: "Log in with Discord" }));

    expect(sessionStorage.getItem(POST_LOGIN_REDIRECT_STORAGE_KEY)).toBeNull();
  });

  it("keeps a pending post-login redirect when the landing page has a ?next param", async () => {
    vi.stubGlobal("location", { ...window.location, assign: assignSpy, pathname: "/", search: "?next=/event/abc" });
    sessionStorage.setItem(POST_LOGIN_REDIRECT_STORAGE_KEY, "/event/abc");
    const user = userEvent.setup();
    render(<DiscordLoginButton />);

    await user.click(screen.getByRole("link", { name: "Log in with Discord" }));

    expect(sessionStorage.getItem(POST_LOGIN_REDIRECT_STORAGE_KEY)).toBe("/event/abc");
  });

  it("keeps a pending post-login redirect when clicked from a non-landing page", async () => {
    vi.stubGlobal("location", { ...window.location, assign: assignSpy, pathname: "/my_account", search: "" });
    sessionStorage.setItem(POST_LOGIN_REDIRECT_STORAGE_KEY, "/event/abc");
    const user = userEvent.setup();
    render(<DiscordLoginButton />);

    await user.click(screen.getByRole("link", { name: "Log in with Discord" }));

    expect(sessionStorage.getItem(POST_LOGIN_REDIRECT_STORAGE_KEY)).toBe("/event/abc");
  });

  it("has no accessibility violations", async () => {
    const { container } = render(<DiscordLoginButton />);
    expect(await axe(container)).toHaveNoViolations();
  });
});
