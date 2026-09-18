// Tests for the create/edit event form's Zod schema and edit-mode schedule validation.
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { buildEventFormSchema, validateEditScheduleDraft } from "./schema";

function validCreatePayload(overrides: Record<string, unknown> = {}) {
  return {
    name: "",
    game_id: "game-1",
    game_mode_id: "mode-1",
    region: "AMER",
    start_time_local: "2026-09-20T12:00",
    sub_min: 1,
    games_to_run: 1,
    registration_open: true,
    sort_logic: "balanced",
    discord_lock: false,
    discord_guild_ids: [],
    ...overrides,
  };
}

describe("buildEventFormSchema (create mode)", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-09-01T00:00:00Z"));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  const schema = buildEventFormSchema("create");

  it("accepts a fully valid payload", () => {
    expect(schema.safeParse(validCreatePayload()).success).toBe(true);
  });

  it("requires a start time", () => {
    const result = schema.safeParse(validCreatePayload({ start_time_local: "" }));
    expect(result.success).toBe(false);
  });

  it("rejects a start time in the past", () => {
    const result = schema.safeParse(validCreatePayload({ start_time_local: "2020-01-01T00:00" }));
    expect(result.success).toBe(false);
  });

  it("requires game_mode_id in create mode", () => {
    const result = schema.safeParse(validCreatePayload({ game_mode_id: "" }));
    expect(result.success).toBe(false);
  });

  it("rejects an invalid region", () => {
    const result = schema.safeParse(validCreatePayload({ region: "MARS" }));
    expect(result.success).toBe(false);
  });

  it("rejects a negative sub_min", () => {
    const result = schema.safeParse(validCreatePayload({ sub_min: -1 }));
    expect(result.success).toBe(false);
  });

  it("rejects games_to_run less than 1", () => {
    const result = schema.safeParse(validCreatePayload({ games_to_run: 0 }));
    expect(result.success).toBe(false);
  });

  it("requires at least one Discord server when discord_lock is enabled", () => {
    const result = schema.safeParse(validCreatePayload({ discord_lock: true, discord_guild_ids: [] }));
    expect(result.success).toBe(false);
  });

  it("accepts discord_lock with at least one server selected", () => {
    const result = schema.safeParse(
      validCreatePayload({ discord_lock: true, discord_guild_ids: ["guild-1"] }),
    );
    expect(result.success).toBe(true);
  });
});

describe("buildEventFormSchema (edit mode)", () => {
  const schema = buildEventFormSchema("edit");

  it("does not require start_time_local or game_mode_id", () => {
    const result = schema.safeParse(
      validCreatePayload({ start_time_local: undefined, game_mode_id: undefined }),
    );
    expect(result.success).toBe(true);
  });
});

describe("validateEditScheduleDraft", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-09-01T00:00:00Z"));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("returns null for a valid schedule", () => {
    const result = validateEditScheduleDraft([
      { eventId: "e1", startLocal: "2026-09-20T12:00", modeId: "mode-1" },
    ]);
    expect(result).toBeNull();
  });

  it("requires a game mode for every row", () => {
    const result = validateEditScheduleDraft([
      { eventId: "e1", startLocal: "2026-09-20T12:00", modeId: "" },
    ]);
    expect(result).toBe("Game mode is required for each scheduled game.");
  });

  it("rejects an invalid start time", () => {
    const result = validateEditScheduleDraft([
      { eventId: "e1", startLocal: "not-a-date", modeId: "mode-1" },
    ]);
    expect(result).toBe("One or more game times are invalid.");
  });

  it("rejects a start time in the past", () => {
    const result = validateEditScheduleDraft([
      { eventId: "e1", startLocal: "2020-01-01T00:00", modeId: "mode-1" },
    ]);
    expect(result).toBe("Start times cannot be in the past.");
  });
});
