// Zod schema + draft types for the create/edit event form.
import { z } from "zod";
import { EVENT_NAME_MAX_RUNES, REGIONS } from "@/app/_lib/constants";
import { optionalFreeTextSchema } from "@/app/_lib/textInput";

export type EventFormEditScheduleRow = {
  id: string;
  start_time: string;
  game_mode_id: string;
};

/** Zod schema for create vs edit: edit omits required start time / mode; Discord lock needs at least one server. */
export function buildEventFormSchema(mode: "create" | "edit") {
  const startTimeField =
    mode === "create"
      ? z
          .string()
          .min(1, "Start time is required.")
          .refine((s) => !Number.isNaN(new Date(s).getTime()), {
            message: "Start time is invalid.",
          })
          .refine((s) => {
            const t = new Date(s).getTime();
            if (Number.isNaN(t)) return true;
            return t >= Date.now();
          }, {
            message: "Start time cannot be in the past.",
          })
      : z.string().optional();

  return z.object({
    name: optionalFreeTextSchema(EVENT_NAME_MAX_RUNES),
    game_id: z.string().min(1, "Game is required."),
    game_mode_id:
      mode === "create"
        ? z.string().min(1, "Game mode is required.")
        : z.string().optional(),
    region: z
      .string()
      .min(1, "Region is required.")
      .refine((s) => (REGIONS as readonly string[]).includes(s), {
        message: "Please select a valid region.",
      }),
    start_time_local: startTimeField,
    sub_min: z
      .number()
      .int()
      .min(0, "Minimum subs per lobby cannot be below 0."),
    games_to_run: z
      .number()
      .int()
      .min(1, "Number of games must be greater than 0."),
    registration_open: z.boolean(),
    sort_logic: z.enum(["balanced", "ranked"]),
    discord_lock: z.boolean(),
    discord_guild_ids: z.array(z.string()),
  }).refine((data) => !data.discord_lock || data.discord_guild_ids.length > 0, {
    message: "Select at least one Discord server.",
    path: ["discord_guild_ids"],
  });
}

export type EventFormValues = z.infer<ReturnType<typeof buildEventFormSchema>>;

export type PerGameDraftRow = {
  eventId: string;
  startLocal: string;
  modeId: string;
};

export function validateEditScheduleDraft(rows: PerGameDraftRow[]): string | null {
  const nowMs = Date.now();
  for (const row of rows) {
    if (!row.modeId) return "Game mode is required for each scheduled game.";
    const t = new Date(row.startLocal).getTime();
    if (Number.isNaN(t)) return "One or more game times are invalid.";
    if (t < nowMs) return "Start times cannot be in the past.";
  }
  return null;
}
