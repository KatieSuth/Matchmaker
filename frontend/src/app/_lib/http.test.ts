// Tests for the shared canceled-request classifier used by every cancel-safe fetch effect.
import { describe, expect, it } from "vitest";
import { isCanceledError } from "./http";

describe("isCanceledError", () => {
  it("returns true for an axios ERR_CANCELED error", () => {
    expect(isCanceledError({ code: "ERR_CANCELED" })).toBe(true);
  });

  it("returns true for a native AbortController CanceledError", () => {
    expect(isCanceledError({ name: "CanceledError" })).toBe(true);
  });

  it("returns false for a regular error", () => {
    expect(isCanceledError(new Error("network down"))).toBe(false);
  });

  it("returns false for null/undefined", () => {
    expect(isCanceledError(null)).toBe(false);
    expect(isCanceledError(undefined)).toBe(false);
  });

  it("returns false for a plain object with unrelated fields", () => {
    expect(isCanceledError({ code: "ERR_NETWORK" })).toBe(false);
  });
});
