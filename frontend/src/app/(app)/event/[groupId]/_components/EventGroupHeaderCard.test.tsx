// Tests for the event group page's title/meta card and its host-only action buttons.
import { describe, expect, it, vi } from "vitest";
import { axe } from "@/test/axe";
import { buildEventGroupDetail, buildEventGroupEvent } from "@/test/fixtures";
import { render, screen, userEvent } from "@/test/render";
import { EventGroupHeaderCard } from "./EventGroupHeaderCard";

function baseProps(overrides: Partial<Parameters<typeof EventGroupHeaderCard>[0]> = {}) {
  return {
    group: buildEventGroupDetail(),
    isHost: false,
    hasAnyLobbies: false,
    working: false,
    firstEventStart: "2026-09-20T18:00:00Z",
    pingStatus: "idle" as const,
    onCopyDiscordPings: vi.fn(),
    shareStatus: "idle" as const,
    onShare: vi.fn(),
    onOpenEditSheet: vi.fn(),
    onLockInClick: vi.fn(),
    ...overrides,
  };
}

describe("EventGroupHeaderCard", () => {
  it("shows the group name and meta line", () => {
    const group = buildEventGroupDetail({ name: "Friday Customs", game_name: "Valorant", game_mode_name: "5v5", region: "AMER" });
    render(<EventGroupHeaderCard {...baseProps({ group })} />);
    expect(screen.getByRole("heading", { name: "Friday Customs" })).toBeInTheDocument();
    expect(screen.getByText("Valorant · 5v5 · AMER")).toBeInTheDocument();
  });

  it("falls back to the game name as the title when the group has no name", () => {
    const group = buildEventGroupDetail({ name: "", game_name: "Valorant", game_mode_name: "5v5", region: "AMER" });
    render(<EventGroupHeaderCard {...baseProps({ group })} />);
    expect(screen.getByRole("heading", { name: "Valorant" })).toBeInTheDocument();
    expect(screen.getByText("5v5 · AMER")).toBeInTheDocument();
  });

  it("shows the first event's scheduled time, or 'Not scheduled' when there isn't one", () => {
    const { rerender } = render(<EventGroupHeaderCard {...baseProps({ firstEventStart: "" })} />);
    expect(screen.getByText(/Not scheduled/)).toBeInTheDocument();

    rerender(<EventGroupHeaderCard {...baseProps({ firstEventStart: "2026-09-20T18:00:00Z" })} />);
    expect(screen.queryByText(/Not scheduled/)).not.toBeInTheDocument();
  });

  it("hides host-only actions for a non-host viewer", () => {
    render(<EventGroupHeaderCard {...baseProps({ isHost: false })} />);
    expect(screen.queryByRole("button", { name: /Lock In|Create teams|Delete teams/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Copy Discord Pings" })).not.toBeInTheDocument();
  });

  it("shows 'Lock In & Create Teams' for a host while registration is open", () => {
    const group = buildEventGroupDetail({ registration_open: true });
    render(<EventGroupHeaderCard {...baseProps({ group, isHost: true })} />);
    expect(screen.getByRole("button", { name: "Lock In & Create Teams" })).toBeInTheDocument();
  });

  it("shows 'Create teams' for a host once registration is closed with no lobbies yet", () => {
    const group = buildEventGroupDetail({ registration_open: false });
    render(<EventGroupHeaderCard {...baseProps({ group, isHost: true, hasAnyLobbies: false })} />);
    expect(screen.getByRole("button", { name: "Create teams" })).toBeInTheDocument();
  });

  it("shows 'Delete teams' for a host once registration is closed and teams exist", () => {
    const group = buildEventGroupDetail({ registration_open: false });
    render(<EventGroupHeaderCard {...baseProps({ group, isHost: true, hasAnyLobbies: true })} />);
    expect(screen.getByRole("button", { name: "Delete teams" })).toBeInTheDocument();
  });

  it("disables the lock-in/team button while working", () => {
    render(<EventGroupHeaderCard {...baseProps({ isHost: true, working: true })} />);
    expect(screen.getByRole("button", { name: "Lock In & Create Teams" })).toBeDisabled();
  });

  it("calls onLockInClick when the lock-in button is clicked", async () => {
    const onLockInClick = vi.fn();
    const user = userEvent.setup();
    render(<EventGroupHeaderCard {...baseProps({ isHost: true, onLockInClick })} />);

    await user.click(screen.getByRole("button", { name: "Lock In & Create Teams" }));

    expect(onLockInClick).toHaveBeenCalledTimes(1);
  });

  it("shows the Copy Discord Pings button only for a host with lobbies, and wires it up", async () => {
    const onCopyDiscordPings = vi.fn();
    const user = userEvent.setup();
    render(<EventGroupHeaderCard {...baseProps({ isHost: true, hasAnyLobbies: true, onCopyDiscordPings })} />);

    await user.click(screen.getByRole("button", { name: "Copy Discord Pings" }));

    expect(onCopyDiscordPings).toHaveBeenCalledTimes(1);
  });

  it("calls onShare when the share button is clicked", async () => {
    const onShare = vi.fn();
    const user = userEvent.setup();
    render(<EventGroupHeaderCard {...baseProps({ onShare })} />);

    await user.click(screen.getByRole("button", { name: "Copy share link" }));

    expect(onShare).toHaveBeenCalledTimes(1);
  });

  it("calls onOpenEditSheet when the settings button is clicked", async () => {
    const onOpenEditSheet = vi.fn();
    const user = userEvent.setup();
    render(<EventGroupHeaderCard {...baseProps({ onOpenEditSheet })} />);

    await user.click(screen.getByRole("button", { name: "Event settings" }));

    expect(onOpenEditSheet).toHaveBeenCalledTimes(1);
  });

  it("shows registration status, team size, and host summary", () => {
    const group = buildEventGroupDetail({
      registration_open: true,
      events: [buildEventGroupEvent({ team_size: 5 })],
      owner_display_name: "Host User",
      owner_name: "HostUser#0001",
    });
    render(<EventGroupHeaderCard {...baseProps({ group, isHost: false })} />);
    expect(screen.getByText("Registration Open")).toBeInTheDocument();
    expect(screen.getByText("Host User @HostUser#0001")).toBeInTheDocument();
  });

  it("shows 'You' as the host summary for the host viewer", () => {
    const group = buildEventGroupDetail({ owner_display_name: "Host User", owner_name: "HostUser#0001" });
    render(<EventGroupHeaderCard {...baseProps({ group, isHost: true })} />);
    expect(screen.getByText("You")).toBeInTheDocument();
  });

  it("has no accessibility violations for a guest viewer", async () => {
    const { container } = render(<EventGroupHeaderCard {...baseProps({ isHost: false })} />);
    expect(await axe(container)).toHaveNoViolations();
  });

  it("has no accessibility violations for a host with teams", async () => {
    const group = buildEventGroupDetail({ registration_open: false });
    const { container } = render(
      <EventGroupHeaderCard {...baseProps({ group, isHost: true, hasAnyLobbies: true })} />,
    );
    expect(await axe(container)).toHaveNoViolations();
  });
});
