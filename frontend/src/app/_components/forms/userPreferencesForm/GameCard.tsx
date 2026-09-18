"use client";

// One row in the "Games" section: fetches ranks for the currently-selected game and renders
// a UserGameEditor bound to the parent form's `games.${index}` field.
import { useEffect, useState } from "react";
import { Control, FieldErrors, UseFormSetValue, useWatch } from "react-hook-form";
import { Game, GameRank } from "@/app/_types/types";
import { UserGameEditor } from "@/app/_components/forms/UserGameEditor";
import { isCanceledError } from "@/app/_lib/http";
import { fetchGameRanks } from "@/app/_services/games";
import { PreferencesFormValues } from "./schema";

interface GameCardProps {
  index: number;
  allGames: Game[];
  control: Control<PreferencesFormValues>;
  setValue: UseFormSetValue<PreferencesFormValues>;
  errors: FieldErrors<PreferencesFormValues>;
  takenGameIds: string[];
  /** Game IDs already saved to the server for this user (dropdown locked for those rows). */
  persistedGameIds: Set<string>;
  onRemove: () => void;
}

export function GameCard({
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

    const startSync = async () => {
      if (!watchedGameId) {
        setRanks([]);
        return;
      }

      setRanksLoading(true);
      try {
        const data = await fetchGameRanks(watchedGameId, ac.signal);
        if (!ignore) setRanks(data);
      } catch (err) {
        if (isCanceledError(err)) return;
        if (!ignore) setRanks([]);
      } finally {
        if (!ignore) setRanksLoading(false);
      }
    };

    startSync();

    return () => {
      ignore = true;
      ac.abort();
    };
  }, [watchedGameId]);

  const gameErrors = errors.games?.[index];
  const lockGameSelect = Boolean(watchedGameId) && persistedGameIds.has(watchedGameId);

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
