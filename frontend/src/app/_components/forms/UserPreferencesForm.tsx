"use client";

// Editable profile: region, pronouns, and per-game accounts (user games). Used as /my_account.
import { useState, useEffect, useMemo } from "react";
import { useRouter } from "next/navigation";
import Image from 'next/image';
import { useForm, useWatch, useFieldArray, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { REGIONS, type Region } from "@/app/_lib/constants";
import { User, Game, GameRank, UserGame } from "@/app/_types/types";
import { useAuth } from "@/app/_context/AuthContext";
import { Select } from "@/app/_components/Select";
import { SectionDivider } from "@/app/_components/SectionDivider";
import { UserGameEditor } from "@/app/_components/forms/UserGameEditor";
import { Field } from "@/app/_components/Field";
import { ToggleRow } from "@/app/_components/ToggleRow";
import { inputCls } from "@/app/_lib/styles";
import { consumePostLoginRedirect } from "@/app/_lib/postLoginRedirect";
import { extractApiError, fetchGameRanks, fetchGames } from "@/app/_services/games";
import { fetchCurrentUser, fetchCurrentUserGames, updateCurrentUserPreferences, upsertCurrentUserGame } from "@/app/_services/users";


// ---------------------------------------------------------------------------
// Zod schema
// ---------------------------------------------------------------------------

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
  setValue: ReturnType<typeof useForm<PreferencesFormValues>>["setValue"];
  errors: ReturnType<typeof useForm<PreferencesFormValues>>["formState"]["errors"];
  takenGameIds: string[];
  onRemove: () => void;
}

function GameCard({
  index,
  allGames,
  control,
  setValue,
  errors,
  takenGameIds,
  onRemove,
}: GameCardProps) {
  const [ranks, setRanks] = useState<GameRank[]>([]);
  const [ranksLoading, setRanksLoading] = useState(false);
  
  const watchedGameId = useWatch({ control, name: `games.${index}.game_id` });
  const watchedValue = useWatch({ control, name: `games.${index}` });

  useEffect(() => {
    let ignore = false;
    const ac = new AbortController();

    const startSync = async() => {
      if (!watchedGameId) {
        setRanks(() => []);
        return;
      }

      setRanksLoading(true);
      try {
        const data = await fetchGameRanks(watchedGameId, ac.signal);
        if (!ignore) setRanks(data);
      } catch (err) {
        if ((err as { code?: string; name?: string })?.code === "ERR_CANCELED" || (err as { name?: string })?.name === "CanceledError") {
          return;
        }
        if (!ignore) setRanks([]);
      } finally {
        if (!ignore) setRanksLoading(false);
      }
    }
    
    startSync();

    return () => {
      ignore = true;
      ac.abort();
    }
  }, [watchedGameId]);

  const gameErrors = errors.games?.[index];

  return (
    <UserGameEditor
      value={{
        game_id: watchedValue?.game_id ?? "",
        in_game_name: watchedValue?.in_game_name ?? "",
        current_rank: watchedValue?.current_rank ?? "",
        peak_rank: watchedValue?.peak_rank ?? "",
        show_rank: watchedValue?.show_rank ?? false,
      }}
      allGames={allGames}
      takenGameIds={takenGameIds}
      ranks={ranks}
      ranksLoading={ranksLoading}
      errors={{
        game_id: gameErrors?.game_id?.message,
        in_game_name: gameErrors?.in_game_name?.message,
        current_rank: gameErrors?.current_rank?.message,
        peak_rank: gameErrors?.peak_rank?.message,
      }}
      onChange={(next) => {
        setValue(`games.${index}.game_id`, next.game_id, { shouldDirty: true, shouldValidate: true });
        setValue(`games.${index}.in_game_name`, next.in_game_name, { shouldDirty: true, shouldValidate: true });
        setValue(`games.${index}.current_rank`, next.current_rank, { shouldDirty: true, shouldValidate: true });
        setValue(`games.${index}.peak_rank`, next.peak_rank, { shouldDirty: true, shouldValidate: true });
        setValue(`games.${index}.show_rank`, next.show_rank, { shouldDirty: true, shouldValidate: true });
      }}
      onRemove={onRemove}
    />
  );
}

// ---------------------------------------------------------------------------
// Main form
// ---------------------------------------------------------------------------

export default function UserPreferencesForm() {
  const { user, setUser, isAuthenticated, isLoading: authLoading } = useAuth();
  const router = useRouter();

  const [allGames, setAllGames] = useState<Game[] | null>(null);
  const [userGames, setUserGames] = useState<UserGame[] | null>(null);
  const [status, setStatus] = useState<"idle" | "saving" | "success" | "error">("idle");
  const [errorMsg, setErrorMsg] = useState("");

  const {
    register,
    control,
    handleSubmit,
    reset,
    setValue,
    formState: { errors, isDirty, isSubmitting },
  } = useForm<PreferencesFormValues>({
    resolver: zodResolver(preferencesSchema),
    defaultValues: { pronouns: "", show_pronouns: false, region: null, games: [] },
  });

  const { fields, append, remove } = useFieldArray({ control, name: "games" });
  const watchedGames = useWatch({
    control,
    name: "games",
  })
  
  const takenGameIds = useMemo(() => 
    watchedGames?.map((g: any) => g.game_id).filter(Boolean) ?? [], 
  [watchedGames]);

  // Fetch game list and user's games in parallel once user is ready
  useEffect(() => {
    if (authLoading || !isAuthenticated || !user) return;
    const ac = new AbortController();
    const { signal } = ac;
    Promise.all([fetchGames(signal), fetchCurrentUserGames(signal)])
      .then(([games, ug]) => {
        setAllGames(games);
        setUserGames(ug);
      })
      .catch((err) => {
        if ((err as { code?: string; name?: string })?.code === "ERR_CANCELED" || (err as { name?: string })?.name === "CanceledError") {
          return;
        }
        console.error(err);
        setAllGames([]);
        setUserGames([]);
      });
    return () => {
      ac.abort();
    };
  }, [authLoading, isAuthenticated, user]);

  // Populate form once both datasets are ready
  useEffect(() => {
    if (!user || userGames === null) return;
    reset({
      pronouns: user.pronouns ?? "",
      show_pronouns: user.show_pronouns,
      region: (user.region as Region) ?? null,
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
      await updateCurrentUserPreferences({
        pronouns: data.pronouns || null,
        show_pronouns: data.show_pronouns,
        region: data.region ?? null,
        games: [],
      });
      for (const game of data.games) {
        await upsertCurrentUserGame(game.game_id, {
          in_game_name: game.in_game_name,
          current_rank: game.current_rank ?? null,
          peak_rank: game.peak_rank ?? null,
          show_rank: game.show_rank,
        });
      }
      reset(data);
      const resolvedUser = await fetchCurrentUser();
      if (resolvedUser) {
        setUser(resolvedUser);
        if (wasNewUser && !resolvedUser.new_user) {
          const next = consumePostLoginRedirect();
          router.push(next ?? "/my_events");
          return;
        }
      }
      setStatus("success");
      setTimeout(() => setStatus("idle"), 3500);
    } catch (err) {
      setErrorMsg(extractApiError(err));
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

  if (authLoading || !user || allGames === null || userGames === null) {
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
            <Image
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
                {"No games added yet. Click \"Add game\" to get started."}
              </div>
            )}

            {fields.map((field, index) => (
              <GameCard
                key={field.id}
                index={index}
                allGames={allGames}
                control={control}
                setValue={setValue}
                errors={errors}
                takenGameIds={takenGameIds}
                onRemove={() => remove(index)}
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
