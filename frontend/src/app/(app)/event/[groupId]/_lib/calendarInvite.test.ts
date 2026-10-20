// Tests for the .ics invitation: time window, title, reminders, RFC 5545 escaping/folding, download.
import { afterEach, describe, expect, it, vi } from "vitest";
import { buildEventGroupDetail, buildEventGroupEvent } from "@/test/fixtures";
import {
  buildCalendarInvite,
  calendarFileName,
  calendarTitle,
  downloadCalendarInvite,
  escapeIcsText,
  foldIcsLine,
  formatIcsDate,
  HOST_GAMES_HEADING,
  registeredGameWindow,
} from "./calendarInvite";

const URL_ = "https://matchmaker.example/event/group-1";
const NOW = new Date("2026-10-19T12:00:00Z");

function threeGameGroup(name = "Friday Customs") {
  return buildEventGroupDetail({
    id: "group-1",
    name,
    game_name: "Valorant",
    events: [
      buildEventGroupEvent({ id: "g1", start_time: "2026-11-06T23:00:00Z", game_mode_name: "5v5", duration_minutes: 45 }),
      buildEventGroupEvent({ id: "g2", start_time: "2026-11-06T23:45:00Z", game_mode_name: "5v5", duration_minutes: 45 }),
      buildEventGroupEvent({ id: "g3", start_time: "2026-11-07T00:30:00Z", game_mode_name: "3v3", duration_minutes: 30 }),
    ],
  });
}

/** Unfolds RFC 5545 continuation lines for easier assertions. */
function unfold(ics: string): string[] {
  return ics.replace(/\r\n /g, "").split("\r\n");
}

describe("calendarTitle", () => {
  it("uses the event name, falling back to the game name like the page header", () => {
    expect(calendarTitle(threeGameGroup())).toBe("Friday Customs");
    expect(calendarTitle(threeGameGroup("   "))).toBe("Valorant");
  });
});

describe("registeredGameWindow", () => {
  it("starts at the first registered game and ends when the last one finishes", () => {
    const window = registeredGameWindow(threeGameGroup(), new Set(["g3", "g2"]));
    expect(window?.start.toISOString()).toBe("2026-11-06T23:45:00.000Z");
    expect(window?.end.toISOString()).toBe("2026-11-07T01:00:00.000Z");
    expect(window?.games.map((g) => g.id)).toEqual(["g2", "g3"]);
  });

  it("ignores games the viewer did not register for", () => {
    const window = registeredGameWindow(threeGameGroup(), new Set(["g1"]));
    expect(window?.start.toISOString()).toBe("2026-11-06T23:00:00.000Z");
    expect(window?.end.toISOString()).toBe("2026-11-06T23:45:00.000Z");
  });

  it("is null when the viewer is registered for nothing", () => {
    expect(registeredGameWindow(threeGameGroup(), new Set())).toBeNull();
  });

  it("treats a missing duration as zero length", () => {
    const group = threeGameGroup();
    group.events[0].duration_minutes = 0;
    const window = registeredGameWindow(group, new Set(["g1"]));
    expect(window?.end.toISOString()).toBe(window?.start.toISOString());
  });
});

describe("ICS formatting helpers", () => {
  it("formats UTC date-times", () => {
    expect(formatIcsDate(new Date("2026-11-06T23:05:09.123Z"))).toBe("20261106T230509Z");
  });

  it("escapes backslashes, semicolons, commas, and newlines", () => {
    expect(escapeIcsText("a\\b;c,d\ne")).toBe("a\\\\b\;c\\,d\\ne");
  });

  it("folds long lines at 75 octets without splitting multi-byte characters", () => {
    const folded = foldIcsLine("SUMMARY:" + "x".repeat(70) + "🎮".repeat(20));
    const physical = folded.split("\r\n");
    expect(physical.length).toBeGreaterThan(1);
    for (const [i, line] of physical.entries()) {
      expect(new TextEncoder().encode(line).length).toBeLessThanOrEqual(75);
      if (i > 0) expect(line.startsWith(" ")).toBe(true);
    }
    expect(physical.map((line, i) => (i > 0 ? line.slice(1) : line)).join("")).toBe(
      "SUMMARY:" + "x".repeat(70) + "🎮".repeat(20),
    );
  });

  it("leaves short lines alone", () => {
    expect(foldIcsLine("VERSION:2.0")).toBe("VERSION:2.0");
  });
});

describe("buildCalendarInvite", () => {
  it("builds one event with the title, registered window, link, and three reminders", () => {
    const ics = buildCalendarInvite(threeGameGroup(), new Set(["g1", "g3"]), URL_, NOW, "UTC");
    expect(ics).not.toBeNull();
    expect(ics!.endsWith("\r\n")).toBe(true);
    expect(ics!.replace(/\r\n/g, "")).not.toMatch(/\n/);

    const lines = unfold(ics!);
    expect(lines).toContain("BEGIN:VCALENDAR");
    expect(lines).toContain("UID:group-1@matchmaker");
    expect(lines).toContain("DTSTAMP:20261019T120000Z");
    expect(lines).toContain("DTSTART:20261106T230000Z");
    expect(lines).toContain("DTEND:20261107T010000Z");
    expect(lines).toContain("SUMMARY:Friday Customs");
    expect(lines).toContain(`URL:${URL_}`);
    expect(lines.filter((l) => l === "BEGIN:VALARM")).toHaveLength(3);
    expect(lines.filter((l) => l.startsWith("TRIGGER:"))).toEqual([
      "TRIGGER:-PT24H",
      "TRIGGER:-PT90M",
      "TRIGGER:-PT10M",
    ]);

    const description = lines.find((l) => l.startsWith("DESCRIPTION:Games"))!;
    expect(description).toContain("Game 1 (5v5)");
    expect(description).toContain("Game 3 (3v3)");
    expect(description).not.toContain("Game 2");
    expect(description).toContain(`Event page: ${URL_}`);
  });

  it("uses the host heading when given, covering every game", () => {
    const all = new Set(["g1", "g2", "g3"]);
    const lines = unfold(buildCalendarInvite(threeGameGroup(), all, URL_, NOW, "UTC", HOST_GAMES_HEADING)!);
    expect(lines).toContain("DTSTART:20261106T230000Z");
    expect(lines).toContain("DTEND:20261107T010000Z");
    const description = lines.find((l) => l.startsWith("DESCRIPTION:Games"))!;
    expect(description.startsWith("DESCRIPTION:Games in this event:")).toBe(true);
    expect(description).toContain("Game 2 (5v5)");
  });

  it("escapes punctuation in the event name", () => {
    const ics = buildCalendarInvite(threeGameGroup("Scrims; Friday, late"), new Set(["g1"]), URL_, NOW);
    expect(unfold(ics!)).toContain("SUMMARY:Scrims\; Friday\\, late");
  });

  it("is null when the viewer is not registered", () => {
    expect(buildCalendarInvite(threeGameGroup(), new Set(), URL_, NOW)).toBeNull();
  });
});

describe("calendarFileName", () => {
  it("slugs the title", () => {
    expect(calendarFileName(threeGameGroup("Friday Customs: Round 2!"))).toBe("friday-customs-round-2.ics");
  });

  it("falls back to event.ics when the title has no letters or digits", () => {
    expect(calendarFileName(threeGameGroup("!!!"))).toBe("event.ics");
  });
});

describe("downloadCalendarInvite", () => {
  // jsdom has no object URLs; stub them per test and put the originals back.
  const originalCreate = URL.createObjectURL;
  const originalRevoke = URL.revokeObjectURL;
  afterEach(() => {
    URL.createObjectURL = originalCreate;
    URL.revokeObjectURL = originalRevoke;
    vi.restoreAllMocks();
  });

  it("downloads the invitation as an .ics file", async () => {
    let blob: Blob | undefined;
    const createObjectURL = vi.fn((b: Blob) => {
      blob = b;
      return "blob:invite";
    });
    const revokeObjectURL = vi.fn();
    URL.createObjectURL = createObjectURL;
    URL.revokeObjectURL = revokeObjectURL;
    const click = vi.spyOn(HTMLAnchorElement.prototype, "click").mockImplementation(function (this: HTMLAnchorElement) {
      expect(this.download).toBe("friday-customs.ics");
      expect(this.href).toBe("blob:invite");
    });

    downloadCalendarInvite(threeGameGroup(), new Set(["g1"]), URL_);

    expect(click).toHaveBeenCalledTimes(1);
    expect(revokeObjectURL).toHaveBeenCalledWith("blob:invite");
    expect(blob?.type).toBe("text/calendar;charset=utf-8");
    expect(await blob!.text()).toContain("SUMMARY:Friday Customs");
    expect(document.querySelector("a[download]")).toBeNull();
  });

  it("does nothing when the viewer is not registered", () => {
    const click = vi.spyOn(HTMLAnchorElement.prototype, "click");
    downloadCalendarInvite(threeGameGroup(), new Set(), URL_);
    expect(click).not.toHaveBeenCalled();
  });
});
