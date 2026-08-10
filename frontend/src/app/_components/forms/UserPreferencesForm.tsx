"use client";

// Editable profile: region, pronouns, and per-game accounts (user games). Used as /my_account.
import { useState, useEffect, useMemo, useSyncExternalStore, useCallback } from "react";
import { useRouter } from "next/navigation";
import Image from 'next/image';
import { useForm, useWatch, useFieldArray, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { REGIONS, DISPLAY_NAME_MAX_RUNES, discordAvatarUrl, type Region } from "@/app/_lib/constants";
import { User, Game, GameRank, UserGame } from "@/app/_types/types";
import { useAuth } from "@/app/_context/AuthContext";
import { Select } from "@/app/_components/Select";
import { SectionDivider } from "@/app/_components/SectionDivider";
import { ResponsiveSheet } from "@/app/_components/ResponsiveSheet";
import { UserGameEditor } from "@/app/_components/forms/UserGameEditor";
import { Field } from "@/app/_components/Field";
import { ToggleRow } from "@/app/_components/ToggleRow";
import { inputCls } from "@/app/_lib/styles";
import { optionalFreeTextSchema, codePointLength } from "@/app/_lib/textInput";
import { consumePostLoginRedirect, peekPostLoginRedirect } from "@/app/_lib/postLoginRedirect";
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
  display_name: optionalFreeTextSchema(DISPLAY_NAME_MAX_RUNES),
  pronouns: z.string().nullable().optional(),
  show_pronouns: z.boolean(),
  region: z.enum(REGIONS).nullable().optional(),
  games: z.array(userGameSchema),
});

type PreferencesFormValues = z.infer<typeof preferencesSchema>;

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
  /** Game IDs already saved to the server for this user (dropdown locked for those rows). */
  persistedGameIds: Set<string>;
  onRemove: () => void;
}

function GameCard({
  index,
  allGames,
  control,
  setValue,
  errors,
  takenGameIds,
  persistedGameIds,
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
  const lockGameSelect =
    Boolean(watchedGameId) && persistedGameIds.has(watchedGameId);

  return (
    <UserGameEditor
      lockGameSelect={lockGameSelect}
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
  /** Game IDs already persisted; game picker stays editable until first successful save for that row. */
  const [persistedGameIds, setPersistedGameIds] = useState<Set<string>>(() => new Set());
  const [status, setStatus] = useState<"idle" | "saving" | "success" | "error">("idle");
  const [errorMsg, setErrorMsg] = useState("");
  /** Non-null while the leave-confirm sheet is open; value is the blocked in-app href. */
  const [leaveHref, setLeaveHref] = useState<string | null>(null);
  // sessionStorage is client-only; server snapshot is false to avoid hydration mismatch
  const hasPendingEventRedirect = useSyncExternalStore(
    () => () => {},
    () => peekPostLoginRedirect() !== null,
    () => false,
  );

  const {
    register,
    control,
    handleSubmit,
    reset,
    setValue,
    getValues,
    formState: { errors, isDirty, isSubmitting },
  } = useForm<PreferencesFormValues>({
    resolver: zodResolver(preferencesSchema),
    defaultValues: { display_name: "", pronouns: "", show_pronouns: false, region: null, games: [] },
  });

  const { fields, append, remove } = useFieldArray({ control, name: "games" });
  const watchedGames = useWatch({
    control,
    name: "games",
  });
  const watchedDisplayName = useWatch({ control, name: "display_name" }) ?? "";
  const displayNameCodePoints = codePointLength(watchedDisplayName);
  const displayNameOverLimit = displayNameCodePoints > DISPLAY_NAME_MAX_RUNES;

  const takenGameIds = useMemo(() => 
    watchedGames?.map((g: any) => g.game_id).filter(Boolean) ?? [], 
  [watchedGames]);

  const closeLeaveSheet = useCallback(() => setLeaveHref(null), []);

  const confirmLeave = useCallback(() => {
    if (!leaveHref) return;
    const href = leaveHref;
    setLeaveHref(null);
    consumePostLoginRedirect();
    router.push(href);
  }, [leaveHref, router]);

  // Warn before abandoning a pending event deep-link via in-app navigation
  useEffect(() => {
    if (!user?.new_user || !hasPendingEventRedirect) return;

    const onClick = (e: MouseEvent) => {
      if (e.defaultPrevented || e.button !== 0) return;
      if (e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return;

      const target = e.target;
      if (!(target instanceof Element)) return;
      const anchor = target.closest("a[href]");
      if (!(anchor instanceof HTMLAnchorElement)) return;

      const hrefAttr = anchor.getAttribute("href");
      if (!hrefAttr || hrefAttr.startsWith("#")) return;

      let url: URL;
      try {
        url = new URL(hrefAttr, window.location.origin);
      } catch {
        return;
      }
      if (url.origin !== window.location.origin) return;
      if (url.pathname === "/my_account") return;

      e.preventDefault();
      e.stopPropagation();
      setLeaveHref(url.pathname + url.search);
    };

    document.addEventListener("click", onClick, true);
    return () => document.removeEventListener("click", onClick, true);
  }, [user?.new_user, hasPendingEventRedirect]);

  // Fetch game list and user's games in parallel once user is ready
  useEffect(() => {
    if (authLoading || !isAuthenticated || !user) return;
    const ac = new AbortController();
    const { signal } = ac;
    Promise.all([fetchGames(signal), fetchCurrentUserGames(signal)])
      .then(([games, ug]) => {
        setAllGames(games);
        setUserGames(ug);
        setPersistedGameIds(new Set(ug.map((g) => g.game_id)));
      })
      .catch((err) => {
        if ((err as { code?: string; name?: string })?.code === "ERR_CANCELED" || (err as { name?: string })?.name === "CanceledError") {
          return;
        }
        console.error(err);
        setAllGames([]);
        setUserGames([]);
        setPersistedGameIds(new Set());
      });
    return () => {
      ac.abort();
    };
  }, [authLoading, isAuthenticated, user]);

  // Populate form once both datasets are ready
  useEffect(() => {
    if (!user || userGames === null) return;
    reset({
      display_name: user.display_name ?? "",
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
        display_name: data.display_name.trim() || null,
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
      setPersistedGameIds(new Set(data.games.map((g) => g.game_id).filter(Boolean)));
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

  const leaveSheet = (
    <ResponsiveSheet
      isOpen={leaveHref !== null}
      onClose={closeLeaveSheet}
      title="Leave profile setup?"
    >
      <div className="flex flex-col gap-4">
        <p className="text-sm text-[var(--color-text-soft)]">
          You haven&apos;t finished setting up your profile yet. If you leave now, you won&apos;t be taken to the event you came here for.
        </p>
        <div className="flex justify-end gap-2">
          <button
            type="button"
            onClick={confirmLeave}
            className="px-3 py-2 rounded-lg text-sm font-medium border border-[var(--color-text-danger)]/40 bg-[var(--color-text-danger)]/10 text-[var(--color-text-danger)] hover:bg-[var(--color-text-danger)]/20 transition-colors"
          >
            Leave
          </button>
          <button
            type="button"
            onClick={closeLeaveSheet}
            className="px-3 py-2 rounded-lg text-sm font-medium border border-[var(--color-accent-blue)]/30 bg-[var(--color-accent-blue)]/10 text-[var(--color-accent-blue)]"
          >
            Stay
          </button>
        </div>
      </div>
    </ResponsiveSheet>
  );

  if (authLoading || !user || allGames === null || userGames === null) {
    return (
      <>
        <div className="flex-1 flex items-center justify-center">
          <p className="text-sm text-[var(--color-text-muted)]">Loading…</p>
        </div>
        {leaveSheet}
      </>
    );
  }

  const avatarUrl = discordAvatarUrl(user.discord_id, user.image_url, 80);

  return (
    <div className="flex-1 flex flex-col items-center py-10 px-4">
      <div
        className="w-full max-w-xl"
        style={{ animation: "var(--animate-rise)" }}
      >
        {/* page header */}
        <div className="mb-8" style={{ animation: "var(--animate-rise-1)" }}>
          <h1 className="text-2xl font-semibold text-[var(--color-text)] tracking-tight">
            {user.new_user ? "Welcome to Matchmaker!" : "Settings"}
          </h1>
          <p className="mt-1 text-sm text-[var(--color-text-muted)]">
            {user.new_user
              ? hasPendingEventRedirect
                ? "Let's get your profile set up. Fill in the info below, click \"Save Profile\", and then you can register for the event you came here for. Don't forget to add the games you want to play!"
                : "Let's get your profile set up. Please fill in the info below. Don't forget to add the games you want to play!"
              : "Manage your profile, pronouns, and game accounts."}
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

            <div className="flex flex-col gap-1.5">
              <label
                htmlFor="preferences-display-name"
                className={`text-xs font-medium tracking-wide ${
                  displayNameOverLimit ? "text-[var(--color-text-danger)]" : "text-[var(--color-text-soft)]"
                }`}
              >
                Display Name ({displayNameCodePoints}/{DISPLAY_NAME_MAX_RUNES})
              </label>
              <Controller
                name="display_name"
                control={control}
                render={({ field }) => (
                  <input
                    {...field}
                    id="preferences-display-name"
                    type="text"
                    autoComplete="off"
                    placeholder="Optional public name"
                    className={inputCls}
                  />
                )}
              />
              {errors.display_name && (
                <p className="text-xs text-[var(--color-text-danger)]">{errors.display_name.message}</p>
              )}
            </div>

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
                persistedGameIds={persistedGameIds}
                onRemove={() => {
                  const gid = getValues(`games.${index}.game_id`);
                  remove(index);
                  if (gid) {
                    setPersistedGameIds((prev) => {
                      const next = new Set(prev);
                      next.delete(gid);
                      return next;
                    });
                  }
                }}
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
              {status === "saving" ? "Saving…" : user.new_user ? "Save Profile" : "Save changes"}
            </button>
          </div>
        </form>
      </div>

      {leaveSheet}
    </div>
  );
}
