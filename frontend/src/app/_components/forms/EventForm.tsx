"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/app/_context/AuthContext";
import { Select } from "@/app/_components/Select";
import { ToggleRow } from "@/app/_components/ToggleRow";
import { inputCls } from "@/app/_lib/styles";
import { createEvent } from "@/app/_services/events";
import { extractApiError, fetchGameModes, fetchGamesForUser } from "@/app/_services/games";
import { Game, GameMode } from "@/app/_types/types";

interface EventFormProps {
  mode: "create" | "edit";
  onCancel: () => void;
}

const REGIONS = ["NA", "EMEA", "APAC"] as const;

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

interface NumberStepperProps {
  label: string;
  value: number;
  min: number;
  onChange: (next: number) => void;
  hint?: string;
}

function NumberStepper({ label, value, min, onChange, hint }: NumberStepperProps) {
  const decrement = () => onChange(Math.max(min, value - 1));
  const increment = () => onChange(value + 1);

  return (
    <div className="flex flex-col gap-1.5">
      <label className="text-xs font-medium tracking-wide text-[var(--color-text-soft)]">{label}</label>
      <div className="flex items-center gap-2">
        <button
          type="button"
          onClick={decrement}
          disabled={value <= min}
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
          onChange={(event) => onChange(Math.max(min, Number(event.target.value) || min))}
          className={`${inputCls} no-native-spinner text-center`}
        />
        <button
          type="button"
          onClick={increment}
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

export function EventForm({ mode, onCancel }: EventFormProps) {
  const router = useRouter();
  const { user } = useAuth();
  const userTz =
    typeof Intl !== "undefined" ? Intl.DateTimeFormat().resolvedOptions().timeZone : "UTC";

  const [games, setGames] = useState<Game[]>([]);
  const [gamesLoading, setGamesLoading] = useState(false);
  const [gamesError, setGamesError] = useState<string | null>(null);

  const [modes, setModes] = useState<GameMode[]>([]);
  const [modesLoading, setModesLoading] = useState(false);
  const [modesError, setModesError] = useState<string | null>(null);

  const [gameId, setGameId] = useState("");
  const [gameModeId, setGameModeId] = useState("");
  const [region, setRegion] = useState<(typeof REGIONS)[number] | "">("");
  const [startTimeLocal, setStartTimeLocal] = useState(() => getInitialStartTimeLocal(mode));
  const [minSubsPerLobby, setMinSubsPerLobby] = useState(1);
  const [registrationOpen, setRegistrationOpen] = useState(true);
  const [gamesToRun, setGamesToRun] = useState(1);

  const [submitAttempted, setSubmitAttempted] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);

  useEffect(() => {
    if (!user?.id) return;

    let ignore = false;
    const loadGames = async () => {
      setGamesLoading(true);
      setGamesError(null);
      try {
        const data = await fetchGamesForUser(user.id);
        if (ignore) return;
        setGames(data);
      } catch {
        if (ignore) return;
        setGames([]);
        setGamesError("Could not load available games.");
      } finally {
        if (ignore) return;
        setGamesLoading(false);
      }
    };

    void loadGames();

    return () => {
      ignore = true;
    };
  }, [user?.id]);

  const handleGameChange = (nextGameId: string) => {
    setGameId(nextGameId);
    setGameModeId("");
    setModes([]);
    setModesError(null);
  };

  useEffect(() => {
    if (!gameId) return;

    let ignore = false;
    const loadModes = async () => {
      setModesLoading(true);
      setModesError(null);
      try {
        const data = await fetchGameModes(gameId);
        if (ignore) return;
        setModes(data);
        setGameModeId((prev) => (data.some((m) => m.id === prev) ? prev : ""));
      } catch {
        if (ignore) return;
        setModes([]);
        setGameModeId("");
        setModesError("Could not load game modes.");
      } finally {
        if (ignore) return;
        setModesLoading(false);
      }
    };

    void loadModes();

    return () => {
      ignore = true;
    };
  }, [gameId]);

  const hasGame = gameId.length > 0;
  const hasMode = gameModeId.length > 0;
  const hasRegion = region.length > 0;
  const hasStartTime =
    startTimeLocal.length > 0 && !Number.isNaN(new Date(startTimeLocal).getTime());
  const hasValidSubMin = minSubsPerLobby > 0;
  const hasValidGamesToRun = gamesToRun > 0;

  const isValid = hasGame && hasMode && hasRegion && hasStartTime && hasValidSubMin && hasValidGamesToRun;

  const onSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setSubmitAttempted(true);
    setSubmitError(null);
    if (!isValid) return;
    if (!user?.id) {
      setSubmitError("You must be signed in to create an event.");
      return;
    }
    if (mode !== "create") return;

    const startDate = new Date(startTimeLocal);
    if (Number.isNaN(startDate.getTime())) {
      setSubmitError("Start time is invalid.");
      return;
    }

    try {
      setIsSubmitting(true);
      const result = await createEvent({
        game_mode_id: gameModeId,
        region,
        start_time: startDate.toISOString(),
        sub_min: minSubsPerLobby,
        games_to_run: gamesToRun,
        registration_open: registrationOpen,
      });
      onCancel();
      router.push(`/event/${result.group_id}`);
    } catch (err) {
      setSubmitError(extractApiError(err, "Could not create event. Please try again."));
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <>
      <form className="flex flex-col gap-4" onSubmit={onSubmit} noValidate>
      <p className="text-xs text-[var(--color-text-muted)]">
        Times are entered in your local timezone: <span className="text-[var(--color-text-soft)]">{userTz}</span>
      </p>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <div className="flex flex-col gap-1.5">
          <label className="text-xs font-medium tracking-wide text-[var(--color-text-soft)]">
            Game *
          </label>
          <Select
            value={gameId}
            onChange={handleGameChange}
            disabled={mode === "edit" || gamesLoading || !user?.id || !!gamesError}
            placeholder={gamesLoading ? "Loading games..." : "Select game"}
            options={games.map((game) => ({ value: game.id, label: game.name }))}
          />
          {!user?.id ? (
            <p className="text-xs text-[var(--color-text-danger)]">
              Unable to load games without a signed-in user.
            </p>
          ) : (
            gamesError && <p className="text-xs text-[var(--color-text-danger)]">{gamesError}</p>
          )}
          {mode === "edit" && (
            <p className="text-xs text-[var(--color-text-faint)]">Game cannot be changed while editing.</p>
          )}
          {submitAttempted && !hasGame && (
            <p className="text-xs text-[var(--color-text-danger)]">Game is required.</p>
          )}
        </div>

        <div className="flex flex-col gap-1.5">
          <label className="text-xs font-medium tracking-wide text-[var(--color-text-soft)]">
            Game mode *
          </label>
          <Select
            value={gameModeId}
            onChange={setGameModeId}
            disabled={!gameId || modesLoading || !!modesError}
            placeholder={
              !gameId ? "Select game first" : modesLoading ? "Loading game modes..." : "Select game mode"
            }
            options={modes.map((gameMode) => ({ value: gameMode.id, label: gameMode.name }))}
          />
          {modesError && <p className="text-xs text-[var(--color-text-danger)]">{modesError}</p>}
          {submitAttempted && !hasMode && (
            <p className="text-xs text-[var(--color-text-danger)]">Game mode is required.</p>
          )}
        </div>

        <div className="flex flex-col gap-1.5">
          <label className="text-xs font-medium tracking-wide text-[var(--color-text-soft)]">
            Region *
          </label>
          <Select
            value={region}
            onChange={(value) => setRegion(value as (typeof REGIONS)[number] | "")}
            placeholder="Select region"
            options={REGIONS.map((r) => ({ value: r, label: r }))}
          />
          {submitAttempted && !hasRegion && (
            <p className="text-xs text-[var(--color-text-danger)]">Region is required.</p>
          )}
        </div>

        <div className="flex flex-col gap-1.5">
          <label className="text-xs font-medium tracking-wide text-[var(--color-text-soft)]">
            First game start time *
          </label>
          <input
            type="datetime-local"
            value={startTimeLocal}
            onChange={(event) => setStartTimeLocal(event.target.value)}
            step={900}
            className={inputCls}
          />
          {submitAttempted && !hasStartTime && (
            <p className="text-xs text-[var(--color-text-danger)]">Start time is required.</p>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <NumberStepper
          label="Minimum subs per lobby"
          value={minSubsPerLobby}
          min={1}
          onChange={setMinSubsPerLobby}
          hint="Additional lobbies are created only after this many subs are available per lobby."
        />
        {submitAttempted && !hasValidSubMin && (
          <p className="text-xs text-[var(--color-text-danger)]">
            Minimum subs per lobby must be greater than 0.
          </p>
        )}

        <NumberStepper
          label="Number of games in event"
          value={gamesToRun}
          min={1}
          onChange={setGamesToRun}
        />
        {submitAttempted && !hasValidGamesToRun && (
          <p className="text-xs text-[var(--color-text-danger)]">
            Number of games must be greater than 0.
          </p>
        )}
      </div>

      <div className="pt-1 border-t border-white/[0.06]">
        <ToggleRow
          label="Registration is open"
          description={
            registrationOpen
              ? "Players can register for this event."
              : "Registration is currently closed."
          }
          checked={registrationOpen}
          onChange={setRegistrationOpen}
        />
      </div>

      {submitError && <p className="text-xs text-[var(--color-text-danger)]">{submitError}</p>}

      <div className="mt-1 flex items-center justify-end gap-2">
        <button
          type="button"
          onClick={onCancel}
          disabled={isSubmitting}
          className="px-3 py-2 rounded-lg text-sm font-medium border border-white/10 bg-white/[0.03] text-[var(--color-text-muted)] hover:text-[var(--color-text-soft)] transition-colors"
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={isSubmitting}
          className="px-3 py-2 rounded-lg text-sm font-medium border border-[var(--color-accent-blue)]/30 bg-[var(--color-accent-blue)]/10 text-[var(--color-accent-blue)]"
        >
          {isSubmitting ? "Creating..." : mode === "create" ? "Create Event" : "Save Settings"}
        </button>
      </div>
      </form>
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
    </>
  );
}
