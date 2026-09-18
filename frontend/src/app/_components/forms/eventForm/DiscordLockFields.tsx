"use client";

// "Lock to Discord servers" toggle + the Discord-guild multi-select that appears when enabled.
import { Control, Controller, UseFormSetValue } from "react-hook-form";
import { MultiSelect } from "@/app/_components/Select";
import { ToggleRow } from "@/app/_components/ToggleRow";
import { EventFormValues } from "./schema";

export function DiscordLockFields({
  control,
  setValue,
  watchedDiscordLock,
  discordGuilds,
  discordGuildsLoading,
  discordGuildsError,
  discordGuildIdsError,
  readOnly,
}: {
  control: Control<EventFormValues>;
  setValue: UseFormSetValue<EventFormValues>;
  watchedDiscordLock: boolean;
  discordGuilds: { id: string; name: string }[];
  discordGuildsLoading: boolean;
  discordGuildsError: string | null;
  discordGuildIdsError?: string;
  readOnly: boolean;
}) {
  return (
    <div className="pt-1 border-t border-white/[0.06] flex flex-col gap-3">
      <Controller
        name="discord_lock"
        control={control}
        render={({ field }) => (
          <ToggleRow
            label="Lock to Discord servers"
            description={
              field.value
                ? "Lock this event to one or more Discord servers you belong to. Only members of those servers can open it or register."
                : "Anyone with the link can open this event."
            }
            checked={field.value}
            onChange={(next) => {
              field.onChange(next);
              if (!next) {
                setValue("discord_guild_ids", []);
              }
            }}
            disabled={readOnly}
          />
        )}
      />
      {watchedDiscordLock && (
        <div className="flex flex-col gap-1">
          <Controller
            name="discord_guild_ids"
            control={control}
            render={({ field }) => (
              <MultiSelect
                value={field.value}
                onChange={field.onChange}
                options={discordGuilds.map((g) => ({ value: g.id, label: g.name }))}
                ariaLabel="Discord servers"
                placeholder={
                  discordGuildsLoading
                    ? "Loading Discord servers..."
                    : "Select Discord servers"
                }
                disabled={readOnly || discordGuildsLoading}
                isLoading={discordGuildsLoading}
              />
            )}
          />
          {discordGuildIdsError && (
            <p className="text-xs text-[var(--color-text-danger)]">{discordGuildIdsError}</p>
          )}
          {discordGuildsError && (
            <p className="text-xs text-[var(--color-text-danger)]">{discordGuildsError}</p>
          )}
        </div>
      )}
    </div>
  );
}
