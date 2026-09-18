// Tests mirroring the Go backend's textinput validation rules this module is meant to match.
import { describe, expect, it } from "vitest";
import { codePointLength, hasDisallowedFreeTextChars, optionalFreeTextSchema } from "./textInput";

describe("codePointLength", () => {
  it("counts ASCII characters 1:1", () => {
    expect(codePointLength("hello")).toBe(5);
  });

  it("counts a multi-byte emoji as a single code point", () => {
    expect(codePointLength("a👨b")).toBe(3);
  });

  it("returns 0 for an empty string", () => {
    expect(codePointLength("")).toBe(0);
  });
});

describe("hasDisallowedFreeTextChars", () => {
  it("returns false for ordinary text", () => {
    expect(hasDisallowedFreeTextChars("Looking for a duo, GM+")).toBe(false);
  });

  it("returns true for a C0 control character", () => {
    expect(hasDisallowedFreeTextChars("hello\x01world")).toBe(true);
  });

  it("returns true for a DEL character", () => {
    expect(hasDisallowedFreeTextChars("hello\x7fworld")).toBe(true);
  });

  it("returns false when only '<' is present without '>'", () => {
    expect(hasDisallowedFreeTextChars("a < b")).toBe(false);
  });

  it("returns false when only '>' is present without '<'", () => {
    expect(hasDisallowedFreeTextChars("a > b")).toBe(false);
  });

  it("returns true when both '<' and '>' are present", () => {
    expect(hasDisallowedFreeTextChars("<script>")).toBe(true);
  });
});

describe("optionalFreeTextSchema", () => {
  const schema = optionalFreeTextSchema(10);

  it("accepts an empty string", () => {
    expect(schema.safeParse("").success).toBe(true);
  });

  it("accepts text within the rune limit", () => {
    expect(schema.safeParse("short").success).toBe(true);
  });

  it("rejects text over the rune limit", () => {
    const result = schema.safeParse("this is way too long");
    expect(result.success).toBe(false);
  });

  it("rejects text with disallowed control characters", () => {
    const result = schema.safeParse("bad\x01text");
    expect(result.success).toBe(false);
  });

  it("counts multi-byte code points (not UTF-16 length) against the limit", () => {
    // 10 emoji code points but 20 UTF-16 code units — must pass since the limit is rune-based.
    const tenEmoji = "👨".repeat(10);
    expect(schema.safeParse(tenEmoji).success).toBe(true);
    expect(schema.safeParse(tenEmoji + "👨").success).toBe(false);
  });
});
