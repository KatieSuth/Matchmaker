import { describe, expect, it } from "vitest";
import { EMPTY_VALUE } from "@/app/_lib/constants";
import { buildGameRank, buildLobbyPlayer } from "@/test/fixtures";
import { nearestRankName, playerSkillOrder, teamAverageRankLabel } from "./ranks";

describe("playerSkillOrder", () => {
  it("returns the player's avg_rank_order when positive", () => {
    expect(playerSkillOrder(buildLobbyPlayer({ avg_rank_order: 5 }))).toBe(5);
  });

  it("returns null when avg_rank_order is 0 or negative (no rank data)", () => {
    expect(playerSkillOrder(buildLobbyPlayer({ avg_rank_order: 0 }))).toBeNull();
    expect(playerSkillOrder(buildLobbyPlayer({ avg_rank_order: -1 }))).toBeNull();
  });
});

describe("nearestRankName", () => {
  const ranks = [
    buildGameRank({ name: "Bronze", order: 1 }),
    buildGameRank({ name: "Silver", order: 2 }),
    buildGameRank({ name: "Gold", order: 3 }),
  ];

  it("returns EMPTY_VALUE when there are no ranks", () => {
    expect(nearestRankName([], 2)).toBe(EMPTY_VALUE);
  });

  it("returns the exact match when the skill order equals a rank's order", () => {
    expect(nearestRankName(ranks, 2)).toBe("Silver");
  });

  it("rounds to the closer neighbor for a fractional skill order", () => {
    expect(nearestRankName(ranks, 2.6)).toBe("Gold");
    expect(nearestRankName(ranks, 2.4)).toBe("Silver");
  });

  it("clamps to the nearest edge rank when out of range", () => {
    expect(nearestRankName(ranks, 100)).toBe("Gold");
    expect(nearestRankName(ranks, -100)).toBe("Bronze");
  });
});

describe("teamAverageRankLabel", () => {
  const ranks = [
    buildGameRank({ name: "Bronze", order: 1 }),
    buildGameRank({ name: "Silver", order: 2 }),
    buildGameRank({ name: "Gold", order: 3 }),
  ];

  it("returns EMPTY_VALUE for an empty roster", () => {
    expect(teamAverageRankLabel([], ranks)).toBe(EMPTY_VALUE);
  });

  it("returns EMPTY_VALUE when no player has rank data", () => {
    const players = [buildLobbyPlayer({ avg_rank_order: 0 }), buildLobbyPlayer({ avg_rank_order: 0 })];
    expect(teamAverageRankLabel(players, ranks)).toBe(EMPTY_VALUE);
  });

  it("averages skill orders across players with rank data, ignoring players without it", () => {
    const players = [
      buildLobbyPlayer({ avg_rank_order: 1 }),
      buildLobbyPlayer({ avg_rank_order: 3 }),
      buildLobbyPlayer({ avg_rank_order: 0 }), // excluded from the average
    ];
    // (1 + 3) / 2 = 2 -> Silver
    expect(teamAverageRankLabel(players, ranks)).toBe("Silver");
  });
});
