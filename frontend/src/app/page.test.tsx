import { describe, expect, it } from "vitest";
import { render, screen } from "@/test/render";
import Page from "./page";

describe("Landing page", () => {
  it("renders the welcome copy and Discord login control", () => {
    render(<Page />);
    expect(screen.getByText(/Welcome to Matchmaker/)).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Log in with Discord" })).toBeInTheDocument();
  });
});
