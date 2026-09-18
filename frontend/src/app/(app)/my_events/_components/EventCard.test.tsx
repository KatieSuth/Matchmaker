// Tests for the single event row in the My Events list.
import { describe, expect, it } from "vitest";
import { buildMyEvent } from "@/test/fixtures";
import { axe } from "@/test/axe";
import { render, screen } from "@/test/render";
import { EventCard } from "./EventCard";

describe("EventCard", () => {
  it("shows the event name, game/mode/region, and registration status", () => {
    const event = buildMyEvent({ name: "Friday Customs", game_name: "Valorant", game_mode: "5v5", region: "AMER", registration_open: true });
    render(<EventCard event={event} isHostingList={true} hostingIds={new Set()} />);

    expect(screen.getByText("Friday Customs")).toBeInTheDocument();
    expect(screen.getByText("Valorant · 5v5 · AMER")).toBeInTheDocument();
    expect(screen.getByText("Open")).toBeInTheDocument();
  });

  it("falls back to the game name as the title when the event has no name", () => {
    const event = buildMyEvent({ name: "", game_name: "Valorant", game_mode: "5v5", region: "AMER" });
    render(<EventCard event={event} isHostingList={true} hostingIds={new Set()} />);

    // Title falls back to the game name; the subtitle then omits the (now-redundant) game name.
    expect(screen.getAllByText("Valorant")).toHaveLength(1);
    expect(screen.getByText("5v5 · AMER")).toBeInTheDocument();
  });

  it("shows 'You' as the host when the current user is the host", () => {
    const event = buildMyEvent({ host_id: "user-1", host_name: "Host#0001", host_display_name: "Host User" });
    render(<EventCard event={event} currentUserId="user-1" isHostingList={true} hostingIds={new Set()} />);
    expect(screen.getByText("You")).toBeInTheDocument();
  });

  it("shows the host's display label when viewing someone else's event", () => {
    const event = buildMyEvent({ host_id: "host-1", host_name: "Host#0001", host_display_name: "Host User" });
    render(<EventCard event={event} currentUserId="viewer-1" isHostingList={false} hostingIds={new Set()} />);
    expect(screen.getByText("Host User @Host#0001")).toBeInTheDocument();
  });

  it("shows a 'Hosting' pill on the registered tab for events the viewer also hosts", () => {
    const event = buildMyEvent({ id: "event-1" });
    render(<EventCard event={event} isHostingList={false} hostingIds={new Set(["event-1"])} />);
    expect(screen.getByText("Hosting")).toBeInTheDocument();
  });

  it("does not show a 'Hosting' pill on the hosting tab itself", () => {
    const event = buildMyEvent({ id: "event-1" });
    render(<EventCard event={event} isHostingList={true} hostingIds={new Set(["event-1"])} />);
    expect(screen.queryByText("Hosting")).not.toBeInTheDocument();
  });

  it("shows the registered player count", () => {
    const event = buildMyEvent({ registered_count: 7 });
    render(<EventCard event={event} isHostingList={true} hostingIds={new Set()} />);
    expect(screen.getByText("7")).toBeInTheDocument();
  });

  it("links to the event group detail page", () => {
    const event = buildMyEvent({ id: "group-42" });
    render(<EventCard event={event} isHostingList={true} hostingIds={new Set()} />);
    expect(screen.getByRole("link")).toHaveAttribute("href", "/event/group-42");
  });

  it("has no accessibility violations", async () => {
    const event = buildMyEvent();
    const { container } = render(<EventCard event={event} isHostingList={true} hostingIds={new Set()} />);
    expect(await axe(container)).toHaveNoViolations();
  });
});
