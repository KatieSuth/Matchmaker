// Tests for the shared footer link cluster (About/Privacy, GitHub, feedback).
import { afterEach, describe, expect, it, vi } from "vitest";
import { GITHUB_REPO_URL, DEFAULT_FEEDBACK_URL } from "@/app/_lib/constants";
import { axe } from "@/test/axe";
import { render, screen } from "@/test/render";
import SiteFooterLinks from "./SiteFooterLinks";

describe("SiteFooterLinks", () => {
  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it("links to the About & Privacy page", () => {
    render(<SiteFooterLinks />);
    expect(screen.getByRole("link", { name: "About & Privacy" })).toHaveAttribute("href", "/about");
  });

  it("links to the GitHub source repo in a new tab", () => {
    render(<SiteFooterLinks />);
    const link = screen.getByRole("link", { name: /Source on GitHub/ });
    expect(link).toHaveAttribute("href", GITHUB_REPO_URL);
    expect(link).toHaveAttribute("target", "_blank");
    expect(link).toHaveAttribute("rel", "noopener noreferrer");
  });

  it("links to the default feedback URL when no override is configured", () => {
    render(<SiteFooterLinks />);
    expect(screen.getByRole("link", { name: "Submit Issues/Feedback" })).toHaveAttribute("href", DEFAULT_FEEDBACK_URL);
  });

  it("links to a custom feedback URL when NEXT_PUBLIC_FEEDBACK_URL is set", () => {
    vi.stubEnv("NEXT_PUBLIC_FEEDBACK_URL", "https://forms.example.com/feedback");
    render(<SiteFooterLinks />);
    expect(screen.getByRole("link", { name: "Submit Issues/Feedback" })).toHaveAttribute(
      "href",
      "https://forms.example.com/feedback",
    );
  });

  it("has no accessibility violations", async () => {
    const { container } = render(<SiteFooterLinks />);
    expect(await axe(container)).toHaveNoViolations();
  });
});
