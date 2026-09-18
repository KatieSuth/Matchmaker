// Tests for a single games-section row: rank fetch on game selection, then UserGameEditor wiring.
import { useForm } from "react-hook-form";
import { http, HttpResponse } from "msw";
import { describe, expect, it, vi } from "vitest";
import { waitFor } from "@testing-library/react";
import { axe } from "@/test/axe";
import { buildGame, buildGameRank } from "@/test/fixtures";
import { server } from "@/test/msw/server";
import { TEST_API_URL } from "@/test/constants";
import { render, screen, userEvent } from "@/test/render";
import { GameCard } from "./GameCard";
import { PreferencesFormValues } from "./schema";

const GAME_ID = "123e4567-e89b-12d3-a456-426614174000";
const OTHER_GAME_ID = "123e4567-e89b-12d3-a456-426614174001";

function Harness({
  persistedGameIds = new Set<string>(),
  onRemove = () => {},
}: {
  persistedGameIds?: Set<string>;
  onRemove?: () => void;
}) {
  const { control, setValue, formState } = useForm<PreferencesFormValues>({
    defaultValues: {
      display_name: "",
      pronouns: "",
      show_pronouns: false,
      region: null,
      games: [
        {
          game_id: GAME_ID,
          in_game_name: "Tag#1",
          current_rank: "",
          peak_rank: "",
          show_rank: true,
          api_permission: false,
        },
      ],
    },
  });

  return (
    <GameCard
      index={0}
      allGames={[buildGame({ id: GAME_ID, name: "Valorant" }), buildGame({ id: OTHER_GAME_ID, name: "League" })]}
      control={control}
      setValue={setValue}
      errors={formState.errors}
      takenGameIds={[GAME_ID]}
      persistedGameIds={persistedGameIds}
      onRemove={onRemove}
    />
  );
}

describe("GameCard", () => {
  it("fetches ranks for the selected game and shows the in-game name field", async () => {
    server.use(
      http.get(`${TEST_API_URL}/games/${GAME_ID}/ranks`, () =>
        HttpResponse.json([buildGameRank({ id: "r1", name: "Gold 1", game_id: GAME_ID, order: 1 })]),
      ),
    );

    render(<Harness />);

    expect(await screen.findByDisplayValue("Tag#1")).toBeInTheDocument();
    expect(screen.getByText("Current rank *")).toBeInTheDocument();
  });

  it("locks the game dropdown for a row whose game is already persisted", async () => {
    server.use(http.get(`${TEST_API_URL}/games/${GAME_ID}/ranks`, () => HttpResponse.json([])));
    render(<Harness persistedGameIds={new Set([GAME_ID])} />);

    await waitFor(() => expect(screen.getByRole("combobox")).toHaveAttribute("aria-readonly", "true"));
  });

  it("calls onRemove from the remove button", async () => {
    server.use(http.get(`${TEST_API_URL}/games/${GAME_ID}/ranks`, () => HttpResponse.json([])));
    const onRemove = vi.fn();
    const user = userEvent.setup();
    render(<Harness onRemove={onRemove} />);

    await user.click(screen.getByRole("button", { name: "Remove game" }));
    expect(onRemove).toHaveBeenCalledTimes(1);
  });

  it("has no accessibility violations", async () => {
    server.use(
      http.get(`${TEST_API_URL}/games/${GAME_ID}/ranks`, () =>
        HttpResponse.json([buildGameRank({ id: "r1", name: "Gold 1", game_id: GAME_ID, order: 1 })]),
      ),
    );
    const { container } = render(<Harness />);
    expect(await screen.findByDisplayValue("Tag#1")).toBeInTheDocument();
    expect(await axe(container)).toHaveNoViolations();
  });
});
