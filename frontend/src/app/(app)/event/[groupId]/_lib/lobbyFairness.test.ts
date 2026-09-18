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
  it("uses the lock-in message when flagged unfair at lock time", () => {
    const lobby = buildEventLobby({ fairness_warning_at_lock: true });
    expect(lobbyFairnessWarningMessage(lobby)).toMatch(/best available balance/);
  });

  it("uses the post-edit message when fair at lock but unfair now", () => {
    const lobby = buildEventLobby({ fairness_warning_at_lock: false });
    expect(lobbyFairnessWarningMessage(lobby)).toMatch(/manual roster change/);
  });
});
