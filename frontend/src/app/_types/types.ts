export interface User {
    id: string;
    discord_id: string;
    discord_name: string | null;
    image_url: string | null;
    pronouns: string | null
    show_pronouns: boolean
    region: string | null
    new_user: boolean
}

export interface Game {
  id: string;
  name: string;
  owner_id: string | null;
  created_at: string;
  updated_at: string;
}

export interface GameMode {
  id: string;
  game_id: string;
  name: string;
  team_size: number;
  owner_id: string | null;
  duration: number;
  created_at: string;
  updated_at: string;
}

export interface GameRank {
  id: string;
  game_id: string | null;
  name: string;
  order: number;
  created_at: string;
  updated_at: string;
}

export interface UserGame {
  user_id: string;
  game_id: string;
  in_game_name: string;
  region: string | null;
  current_rank: string | null;
  peak_rank: string | null;
  show_rank: boolean;
  api_permission: boolean;
  api_links_id: string | null;
  created_at: string;
  updated_at: string;
}