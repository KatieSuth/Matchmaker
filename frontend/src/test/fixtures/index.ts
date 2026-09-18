// Factory functions for the DTOs in `@/app/_types/types.ts`, used across MSW handlers and
// component/hook/page tests. Each factory returns a fully-populated, realistic default object;
// pass a partial override for the fields a given test cares about instead of hand-rolling the
// whole shape. Keep these in sync with `_types/types.ts` when the backend contract changes.
import {
  CompleteAuthResponse,
  CreateTeamsResponse,
  DiscordGuild,
  Event,
  EventGroupDetail,
  EventGroupEvent,
  EventLobby,
  EventRegistration,
  EventsPage,
  EventTeam,
  Game,
  GameMode,
  GameRank,
  LobbyPlayer,
  User,
  UserGame,
} from "@/app/_types/types";

let idCounter = 0;
/** Deterministic-but-unique id generator so fixtures never collide within a test file. */
function nextId(prefix: string): string {
  idCounter += 1;
  return `${prefix}-${idCounter}`;
}

export function buildUser(overrides: Partial<User> = {}): User {
  return {
    id: nextId("user"),
    discord_id: "111111111111111111",
    discord_name: "TestUser#0001",
    image_url: null,
    display_name: "Test User",
    pronouns: "they/them",
    show_pronouns: true,
    region: "AMER",
    new_user: false,
    ...overrides,
  };
}

export function buildGame(overrides: Partial<Game> = {}): Game {
  return {
    id: nextId("game"),
    name: "Test Game",
    owner_id: null,
    join_link_base: null,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

export function buildGameMode(overrides: Partial<GameMode> = {}): GameMode {
  return {
    id: nextId("game-mode"),
    game_id: nextId("game"),
    name: "5v5",
    team_size: 5,
    owner_id: null,
    duration: 30,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

export function buildGameRank(overrides: Partial<GameRank> = {}): GameRank {
  return {
    id: nextId("rank"),
    game_id: nextId("game"),
    name: "Gold",
    order: 1,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

export function buildUserGame(overrides: Partial<UserGame> = {}): UserGame {
  return {
    user_id: nextId("user"),
    game_id: nextId("game"),
    in_game_name: "TestIGN",
    game_name: "Test Game",
    current_rank: nextId("rank"),
    current_rank_name: "Gold",
    peak_rank: nextId("rank"),
    peak_rank_name: "Platinum",
    avg_rank: null,
    avg_rank_name: undefined,
    show_rank: true,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

export function buildEventRegistration(overrides: Partial<EventRegistration> = {}): EventRegistration {
  return {
    event_id: nextId("event"),
    user_id: nextId("user"),
    discord_name: "TestUser#0001",
    display_name: "Test User",
    in_game_name: "TestIGN",
    pronouns: "they/them",
    current_rank_name: "Gold",
    peak_rank_name: "Platinum",
    avg_rank_name: "Gold",
    can_substitute: true,
    can_lobby_host: false,
    duo_request: null,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

export function buildLobbyPlayer(overrides: Partial<LobbyPlayer> = {}): LobbyPlayer {
  return {
    user_id: nextId("user"),
    discord_name: "TestUser#0001",
    display_name: "Test User",
    in_game_name: "TestIGN",
    pronouns: "they/them",
    current_rank_name: "Gold",
    current_rank_order: 1,
    peak_rank_name: "Platinum",
    peak_rank_order: 2,
    avg_rank_name: "Gold",
    avg_rank_order: 1,
    can_substitute: true,
    can_lobby_host: false,
    duo_request: null,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

export function buildEventTeam(overrides: Partial<EventTeam> = {}): EventTeam {
  return {
    team_number: 1,
    players: [buildLobbyPlayer()],
    ...overrides,
  };
}

export function buildEventLobby(overrides: Partial<EventLobby> = {}): EventLobby {
  return {
    id: nextId("lobby"),
    host_id: null,
    join_code: null,
    fairness_warning: false,
    fairness_warning_at_lock: false,
    teams: [buildEventTeam()],
    subs: [],
    ...overrides,
  };
}

export function buildEventGroupEvent(overrides: Partial<EventGroupEvent> = {}): EventGroupEvent {
  return {
    id: nextId("event"),
    start_time: "2026-09-14T20:00:00Z",
    game_mode_id: nextId("game-mode"),
    game_mode_name: "5v5",
    team_size: 5,
    registered_count: 1,
    lobbies_count: 0,
    player_registered: false,
    registrations: [buildEventRegistration()],
    lobbies: [],
    unplaced: [],
    ...overrides,
  };
}

export function buildDiscordGuild(overrides: Partial<DiscordGuild> = {}): DiscordGuild {
  return {
    id: nextId("guild"),
    name: "Test Discord Server",
    ...overrides,
  };
}

export function buildEventGroupDetail(overrides: Partial<EventGroupDetail> = {}): EventGroupDetail {
  return {
    id: nextId("group"),
    name: "Test Event Group",
    owner_id: nextId("user"),
    owner_name: "HostUser#0001",
    owner_display_name: "Host User",
    owner_pronouns: "",
    game_mode_id: nextId("game-mode"),
    game_mode_name: "5v5",
    game_id: nextId("game"),
    game_name: "Test Game",
    join_link_base: null,
    team_size: 5,
    sub_min: 1,
    registration_open: true,
    region: "AMER",
    sort_logic: "balanced",
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    events: [buildEventGroupEvent()],
    discord_guilds: [],
    ...overrides,
  };
}

export function buildMyEvent(overrides: Partial<Event> = {}): Event {
  return {
    id: nextId("group"),
    name: "Test Event Group",
    game_name: "Test Game",
    game_mode: "5v5",
    region: "AMER",
    event_date: "2026-09-14T20:00:00Z",
    host_name: "HostUser#0001",
    host_display_name: "Host User",
    host_id: nextId("user"),
    registered_count: 1,
    registration_open: true,
    ...overrides,
  };
}

export function buildEventsPage(overrides: Partial<EventsPage> = {}): EventsPage {
  return {
    event_groups: [buildMyEvent()],
    next_cursor: null,
    has_more: false,
    ...overrides,
  };
}

export function buildCreateTeamsResponse(overrides: Partial<CreateTeamsResponse> = {}): CreateTeamsResponse {
  return {
    sub_capacity_adjusted: false,
    ...overrides,
  };
}

export function buildCompleteAuthResponse(overrides: Partial<CompleteAuthResponse> = {}): CompleteAuthResponse {
  return {
    access_token: "test-access-token",
    ...overrides,
  };
}
