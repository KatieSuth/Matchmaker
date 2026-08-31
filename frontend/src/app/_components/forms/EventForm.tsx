"use client";

// Create or edit an event group: zod + react-hook-form, game/mode pickers, and API mutations.
import { useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import DatePicker from "react-datepicker";
import { Controller, useForm, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import "react-datepicker/dist/react-datepicker.css";
import { useAuth } from "@/app/_context/AuthContext";
import { Select, MultiSelect } from "@/app/_components/Select";
import { ToggleRow } from "@/app/_components/ToggleRow";
import { EVENT_NAME_MAX_RUNES, REGIONS } from "@/app/_lib/constants";
import { DATEPICKER_PORTAL_ID, datepickerStyles, inputCls } from "@/app/_lib/styles";
import { codePointLength, optionalFreeTextSchema } from "@/app/_lib/textInput";
import { createEvent, deleteEventGroup, updateEventGroup } from "@/app/_services/events";
import { extractApiError, fetchGameModes, fetchGamesForUser } from "@/app/_services/games";
import { fetchMyDiscordGuilds } from "@/app/_services/users";
import { Game, GameMode } from "@/app/_types/types";

export type EventFormEditScheduleRow = {
  id: string;
  start_time: string;
  game_mode_id: string;
};

/** Zod schema for create vs edit: edit omits required start time / mode; Discord lock needs at least one server. */
function buildEventFormSchema(mode: "create" | "edit") {
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

function toDateTimeLocalValue(date: Date): string {
  const pad = (n: number) => n.toString().padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(
    date.getHours()
  )}:${pad(date.getMinutes())}`;
}

function roundUpToQuarterHour(date: Date): Date {
  const rounded = new Date(date);
  rounded.setSeconds(0, 0);
  const minutes = rounded.getMinutes();
  const remainder = minutes % 15;
  if (remainder !== 0) rounded.setMinutes(minutes + (15 - remainder));
  return rounded;
}

function getInitialStartTimeLocal(mode: "create" | "edit"): string {
  if (mode !== "create") return "";
  const now = roundUpToQuarterHour(new Date(Date.now() + 30 * 60 * 1000));
  return toDateTimeLocalValue(now);
}

function parseLocalDateTimeString(s: string): Date | null {
  if (!s?.trim()) return null;
  const d = new Date(s);
  return Number.isNaN(d.getTime()) ? null : d;
}

function startOfToday(): Date {
  const d = new Date();
  d.setHours(0, 0, 0, 0);
  return d;
}

/** Local datetime string (same shape as `datetime-local`) for API + zod; 15-minute steps. */
function EventFormDateTimePicker({
  value,
  onChange,
  onBlur,
  name,
  id,
  disallowPast,
  placeholderText = "Select date & time",
  disabled = false,
}: {
  value: string;
  onChange: (v: string) => void;
  onBlur?: () => void;
  name?: string;
  id?: string;
  disallowPast?: boolean;
  placeholderText?: string;
  disabled?: boolean;
}) {
  const selected = parseLocalDateTimeString(value);
  const filterTime = (time: Date) => {
    if (!disallowPast) return true;
    return time.getTime() > Date.now();
  };

  return (
    <DatePicker
      id={id}
      name={name}
      onBlur={onBlur}
      portalId={DATEPICKER_PORTAL_ID}
      selected={selected}
      onChange={(date: Date | null) => {
        onChange(date ? toDateTimeLocalValue(date) : "");
      }}
      showTimeSelect
      timeIntervals={15}
      timeCaption="Time"
      showMonthDropdown
      showYearDropdown
      dropdownMode="select"
      dateFormat="MMM d, yyyy h:mm aa"
      placeholderText={placeholderText}
      className={inputCls}
      minDate={disallowPast ? startOfToday() : undefined}
      filterTime={disallowPast ? filterTime : undefined}
      popperPlacement="bottom-start"
      popperProps={{ strategy: "fixed" }}
      disabled={disabled}
    />
  );
}

interface NumberStepperProps {
  label: string;
  value: number;
  min: number;
  onChange: (next: number) => void;
  hint?: string;
  disabled?: boolean;
}

function NumberStepper({ label, value, min, onChange, hint, disabled = false }: NumberStepperProps) {
  const decrement = () => onChange(Math.max(min, value - 1));
  const increment = () => onChange(value + 1);

  return (
    <div className="flex flex-col gap-1.5">
      <label className="text-xs font-medium tracking-wide text-[var(--color-text-soft)]">{label}</label>
      <div className="flex items-center gap-2">
        <button
          type="button"
          onClick={decrement}
          disabled={disabled || value <= min}
          className="h-9 w-9 rounded-lg border border-white/10 bg-white/[0.03] text-[var(--color-text-soft)] transition-colors hover:bg-white/[0.08] disabled:opacity-40 disabled:cursor-not-allowed"
          aria-label={`Decrease ${label}`}
        >
          <svg width="12" height="12" viewBox="0 0 12 12" fill="none" className="mx-auto">
            <path d="M2.25 6h7.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
          </svg>
        </button>
        <input
          type="number"
          min={min}
          value={value}
          disabled={disabled}
          onChange={(event) => onChange(Math.max(min, Number(event.target.value) || min))}
          className={`${inputCls} no-native-spinner text-center`}
        />
        <button
          type="button"
          onClick={increment}
          disabled={disabled}
          className="h-9 w-9 rounded-lg border border-white/10 bg-white/[0.03] text-[var(--color-text-soft)] transition-colors hover:bg-white/[0.08]"
          aria-label={`Increase ${label}`}
        >
          <svg width="12" height="12" viewBox="0 0 12 12" fill="none" className="mx-auto">
            <path
              d="M6 2.25v7.5M2.25 6h7.5"
              stroke="currentColor"
              strokeWidth="1.5"
              strokeLinecap="round"
            />
          </svg>
        </button>
      </div>
      {hint && <p className="text-xs text-[var(--color-text-faint)]">{hint}</p>}
    </div>
  );
}

interface MatchmakingModeOption {
  value: "balanced" | "ranked";
  title: string;
  description: string;
}

const MATCHMAKING_MODE_OPTIONS: readonly MatchmakingModeOption[] = [
  {
    value: "balanced",
    title: "Balanced",
    description:
      "Puts similar ranks on opposite teams (one rank band per player slot) so each side gets a matching mix. Best for casual games.",
  },
  {
    value: "ranked",
    title: "Rank Grouping",
    description:
      "Keeps players of similar skill in the same lobby. Best for serious practice.",
  },
];

interface MatchmakingModeFieldProps {
  value: "balanced" | "ranked";
  onChange: (next: "balanced" | "ranked") => void;
  disabled?: boolean;
}

function MatchmakingModeField({ value, onChange, disabled = false }: MatchmakingModeFieldProps) {
  return (
    <div
      className="flex flex-col gap-2"
      role="radiogroup"
      aria-labelledby="matchmaking-mode-label"
    >
      {MATCHMAKING_MODE_OPTIONS.map((opt) => {
        const selected = value === opt.value;
        return (
          <label
            key={opt.value}
            className={`flex gap-3 rounded-lg border p-3 transition-colors ${
              disabled
                ? "cursor-not-allowed opacity-60"
                : "cursor-pointer"
            } ${
              selected
                ? "border-[var(--color-accent-blue)]/50 bg-[var(--color-accent-blue)]/10"
                : disabled
                  ? "border-white/10 bg-white/[0.02]"
                  : "border-white/10 bg-white/[0.02] hover:bg-white/[0.04]"
            }`}
          >
            <input
              type="radio"
              name="sort_logic_form"
              value={opt.value}
              checked={selected}
              disabled={disabled}
              onChange={() => onChange(opt.value)}
              className="mt-1 h-4 w-4 shrink-0 accent-[var(--color-accent-blue)]"
            />
            <div className="min-w-0 flex flex-col gap-1">
              <span className="text-sm font-medium text-[var(--color-text-soft)]">{opt.title}</span>
              <p className="text-xs leading-relaxed text-[var(--color-text-faint)]">{opt.description}</p>
            </div>
          </label>
        );
      })}
    </div>
  );
}

type PerGameDraftRow = {
  eventId: string;
  startLocal: string;
  modeId: string;
};

function validateEditScheduleDraft(rows: PerGameDraftRow[]): string | null {
  const nowMs = Date.now();
  for (const row of rows) {
    if (!row.modeId) return "Game mode is required for each scheduled game.";
    const t = new Date(row.startLocal).getTime();
    if (Number.isNaN(t)) return "One or more game times are invalid.";
    if (t < nowMs) return "Start times cannot be in the past.";
  }
  return null;
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
  const userTz =
    typeof Intl !== "undefined" ? Intl.DateTimeFormat().resolvedOptions().timeZone : "UTC";

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
    [initialValues, mode]
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
        : []
    );
  }

  useEffect(() => {
    if (authLoading || !isAuthenticated || !user?.id) return;

    const ac = new AbortController();
    const { signal } = ac;
    const loadGames = async () => {
      setGamesLoading(true);
      setGamesError(null);
      try {
        const data = await fetchGamesForUser(user.id, signal);
        if (signal.aborted) return;
        setGames(data);
      } catch (err) {
        if (
          signal.aborted ||
          (err as { code?: string; name?: string })?.code === "ERR_CANCELED" ||
          (err as { name?: string })?.name === "CanceledError"
        ) {
          return;
        }
        setGames([]);
        setGamesError("Could not load available games.");
      } finally {
        if (!signal.aborted) {
          setGamesLoading(false);
        }
      }
    };

    void loadGames();

    return () => {
      ac.abort();
    };
  }, [authLoading, isAuthenticated, user?.id]);

  useEffect(() => {
    if (authLoading || !isAuthenticated) return;
    if (!watchedGameId) return;

    const ac = new AbortController();
    const { signal } = ac;
    const loadModes = async () => {
      setModesLoading(true);
      setModesError(null);
      try {
        const data = await fetchGameModes(watchedGameId, signal);
        if (signal.aborted) return;
        setModes(data);
        const prevModeId = getValues("game_mode_id");
        if (mode !== "edit") {
          setValue("game_mode_id", data.some((m) => m.id === prevModeId) ? prevModeId : "");
        }
      } catch (err) {
        if (
          signal.aborted ||
          (err as { code?: string; name?: string })?.code === "ERR_CANCELED" ||
          (err as { name?: string })?.name === "CanceledError"
        ) {
          return;
        }
        setModes([]);
        setValue("game_mode_id", "");
        setModesError("Could not load game modes.");
      } finally {
        if (!signal.aborted) {
          setModesLoading(false);
        }
      }
    };

    void loadModes();

    return () => {
      ac.abort();
    };
  }, [watchedGameId, getValues, setValue, authLoading, isAuthenticated, mode]);

  useEffect(() => {
    if (authLoading || !isAuthenticated) return;
    if (!watchedDiscordLock) return;

    const ac = new AbortController();
    const { signal } = ac;
    const loadGuilds = async () => {
      setDiscordGuildsLoading(true);
      setDiscordGuildsError(null);
      try {
        const data = await fetchMyDiscordGuilds(signal);
        if (signal.aborted) return;
        setDiscordGuilds(
          [...data].sort((a, b) => a.name.localeCompare(b.name, undefined, { sensitivity: "base" })),
        );
      } catch (err) {
        if (
          signal.aborted ||
          (err as { code?: string; name?: string })?.code === "ERR_CANCELED" ||
          (err as { name?: string })?.name === "CanceledError"
        ) {
          return;
        }
        setDiscordGuilds([]);
        setDiscordGuildsError(extractApiError(err, "Could not load Discord servers."));
      } finally {
        if (!signal.aborted) {
          setDiscordGuildsLoading(false);
        }
      }
    };

    void loadGuilds();
    return () => {
      ac.abort();
    };
  }, [authLoading, isAuthenticated, watchedDiscordLock]);

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
          mode === "create" ? "Could not create event. Please try again." : "Could not update event settings."
        )
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
            <label className="text-xs font-medium tracking-wide text-[var(--color-text-soft)]">
              Game *
            </label>
            <Controller
              name="game_id"
              control={control}
              render={({ field }) => (
                <Select
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
              <label className="text-xs font-medium tracking-wide text-[var(--color-text-soft)]">
                Game mode *
              </label>
              <Controller
                name="game_mode_id"
                control={control}
                render={({ field }) => (
                  <Select
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
            <label className="text-xs font-medium tracking-wide text-[var(--color-text-soft)]">
              Region *
            </label>
            <Controller
              name="region"
              control={control}
              render={({ field }) => (
                <Select
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

        {mode === "edit" && perGameDraft.length > 0 && (
          <div className="flex flex-col gap-3 pt-2 border-t border-white/[0.06]">
            <p className="text-xs font-medium tracking-wide text-[var(--color-text-soft)]">
              Games in this series
            </p>
            <p className="text-xs text-[var(--color-text-faint)]">
              {readOnly
                ? `Each row is one scheduled game. Times use your local timezone (${userTz}).`
                : `Each row applies to one scheduled game. Times use your local timezone (${userTz}).`}
            </p>
            <div className="flex flex-col gap-4">
              {perGameDraft.map((row, index) => (
                <div
                  key={row.eventId}
                  className="rounded-lg border border-white/[0.08] bg-white/[0.02] p-3 flex flex-col gap-2"
                >
                  <p className="text-xs font-semibold text-[var(--color-text-soft)]">Game {index + 1}</p>
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                    <div className="flex flex-col gap-1.5">
                      <label
                        htmlFor={`event-form-game-start-${row.eventId}`}
                        className="text-xs font-medium tracking-wide text-[var(--color-text-soft)]"
                      >
                        Start time *
                      </label>
                      <EventFormDateTimePicker
                        id={`event-form-game-start-${row.eventId}`}
                        value={row.startLocal}
                        onChange={(v) => {
                          setPerGameDraft((prev) =>
                            prev.map((r) => (r.eventId === row.eventId ? { ...r, startLocal: v } : r))
                          );
                        }}
                        disallowPast={false}
                        disabled={readOnly}
                      />
                    </div>
                    <div className="flex flex-col gap-1.5">
                      <label className="text-xs font-medium tracking-wide text-[var(--color-text-soft)]">
                        Game mode *
                      </label>
                      <Select
                        value={row.modeId}
                        onChange={(nextMode) => {
                          setPerGameDraft((prev) =>
                            prev.map((r) => (r.eventId === row.eventId ? { ...r, modeId: nextMode } : r))
                          );
                        }}
                        disabled={readOnly || !watchedGameId || modesLoading || !!modesError}
                        placeholder={
                          !watchedGameId
                            ? "Select game first"
                            : modesLoading
                              ? "Loading game modes..."
                              : "Select game mode"
                        }
                        options={modes.map((gameMode) => ({ value: gameMode.id, label: gameMode.name }))}
                      />
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
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
              {errors.discord_guild_ids && (
                <p className="text-xs text-[var(--color-text-danger)]">{errors.discord_guild_ids.message}</p>
              )}
              {discordGuildsError && (
                <p className="text-xs text-[var(--color-text-danger)]">{discordGuildsError}</p>
              )}
            </div>
          )}
        </div>

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
      {!readOnly && isDeleteConfirmOpen && (
        <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/70 px-4">
          <div className="w-full max-w-md rounded-xl border border-white/10 bg-[var(--color-bg)] p-5 shadow-[0_30px_90px_rgba(0,0,0,0.7)]">
            <h3 className="text-base font-semibold text-[var(--color-text)]">Delete Event</h3>
            <p className="mt-2 text-sm text-[var(--color-text-soft)]">
              This action cannot be undone. All games, registrations, and teams in this event group will be permanently deleted.
            </p>
            {deleteError && <p className="mt-3 text-xs text-[var(--color-text-danger)]">{deleteError}</p>}
            <div className="mt-4 flex justify-end gap-2">
              <button
                type="button"
                onClick={() => setIsDeleteConfirmOpen(false)}
                disabled={isDeleting}
                className="px-3 py-2 rounded-lg text-sm font-medium border border-white/10 bg-white/[0.03] text-[var(--color-text-muted)] hover:text-[var(--color-text-soft)] transition-colors disabled:opacity-50"
              >
                Cancel
              </button>
              <button
                type="button"
                onClick={onConfirmDelete}
                disabled={isDeleting}
                className="px-3 py-2 rounded-lg text-sm font-medium border border-[var(--color-text-danger)]/40 bg-[var(--color-text-danger)]/10 text-[var(--color-text-danger)] hover:bg-[var(--color-text-danger)]/20 transition-colors disabled:opacity-50"
              >
                {isDeleting ? "Deleting..." : "Delete Permanently"}
              </button>
            </div>
          </div>
        </div>
      )}
      <style jsx global>{`
        .no-native-spinner {
          -moz-appearance: textfield;
          appearance: textfield;
        }
        .no-native-spinner::-webkit-outer-spin-button,
        .no-native-spinner::-webkit-inner-spin-button {
          -webkit-appearance: none;
          margin: 0;
        }
      `}</style>
      <style>{datepickerStyles}</style>
    </>
  );
}
