import { describe, expect, it } from "vitest";
import { formatMyEventsDate } from "./dateFmt";

describe("formatMyEventsDate", () => {
  it("formats using an explicit time zone", () => {
    expect(formatMyEventsDate(new Date("2026-09-14T23:30:00Z"), "UTC")).toBe("Sep 14, 2026");
  });

  it("shifts to the local day for a non-UTC zone", () => {
    expect(formatMyEventsDate(new Date("2026-09-14T23:30:00Z"), "Asia/Tokyo")).toBe("Sep 15, 2026");
  });
});
