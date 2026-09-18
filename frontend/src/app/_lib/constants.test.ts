import { describe, expect, it } from "vitest";
import { discordAvatarUrl, DISCORD_DEFAULT_AVATAR_URL } from "./constants";

describe("discordAvatarUrl", () => {
  it("builds a CDN URL when both discordId and avatarHash are present", () => {
    expect(discordAvatarUrl("123456789", "abcdef", 80)).toBe(
      "https://cdn.discordapp.com/avatars/123456789/abcdef.webp?size=80",
    );
  });

  it("falls back to the default avatar when discordId is missing", () => {
    expect(discordAvatarUrl(null, "abcdef", 80)).toBe(DISCORD_DEFAULT_AVATAR_URL);
  });

  it("falls back to the default avatar when avatarHash is missing", () => {
    expect(discordAvatarUrl("123456789", null, 80)).toBe(DISCORD_DEFAULT_AVATAR_URL);
    expect(discordAvatarUrl("123456789", undefined, 80)).toBe(DISCORD_DEFAULT_AVATAR_URL);
  });
});
