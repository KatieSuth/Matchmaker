// Tests for the create/edit event form's local-datetime helpers. Time-dependent functions use
// `vi.setSystemTime` for determinism instead of depending on the machine's real clock.
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  getInitialStartTimeLocal,
  parseLocalDateTimeString,
  roundUpToQuarterHour,
  startOfToday,
  toDateTimeLocalValue,
} from "./dateTime";

describe("toDateTimeLocalValue", () => {
  it("formats a Date as YYYY-MM-DDTHH:mm in local time, zero-padded", () => {
    const date = new Date(2026, 8, 4, 9, 5); // Sep 4, 2026, 09:05 local
    expect(toDateTimeLocalValue(date)).toBe("2026-09-04T09:05");
  });
});

describe("roundUpToQuarterHour", () => {
  it("rounds up to the next quarter hour and zeroes seconds/ms", () => {
    const date = new Date(2026, 8, 4, 9, 7, 30);
    const rounded = roundUpToQuarterHour(date);
    expect(rounded.getMinutes()).toBe(15);
    expect(rounded.getSeconds()).toBe(0);
    expect(rounded.getMilliseconds()).toBe(0);
  });

  it("leaves an already-on-the-quarter-hour time unchanged (aside from seconds/ms)", () => {
    const date = new Date(2026, 8, 4, 9, 30, 45);
    const rounded = roundUpToQuarterHour(date);
    expect(rounded.getMinutes()).toBe(30);
  });

  it("rolls over into the next hour when rounding up from :50", () => {
    const date = new Date(2026, 8, 4, 9, 50);
    const rounded = roundUpToQuarterHour(date);
    expect(rounded.getHours()).toBe(10);
    expect(rounded.getMinutes()).toBe(0);
  });
});

describe("getInitialStartTimeLocal", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date(2026, 8, 4, 9, 0, 0));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("returns an empty string for edit mode", () => {
    expect(getInitialStartTimeLocal("edit")).toBe("");
  });

  it("returns 'now + 30 minutes' rounded up to the next quarter hour for create mode", () => {
    // 09:00 + 30min = 09:30, already on a quarter hour.
    expect(getInitialStartTimeLocal("create")).toBe("2026-09-04T09:30");
  });
});

describe("parseLocalDateTimeString", () => {
  it("parses a valid datetime-local string", () => {
    const parsed = parseLocalDateTimeString("2026-09-04T09:30");
    expect(parsed).toBeInstanceOf(Date);
    expect(parsed?.getFullYear()).toBe(2026);
  });

  it("returns null for an empty/whitespace string", () => {
    expect(parseLocalDateTimeString("")).toBeNull();
    expect(parseLocalDateTimeString("   ")).toBeNull();
  });

  it("returns null for an unparsable string", () => {
    expect(parseLocalDateTimeString("not-a-date")).toBeNull();
  });
});

describe("startOfToday", () => {
  it("returns today's date with the time zeroed out", () => {
    const start = startOfToday();
    expect(start.getHours()).toBe(0);
    expect(start.getMinutes()).toBe(0);
    expect(start.getSeconds()).toBe(0);
    expect(start.getMilliseconds()).toBe(0);
  });
});
