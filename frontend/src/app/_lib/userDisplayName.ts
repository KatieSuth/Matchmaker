/** Formats a user label as "Display @discord" or "@discord" when no display name. */
export function formatUserDisplayLabel(
  displayName?: string | null,
  discordName?: string | null,
): string {
  const discord = (discordName ?? "").trim();
  const display = (displayName ?? "").trim();
  if (!discord) {
    return display || "Unknown user";
  }
  if (display) {
    return `${display} @${discord}`;
  }
  return `@${discord}`;
}
