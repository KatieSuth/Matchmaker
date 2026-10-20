import { describe, expect, it } from "vitest";
import { buildEventGroupEvent, buildEventLobby } from "@/test/fixtures";
import { eventHasUnfairLobby, lobbyFairnessWarningMessage } from "./lobbyFairness";

describe("eventHasUnfairLobby", () => {
  it("returns false when no lobby is flagged", () => {
    const event = buildEventGroupEvent({ lobbies: [buildEventLobby({ fairness_warning: false })] });
    expect(eventHasUnfairLobby(event)).toBe(false);
  });

  it("returns true when at least one lobby is flagged", () => {
    const event = buildEventGroupEvent({
      lobbies: [buildEventLobby({ fairness_warning: false }), buildEventLobby({ fairness_warning: true })],
    });
    expect(eventHasUnfairLobby(event)).toBe(true);
  });

  it("returns false when there are no lobbies yet", () => {
    const event = buildEventGroupEvent({ lobbies: [] });
    expect(eventHasUnfairLobby(event)).toBe(false);
  });
});

describe("lobbyFairnessWarningMessage", () => {
  it.each([
    [{ fairness_outlier: true }, "Teams were formed with the best available balance, but one player is ranked far above everyone else in this lobby."],
    [{ fairness_team_separation: true }, "Teams were formed with the best available balance, but the two teams' average ranks are too far apart."],
    [
      { fairness_outlier: true, fairness_team_separation: true },
      "Teams were formed with the best available balance, but one player is ranked far above everyone else in this lobby, and the two teams' average ranks are too far apart.",
    ],
  ])("names the failed check at lock-in (%o)", (reasons, expected) => {
    const lobby = buildEventLobby({ fairness_warning: true, fairness_warning_at_lock: true, ...reasons });
    expect(lobbyFairnessWarningMessage(lobby)).toBe(expected);
  });

  it("names the failed check after a manual roster change", () => {
    const lobby = buildEventLobby({ fairness_warning: true, fairness_warning_at_lock: false, fairness_team_separation: true });
    expect(lobbyFairnessWarningMessage(lobby)).toBe(
      "This lobby was fair when teams were locked in, but after a manual roster change the two teams' average ranks are too far apart.",
    );
  });

  it("uses the lock-in message when flagged unfair at lock time", () => {
    const lobby = buildEventLobby({ fairness_warning_at_lock: true });
    expect(lobbyFairnessWarningMessage(lobby)).toMatch(/best available balance/);
  });

  // Lobbies locked in before reasons were recorded have both flags false.
  it("uses the post-edit message when fair at lock but unfair now", () => {
    const lobby = buildEventLobby({ fairness_warning_at_lock: false });
    expect(lobbyFairnessWarningMessage(lobby)).toMatch(/manual roster change/);
  });
});
