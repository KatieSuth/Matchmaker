// Tests for the registration card's display fields and its context-sensitive ellipsis menu.
import { describe, expect, it, vi } from "vitest";
import { buildEventRegistration } from "@/test/fixtures";
import { axe } from "@/test/axe";
import { render, screen, userEvent } from "@/test/render";
import { PlayerPlacement } from "../_types";
import { PlayerCard } from "./PlayerCard";

const teamPlacement: PlayerPlacement = {
  eventId: "event-1",
  userId: "user-1",
  discordName: "TestUser#0001",
  lobbyId: "lobby-1",
  sourceLobbyIndex: 0,
  teamNumber: 1,
};
const subPlacement: PlayerPlacement = { ...teamPlacement, teamNumber: null };
const unplacedPlacement: PlayerPlacement = { ...teamPlacement, lobbyId: null, teamNumber: undefined };

function baseProps(overrides: Partial<Parameters<typeof PlayerCard>[0]> = {}) {
  return {
    registration: buildEventRegistration({ user_id: "user-1" }),
    gameNumber: 1,
    eventRegion: "AMER",
    isHostView: false,
    canEditRegistration: false,
    onShowDetails: vi.fn(),
    onDeleteRegistrationForGame: vi.fn(),
    onDeleteAllFromUser: vi.fn(),
    ...overrides,
  };
}

describe("PlayerCard", () => {
  it("shows the player's display name, pronouns, in-game name, rank, and substitute flag", () => {
    const registration = buildEventRegistration({
      display_name: "Ash",
      discord_name: "ash#0001",
      pronouns: "he/him",
      in_game_name: "AshK",
      avg_rank_name: "Gold",
      can_substitute: true,
    });
    render(<PlayerCard {...baseProps({ registration })} />);

    expect(screen.getByText("Ash @ash#0001")).toBeInTheDocument();
    expect(screen.getByText("he/him")).toBeInTheDocument();
    expect(screen.getByText("AshK")).toBeInTheDocument();
    expect(screen.getByText("Gold")).toBeInTheDocument();
    expect(screen.getByText("Yes")).toBeInTheDocument();
  });

  it("shows an em dash placeholder for missing optional fields", () => {
    const registration = buildEventRegistration({ pronouns: "", in_game_name: "", avg_rank_name: undefined });
    render(<PlayerCard {...baseProps({ registration })} />);
    expect(screen.getAllByText("—").length).toBeGreaterThanOrEqual(2);
  });

  it("shows the duo request field only when showDuoRequest is true", () => {
    const registration = buildEventRegistration({ duo_request: "buddy#0001" });
    const { rerender } = render(<PlayerCard {...baseProps({ registration, showDuoRequest: false })} />);
    expect(screen.queryByText("Duo request")).not.toBeInTheDocument();

    rerender(<PlayerCard {...baseProps({ registration, showDuoRequest: true })} />);
    expect(screen.getByText("buddy#0001")).toBeInTheDocument();
  });

  it("flags a region mismatch banner only when editable and the viewer's region differs", () => {
    const { rerender } = render(
      <PlayerCard {...baseProps({ canEditRegistration: true, currentUserRegion: "EU", eventRegion: "AMER" })} />,
    );
    expect(screen.getByText(/Region: EU/)).toBeInTheDocument();

    rerender(<PlayerCard {...baseProps({ canEditRegistration: false, currentUserRegion: "EU", eventRegion: "AMER" })} />);
    expect(screen.queryByText(/Region:/)).not.toBeInTheDocument();
  });

  it("hides the menu entirely for a read-only, non-host, non-owning viewer", () => {
    render(<PlayerCard {...baseProps({ isHostView: false, canEditRegistration: false })} />);
    expect(screen.queryByRole("button", { name: "Registration actions" })).not.toBeInTheDocument();
  });

  it("shows only 'Show More Details' for a non-host viewing another player's read-only card, without delete", async () => {
    const user = userEvent.setup();
    render(
      <PlayerCard
        {...baseProps({ isHostView: false, canEditRegistration: true, registration: buildEventRegistration({ user_id: "someone-else" }) })}
      />,
    );

    await user.click(screen.getByRole("button", { name: "Registration actions" }));

    expect(screen.getByText("Show More Details")).toBeInTheDocument();
    expect(screen.queryByText(/Delete/)).not.toBeInTheDocument();
  });

  it("lets the registration's own owner delete it even without host view", async () => {
    const user = userEvent.setup();
    render(<PlayerCard {...baseProps({ isHostView: false, canEditRegistration: true, currentUserId: "user-1" })} />);

    await user.click(screen.getByRole("button", { name: "Registration actions" }));

    expect(screen.getByText("Delete for Game 1")).toBeInTheDocument();
    expect(screen.getByText("Delete All")).toBeInTheDocument();
  });

  it("suppresses delete options when allowRegistrationDelete is false", async () => {
    const user = userEvent.setup();
    render(<PlayerCard {...baseProps({ isHostView: false, canEditRegistration: true, currentUserId: "user-1", allowRegistrationDelete: false })} />);

    await user.click(screen.getByRole("button", { name: "Registration actions" }));

    expect(screen.queryByText(/Delete/)).not.toBeInTheDocument();
  });

  it("calls onDeleteRegistrationForGame / onDeleteAllFromUser for the right menu entries", async () => {
    const onDeleteRegistrationForGame = vi.fn();
    const onDeleteAllFromUser = vi.fn();
    const registration = buildEventRegistration({ user_id: "user-1" });
    const user = userEvent.setup();
    render(
      <PlayerCard
        {...baseProps({
          isHostView: false,
          canEditRegistration: true,
          currentUserId: "user-1",
          registration,
          onDeleteRegistrationForGame,
          onDeleteAllFromUser,
        })}
      />,
    );

    await user.click(screen.getByRole("button", { name: "Registration actions" }));
    await user.click(screen.getByText("Delete for Game 1"));
    expect(onDeleteRegistrationForGame).toHaveBeenCalledWith(registration, 1);

    await user.click(screen.getByRole("button", { name: "Registration actions" }));
    await user.click(screen.getByText("Delete All"));
    expect(onDeleteAllFromUser).toHaveBeenCalledWith(registration, 1);
  });

  describe("host-only roster actions", () => {
    it("offers Swap and Make Lobby Host for a team-assigned placement (non-lobby-host)", async () => {
      const onSwap = vi.fn();
      const onMakeLobbyHost = vi.fn();
      const user = userEvent.setup();
      render(
        <PlayerCard
          {...baseProps({
            isHostView: true,
            placement: teamPlacement,
            onSwap,
            onMakeLobbyHost,
            lobbyHostId: "someone-else",
          })}
        />,
      );

      await user.click(screen.getByRole("button", { name: "Registration actions" }));

      expect(screen.getByText("Swap")).toBeInTheDocument();
      expect(screen.getByText("Make Lobby Host")).toBeInTheDocument();
      await user.click(screen.getByText("Swap"));
      expect(onSwap).toHaveBeenCalledWith(teamPlacement);
    });

    it("hides Make Lobby Host when the placement is already the lobby host", async () => {
      const user = userEvent.setup();
      render(
        <PlayerCard
          {...baseProps({ isHostView: true, placement: teamPlacement, onMakeLobbyHost: vi.fn(), lobbyHostId: teamPlacement.userId })}
        />,
      );
      await user.click(screen.getByRole("button", { name: "Registration actions" }));
      expect(screen.queryByText("Make Lobby Host")).not.toBeInTheDocument();
    });

    it("offers Move to Unplaced for a sub placement", async () => {
      const onMoveToUnplaced = vi.fn();
      const user = userEvent.setup();
      render(<PlayerCard {...baseProps({ isHostView: true, placement: subPlacement, onMoveToUnplaced })} />);

      await user.click(screen.getByRole("button", { name: "Registration actions" }));
      await user.click(screen.getByText("Move to Unplaced"));

      expect(onMoveToUnplaced).toHaveBeenCalledWith(subPlacement);
    });

    it("offers Move to Substitutes for an unplaced, substitute-eligible placement", async () => {
      const onMoveToSubs = vi.fn();
      const registration = buildEventRegistration({ can_substitute: true });
      const user = userEvent.setup();
      render(<PlayerCard {...baseProps({ isHostView: true, registration, placement: unplacedPlacement, onMoveToSubs })} />);

      await user.click(screen.getByRole("button", { name: "Registration actions" }));
      await user.click(screen.getByText("Move to Substitutes"));

      expect(onMoveToSubs).toHaveBeenCalledWith(unplacedPlacement);
    });

    it("hides Move to Substitutes when the unplaced player opted out of subbing", async () => {
      const registration = buildEventRegistration({ can_substitute: false });
      const user = userEvent.setup();
      render(<PlayerCard {...baseProps({ isHostView: true, registration, placement: unplacedPlacement, onMoveToSubs: vi.fn() })} />);

      await user.click(screen.getByRole("button", { name: "Registration actions" }));

      expect(screen.queryByText("Move to Substitutes")).not.toBeInTheDocument();
    });
  });

  it("has no accessibility violations", async () => {
    const { container } = render(<PlayerCard {...baseProps({ isHostView: true, placement: teamPlacement, onSwap: vi.fn() })} />);
    expect(await axe(container)).toHaveNoViolations();
  });
});
