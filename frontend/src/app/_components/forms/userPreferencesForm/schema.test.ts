import { describe, expect, it } from "vitest";
import { preferencesSchema, userGameSchema } from "./schema";

function validUserGame(overrides: Record<string, unknown> = {}) {
  return {
    game_id: "123e4567-e89b-12d3-a456-426614174000",
    in_game_name: "TestIGN",
    current_rank: "Gold",
    peak_rank: "Platinum",
    show_rank: true,
    api_permission: false,
    ...overrides,
  };
}

describe("userGameSchema", () => {
  it("accepts a fully valid game entry", () => {
    expect(userGameSchema.safeParse(validUserGame()).success).toBe(true);
  });

  it("rejects a non-UUID game_id", () => {
    expect(userGameSchema.safeParse(validUserGame({ game_id: "not-a-uuid" })).success).toBe(false);
  });

  it("requires a non-empty in_game_name", () => {
    expect(userGameSchema.safeParse(validUserGame({ in_game_name: "" })).success).toBe(false);
  });

  it("requires current_rank and peak_rank", () => {
    expect(userGameSchema.safeParse(validUserGame({ current_rank: "" })).success).toBe(false);
    expect(userGameSchema.safeParse(validUserGame({ peak_rank: "" })).success).toBe(false);
  });
});

describe("preferencesSchema", () => {
  function validPreferences(overrides: Record<string, unknown> = {}) {
    return {
      display_name: "",
      pronouns: "",
      show_pronouns: false,
      region: null,
      games: [],
      ...overrides,
    };
  }

  it("accepts a minimal valid payload", () => {
    expect(preferencesSchema.safeParse(validPreferences()).success).toBe(true);
  });

  it("accepts a payload with games and a region", () => {
    expect(
      preferencesSchema.safeParse(validPreferences({ region: "AMER", games: [validUserGame()] })).success,
    ).toBe(true);
  });

  it("rejects a display_name over the rune limit", () => {
    expect(
      preferencesSchema.safeParse(validPreferences({ display_name: "a".repeat(51) })).success,
    ).toBe(false);
  });

  it("rejects an invalid region", () => {
    expect(preferencesSchema.safeParse(validPreferences({ region: "MARS" })).success).toBe(false);
  });

  it("rejects an invalid nested game entry", () => {
    expect(
      preferencesSchema.safeParse(validPreferences({ games: [validUserGame({ in_game_name: "" })] })).success,
    ).toBe(false);
  });
});
