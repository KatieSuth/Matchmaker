// Allowed event region values; keep in sync with how hosts pick regions in forms.
export const REGIONS = ["AMER", "EMEA", "APAC"] as const;
export const EMPTY_VALUE = "—";
export const NO_SUBSTITUTES_MESSAGE = "There are no substitutes available";
export const GITHUB_REPO_URL = "https://github.com/KatieSuth/Matchmaker";
export const DEFAULT_FEEDBACK_URL = "https://github.com/KatieSuth/Matchmaker/issues";

export const NAV_LINKS = [
  { label: "Events", href: "/my_events" },
  //{ label: "Games",  href: "/games"  },
] as const;

export const EVENT_NAME_MAX_RUNES = 50;
export const DISPLAY_NAME_MAX_RUNES = 50;

/** Backend always forms two teams per lobby; display numbers continue across lobbies within a game. */
export const TEAMS_PER_LOBBY = 2;

export const DISCORD_CDN_BASE = "https://cdn.discordapp.com";
export const DISCORD_DEFAULT_AVATAR_URL = `${DISCORD_CDN_BASE}/embed/avatars/0.png`;

export type Region = (typeof REGIONS)[number];

export function discordAvatarUrl(
  discordId: string | null | undefined,
  avatarHash: string | null | undefined,
  size: number,
): string {
  if (!discordId || !avatarHash) return DISCORD_DEFAULT_AVATAR_URL;
  return `${DISCORD_CDN_BASE}/avatars/${discordId}/${avatarHash}.webp?size=${size}`;
}
