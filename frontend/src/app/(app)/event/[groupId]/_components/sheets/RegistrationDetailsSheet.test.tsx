// Tests for the read-only "registration details" sheet.
import { describe, expect, it, vi } from "vitest";
import { buildEventRegistration } from "@/test/fixtures";
import { axe } from "@/test/axe";
import { render, screen } from "@/test/render";
import { RegistrationDetailsSheet } from "./RegistrationDetailsSheet";

describe("RegistrationDetailsSheet", () => {
  it("shows a fallback message when there is no registration", () => {
    render(<RegistrationDetailsSheet isOpen={true} onClose={vi.fn()} registration={null} />);
    expect(screen.getByText("No registration selected.")).toBeInTheDocument();
  });

  it("shows all registration fields", () => {
    const registration = buildEventRegistration({
      discord_name: "alice#0001",
      display_name: "Alice",
      in_game_name: "AliceIGN",
      pronouns: "she/her",
      current_rank_name: "Gold",
      avg_rank_name: "Platinum",
      peak_rank_name: "Diamond",
      can_lobby_host: true,
      can_substitute: false,
      duo_request: "Bob#0001",
    });
    render(<RegistrationDetailsSheet isOpen={true} onClose={vi.fn()} registration={registration} />);

    expect(screen.getByText("alice#0001")).toBeInTheDocument();
    expect(screen.getByText("Alice")).toBeInTheDocument();
    expect(screen.getByText("AliceIGN")).toBeInTheDocument();
    expect(screen.getByText("she/her")).toBeInTheDocument();
    expect(screen.getByText("Gold")).toBeInTheDocument();
    expect(screen.getByText("Platinum")).toBeInTheDocument();
    expect(screen.getByText("Diamond")).toBeInTheDocument();
    expect(screen.getByText("Bob#0001")).toBeInTheDocument();
    expect(screen.getAllByText("Yes")).toHaveLength(1);
    expect(screen.getAllByText("No")).toHaveLength(1);
  });

  it("falls back to an em dash for missing optional fields", () => {
    const registration = buildEventRegistration({ pronouns: "", duo_request: null });
    render(<RegistrationDetailsSheet isOpen={true} onClose={vi.fn()} registration={registration} />);
    expect(screen.getAllByText("—").length).toBeGreaterThanOrEqual(2);
  });

  it("has no accessibility violations", async () => {
    const registration = buildEventRegistration();
    render(<RegistrationDetailsSheet isOpen={true} onClose={vi.fn()} registration={registration} />);
    // The sheet portals to document.body, so axe needs to scan the whole document.
    expect(await axe(document.body)).toHaveNoViolations();
  });
});
