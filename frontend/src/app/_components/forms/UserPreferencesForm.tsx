"use client";

// Editable profile: region, pronouns, and per-game accounts (user games). Used as /my_account.
import { useState, useEffect, useMemo, useSyncExternalStore, useRef } from "react";
import { useRouter } from "next/navigation";
import Image from "next/image";
import { useForm, useWatch, useFieldArray, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { REGIONS, DISPLAY_NAME_MAX_RUNES, discordAvatarUrl, type Region } from "@/app/_lib/constants";
import { useAuth } from "@/app/_context/AuthContext";
import { Select } from "@/app/_components/Select";
import { SectionDivider } from "@/app/_components/SectionDivider";
import { Field } from "@/app/_components/Field";
import { ToggleRow } from "@/app/_components/ToggleRow";
import { inputCls } from "@/app/_lib/styles";
import { codePointLength } from "@/app/_lib/textInput";
import { peekPostLoginRedirect, consumePostLoginRedirect } from "@/app/_lib/postLoginRedirect";
import { useCancelableFetch } from "@/app/_hooks/useCancelableFetch";
import { extractApiError, fetchGames } from "@/app/_services/games";
import { fetchCurrentUser, fetchCurrentUserGames, updateCurrentUserPreferences, upsertCurrentUserGame, deleteCurrentUserGame } from "@/app/_services/users";
import { Game, UserGame } from "@/app/_types/types";
import { preferencesSchema, PreferencesFormValues } from "./userPreferencesForm/schema";
import { GameCard } from "./userPreferencesForm/GameCard";
import { LeaveProfileSetupSheet } from "./userPreferencesForm/LeaveProfileSetupSheet";
import { useLeaveConfirmGuard } from "./userPreferencesForm/useLeaveConfirmGuard";

export default function UserPreferencesForm() {
  const { user, setUser, isAuthenticated, isLoading: authLoading } = useAuth();
  const router = useRouter();

  const [allGames, setAllGames] = useState<Game[] | null>(null);
  const [userGames, setUserGames] = useState<UserGame[] | null>(null);
  /** Game IDs already persisted; game picker stays editable until first successful save for that row. */
  const [persistedGameIds, setPersistedGameIds] = useState<Set<string>>(() => new Set());
  const [status, setStatus] = useState<"idle" | "saving" | "success" | "error">("idle");
  const [errorMsg, setErrorMsg] = useState("");
  // sessionStorage is client-only; server snapshot is false to avoid hydration mismatch
  const hasPendingEventRedirect = useSyncExternalStore(
    () => () => {},
    () => peekPostLoginRedirect() !== null,
    () => false,
  );
  const { leaveHref, closeLeaveSheet, confirmLeave } = useLeaveConfirmGuard(user?.new_user, hasPendingEventRedirect);

  const {
    register,
    control,
    handleSubmit,
    reset,
    setValue,
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
  /** Only hydrate from the server once per user; setUser after save must not reset the field array. */
  const hydratedUserIdRef = useRef<string | null>(null);

  // Auto-reset the "success" status back to idle after a few seconds (mirrors the event-group
  // page's toast auto-dismiss effect). Keying off `status` — rather than setting an untracked
  // `setTimeout` inline in `onSubmit` — means the cleanup function automatically cancels a
  // pending reset if the component unmounts or the status changes again before it fires.
  useEffect(() => {
    if (status !== "success") return;
    const timer = window.setTimeout(() => setStatus("idle"), 3500);
    return () => window.clearTimeout(timer);
  }, [status]);

  const takenGameIds = useMemo(
    () => watchedGames?.map((g) => g.game_id).filter(Boolean) ?? [],
    [watchedGames]
  );

  // Fetch game list and user's games in parallel once user is ready
  useCancelableFetch({
    fetcher: (signal) => Promise.all([fetchGames(signal), fetchCurrentUserGames(signal)]),
    enabled: !authLoading && isAuthenticated && Boolean(user?.id),
    onSuccess: ([games, ug]) => {
      setAllGames(games);
      setUserGames(ug);
      setPersistedGameIds(new Set(ug.map((g) => g.game_id)));
    },
    onError: (err) => {
      console.error(err);
      setAllGames([]);
      setUserGames([]);
      setPersistedGameIds(new Set());
    },
    deps: [authLoading, isAuthenticated, user?.id],
  });

  // Populate form once both datasets are ready
  useEffect(() => {
    if (!user || userGames === null) return;
    if (hydratedUserIdRef.current === user.id) return;
    hydratedUserIdRef.current = user.id;
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
      const remainingIds = new Set(data.games.map((g) => g.game_id).filter(Boolean));
      const removedIds = [...persistedGameIds].filter((id) => !remainingIds.has(id));
      for (const gameId of removedIds) {
        await deleteCurrentUserGame(gameId);
      }
      // keepValues: reset() otherwise regenerates useFieldArray ids and remounts GameCards.
      reset(data, { keepValues: true });
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
    <LeaveProfileSetupSheet
      isOpen={leaveHref !== null}
      onClose={closeLeaveSheet}
      onConfirmLeave={confirmLeave}
    />
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
              htmlFor="preferences-pronouns"
              error={errors.pronouns?.message}
              hint="Shown on your public profile when display is enabled."
            >
              <input
                {...register("pronouns")}
                id="preferences-pronouns"
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

            <Field label="Region" htmlFor="preferences-region" error={errors.region?.message}>
              <Controller
                control={control}
                name="region"
                render={({ field }) => (
                  <Select
                    inputId="preferences-region"
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
                key={field.game_id || field.id}
                index={index}
                allGames={allGames}
                control={control}
                setValue={setValue}
                errors={errors}
                takenGameIds={takenGameIds}
                persistedGameIds={persistedGameIds}
                onRemove={() => {
                  remove(index);
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
