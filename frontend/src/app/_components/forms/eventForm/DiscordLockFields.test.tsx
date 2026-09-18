// Tests for the "Lock to Discord servers" toggle + guild multi-select. Wrapped in a minimal
// react-hook-form harness since the component is driven by RHF `Control`/`setValue`.
import { useForm, useWatch } from "react-hook-form";
import { describe, expect, it, vi } from "vitest";
import { axe } from "@/test/axe";
import { render, screen, userEvent } from "@/test/render";
import { EventFormValues } from "./schema";
import { DiscordLockFields } from "./DiscordLockFields";

function Harness({
  discordGuilds = [{ id: "guild-1", name: "Server One" }],
  discordGuildsLoading = false,
  discordGuildsError = null,
  discordGuildIdsError,
  readOnly = false,
  defaultDiscordLock = false,
}: {
  discordGuilds?: { id: string; name: string }[];
  discordGuildsLoading?: boolean;
  discordGuildsError?: string | null;
  discordGuildIdsError?: string;
  readOnly?: boolean;
  defaultDiscordLock?: boolean;
}) {
  const { control, setValue } = useForm<EventFormValues>({
    defaultValues: {
      name: "",
      game_id: "game-1",
      game_mode_id: "mode-1",
      region: "AMER",
      start_time_local: "",
      sub_min: 0,
      games_to_run: 1,
      registration_open: true,
      sort_logic: "balanced",
      discord_lock: defaultDiscordLock,
      discord_guild_ids: [],
    },
  });
  const watchedDiscordLock = useWatch({ control, name: "discord_lock" });

  return (
    <DiscordLockFields
      control={control}
      setValue={setValue}
      watchedDiscordLock={watchedDiscordLock}
      discordGuilds={discordGuilds}
      discordGuildsLoading={discordGuildsLoading}
      discordGuildsError={discordGuildsError}
      discordGuildIdsError={discordGuildIdsError}
      readOnly={readOnly}
    />
  );
}

describe("DiscordLockFields", () => {
  it("hides the guild multi-select until the lock toggle is on", () => {
    render(<Harness />);
    expect(screen.queryByRole("combobox")).not.toBeInTheDocument();
  });

  it("shows the guild multi-select once the lock toggle is on", () => {
    render(<Harness defaultDiscordLock={true} />);
    expect(screen.getByRole("combobox")).toBeInTheDocument();
  });

  it("turning the toggle on reveals the multi-select", async () => {
    const user = userEvent.setup();
    render(<Harness />);

    await user.click(screen.getByRole("switch"));

    expect(screen.getByRole("combobox")).toBeInTheDocument();
  });

  it("shows a Discord-guild fetch error when present", () => {
    render(<Harness defaultDiscordLock={true} discordGuildsError="Could not load Discord servers." />);
    expect(screen.getByText("Could not load Discord servers.")).toBeInTheDocument();
  });

  it("shows a validation error for discord_guild_ids when present", () => {
    render(<Harness defaultDiscordLock={true} discordGuildIdsError="Select at least one Discord server." />);
    expect(screen.getByText("Select at least one Discord server.")).toBeInTheDocument();
  });

  it("disables the toggle and select when readOnly", () => {
    render(<Harness defaultDiscordLock={true} readOnly={true} />);
    expect(screen.getByRole("switch")).toBeDisabled();
    // A disabled native form control is excluded from the accessibility tree by default (matching
    // real AT behavior), so `hidden: true` is needed to still query it here.
    expect(screen.getByRole("combobox", { hidden: true })).toBeDisabled();
  });

  it("has no accessibility violations", async () => {
    const { container } = render(<Harness defaultDiscordLock={true} />);
    expect(await axe(container)).toHaveNoViolations();
  });
});
