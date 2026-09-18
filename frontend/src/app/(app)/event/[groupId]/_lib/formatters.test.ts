// Tests for the event group detail page's pure display-formatting helpers.
import { describe, expect, it } from "vitest";
import { EMPTY_VALUE } from "@/app/_lib/constants";
import {
  discordLockLeadSentence,
  formatDateTime,
  formatGameModeAndTime,
  formatGroupTeamSizeLabel,
  formatHostDisplayLabel,
  formatPlacementCategory,
  formatPlayerCount,
  formatSwapCandidateLabel,
  joinDiscordGuildNames,
} from "./formatters";

describe("formatDateTime", () => {
  it("returns EMPTY_VALUE for an empty string", () => {
    expect(formatDateTime("")).toBe(EMPTY_VALUE);
    expect(formatDateTime("   ")).toBe(EMPTY_VALUE);
  });

  it("returns EMPTY_VALUE for an unparsable date string", () => {
    expect(formatDateTime("not-a-date")).toBe(EMPTY_VALUE);
  });

  it("formats a valid ISO string using an explicit time zone", () => {
    expect(formatDateTime("2026-09-14T15:05:00Z", "UTC")).toBe("Sep 14, 2026, 3:05 PM");
  });
});

describe("joinDiscordGuildNames", () => {
  it("returns a generic phrase for an empty list", () => {
    expect(joinDiscordGuildNames([])).toBe("a required Discord server");
  });

  it("returns the single name for a one-item list", () => {
    expect(joinDiscordGuildNames(["Server A"])).toBe("Server A");
  });

  it("joins two names with 'or'", () => {
    expect(joinDiscordGuildNames(["Server A", "Server B"])).toBe("Server A or Server B");
  });

  it("Oxford-commas three or more names", () => {
    expect(joinDiscordGuildNames(["A", "B", "C"])).toBe("A, B, or C");
  });
});

describe("discordLockLeadSentence", () => {
  it("uses the event title when named", () => {
    expect(discordLockLeadSentence(true, "Friday Customs", "Server A")).toBe("Friday Customs is locked to Server A.");
  });

  it("falls back to the game name when unnamed", () => {
    expect(discordLockLeadSentence(false, "Valorant", "Server A")).toBe(
      "This Valorant event is locked to Server A.",
    );
  });

  it("falls back to 'game' when both title and named are missing", () => {
    expect(discordLockLeadSentence(true, "", "Server A")).toBe("This game event is locked to Server A.");
  });
});

describe("formatPlayerCount", () => {
  it("pluralizes correctly", () => {
    expect(formatPlayerCount(0)).toBe("0 Players");
    expect(formatPlayerCount(1)).toBe("1 Player");
    expect(formatPlayerCount(2)).toBe("2 Players");
  });
});

describe("formatGroupTeamSizeLabel", () => {
  it("returns '—' for no events", () => {
    expect(formatGroupTeamSizeLabel([])).toBe("—");
  });

  it("returns the shared size when all events match", () => {
    expect(
      formatGroupTeamSizeLabel([
        { team_size: 5 } as never,
        { team_size: 5 } as never,
      ]),
    ).toBe("5");
  });

  it("returns 'Varies' when events have different team sizes", () => {
    expect(
      formatGroupTeamSizeLabel([
        { team_size: 5 } as never,
        { team_size: 3 } as never,
      ]),
    ).toBe("Varies");
  });
});

describe("formatHostDisplayLabel", () => {
  it("returns 'You' when the viewer is the host", () => {
    expect(formatHostDisplayLabel(true, "Display", "discord#0001", "they/them")).toBe("You");
  });

  it("includes pronouns when set and viewer is not the host", () => {
    expect(formatHostDisplayLabel(false, "Display", "discord#0001", "they/them")).toBe(
      "Display @discord#0001 (they/them)",
    );
  });

  it("omits pronouns parenthetical when empty", () => {
    expect(formatHostDisplayLabel(false, "Display", "discord#0001", "")).toBe("Display @discord#0001");
  });
});

describe("formatGameModeAndTime", () => {
  it("joins mode name and formatted time with a middot", () => {
    expect(formatGameModeAndTime("5v5", "2026-09-14T15:05:00Z")).toBe(
      `5v5 · ${formatDateTime("2026-09-14T15:05:00Z")}`,
    );
  });
});

describe("formatPlacementCategory", () => {
  it("returns 'Unplaced' when teamNumber is undefined", () => {
    expect(formatPlacementCategory(null, 0, undefined)).toBe("Unplaced");
  });

  it("returns 'Subs' when teamNumber is null", () => {
    expect(formatPlacementCategory(null, 0, null)).toBe("Subs");
  });

  it("returns a bare 'Team N' label when the target lobby is the source lobby", () => {
    expect(formatPlacementCategory(0, 0, 1)).toBe("Team 1");
  });

  it("returns a 'Lobby N · Team M' label when the target lobby differs from the source", () => {
    expect(formatPlacementCategory(0, 1, 1)).toBe("Lobby 2 · Team 3");
  });

  it("uses the lobby-qualified label when there is no source lobby (unplaced source)", () => {
    expect(formatPlacementCategory(null, 0, 2)).toBe("Lobby 1 · Team 2");
  });
});

describe("formatSwapCandidateLabel", () => {
  it("formats name, category, and rank", () => {
    expect(formatSwapCandidateLabel("Alex", "alex#0001", "Team 1", "Gold")).toBe(
      "Alex @alex#0001 (Team 1) · Gold",
    );
  });

  it("falls back to EMPTY_VALUE when rank is missing", () => {
    expect(formatSwapCandidateLabel("Alex", "alex#0001", "Subs", undefined)).toBe(
      `Alex @alex#0001 (Subs) · ${EMPTY_VALUE}`,
    );
  });
});
