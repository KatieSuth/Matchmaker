// Tests for the pre-lock-in flat registration list.
import { describe, expect, it, vi } from "vitest";
import { axe } from "@/test/axe";
import { buildEventGroupEvent, buildEventRegistration } from "@/test/fixtures";
import { render, screen, userEvent } from "@/test/render";
import { EventPanel } from "./EventPanel";

function baseProps(overrides: Partial<Parameters<typeof EventPanel>[0]> = {}) {
  return {
    event: buildEventGroupEvent(),
    gameNumber: 1,
    eventRegion: "AMER",
    isHostView: false,
    onShowDetails: vi.fn(),
    onDeleteRegistrationForGame: vi.fn(),
    onDeleteAllFromUser: vi.fn(),
    ...overrides,
  };
}

describe("EventPanel", () => {
  it("shows an empty-state message when there are no registrations", () => {
    const event = buildEventGroupEvent({ registrations: [] });
    render(<EventPanel {...baseProps({ event })} />);
    expect(screen.getByText("No registered players yet.")).toBeInTheDocument();
  });

  it("renders a PlayerCard for every registration", () => {
    const event = buildEventGroupEvent({
      registrations: [
        buildEventRegistration({ user_id: "u1", display_name: "Alice" }),
        buildEventRegistration({ user_id: "u2", display_name: "Bob" }),
      ],
    });
    render(<EventPanel {...baseProps({ event })} />);
    expect(screen.getByText(/Alice/)).toBeInTheDocument();
    expect(screen.getByText(/Bob/)).toBeInTheDocument();
  });

  it("allows registration deletion only when no lobbies exist yet", async () => {
    const registration = buildEventRegistration({ user_id: "current-user" });
    const event = buildEventGroupEvent({ registrations: [registration], lobbies_count: 1 });
    const { rerender } = render(
      <EventPanel {...baseProps({ event, currentUserId: "current-user" })} />,
    );
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "Registration actions" }));
    expect(screen.queryByText(/Delete/)).not.toBeInTheDocument();

    // The menu's open/closed state lives in the still-mounted EllipsisMenu instance, so it stays
    // open across this rerender — no need to click again — and should now show the delete options.
    rerender(
      <EventPanel {...baseProps({ event: { ...event, lobbies_count: 0 }, currentUserId: "current-user" })} />,
    );
    expect(screen.getAllByText(/Delete/).length).toBe(2);
  });

  it("has no accessibility violations when empty", async () => {
    const event = buildEventGroupEvent({ registrations: [] });
    const { container } = render(<EventPanel {...baseProps({ event })} />);
    expect(await axe(container)).toHaveNoViolations();
  });

  it("has no accessibility violations for a guest with registrations", async () => {
    const event = buildEventGroupEvent({
      registrations: [buildEventRegistration({ display_name: "Alice" })],
    });
    const { container } = render(<EventPanel {...baseProps({ event, isHostView: false })} />);
    expect(await axe(container)).toHaveNoViolations();
  });

  it("has no accessibility violations for a host with registrations", async () => {
    const event = buildEventGroupEvent({
      registrations: [buildEventRegistration({ display_name: "Alice" })],
    });
    const { container } = render(<EventPanel {...baseProps({ event, isHostView: true })} />);
    expect(await axe(container)).toHaveNoViolations();
  });
});
