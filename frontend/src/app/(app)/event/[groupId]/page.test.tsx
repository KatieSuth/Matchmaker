// Integration tests for the event group page: auth/access gates, host vs participant visibility,
// and the click wiring that opens sheets / copy actions / registration (hook internals stay in
// their own files).
import { http, HttpResponse } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  buildEventGroupDetail,
  buildEventGroupEvent,
  buildEventLobby,
  buildEventRegistration,
  buildEventTeam,
  buildLobbyPlayer,
  buildUser,
  buildUserGame,
} from "@/test/fixtures";
import { server } from "@/test/msw/server";
import { axe } from "@/test/axe";
import { TEST_API_URL } from "@/test/constants";
import { resetNextNavigationMock, setMockParams } from "@/test/nextNavigationMock";
import { renderWithProviders, screen, userEvent, waitFor, within } from "@/test/render";
import EventGroupPage from "./page";

vi.mock("next/navigation", () => import("@/test/nextNavigationMock"));

const HOST = buildUser({ id: "host-1", display_name: "Hosty" });
const GUEST = buildUser({ id: "guest-1", display_name: "Guest" });
const GROUP_ID = "group-1";

function groupPayload(overrides: Parameters<typeof buildEventGroupDetail>[0] = {}) {
  return buildEventGroupDetail({
    id: GROUP_ID,
    name: "Friday Customs",
    owner_id: HOST.id,
    owner_display_name: "Hosty",
    owner_name: "hosty#0001",
    game_name: "Valorant",
    game_mode_name: "5v5",
    region: "AMER",
    registration_open: true,
    events: [
      buildEventGroupEvent({
        id: "event-1",
        start_time: "2026-12-01T18:00:00Z",
        game_mode_name: "5v5",
        team_size: 5,
        registered_count: 1,
        lobbies_count: 0,
        player_registered: false,
        registrations: [buildEventRegistration({ user_id: GUEST.id, display_name: "Guest" })],
        lobbies: [],
        unplaced: [],
      }),
    ],
    ...overrides,
  });
}

function stubGroup(payload: ReturnType<typeof buildEventGroupDetail>) {
  server.use(
    http.get(`${TEST_API_URL}/events/${GROUP_ID}/access`, () => new HttpResponse(null, { status: 204 })),
    http.get(`${TEST_API_URL}/events/${GROUP_ID}`, () => HttpResponse.json(payload)),
    http.get(`${TEST_API_URL}/games/:gameId/ranks`, () => HttpResponse.json([])),
  );
}

/** MSW 403 with Discord-guild restriction details for the page's lock-denial branch. */
function stubAccessDenied(details: {
  event_title: string;
  event_named: boolean;
  discord_guilds: { id: string; name: string }[];
}) {
  server.use(
    http.get(`${TEST_API_URL}/events/${GROUP_ID}/access`, () =>
      HttpResponse.json(
        { message: "restricted", details: { code: "discord_guild_restricted", ...details } },
        { status: 403 },
      ),
    ),
  );
}

/** Locked-in group with two lobbies, a sub, and an unplaced player so host roster menus exist. */
function lockedInPayload() {
  const alice = buildLobbyPlayer({
    user_id: "alice-1",
    display_name: "Alice",
    discord_name: "alice#0001",
    can_lobby_host: false,
  });
  const bob = buildLobbyPlayer({
    user_id: "bob-1",
    display_name: "Bob",
    discord_name: "bob#0001",
    can_lobby_host: true,
  });
  const subby = buildLobbyPlayer({
    user_id: "sub-1",
    display_name: "Subby",
    discord_name: "sub#0001",
    can_substitute: true,
  });
  const unplaced = buildEventRegistration({
    user_id: "loose-1",
    display_name: "Loose Player",
    discord_name: "loose#0001",
    can_substitute: true,
  });
  const lobbyA = buildEventLobby({
    id: "lobby-1",
    host_id: HOST.id,
    join_code: "ABC123",
    teams: [
      buildEventTeam({ team_number: 1, players: [alice] }),
      buildEventTeam({ team_number: 2, players: [bob] }),
    ],
    subs: [subby],
  });
  const lobbyB = buildEventLobby({
    id: "lobby-2",
    host_id: null,
    join_code: null,
    teams: [buildEventTeam({ team_number: 1, players: [buildLobbyPlayer({ display_name: "Cara" })] })],
    subs: [],
  });
  return groupPayload({
    registration_open: false,
    events: [
      buildEventGroupEvent({
        id: "event-1",
        start_time: "2026-12-01T18:00:00Z",
        lobbies_count: 2,
        lobbies: [lobbyA, lobbyB],
        unplaced: [unplaced],
        registrations: [unplaced],
      }),
      buildEventGroupEvent({
        id: "event-2",
        start_time: "2026-12-01T19:00:00Z",
        game_mode_name: "Deathmatch",
        lobbies_count: 0,
        lobbies: [],
        unplaced: [],
        registrations: [],
      }),
    ],
  });
}

/** Opens the ellipsis menu on the player card whose display name starts with `name`. */
async function openRegistrationMenu(user: ReturnType<typeof userEvent.setup>, name: string) {
  const label = screen.getByText(new RegExp(`^${name}\\b`));
  const card = label.closest("div.card");
  if (!card) {
    throw new Error(`Could not find player card for ${name}`);
  }
  await user.click(within(card as HTMLElement).getByRole("button", { name: "Registration actions" }));
}

/** jsdom's navigator.clipboard is a getter-only property; spy the existing writeText when present. */
function stubClipboard() {
  const writeText = vi.fn().mockResolvedValue(undefined);
  if (!navigator.clipboard) {
    Object.defineProperty(navigator, "clipboard", {
      configurable: true,
      value: { writeText },
    });
    return writeText;
  }
  vi.spyOn(navigator.clipboard, "writeText").mockImplementation(writeText);
  return writeText;
}

describe("EventGroupPage", () => {
  beforeEach(() => {
    resetNextNavigationMock();
    setMockParams({ groupId: GROUP_ID });
  });

  it("shows host-only lock-in controls for the owner", async () => {
    stubGroup(groupPayload());
    renderWithProviders(<EventGroupPage />, { user: HOST, isAuthenticated: true });

    expect(await screen.findByRole("heading", { name: "Friday Customs" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Lock In & Create Teams" })).toBeInTheDocument();
  });

  it("hides lock-in and Discord-ping controls for a non-host participant", async () => {
    stubGroup(groupPayload());
    renderWithProviders(<EventGroupPage />, { user: GUEST, isAuthenticated: true });

    expect(await screen.findByRole("heading", { name: "Friday Customs" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Lock In & Create Teams" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Copy Discord Pings" })).not.toBeInTheDocument();
  });

  it("hides unplaced players from a non-host after lock-in", async () => {
    const player = buildLobbyPlayer({ display_name: "Alice" });
    const unplaced = buildEventRegistration({ display_name: "Loose Player" });
    const lobby = buildEventLobby({ teams: [buildEventTeam({ players: [player] })] });
    stubGroup(
      groupPayload({
        registration_open: false,
        events: [
          buildEventGroupEvent({
            id: "event-1",
            start_time: "2026-12-01T18:00:00Z",
            lobbies_count: 1,
            lobbies: [lobby],
            unplaced: [unplaced],
            registrations: [unplaced],
          }),
        ],
      }),
    );

    renderWithProviders(<EventGroupPage />, { user: GUEST, isAuthenticated: true });
    expect(await screen.findByText(/Alice/)).toBeInTheDocument();
    expect(screen.queryByText(/Loose Player/)).not.toBeInTheDocument();
    expect(screen.queryByText(/Unplaced/)).not.toBeInTheDocument();
  });

  it("shows unplaced players to the host after lock-in", async () => {
    const player = buildLobbyPlayer({ display_name: "Alice" });
    const unplaced = buildEventRegistration({ display_name: "Loose Player" });
    const lobby = buildEventLobby({ teams: [buildEventTeam({ players: [player] })] });
    stubGroup(
      groupPayload({
        registration_open: false,
        events: [
          buildEventGroupEvent({
            id: "event-1",
            start_time: "2026-12-01T18:00:00Z",
            lobbies_count: 1,
            lobbies: [lobby],
            unplaced: [unplaced],
            registrations: [unplaced],
          }),
        ],
      }),
    );

    renderWithProviders(<EventGroupPage />, { user: HOST, isAuthenticated: true });
    expect(await screen.findByText(/Unplaced · 1/)).toBeInTheDocument();
    expect(screen.getByText(/Loose Player/)).toBeInTheDocument();
  });

  it("has no accessibility violations for the host", async () => {
    stubGroup(groupPayload());
    const { container } = renderWithProviders(<EventGroupPage />, { user: HOST, isAuthenticated: true });
    expect(await screen.findByRole("heading", { name: "Friday Customs" })).toBeInTheDocument();
    expect(await axe(container)).toHaveNoViolations();
  });

  it("has no accessibility violations for a guest", async () => {
    stubGroup(groupPayload());
    const { container } = renderWithProviders(<EventGroupPage />, { user: GUEST, isAuthenticated: true });
    expect(await screen.findByRole("heading", { name: "Friday Customs" })).toBeInTheDocument();
    expect(await axe(container)).toHaveNoViolations();
  });

  it("shows a loading gate while auth is still resolving", () => {
    renderWithProviders(<EventGroupPage />, { user: HOST, isAuthenticated: true, isLoading: true });
    expect(screen.getByText("Loading event group...")).toBeInTheDocument();
  });

  it("asks the viewer to sign in when they are not authenticated", () => {
    renderWithProviders(<EventGroupPage />, { isAuthenticated: false, user: null });
    expect(screen.getByText("Please sign in to view this page.")).toBeInTheDocument();
  });

  it("explains a Discord lock with a single required server", async () => {
    stubAccessDenied({
      event_title: "Friday Customs",
      event_named: true,
      discord_guilds: [{ id: "g1", name: "Server A" }],
    });
    renderWithProviders(<EventGroupPage />, { user: GUEST, isAuthenticated: true });

    expect(await screen.findByRole("heading", { name: "This event is locked" })).toBeInTheDocument();
    expect(screen.getByText(/Friday Customs is locked to Server A/)).toBeInTheDocument();
    expect(screen.getByText(/You need to be a member of that Discord server/)).toBeInTheDocument();
  });

  it("explains a Discord lock with multiple required servers", async () => {
    stubAccessDenied({
      event_title: "Valorant",
      event_named: false,
      discord_guilds: [
        { id: "g1", name: "Server A" },
        { id: "g2", name: "Server B" },
      ],
    });
    renderWithProviders(<EventGroupPage />, { user: GUEST, isAuthenticated: true });

    expect(await screen.findByRole("heading", { name: "This event is locked" })).toBeInTheDocument();
    expect(screen.getByText(/This Valorant event is locked to Server A or Server B/)).toBeInTheDocument();
    expect(screen.getByText(/You need to be a member of at least one of those Discord servers/)).toBeInTheDocument();
    expect(screen.getByText(/these Discord servers/)).toBeInTheDocument();
  });

  it("shows the load error when the group cannot be fetched", async () => {
    server.use(
      http.get(`${TEST_API_URL}/events/${GROUP_ID}/access`, () => new HttpResponse(null, { status: 204 })),
      http.get(`${TEST_API_URL}/events/${GROUP_ID}`, () => HttpResponse.json({ message: "missing" }, { status: 404 })),
    );
    renderWithProviders(<EventGroupPage />, { user: HOST, isAuthenticated: true });

    expect(await screen.findByText("Could not load this event group.")).toBeInTheDocument();
  });

  it("locks in teams and surfaces the sub-capacity notice when the API adjusted capacity", async () => {
    stubGroup(groupPayload());
    server.use(
      http.post(`${TEST_API_URL}/events/${GROUP_ID}/teams`, () =>
        HttpResponse.json({ sub_capacity_adjusted: true }),
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<EventGroupPage />, { user: HOST, isAuthenticated: true });
    expect(await screen.findByRole("heading", { name: "Friday Customs" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Lock In & Create Teams" }));

    expect(await screen.findByRole("dialog", { name: "Substitute minimum changed the teams" })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Okay" }));
    expect(screen.queryByRole("dialog", { name: "Substitute minimum changed the teams" })).not.toBeInTheDocument();
  });

  it("opens the delete-teams warning after lock-in and confirms deletion", async () => {
    stubGroup(lockedInPayload());
    const user = userEvent.setup();
    renderWithProviders(<EventGroupPage />, { user: HOST, isAuthenticated: true });
    expect(await screen.findByRole("heading", { name: "Friday Customs" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Delete teams" }));
    expect(await screen.findByRole("dialog", { name: "Delete teams" })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Cancel" }));
    expect(screen.queryByRole("dialog", { name: "Delete teams" })).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Delete teams" }));
    await user.click(screen.getByRole("button", { name: "Delete Teams" }));
    await waitFor(() => expect(screen.queryByRole("dialog", { name: "Delete teams" })).not.toBeInTheDocument());
  });

  it("copies the share link and Discord pings, and scrolls back to the top", async () => {
    const writeText = stubClipboard();
    const scrollIntoView = vi.fn();
    Element.prototype.scrollIntoView = scrollIntoView;
    stubGroup(lockedInPayload());
    const user = userEvent.setup();
    renderWithProviders(<EventGroupPage />, { user: HOST, isAuthenticated: true });
    expect(await screen.findByRole("heading", { name: "Friday Customs" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Show one at a time" }));
    expect(screen.getByRole("heading", { name: /Game 1/ })).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: /Game 2/ })).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /Game 2 ·/ }));
    expect(screen.getByRole("heading", { name: /Game 2/ })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /Game 1 ·/ }));

    await user.click(screen.getByRole("button", { name: "Copy share link" }));
    await waitFor(() => expect(writeText).toHaveBeenCalled());

    await user.click(screen.getByRole("button", { name: "Copy Discord Pings" }));
    await waitFor(() => expect(writeText).toHaveBeenCalledTimes(2));

    await user.click(screen.getAllByRole("button", { name: "Back to top" })[0]);
    expect(scrollIntoView).toHaveBeenCalled();
  });

  it("opens event settings and toasts after a successful save", async () => {
    stubGroup(groupPayload());
    const user = userEvent.setup();
    renderWithProviders(<EventGroupPage />, { user: HOST, isAuthenticated: true });
    expect(await screen.findByRole("heading", { name: "Friday Customs" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Event settings" }));
    expect(await screen.findByRole("dialog", { name: "Edit event settings" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Save Settings" }));
    expect(await screen.findByText("Event settings updated.")).toBeInTheDocument();
  });

  it("switches to a single-game view from the tab strip", async () => {
    stubGroup(
      groupPayload({
        events: [
          buildEventGroupEvent({
            id: "event-1",
            start_time: "2026-12-01T18:00:00Z",
            game_mode_name: "5v5",
            registrations: [buildEventRegistration({ user_id: GUEST.id, display_name: "Guest" })],
          }),
          buildEventGroupEvent({
            id: "event-2",
            start_time: "2026-12-01T19:00:00Z",
            game_mode_name: "Deathmatch",
            registrations: [],
          }),
        ],
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<EventGroupPage />, { user: HOST, isAuthenticated: true });
    expect(await screen.findByRole("heading", { name: "Friday Customs" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: /Game 1/ })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: /Game 2/ })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Show one at a time" }));
    await user.click(screen.getByRole("button", { name: /Game 2 ·/ }));

    expect(screen.getByRole("heading", { name: /Game 2/ })).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: /Game 1/ })).not.toBeInTheDocument();
  });

  it("opens the registration editor and save footer for the host", async () => {
    const payload = groupPayload({ game_id: "game-valorant" });
    stubGroup(payload);
    server.use(
      http.get(`${TEST_API_URL}/users/me/games`, () =>
        HttpResponse.json([
          buildUserGame({
            game_id: "game-valorant",
            in_game_name: "HostIGN",
            current_rank: "rank-1",
            peak_rank: "rank-2",
          }),
        ]),
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<EventGroupPage />, { user: HOST, isAuthenticated: true });
    expect(await screen.findByRole("heading", { name: "Friday Customs" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Register Now" }));
    expect(await screen.findByRole("heading", { name: "Register" })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Cancel Registration" }));
    await waitFor(() => expect(screen.queryByRole("heading", { name: "Register" })).not.toBeInTheDocument());

    await user.click(screen.getByRole("button", { name: "Register Now" }));
    expect(await screen.findByRole("heading", { name: "Register" })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Save Registration" }));
    await waitFor(() => expect(screen.queryByRole("heading", { name: "Register" })).not.toBeInTheDocument());
  });

  it("opens player details and host roster sheets from the locked-in teams view", async () => {
    stubGroup(lockedInPayload());
    stubClipboard();
    const user = userEvent.setup();
    renderWithProviders(<EventGroupPage />, { user: HOST, isAuthenticated: true });
    expect(await screen.findByText(/Alice @alice#0001/)).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Show one at a time" }));

    await openRegistrationMenu(user, "Alice");
    await user.click(screen.getByRole("button", { name: "Show More Details" }));
    expect(await screen.findByRole("dialog", { name: "Registration Details" })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Close sheet" }));

    await openRegistrationMenu(user, "Alice");
    await user.click(screen.getByRole("button", { name: "Swap" }));
    expect(await screen.findByRole("dialog", { name: /Swap / })).toBeInTheDocument();
    await user.click(screen.getByRole("combobox", { name: "Player to swap with" }));
    await user.click(await screen.findByRole("option", { name: /Bob/ }));
    await user.click(screen.getByRole("button", { name: "Submit" }));
    await waitFor(() => expect(screen.queryByRole("dialog", { name: /Swap / })).not.toBeInTheDocument());

    await openRegistrationMenu(user, "Alice");
    await user.click(screen.getByRole("button", { name: "Make Lobby Host" }));
    expect(await screen.findByRole("dialog", { name: /Make .* Lobby Host/ })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Yes, make lobby host" }));
    await waitFor(() => expect(screen.queryByRole("dialog", { name: /Make .* Lobby Host/ })).not.toBeInTheDocument());

    await openRegistrationMenu(user, "Subby");
    await user.click(screen.getByRole("button", { name: "Move to Unplaced" }));

    await openRegistrationMenu(user, "Loose Player");
    await user.click(screen.getByRole("button", { name: "Move to Substitutes" }));
    expect(await screen.findByRole("dialog", { name: /Move .* to subs/ })).toBeInTheDocument();
    await user.click(screen.getByRole("combobox", { name: "Lobby sub pool" }));
    await user.click(await screen.findByRole("option", { name: "Lobby 1" }));
    await user.click(screen.getByRole("button", { name: "Submit" }));
    await waitFor(() => expect(screen.queryByRole("dialog", { name: /Move .* to subs/ })).not.toBeInTheDocument());

    await user.click(screen.getAllByRole("button", { name: "Join Lobby" })[0]);
    expect(await screen.findByRole("dialog", { name: /Join Lobby 1/ })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Copy lobby join info" }));
    await user.click(screen.getByRole("button", { name: "Save" }));
    await waitFor(() => expect(screen.queryByRole("dialog", { name: /Join Lobby 1/ })).not.toBeInTheDocument());
  });

  it("opens delete-registration confirmation from the open-registration list", async () => {
    stubGroup(groupPayload());
    const user = userEvent.setup();
    renderWithProviders(<EventGroupPage />, { user: HOST, isAuthenticated: true });
    expect(await screen.findByRole("heading", { name: "Friday Customs" })).toBeInTheDocument();

    await openRegistrationMenu(user, "Guest");
    await user.click(screen.getByRole("button", { name: "Delete for Game 1" }));
    expect(await screen.findByRole("dialog", { name: "Delete Registration for Game 1" })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Delete Registration" }));
    await waitFor(() =>
      expect(screen.queryByRole("dialog", { name: "Delete Registration for Game 1" })).not.toBeInTheDocument(),
    );
  });

  it("opens delete-all confirmation from the open-registration list", async () => {
    stubGroup(groupPayload());
    const user = userEvent.setup();
    renderWithProviders(<EventGroupPage />, { user: HOST, isAuthenticated: true });
    expect(await screen.findByRole("heading", { name: "Friday Customs" })).toBeInTheDocument();

    await openRegistrationMenu(user, "Guest");
    await user.click(screen.getByRole("button", { name: "Delete All" }));
    expect(await screen.findByRole("dialog", { name: /Delete All Registrations From / })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Cancel" }));
    expect(screen.queryByRole("dialog", { name: /Delete All Registrations From / })).not.toBeInTheDocument();
  });
});
