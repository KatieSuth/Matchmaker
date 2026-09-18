// Tests for the shared lobby-host duty reminder list (registration hint + event page banner).
import { describe, expect, it } from "vitest";
import { axe } from "@/test/axe";
import { render, screen } from "@/test/render";
import { LobbyHostResponsibilitiesList } from "./LobbyHostResponsibilities";

describe("LobbyHostResponsibilitiesList", () => {
  it("renders as a list with all six responsibility items", () => {
    render(<LobbyHostResponsibilitiesList />);
    expect(screen.getByRole("list")).toBeInTheDocument();
    expect(screen.getAllByRole("listitem")).toHaveLength(6);
  });

  it("mentions contacting the event host for help", () => {
    render(<LobbyHostResponsibilitiesList />);
    expect(screen.getByText(/Contact the event host if you need help/)).toBeInTheDocument();
  });

  it("applies a custom className to the list", () => {
    render(<LobbyHostResponsibilitiesList className="text-red-500" />);
    expect(screen.getByRole("list")).toHaveClass("text-red-500");
  });

  it("has no accessibility violations", async () => {
    const { container } = render(<LobbyHostResponsibilitiesList />);
    expect(await axe(container)).toHaveNoViolations();
  });
});
