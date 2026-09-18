// Tests for the landing-page effect that stashes ?next= before Discord OAuth kicks the user out.
import { afterEach, describe, expect, it, vi } from "vitest";
import { POST_LOGIN_REDIRECT_STORAGE_KEY } from "@/app/_lib/constants";
import { render } from "@/test/render";
import { PostLoginRedirectCapture } from "./PostLoginRedirectCapture";

describe("PostLoginRedirectCapture", () => {
  afterEach(() => {
    sessionStorage.clear();
    vi.unstubAllGlobals();
  });

  it("renders nothing", () => {
    const { container } = render(<PostLoginRedirectCapture />);
    expect(container).toBeEmptyDOMElement();
  });

  it("persists an allow-listed ?next= path from the URL on mount", () => {
    vi.stubGlobal("location", { ...window.location, search: "?next=/event/abc123" });
    render(<PostLoginRedirectCapture />);
    expect(sessionStorage.getItem(POST_LOGIN_REDIRECT_STORAGE_KEY)).toBe("/event/abc123");
  });

  it("does nothing when there is no ?next= param", () => {
    vi.stubGlobal("location", { ...window.location, search: "" });
    render(<PostLoginRedirectCapture />);
    expect(sessionStorage.getItem(POST_LOGIN_REDIRECT_STORAGE_KEY)).toBeNull();
  });
});
