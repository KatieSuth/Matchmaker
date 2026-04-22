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

export interface EventRegistration {
  event_id: string;
  user_id: string;
  discord_name: string;
  pronouns: string;
  current_rank_name: string;
  can_substitute: boolean;
  can_lobby_host: boolean;
  duo_request: string | null;
  created_at: string;
  updated_at: string;
}

export interface EventGroupEvent {
  id: string;
  start_time: string;
  registered_count: number;
  lobbies_count: number;
  player_registered: boolean;
  registrations: EventRegistration[];
}

export interface EventGroupDetail {
  id: string;
  owner_id: string;
  owner_name: string;
  game_mode_id: string;
  game_mode_name: string;
  game_id: string;
  game_name: string;
  team_size: number;
  sub_min: number;
  registration_open: boolean;
  region: string;
  created_at: string;
  updated_at: string;
  events: EventGroupEvent[];
}