// Zod schema and derived types for the user preferences (profile settings) form.
import { z } from "zod";
import { REGIONS, DISPLAY_NAME_MAX_RUNES } from "@/app/_lib/constants";
import { optionalFreeTextSchema } from "@/app/_lib/textInput";

export const userGameSchema = z.object({
  game_id: z.string().uuid("Please select a game"),
  in_game_name: z.string().min(1, "In-game name is required"),
  current_rank: z.string().min(1, "Current rank is required"),
  peak_rank: z.string().min(1, "Peak rank is required"),
  show_rank: z.boolean(),
  api_permission: z.boolean(),
});

export const preferencesSchema = z.object({
  display_name: optionalFreeTextSchema(DISPLAY_NAME_MAX_RUNES),
  pronouns: z.string().nullable().optional(),
  show_pronouns: z.boolean(),
  region: z.enum(REGIONS).nullable().optional(),
  games: z.array(userGameSchema),
});

export type PreferencesFormValues = z.infer<typeof preferencesSchema>;
