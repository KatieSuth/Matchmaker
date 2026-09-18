// Tests for the "Games in this group" chip row (active-tab switching / show-all toggle).
import { describe, expect, it, vi } from "vitest";
import { axe } from "@/test/axe";
import { buildEventGroupEvent, buildEventLobby } from "@/test/fixtures";
import { render, screen, userEvent } from "@/test/render";
import { GameTabsStrip } from "./GameTabsStrip";

describe("GameTabsStrip", () => {
  it("renders a chip for every event, labeled by game number", () => {
    const events = [buildEventGroupEvent({ id: "e1" }), buildEventGroupEvent({ id: "e2" })];
    render(
      <GameTabsStrip
        events={events}
        activeEventId="e1"
        showAllEvents={false}
        registrationEditorOpen={false}
        onToggleShowAll={vi.fn()}
        onSelectEvent={vi.fn()}
        onScrollToEvent={vi.fn()}
      />,
    );
    expect(screen.getByText(/Game 1/)).toBeInTheDocument();
    expect(screen.getByText(/Game 2/)).toBeInTheDocument();
  });

  it("does not show the view-all toggle for a single event", () => {
    render(
      <GameTabsStrip
        events={[buildEventGroupEvent()]}
        activeEventId="e1"
        showAllEvents={false}
        registrationEditorOpen={false}
        onToggleShowAll={vi.fn()}
        onSelectEvent={vi.fn()}
        onScrollToEvent={vi.fn()}
      />,
    );
    expect(screen.queryByText("View all")).not.toBeInTheDocument();
  });

  it("selects a game tab on click when not showing all", async () => {
    const onSelectEvent = vi.fn();
    const user = userEvent.setup();
    const events = [buildEventGroupEvent({ id: "e1" }), buildEventGroupEvent({ id: "e2" })];
    render(
      <GameTabsStrip
        events={events}
        activeEventId="e1"
        showAllEvents={false}
        registrationEditorOpen={false}
        onToggleShowAll={vi.fn()}
        onSelectEvent={onSelectEvent}
        onScrollToEvent={vi.fn()}
      />,
    );

    await user.click(screen.getByText(/Game 2/));

    expect(onSelectEvent).toHaveBeenCalledWith("e2");
  });

  it("scrolls to the event instead of selecting a tab when showing all", async () => {
    const onScrollToEvent = vi.fn();
    const onSelectEvent = vi.fn();
    const user = userEvent.setup();
    const events = [buildEventGroupEvent({ id: "e1" }), buildEventGroupEvent({ id: "e2" })];
    render(
      <GameTabsStrip
        events={events}
        activeEventId="e1"
        showAllEvents={true}
        registrationEditorOpen={false}
        onToggleShowAll={vi.fn()}
        onSelectEvent={onSelectEvent}
        onScrollToEvent={onScrollToEvent}
      />,
    );

    await user.click(screen.getByText(/Game 2/));

    expect(onScrollToEvent).toHaveBeenCalledWith("e2");
    expect(onSelectEvent).not.toHaveBeenCalled();
  });

  it("toggles the view-all label and calls onToggleShowAll", async () => {
    const onToggleShowAll = vi.fn();
    const user = userEvent.setup();
    const events = [buildEventGroupEvent(), buildEventGroupEvent()];
    const { rerender } = render(
      <GameTabsStrip
        events={events}
        activeEventId="e1"
        showAllEvents={false}
        registrationEditorOpen={false}
        onToggleShowAll={onToggleShowAll}
        onSelectEvent={vi.fn()}
        onScrollToEvent={vi.fn()}
      />,
    );
    await user.click(screen.getByText("View all"));
    expect(onToggleShowAll).toHaveBeenCalledTimes(1);

    rerender(
      <GameTabsStrip
        events={events}
        activeEventId="e1"
        showAllEvents={true}
        registrationEditorOpen={false}
        onToggleShowAll={onToggleShowAll}
        onSelectEvent={vi.fn()}
        onScrollToEvent={vi.fn()}
      />,
    );
    expect(screen.getByText("Show one at a time")).toBeInTheDocument();
  });

  it("disables the view-all toggle while the registration editor is open", () => {
    const events = [buildEventGroupEvent(), buildEventGroupEvent()];
    render(
      <GameTabsStrip
        events={events}
        activeEventId="e1"
        showAllEvents={false}
        registrationEditorOpen={true}
        onToggleShowAll={vi.fn()}
        onSelectEvent={vi.fn()}
        onScrollToEvent={vi.fn()}
      />,
    );
    expect(screen.getByRole("button", { name: "View all" })).toBeDisabled();
  });

  it("shows an unfair-lobby warning icon for events with an unfair lobby", () => {
    const unfairEvent = buildEventGroupEvent({
      id: "e1",
      lobbies: [buildEventLobby({ fairness_warning: true })],
    });
    render(
      <GameTabsStrip
        events={[unfairEvent]}
        activeEventId="e1"
        showAllEvents={false}
        registrationEditorOpen={false}
        onToggleShowAll={vi.fn()}
        onSelectEvent={vi.fn()}
        onScrollToEvent={vi.fn()}
      />,
    );
    expect(screen.getByLabelText("Contains unfair lobby")).toBeInTheDocument();
  });

  it("has no accessibility violations", async () => {
    const events = [
      buildEventGroupEvent({ id: "e1" }),
      buildEventGroupEvent({
        id: "e2",
        lobbies: [buildEventLobby({ fairness_warning: true })],
      }),
    ];
    const { container } = render(
      <GameTabsStrip
        events={events}
        activeEventId="e1"
        showAllEvents={false}
        registrationEditorOpen={false}
        onToggleShowAll={vi.fn()}
        onSelectEvent={vi.fn()}
        onScrollToEvent={vi.fn()}
      />,
    );
    expect(await axe(container)).toHaveNoViolations();
  });
});
