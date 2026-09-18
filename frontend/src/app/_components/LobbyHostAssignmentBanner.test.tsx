// Tests for the "you are the lobby host" reminder banner on the event group page.
import { describe, expect, it } from "vitest";
import { axe } from "@/test/axe";
import { render, screen } from "@/test/render";
import { LobbyHostAssignmentBanner } from "./LobbyHostAssignmentBanner";

describe("LobbyHostAssignmentBanner", () => {
  it("renders nothing when there are no assignments", () => {
    const { container } = render(<LobbyHostAssignmentBanner assignments={[]} />);
    expect(container).toBeEmptyDOMElement();
  });

  it("renders a single-assignment sentence without a per-assignment list for one assignment", () => {
    render(<LobbyHostAssignmentBanner assignments={[{ gameNumber: 1, lobbyNumber: 2 }]} />);
    expect(screen.getByText("You are the lobby host for Lobby 2 in Game 1.")).toBeInTheDocument();
    // Only the shared responsibilities list should be present — no per-assignment bullet list.
    expect(screen.getAllByRole("list")).toHaveLength(1);
  });

  it("renders a bulleted list for multiple assignments, alongside the shared responsibilities list", () => {
    render(
      <LobbyHostAssignmentBanner
        assignments={[
          { gameNumber: 1, lobbyNumber: 2 },
          { gameNumber: 3, lobbyNumber: 1 },
        ]}
      />,
    );
    expect(screen.getByText("You are the lobby host for the following:")).toBeInTheDocument();
    expect(screen.getAllByRole("list")).toHaveLength(2);
    expect(screen.getByText("Lobby 2 in Game 1")).toBeInTheDocument();
    expect(screen.getByText("Lobby 1 in Game 3")).toBeInTheDocument();
  });

  it("always shows the shared responsibilities list", () => {
    render(<LobbyHostAssignmentBanner assignments={[{ gameNumber: 1, lobbyNumber: 1 }]} />);
    expect(screen.getByText(/Contact the event host if you need help/)).toBeInTheDocument();
  });

  it("has no accessibility violations", async () => {
    const { container } = render(
      <LobbyHostAssignmentBanner assignments={[{ gameNumber: 1, lobbyNumber: 2 }]} />,
    );
    expect(await axe(container)).toHaveNoViolations();
  });
});
