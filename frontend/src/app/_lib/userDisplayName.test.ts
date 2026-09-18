import { describe, expect, it } from "vitest";
import { formatUserDisplayLabel } from "./userDisplayName";

describe("formatUserDisplayLabel", () => {
  it("formats as 'Display @discord' when both are present", () => {
    expect(formatUserDisplayLabel("Alex", "alex#0001")).toBe("Alex @alex#0001");
  });

  it("formats as '@discord' when there is no display name", () => {
    expect(formatUserDisplayLabel(null, "alex#0001")).toBe("@alex#0001");
    expect(formatUserDisplayLabel("", "alex#0001")).toBe("@alex#0001");
    expect(formatUserDisplayLabel(undefined, "alex#0001")).toBe("@alex#0001");
  });

  it("falls back to the display name alone when there is no discord name", () => {
    expect(formatUserDisplayLabel("Alex", null)).toBe("Alex");
  });

  it("falls back to 'Unknown user' when neither is present", () => {
    expect(formatUserDisplayLabel(null, null)).toBe("Unknown user");
    expect(formatUserDisplayLabel("   ", "   ")).toBe("Unknown user");
  });

  it("trims whitespace from both inputs", () => {
    expect(formatUserDisplayLabel("  Alex  ", "  alex#0001  ")).toBe("Alex @alex#0001");
  });
});
