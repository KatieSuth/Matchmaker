// Rank/skill math for the event group detail page (team averages, nearest-rank lookup).
import { EMPTY_VALUE } from "@/app/_lib/constants";
import { GameRank, LobbyPlayer } from "@/app/_types/types";

/** Per-player skill value used for team averages — mirrors backend stored avg_rank_order. */
export function playerSkillOrder(player: LobbyPlayer): number | null {
  if (player.avg_rank_order <= 0) {
    return null;
  }
  return player.avg_rank_order;
}

/** Maps a numeric skill value to the closest game rank name by order distance. */
export function nearestRankName(ranks: GameRank[], skillOrder: number): string {
  if (ranks.length === 0) return EMPTY_VALUE;
  let nearest = ranks[0];
  let bestDistance = Math.abs(nearest.order - skillOrder);
  for (const rank of ranks) {
    const distance = Math.abs(rank.order - skillOrder);
    if (distance < bestDistance) {
      bestDistance = distance;
      nearest = rank;
    }
  }
  return nearest.name;
}

/** Returns the nearest rank name for a team's mean skill, or EMPTY_VALUE when no rank data exists. */
export function teamAverageRankLabel(players: LobbyPlayer[], ranks: GameRank[]): string {
  const skillOrders = players
    .map(playerSkillOrder)
    .filter((order): order is number => order !== null);
  if (skillOrders.length === 0) {
    return EMPTY_VALUE;
  }
  const average =
    skillOrders.reduce((sum, order) => sum + order, 0) / skillOrders.length;
  return nearestRankName(ranks, average);
}
