"use client"

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { useForm, useFieldArray, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import api from "@/app/_lib/axios";
import { User, Game, GameRank, UserGame } from "@/app/_types/types";
import { useAuth } from "@/app/_context/AuthContext";
import { Select } from "@/app/_components/Select";
import { SectionDivider } from "@/app/_components/SectionDivider";
import { Field } from "@/app/_components/Field";
import { ToggleRow } from "@/app/_components/ToggleRow";
import { inputCls } from "@/app/_lib/styles";
import { fetchGames, extractApiError } from "@/app/_services/games";


// ---------------------------------------------------------------------------
// Zod schema
// ---------------------------------------------------------------------------

const REGIONS = ["NA", "EMEA", "APAC"] as const;

const userGameSchema = z.object({
  game_id: z.string().uuid("Please select a game"),
  in_game_name: z.string().min(1, "In-game name is required"),
  current_rank: z.string().min(1, "Current rank is required"),
  peak_rank: z.string().min(1, "Peak rank is required"),
  show_rank: z.boolean(),
  api_permission: z.boolean(),
});

const preferencesSchema = z.object({
  pronouns: z.string().nullable().optional(),
  show_pronouns: z.boolean(),
  region: z.enum(REGIONS).nullable().optional(),
  games: z.array(userGameSchema),
});

type PreferencesFormValues = z.infer<typeof preferencesSchema>;

// ---------------------------------------------------------------------------
// API helpers
// ---------------------------------------------------------------------------

async function fetchRanksForGame(gameId: string): Promise<GameRank[]> {
  const res = await api.get<GameRank[]>(`/games/${gameId}/ranks`);
  return res.data.sort((a, b) => a.order - b.order);
}

async function fetchUserGames(): Promise<UserGame[]> {
  const res = await api.get<UserGame[]>("/users/me/games");
  return res.data;
}

async function saveUserPreferences(data: PreferencesFormValues): Promise<void> {
  try {
    await api.put("/users/me", {
      pronouns: data.pronouns || null,
      show_pronouns: data.show_pronouns,
      region: data.region ?? null,
      games: data.games.map((g) => ({
        game_id: g.game_id,
        in_game_name: g.in_game_name,
        current_rank: g.current_rank ?? null,
        peak_rank: g.peak_rank ?? null,
        show_rank: g.show_rank,
      })),
    });
  } catch (err: unknown) {
    throw new Error(extractApiError(err));
  }
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function discordAvatarUrl(discordId: string, hash: string | null): string {
  if (!hash) return "https://cdn.discordapp.com/embed/avatars/0.png";
  return `https://cdn.discordapp.com/avatars/${discordId}/${hash}.webp?size=80`;
}

// ---------------------------------------------------------------------------
// GameCard — fetches its own ranks when a game is selected
// ---------------------------------------------------------------------------

interface GameCardProps {
  index: number;
  allGames: Game[];
  control: ReturnType<typeof useForm<PreferencesFormValues>>["control"];
  register: ReturnType<typeof useForm<PreferencesFormValues>>["register"];
  errors: ReturnType<typeof useForm<PreferencesFormValues>>["formState"]["errors"];
  watchedGameId: string;
  takenGameIds: string[];
  onRemove: () => void;
  onGameChange: (gameId: string) => void;
}

function GameCard({
  index,
  allGames,
  control,
  register,
  errors,
  watchedGameId,
  takenGameIds,
  onRemove,
  onGameChange,
}: GameCardProps) {
  const [ranks, setRanks] = useState<GameRank[]>([]);
  const [ranksLoading, setRanksLoading] = useState(false);

  useEffect(() => {
    if (!watchedGameId) {
      setRanks([]);
      return;
    }
    setRanksLoading(true);
    fetchRanksForGame(watchedGameId)
      .then(setRanks)
      .catch(() => setRanks([]))
      .finally(() => setRanksLoading(false));
  }, [watchedGameId]);

  const gameErrors = errors.games?.[index];

  return (
    <div className="card rounded-xl p-4 flex flex-col gap-4 relative overflow-hidden">
      <div className="absolute top-0 left-4 right-4 h-px bg-top-edge opacity-20 rounded-full" />

      {/* game picker + remove */}
      <div className="flex items-center gap-2">
        <Controller
          control={control}
          name={`games.${index}.game_id`}
          render={({ field }) => (
            <div className="flex-1">
              <Select
                value={field.value}
                onChange={(v) => {
                  field.onChange(v);
                  onGameChange(v);
                }}
                placeholder="— Select a game —"
                options={allGames
                  .filter((g) => g.id === watchedGameId || !takenGameIds.includes(g.id))
                  .map((g) => ({ value: g.id, label: g.name }))}
              />
            </div>
          )}
        />
        <button
          type="button"
          onClick={onRemove}
          aria-label="Remove game"
          className="flex-shrink-0 w-8 h-8 rounded-lg flex items-center justify-center
                     border border-white/10 bg-white/5 text-[var(--color-text-muted)] text-xs
                     hover:border-[var(--color-text-danger)]/40
                     hover:text-[var(--color-text-danger)]
                     hover:bg-[rgba(255,100,80,0.08)]
                     transition-all duration-150"
        >
          ✕
        </button>
      </div>
      {gameErrors?.game_id && (
        <p className="text-xs text-[var(--color-text-danger)] -mt-2">
          {gameErrors.game_id.message}
        </p>
      )}

      {/* fields — only shown once a game is selected */}
      {!watchedGameId && (
        <div className="flex items-center gap-2.5 px-1 py-0.5">
          <div className="w-1 h-1 rounded-full bg-white/15 flex-shrink-0" />
          <p className="text-xs text-[var(--color-text-faint)]">
            Select a game above to continue
          </p>
        </div>
      )}

      {watchedGameId && ranksLoading && (
        <div className="flex flex-col gap-3 animate-pulse">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div className="h-9 rounded-lg bg-white/5 border border-white/[0.06]" />
            <div className="h-9 rounded-lg bg-white/5 border border-white/[0.06]" />
            <div className="h-9 rounded-lg bg-white/5 border border-white/[0.06]" />
            <div className="h-9 rounded-lg bg-white/5 border border-white/[0.06]" />
          </div>
          <div className="flex flex-col gap-3 pt-3 border-t border-white/[0.06]">
            <div className="h-4 w-40 rounded bg-white/5" />
            <div className="h-4 w-48 rounded bg-white/5" />
          </div>
        </div>
      )}

      {watchedGameId && !ranksLoading && (
        <>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <Field label="In-game name *" error={gameErrors?.in_game_name?.message}>
              <input
                {...register(`games.${index}.in_game_name`)}
                placeholder="YourTag#1234"
                className={inputCls}
              />
            </Field>

            <br />

            {ranks.length > 0 && (
              <>
                <Field label="Current rank *" error={gameErrors?.current_rank?.message}>
                  <Controller
                    control={control}
                    name={`games.${index}.current_rank`}
                    render={({ field }) => (
                      <Select
                        value={field.value ?? ""}
                        onChange={(v) => field.onChange(v || null)}
                        placeholder="— Select rank —"
                        options={ranks.map((r) => ({ value: r.id, label: r.name }))}
                      />
                    )}
                  />
                </Field>

                <Field label="Peak rank *" error={gameErrors?.peak_rank?.message}>
                  <Controller
                    control={control}
                    name={`games.${index}.peak_rank`}
                    render={({ field }) => (
                      <Select
                        value={field.value ?? ""}
                        onChange={(v) => field.onChange(v || null)}
                        placeholder="— Select rank —"
                        options={ranks.map((r) => ({ value: r.id, label: r.name }))}
                      />
                    )}
                  />
                </Field>

                <p className="col-span-full text-xs text-[var(--color-text-muted)] leading-relaxed">
                  If unranked, provide the most recent rank you had or one you believe you would be placed into if you were to play.
                </p>
              </>
            )}
          </div>

          <div className="flex flex-col gap-3 pt-3 border-t border-white/[0.06]">
            <Controller
              control={control}
              name={`games.${index}.show_rank`}
              render={({ field }) => (
                <ToggleRow
                  label="Show rank publicly"
                  description={
                    field.value
                      ? "Visible to everyone in games you register for"
                      : "Visible only to the host of games you register for"
                  }
                  checked={field.value}
                  onChange={field.onChange}
                />
              )}
            />
          </div>
        </>
      )}
    </div>
  );
}

// ---------------------------------------------------------------------------
// Main form
// ---------------------------------------------------------------------------

export default function UserPreferencesForm() {
  const { user, setUser } = useAuth();
  const router = useRouter();

  const [allGames, setAllGames] = useState<Game[] | null>(null);
  const [userGames, setUserGames] = useState<UserGame[] | null>(null);
  const [status, setStatus] = useState<"idle" | "saving" | "success" | "error">("idle");
  const [errorMsg, setErrorMsg] = useState("");

  const {
    register,
    control,
    handleSubmit,
    watch,
    reset,
    setValue,
    formState: { errors, isDirty, isSubmitting },
  } = useForm<PreferencesFormValues>({
    resolver: zodResolver(preferencesSchema),
    defaultValues: { pronouns: "", show_pronouns: false, region: null, games: [] },
  });

  const { fields, append, remove } = useFieldArray({ control, name: "games" });
  const watchedGames = watch("games");

  // Fetch game list and user's games in parallel once user is ready
  useEffect(() => {
    if (!user) return;
    Promise.all([fetchGames(), fetchUserGames()])
      .then(([games, ug]) => {
        setAllGames(games);
        setUserGames(ug);
      })
      .catch((err) => {
        console.error(err);
        setAllGames([]);
        setUserGames([]);
      });
  }, [user]);

  // Populate form once both datasets are ready
  useEffect(() => {
    if (!user || userGames === null) return;
    reset({
      pronouns: user.pronouns ?? "",
      show_pronouns: user.show_pronouns,
      region: (user.region as (typeof REGIONS)[number]) ?? null,
      games: userGames.map((ug) => ({
        game_id: ug.game_id,
        in_game_name: ug.in_game_name ?? "",
        current_rank: ug.current_rank ?? "",
        peak_rank: ug.peak_rank ?? "",
        show_rank: ug.show_rank,
        api_permission: false,
      })),
    });
  }, [user, userGames, reset]);

  const onSubmit = async (data: PreferencesFormValues) => {
    setStatus("saving");
    setErrorMsg("");
    const wasNewUser = user?.new_user ?? false;
    try {
      await saveUserPreferences(data);
      reset(data);
      const res = await api.get<User>("/users/me");
      if (res.data) {
        setUser(res.data);
        if (wasNewUser && !res.data.new_user) {
          router.push("/events");
          return;
        }
      }
      setStatus("success");
      setTimeout(() => setStatus("idle"), 3500);
    } catch (err) {
      setErrorMsg(err instanceof Error ? err.message : "Something went wrong. Please try again.");
      setStatus("error");
    }
  };

  const addGame = () =>
    append({
      game_id: "",
      in_game_name: "",
      current_rank: "",
      peak_rank: "",
      show_rank: true,
      api_permission: false,
    });

  const handleGameChange = (index: number) => {
    setValue(`games.${index}.current_rank`, "");
    setValue(`games.${index}.peak_rank`, "");
  };

  if (!user || allGames === null || userGames === null) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <p className="text-sm text-[var(--color-text-muted)]">Loading…</p>
      </div>
    );
  }

  const avatarUrl = discordAvatarUrl(user.discord_id, user.image_url);

  return (
    <div className="flex-1 flex flex-col items-center py-10 px-4">
      <div
        className="w-full max-w-xl"
        style={{ animation: "var(--animate-rise)" }}
      >
        {/* page header */}
        <div className="mb-8" style={{ animation: "var(--animate-rise-1)" }}>
          <h1 className="text-2xl font-semibold text-[var(--color-text)] tracking-tight">
            Settings
          </h1>
          <p className="mt-1 text-sm text-[var(--color-text-muted)]">
            Manage your profile, pronouns, and game accounts.
          </p>
        </div>

        <form
          onSubmit={handleSubmit(onSubmit)}
          noValidate
          className="flex flex-col gap-5"
          style={{ animation: "var(--animate-rise-2)" }}
        >
          {/* ── Discord identity ─────────────────────────────── */}
          <div className="card rounded-xl p-4 flex items-center gap-4 relative overflow-hidden">
            <div className="absolute top-0 left-4 right-4 h-px bg-top-edge opacity-20 rounded-full" />
            <img
              src={avatarUrl}
              alt={user.discord_name ?? "Discord avatar"}
              width={52}
              height={52}
              className="rounded-full flex-shrink-0 ring-1 ring-white/10"
            />
            <div className="min-w-0">
              <p className="font-medium text-[var(--color-text)] truncate">
                {user.discord_name ?? "Unknown"}
              </p>
              <p className="text-xs text-[var(--color-text-muted)] mt-0.5 flex items-center gap-1.5">
                <span
                  className="inline-block w-1.5 h-1.5 rounded-full flex-shrink-0"
                  style={{ background: "var(--color-discord)" }}
                />
                Connected via Discord · read-only
              </p>
            </div>
          </div>

          {/* ── Preferences ──────────────────────────────────── */}
          <div className="card rounded-xl p-5 flex flex-col gap-5 relative overflow-hidden">
            <div className="absolute top-0 left-4 right-4 h-px bg-top-edge opacity-20 rounded-full" />
            <SectionDivider title="Preferences" />

            <Field
              label="Pronouns"
              error={errors.pronouns?.message}
              hint="Shown on your public profile when display is enabled."
            >
              <input
                {...register("pronouns")}
                placeholder="e.g. they/them"
                className={inputCls}
              />
            </Field>

            <Controller
              control={control}
              name="show_pronouns"
              render={({ field }) => (
                <ToggleRow
                  label="Display pronouns publicly"
                  description={
                    field.value
                      ? "Visible to everyone in games you register for"
                      : "Visible only to the host of games you register for"
                  }
                  checked={field.value}
                  onChange={field.onChange}
                />
              )}
            />

            <Field label="Region" error={errors.region?.message}>
              <Controller
                control={control}
                name="region"
                render={({ field }) => (
                  <Select
                    value={field.value ?? ""}
                    onChange={(v) => field.onChange(v || null)}
                    placeholder="— No preference —"
                    options={REGIONS.map((r) => ({ value: r, label: r }))}
                  />
                )}
              />
            </Field>
          </div>

          {/* ── Games ────────────────────────────────────────── */}
          <div className="flex flex-col gap-3">
            <SectionDivider title="Games" />

            {fields.length === 0 && (
              <div className="text-center py-8 rounded-xl border border-dashed border-white/10 text-sm text-[var(--color-text-muted)]">
                No games added yet. Click "Add game" to get started.
              </div>
            )}

            {fields.map((field, index) => (
              <GameCard
                key={field.id}
                index={index}
                allGames={allGames}
                control={control}
                register={register}
                errors={errors}
                watchedGameId={watchedGames?.[index]?.game_id ?? ""}
                takenGameIds={watchedGames?.map((g) => g.game_id).filter(Boolean) ?? []}
                onRemove={() => remove(index)}
                onGameChange={() => handleGameChange(index)}
              />
            ))}

            <button
              type="button"
              onClick={addGame}
              className="self-start flex items-center gap-2 px-4 py-2 rounded-lg text-sm
                         border border-white/10 bg-white/[0.03] text-[var(--color-text-soft)]
                         hover:bg-white/[0.07] hover:border-white/20 hover:text-[var(--color-text)]
                         transition-all duration-150"
            >
              <span className="text-base leading-none text-[var(--color-accent-blue)]">+</span>
              Add game
            </button>
          </div>

          {/* ── Save footer ───────────────────────────────────── */}
          <div className="flex items-center justify-between gap-4 pt-1">
            <div className="text-sm min-h-5">
              {status === "success" && (
                <span className="text-emerald-400">Changes saved.</span>
              )}
              {status === "error" && (
                <span className="text-[var(--color-text-danger)]">{errorMsg}</span>
              )}
            </div>

            <button
              type="submit"
              disabled={!isDirty || isSubmitting || status === "saving"}
              className="relative overflow-hidden px-5 py-2 rounded-lg text-sm font-medium
                         border border-white/10 bg-white/[0.04] text-[var(--color-text)]
                         hover:bg-white/[0.09] hover:border-[var(--color-accent-blue)]/40
                         disabled:opacity-40 disabled:cursor-not-allowed
                         focus-visible:outline-none
                         focus-visible:ring-2 focus-visible:ring-[var(--color-accent-blue)]/40
                         transition-all duration-150"
            >
              {isDirty && status !== "saving" && (
                <span className="absolute top-0 left-0 right-0 h-px bg-top-edge opacity-40" />
              )}
              {status === "saving" ? "Saving…" : "Save changes"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
