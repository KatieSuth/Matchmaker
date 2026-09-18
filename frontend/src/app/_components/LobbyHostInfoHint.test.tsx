// Tests for the "What is a lobby host?" hover/tap info popover (portaled dialog).
import { waitFor } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { axe } from "@/test/axe";
import { render, screen, userEvent } from "@/test/render";
import { LobbyHostInfoHint } from "./LobbyHostInfoHint";

describe("LobbyHostInfoHint", () => {
  it("hides the responsibilities panel until opened", () => {
    render(<LobbyHostInfoHint />);
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "What is a lobby host?" })).toHaveAttribute("aria-expanded", "false");
  });

  it("opens the panel on click (pinned) and shows the responsibilities list", async () => {
    const user = userEvent.setup();
    render(<LobbyHostInfoHint />);

    await user.click(screen.getByRole("button", { name: "What is a lobby host?" }));

    expect(screen.getByRole("dialog", { name: "Lobby host responsibilities" })).toBeInTheDocument();
    expect(screen.getByText(/The lobby host is responsible for/)).toBeInTheDocument();
  });

  it("toggles closed when clicked again", async () => {
    const user = userEvent.setup();
    render(<LobbyHostInfoHint />);
    const button = screen.getByRole("button", { name: "What is a lobby host?" });
    await user.click(button);
    expect(screen.getByRole("dialog")).toBeInTheDocument();

    await user.click(button);

    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("opens on hover and closes shortly after the pointer leaves", async () => {
    // Real timers throughout: user-event's hover/unhover pointer sequencing doesn't play well
    // with Vitest's fake timers (its internal waits can hang indefinitely), so the ~220ms hide
    // debounce is awaited for real here instead of faking/advancing the clock.
    const user = userEvent.setup();
    render(<LobbyHostInfoHint />);
    const button = screen.getByRole("button", { name: "What is a lobby host?" });

    await user.hover(button);
    expect(screen.getByRole("dialog")).toBeInTheDocument();

    await user.unhover(button);
    // Still visible immediately after unhover — hide is debounced by ~220ms.
    expect(screen.getByRole("dialog")).toBeInTheDocument();

    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
  });

  it("closes when clicking outside", async () => {
    const user = userEvent.setup();
    render(
      <div>
        <LobbyHostInfoHint />
        <button>Outside</button>
      </div>,
    );
    await user.click(screen.getByRole("button", { name: "What is a lobby host?" }));
    expect(screen.getByRole("dialog")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Outside" }));

    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("has no accessibility violations when closed", async () => {
    const { container } = render(<LobbyHostInfoHint />);
    expect(await axe(container)).toHaveNoViolations();
  });

  it("has no accessibility violations when open", async () => {
    const user = userEvent.setup();
    render(<LobbyHostInfoHint />);

    await user.click(screen.getByRole("button", { name: "What is a lobby host?" }));

    // The popover portals to document.body, so axe needs to scan the whole document.
    expect(await axe(document.body)).toHaveNoViolations();
  });
});
