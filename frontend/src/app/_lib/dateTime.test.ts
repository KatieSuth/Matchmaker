// Tests for shared date/time helpers, including the timeZone-override behavior added specifically
// so date formatting is deterministic/testable regardless of the host machine's local zone.
import { describe, expect, it } from "vitest";
import { formatDate, formatDateTime, getUserTimeZone } from "./dateTime";

describe("getUserTimeZone", () => {
  it("returns a non-empty IANA time zone string in a jsdom environment", () => {
    expect(getUserTimeZone().length).toBeGreaterThan(0);
  });
});

describe("formatDate", () => {
  const date = new Date("2026-09-14T23:30:00Z");

  it("formats using an explicit time zone, independent of the host machine's local zone", () => {
    expect(formatDate(date, "UTC")).toBe("Sep 14, 2026");
  });

  it("shifts to the previous/next local day depending on the requested zone", () => {
    // 23:30 UTC on Sep 14 is already Sep 15 in a positive-offset zone like Tokyo.
    expect(formatDate(date, "Asia/Tokyo")).toBe("Sep 15, 2026");
  });
});

describe("formatDateTime", () => {
  const date = new Date("2026-09-14T15:05:00Z");

  it("formats date + time using an explicit time zone", () => {
    expect(formatDateTime(date, "UTC")).toBe("Sep 14, 2026, 3:05 PM");
  });

  it("defaults to the caller's local time zone when none is passed", () => {
    // Can't assert an exact string without knowing the host zone, but it should match what an
    // explicit call with the resolved local zone produces.
    expect(formatDateTime(date)).toBe(formatDateTime(date, getUserTimeZone()));
  });
});
