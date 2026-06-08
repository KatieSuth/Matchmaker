// DTOs mirroring the Go API’s JSON field names. Nullable fields use null where the server sends null.

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
  game_name?: string;
  current_rank: string | null;
  current_rank_name?: string;
  peak_rank: string | null;
  peak_rank_name?: string;
  show_rank: boolean;
  created_at: string;
  updated_at: string;
}

export interface EventRegistration {
  event_id: string;
  user_id: string;
  discord_name: string;
  pronouns: string;
  current_rank_name: string;
  peak_rank_name?: string;
  can_substitute: boolean;
  can_lobby_host: boolean;
  duo_request: string | null;
  created_at: string;
  updated_at: string;
}

export interface LobbyPlayer {
  user_id: string;
  discord_name: string;
  pronouns: string;
  current_rank_name: string;
  current_rank_order: number;
  peak_rank_order: number;
  can_substitute: boolean;
  can_lobby_host: boolean;
  created_at: string;
  updated_at: string;
}

export interface EventTeam {
  team_number: number;
  players: LobbyPlayer[];
}

export interface EventLobby {
  id: string;
  host_id: string | null;
  fairness_warning: boolean;
  teams: EventTeam[];
  subs: LobbyPlayer[];
}

export interface EventGroupEvent {
  id: string;
  start_time: string;
  game_mode_id: string;
  game_mode_name: string;
  team_size: number;
  registered_count: number;
  lobbies_count: number;
  player_registered: boolean;
  registrations: EventRegistration[];
  lobbies: EventLobby[];
  unplaced: EventRegistration[];
}

export type EventSortLogic = "balanced" | "ranked";

export interface EventGroupDetail {
  id: string;
  owner_id: string;
  owner_name: string;
  /** Non-empty when the host has enabled public pronouns; otherwise "". */
  owner_pronouns: string;
  game_mode_id: string;
  game_mode_name: string;
  game_id: string;
  game_name: string;
  team_size: number;
  sub_min: number;
  registration_open: boolean;
  region: string;
  sort_logic: EventSortLogic;
  created_at: string;
  updated_at: string;
  events: EventGroupEvent[];
}

export interface GroupRegistrationEventInput {
  event_id: string;
  can_substitute: boolean;
  can_lobby_host: boolean;
}

export interface UpsertGroupRegistrationRequest {
  duo_request: string;
  events: GroupRegistrationEventInput[];
}

export interface Event {
  id: string;
  game_name: string;
  game_mode: string | null;
  event_date: string;
  host_name: string;
  host_id: string;
  registered_count: number;
  registration_open: boolean;
}

export interface EventsPage {
  event_groups: Event[];
  next_cursor: string | null;
  has_more: boolean;
}

export interface CompleteAuthResponse {
  access_token: string;
}