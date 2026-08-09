import { z } from "zod";

/** Matches Go utf8.RuneCountInString (Unicode code points). Example: "a👨b" → 3. */
export function codePointLength(s: string): number {
  let count = 0;
  for (const _ of s) {
    count += 1;
  }
  return count;
}

/** True when s contains C0/DEL, or both ASCII < and > (mirrors Go textinput). */
export function hasDisallowedFreeTextChars(s: string): boolean {
  let hasLT = false;
  let hasGT = false;
  for (const ch of s) {
    const code = ch.codePointAt(0)!;
    if (code < 0x20 || code === 0x7f) {
      return true;
    }
    if (ch === "<") hasLT = true;
    if (ch === ">") hasGT = true;
  }
  return hasLT && hasGT;
}

/** Optional free-entry text: empty allowed; mirrors backend textinput.NormalizeOptional rules. */
export function optionalFreeTextSchema(maxRunes: number) {
  return z
    .string()
    .refine((s) => !hasDisallowedFreeTextChars(s), {
      message: "Cannot include control characters, or both < and >.",
    })
    .refine((s) => codePointLength(s) <= maxRunes, {
      message: `Must be at most ${maxRunes} characters.`,
    });
}
