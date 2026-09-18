"use client";

// Create or edit an event group: zod + react-hook-form, game/mode pickers, and API mutations.
// Presentational subpieces and pure helpers live in ./eventForm/; this file wires them together.
import { useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { Controller, useForm, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import "react-datepicker/dist/react-datepicker.css";
import { useAuth } from "@/app/_context/AuthContext";
import { NumberStepper } from "@/app/_components/NumberStepper";
import { Select, MultiSelect } from "@/app/_components/Select";
import { ToggleRow } from "@/app/_components/ToggleRow";
import { useCancelableFetch } from "@/app/_hooks/useCancelableFetch";
import { EVENT_NAME_MAX_RUNES, REGIONS } from "@/app/_lib/constants";
import { getUserTimeZone } from "@/app/_lib/dateTime";
import { datepickerStyles, inputCls } from "@/app/_lib/styles";
import { codePointLength } from "@/app/_lib/textInput";
import { createEvent, deleteEventGroup, updateEventGroup } from "@/app/_services/events";
import { extractApiError, fetchGameModes, fetchGamesForUser } from "@/app/_services/games";
import { fetchMyDiscordGuilds } from "@/app/_services/users";
import { Game, GameMode } from "@/app/_types/types";

import { DeleteEventDialog } from "./eventForm/DeleteEventDialog";
import { DiscordLockFields } from "./eventForm/DiscordLockFields";
import { EventFormDateTimePicker } from "./eventForm/EventFormDateTimePicker";
import { MatchmakingModeField } from "./eventForm/MatchmakingModeField";
import { PerGameScheduleEditor } from "./eventForm/PerGameScheduleEditor";
import {
  EventFormEditScheduleRow,
  EventFormValues,
  PerGameDraftRow,
  buildEventFormSchema,
  validateEditScheduleDraft,
} from "./eventForm/schema";
import { getInitialStartTimeLocal, toDateTimeLocalValue } from "./eventForm/dateTime";

export type { EventFormEditScheduleRow, EventFormValues };

interface EventFormProps {
  mode: "create" | "edit";
  onCancel: () => void;
  eventGroupId?: string;
  initialValues?: Partial<EventFormValues>;
  /** Edit only: current games from GET /events/:groupId (drives per-game time + mode UI). */
  editSchedule?: EventFormEditScheduleRow[];
  onSubmitted?: () => void;
  /** Edit only: view settings without mutation controls (non-host). */
  readOnly?: boolean;
}

export function EventForm({
  mode,
  onCancel,
  eventGroupId,
  initialValues,
  editSchedule,
  onSubmitted,
  readOnly = false,
}: EventFormProps) {
  const router = useRouter();
  const { user, isAuthenticated, isLoading: authLoading } = useAuth();
  const userTz = getUserTimeZone();

  const eventFormSchema = useMemo(() => buildEventFormSchema(mode), [mode]);

  const defaultValues = useMemo(
    (): EventFormValues => ({
      name: initialValues?.name ?? "",
      game_id: initialValues?.game_id ?? "",
      game_mode_id: initialValues?.game_mode_id ?? "",
      region: initialValues?.region ?? "",
      start_time_local: initialValues?.start_time_local ?? getInitialStartTimeLocal(mode),
      sub_min: initialValues?.sub_min ?? 0,
      games_to_run: initialValues?.games_to_run ?? 1,
      registration_open: initialValues?.registration_open ?? true,
      sort_logic: initialValues?.sort_logic ?? "balanced",
      discord_lock: initialValues?.discord_lock ?? false,
      discord_guild_ids: initialValues?.discord_guild_ids ?? [],
    }),
    [initialValues, mode],
  );

  const {
    control,
    handleSubmit,
    setValue,
    getValues,
    formState: { errors },
  } = useForm<EventFormValues>({
    resolver: zodResolver(eventFormSchema),
    defaultValues,
  });

  const watchedGameId = useWatch({ control, name: "game_id" });
  const watchedName = useWatch({ control, name: "name" }) ?? "";
  const watchedDiscordLock = useWatch({ control, name: "discord_lock" });
  const nameCodePoints = codePointLength(watchedName);
  const nameOverLimit = nameCodePoints > EVENT_NAME_MAX_RUNES;

  const [games, setGames] = useState<Game[]>([]);
  const [gamesLoading, setGamesLoading] = useState(false);
  const [gamesError, setGamesError] = useState<string | null>(null);

  const [modes, setModes] = useState<GameMode[]>([]);
  const [modesLoading, setModesLoading] = useState(false);
  const [modesError, setModesError] = useState<string | null>(null);

  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [isDeleteConfirmOpen, setIsDeleteConfirmOpen] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  const [discordGuilds, setDiscordGuilds] = useState<{ id: string; name: string }[]>([]);
  const [discordGuildsLoading, setDiscordGuildsLoading] = useState(false);
  const [discordGuildsError, setDiscordGuildsError] = useState<string | null>(null);

  const editScheduleSig =
    mode === "edit" && editSchedule?.length
      ? editSchedule.map((row) => `${row.id}:${row.start_time}:${row.game_mode_id}`).join("|")
      : "";

  const [syncedEditScheduleSig, setSyncedEditScheduleSig] = useState<string | null>(null);
  const [perGameDraft, setPerGameDraft] = useState<PerGameDraftRow[]>([]);

  if (editScheduleSig !== syncedEditScheduleSig) {
    setSyncedEditScheduleSig(editScheduleSig);
    setPerGameDraft(
      mode === "edit" && editSchedule?.length
        ? editSchedule.map((row) => ({
            eventId: row.id,
            startLocal: toDateTimeLocalValue(new Date(row.start_time)),
            modeId: row.game_mode_id,
          }))
        : [],
    );
  }

  useCancelableFetch({
    enabled: !authLoading && isAuthenticated && !!user?.id,
    fetcher: (signal) => fetchGamesForUser(user?.id ?? "", signal),
    onStart: () => {
      setGamesLoading(true);
      setGamesError(null);
    },
    onSuccess: (data) => setGames(data),
    onError: () => {
      setGames([]);
      setGamesError("Could not load available games.");
    },
    onSettled: () => setGamesLoading(false),
    deps: [authLoading, isAuthenticated, user?.id],
  });

  useCancelableFetch({
    enabled: !authLoading && isAuthenticated && !!watchedGameId,
    fetcher: (signal) => fetchGameModes(watchedGameId, signal),
    onStart: () => {
      setModesLoading(true);
      setModesError(null);
    },
    onSuccess: (data) => {
      setModes(data);
      const prevModeId = getValues("game_mode_id");
      if (mode !== "edit") {
        setValue("game_mode_id", data.some((m) => m.id === prevModeId) ? prevModeId : "");
      }
    },
    onError: () => {
      setModes([]);
      setValue("game_mode_id", "");
      setModesError("Could not load game modes.");
    },
    onSettled: () => setModesLoading(false),
    deps: [watchedGameId, getValues, setValue, authLoading, isAuthenticated, mode],
  });

  useCancelableFetch({
    enabled: !authLoading && isAuthenticated && !!watchedDiscordLock,
    fetcher: (signal) => fetchMyDiscordGuilds(signal),
    onStart: () => {
      setDiscordGuildsLoading(true);
      setDiscordGuildsError(null);
    },
    onSuccess: (data) => {
      setDiscordGuilds(
        [...data].sort((a, b) => a.name.localeCompare(b.name, undefined, { sensitivity: "base" })),
      );
    },
    onError: (err) => {
      setDiscordGuilds([]);
      setDiscordGuildsError(extractApiError(err, "Could not load Discord servers."));
    },
    onSettled: () => setDiscordGuildsLoading(false),
    deps: [authLoading, isAuthenticated, watchedDiscordLock],
  });

  const onValidSubmit = async (data: EventFormValues) => {
    if (readOnly) return;
    setSubmitError(null);
    if (!user?.id) {
      setSubmitError("You must be signed in to create an event.");
      return;
    }
    try {
      setIsSubmitting(true);
      if (mode === "create") {
        if (!data.game_mode_id || !data.start_time_local) {
          setSubmitError("Game mode and start time are required.");
          return;
        }
        const startDate = new Date(data.start_time_local);
        const result = await createEvent({
          game_mode_id: data.game_mode_id,
          region: data.region,
          start_time: startDate.toISOString(),
          sub_min: data.sub_min,
          games_to_run: data.games_to_run,
          registration_open: data.registration_open,
          sort_logic: data.sort_logic,
          name: data.name,
          discord_guild_ids: data.discord_lock ? data.discord_guild_ids : [],
        });
        onCancel();
        router.push(`/event/${result.group_id}`);
        return;
      }

      if (!eventGroupId) {
        setSubmitError("Missing event group id for edit mode.");
        return;
      }

      if (perGameDraft.length === 0) {
        setSubmitError("Could not load games for this event.");
        return;
      }

      const scheduleErr = validateEditScheduleDraft(perGameDraft);
      if (scheduleErr) {
        setSubmitError(scheduleErr);
        return;
      }

      await updateEventGroup(eventGroupId, {
        region: data.region,
        sub_min: data.sub_min,
        sort_logic: data.sort_logic,
        registration_open: data.registration_open,
        name: data.name,
        discord_guild_ids: data.discord_lock ? data.discord_guild_ids : [],
        events: perGameDraft.map((row) => ({
          event_id: row.eventId,
          start_time: new Date(row.startLocal).toISOString(),
          game_mode_id: row.modeId,
        })),
      });
      onSubmitted?.();
      onCancel();
    } catch (err) {
      setSubmitError(
        extractApiError(
          err,
          mode === "create" ? "Could not create event. Please try again." : "Could not update event settings.",
        ),
      );
    } finally {
      setIsSubmitting(false);
    }
  };

  const onConfirmDelete = async () => {
    if (readOnly || mode !== "edit" || !eventGroupId) return;
    setDeleteError(null);
    try {
      setIsDeleting(true);
      await deleteEventGroup(eventGroupId);
      setIsDeleteConfirmOpen(false);
      onCancel();
      router.push("/my_events");
    } catch (err) {
      setDeleteError(extractApiError(err, "Could not delete this event group."));
    } finally {
      setIsDeleting(false);
    }
  };

  return (
    <>
      <form
        className="flex flex-col gap-4"
        onSubmit={readOnly ? (e) => e.preventDefault() : handleSubmit(onValidSubmit)}
        noValidate
      >
        <p className="text-xs text-[var(--color-text-muted)]">
          {readOnly ? "Times are shown in your local timezone: " : "Times are entered in your local timezone: "}
          <span className="text-[var(--color-text-soft)]">{userTz}</span>
        </p>

        <div className="flex flex-col gap-1.5">
          <label
            htmlFor="event-form-name"
            className={`text-xs font-medium tracking-wide ${
              nameOverLimit ? "text-[var(--color-text-danger)]" : "text-[var(--color-text-soft)]"
            }`}
          >
            Event name ({nameCodePoints}/{EVENT_NAME_MAX_RUNES})
          </label>
          <Controller
            name="name"
            control={control}
            render={({ field }) => (
              <input
                {...field}
                id="event-form-name"
                type="text"
                autoComplete="off"
                placeholder="Optional custom title"
                className={inputCls}
                disabled={readOnly}
                readOnly={readOnly}
              />
            )}
          />
          {errors.name && (
            <p className="text-xs text-[var(--color-text-danger)]">{errors.name.message}</p>
          )}
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <div className="flex flex-col gap-1.5">
            <label
              htmlFor="event-form-game"
              className="text-xs font-medium tracking-wide text-[var(--color-text-soft)]"
            >
              Game *
            </label>
            <Controller
              name="game_id"
              control={control}
              render={({ field }) => (
                <Select
                  inputId="event-form-game"
                  value={field.value ?? ""}
                  onChange={(nextId) => {
                    field.onChange(nextId);
                    setValue("game_mode_id", "");
                    setModes([]);
                    setModesError(null);
                  }}
                  disabled={readOnly || mode === "edit" || gamesLoading || !user?.id || !!gamesError}
                  placeholder={gamesLoading ? "Loading games..." : "Select game"}
                  options={games.map((game) => ({ value: game.id, label: game.name }))}
                />
              )}
            />
            {!user?.id ? (
              <p className="text-xs text-[var(--color-text-danger)]">
                Unable to load games without a signed-in user.
              </p>
            ) : (
              gamesError && <p className="text-xs text-[var(--color-text-danger)]">{gamesError}</p>
            )}
            {mode === "edit" && !readOnly && (
              <p className="text-xs text-[var(--color-text-faint)]">
                Game cannot be changed. Adjust date, time, and mode for each scheduled game below.
              </p>
            )}
            {errors.game_id && (
              <p className="text-xs text-[var(--color-text-danger)]">{errors.game_id.message}</p>
            )}
          </div>

          {mode !== "edit" && (
            <div className="flex flex-col gap-1.5">
              <label
                htmlFor="event-form-game-mode"
                className="text-xs font-medium tracking-wide text-[var(--color-text-soft)]"
              >
                Game mode *
              </label>
              <Controller
                name="game_mode_id"
                control={control}
                render={({ field }) => (
                  <Select
                    inputId="event-form-game-mode"
                    value={field.value ?? ""}
                    onChange={field.onChange}
                    disabled={!watchedGameId || modesLoading || !!modesError}
                    placeholder={
                      !watchedGameId
                        ? "Select game first"
                        : modesLoading
                          ? "Loading game modes..."
                          : "Select game mode"
                    }
                    options={modes.map((gameMode) => ({ value: gameMode.id, label: gameMode.name }))}
                  />
                )}
              />
              {modesError && <p className="text-xs text-[var(--color-text-danger)]">{modesError}</p>}
              {errors.game_mode_id && (
                <p className="text-xs text-[var(--color-text-danger)]">{errors.game_mode_id.message}</p>
              )}
            </div>
          )}

          <div className={`flex flex-col gap-1.5${mode === "edit" ? " sm:col-span-2" : ""}`}>
            <label
              htmlFor="event-form-region"
              className="text-xs font-medium tracking-wide text-[var(--color-text-soft)]"
            >
              Region *
            </label>
            <Controller
              name="region"
              control={control}
              render={({ field }) => (
                <Select
                  inputId="event-form-region"
                  value={field.value ?? ""}
                  onChange={field.onChange}
                  disabled={readOnly}
                  placeholder="Select region"
                  options={REGIONS.map((r) => ({ value: r, label: r }))}
                />
              )}
            />
            {errors.region && (
              <p className="text-xs text-[var(--color-text-danger)]">{errors.region.message}</p>
            )}
          </div>

          {mode !== "edit" && (
            <div className="flex flex-col gap-1.5">
              <label
                htmlFor="event-form-start-time"
                className="text-xs font-medium tracking-wide text-[var(--color-text-soft)]"
              >
                First game start time *
              </label>
              <Controller
                name="start_time_local"
                control={control}
                render={({ field }) => (
                  <EventFormDateTimePicker
                    id="event-form-start-time"
                    value={field.value ?? ""}
                    onChange={field.onChange}
                    onBlur={field.onBlur}
                    disallowPast
                  />
                )}
              />
              {errors.start_time_local && (
                <p className="text-xs text-[var(--color-text-danger)]">{errors.start_time_local.message}</p>
              )}
            </div>
          )}

          <div className="flex flex-col gap-1.5 sm:col-span-2">
            <span
              id="matchmaking-mode-label"
              className="text-xs font-medium tracking-wide text-[var(--color-text-soft)]"
            >
              Matchmaking Mode
            </span>
            <Controller
              name="sort_logic"
              control={control}
              render={({ field }) => (
                <MatchmakingModeField
                  value={field.value}
                  onChange={field.onChange}
                  disabled={readOnly}
                />
              )}
            />
            {errors.sort_logic && (
              <p className="text-xs text-[var(--color-text-danger)]">{errors.sort_logic.message}</p>
            )}
          </div>

          {mode === "edit" && (
            <div className="flex flex-col gap-1.5">
              <Controller
                name="sub_min"
                control={control}
                render={({ field }) => (
                  <NumberStepper
                    label="Minimum subs per lobby"
                    value={field.value}
                    min={0}
                    onChange={field.onChange}
                    disabled={readOnly}
                    hint="Additional lobbies are created only after this many subs are available per lobby."
                  />
                )}
              />
              {errors.sub_min && (
                <p className="text-xs text-[var(--color-text-danger)]">{errors.sub_min.message}</p>
              )}
            </div>
          )}
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          {mode !== "edit" && (
            <div className="flex flex-col gap-1.5">
              <Controller
                name="sub_min"
                control={control}
                render={({ field }) => (
                  <NumberStepper
                    label="Minimum subs per lobby"
                    value={field.value}
                    min={0}
                    onChange={field.onChange}
                    hint="Additional lobbies are created only after this many subs are available per lobby."
                  />
                )}
              />
              {errors.sub_min && (
                <p className="text-xs text-[var(--color-text-danger)]">{errors.sub_min.message}</p>
              )}
            </div>
          )}

          {mode !== "edit" && (
            <div className="flex flex-col gap-1.5">
              <Controller
                name="games_to_run"
                control={control}
                render={({ field }) => (
                  <NumberStepper
                    label="Number of games in event"
                    value={field.value}
                    min={1}
                    onChange={field.onChange}
                  />
                )}
              />
              {errors.games_to_run && (
                <p className="text-xs text-[var(--color-text-danger)]">{errors.games_to_run.message}</p>
              )}
            </div>
          )}
        </div>

        {mode === "edit" && (
          <PerGameScheduleEditor
            perGameDraft={perGameDraft}
            setPerGameDraft={setPerGameDraft}
            modes={modes}
            modesLoading={modesLoading}
            modesError={modesError}
            watchedGameId={watchedGameId}
            readOnly={readOnly}
            userTz={userTz}
          />
        )}

        <div className="pt-1 border-t border-white/[0.06]">
          <Controller
            name="registration_open"
            control={control}
            render={({ field }) => (
              <ToggleRow
                label="Registration Status"
                description={
                  field.value
                    ? "Players can register for this event."
                    : "Registration is currently closed."
                }
                checked={field.value}
                onChange={field.onChange}
                disabled={readOnly}
              />
            )}
          />
        </div>

        <DiscordLockFields
          control={control}
          setValue={setValue}
          watchedDiscordLock={watchedDiscordLock}
          discordGuilds={discordGuilds}
          discordGuildsLoading={discordGuildsLoading}
          discordGuildsError={discordGuildsError}
          discordGuildIdsError={errors.discord_guild_ids?.message}
          readOnly={readOnly}
        />

        {submitError && <p className="text-xs text-[var(--color-text-danger)]">{submitError}</p>}

        {readOnly ? (
          <div className="mt-1 flex justify-end">
            <button
              type="button"
              onClick={onCancel}
              className="px-3 py-2 rounded-lg text-sm font-medium border border-white/10 bg-white/[0.03] text-[var(--color-text-muted)] hover:text-[var(--color-text-soft)] transition-colors"
            >
              Close
            </button>
          </div>
        ) : (
          <div className="mt-1 flex items-center justify-between gap-2">
            <div>
              {mode === "edit" && (
                <button
                  type="button"
                  onClick={() => setIsDeleteConfirmOpen(true)}
                  disabled={isSubmitting || isDeleting}
                  className="px-3 py-2 rounded-lg text-sm font-medium border border-[var(--color-text-danger)]/40 bg-[var(--color-text-danger)]/10 text-[var(--color-text-danger)] hover:bg-[var(--color-text-danger)]/20 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Delete Event
                </button>
              )}
            </div>
            <div className="flex items-center gap-2">
              <button
                type="button"
                onClick={onCancel}
                disabled={isSubmitting || isDeleting}
                className="px-3 py-2 rounded-lg text-sm font-medium border border-white/10 bg-white/[0.03] text-[var(--color-text-muted)] hover:text-[var(--color-text-soft)] transition-colors"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={isSubmitting || isDeleting}
                className="px-3 py-2 rounded-lg text-sm font-medium border border-[var(--color-accent-blue)]/30 bg-[var(--color-accent-blue)]/10 text-[var(--color-accent-blue)] hover:bg-[var(--color-accent-blue)]/20 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isSubmitting
                  ? mode === "create"
                    ? "Creating..."
                    : "Saving..."
                  : mode === "create"
                    ? "Create Event"
                    : "Save Settings"}
              </button>
            </div>
          </div>
        )}
      </form>
      {!readOnly && (
        <DeleteEventDialog
          isOpen={isDeleteConfirmOpen}
          deleteError={deleteError}
          isDeleting={isDeleting}
          onCancel={() => setIsDeleteConfirmOpen(false)}
          onConfirm={onConfirmDelete}
        />
      )}
      <style>{datepickerStyles}</style>
    </>
  );
}
