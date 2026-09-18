// Tests for the per-game profile editor used by both the settings form and the registration sheet.
import { describe, expect, it, vi } from "vitest";
import { axe } from "@/test/axe";
import { buildGame, buildGameRank } from "@/test/fixtures";
import { render, screen, userEvent } from "@/test/render";
import { UserGameEditor, UserGameEditorValue } from "./UserGameEditor";

const emptyValue: UserGameEditorValue = {
  game_id: "",
  in_game_name: "",
  current_rank: "",
  peak_rank: "",
  show_rank: false,
};

describe("UserGameEditor", () => {
  it("prompts the user to pick a game before showing name/rank fields", () => {
    render(<UserGameEditor value={emptyValue} ranks={[]} onChange={vi.fn()} />);
    expect(screen.getByText("Select a game above to continue")).toBeInTheDocument();
    expect(screen.queryByText("In-game name *")).not.toBeInTheDocument();
  });

  it("shows a pulse placeholder while ranks are loading for a selected game", () => {
    render(
      <UserGameEditor
        value={{ ...emptyValue, game_id: "game-1" }}
        ranks={[]}
        ranksLoading
        onChange={vi.fn()}
      />,
    );
    expect(screen.queryByText("In-game name *")).not.toBeInTheDocument();
    expect(document.querySelector(".animate-pulse")).toBeInTheDocument();
  });

  it("renders in-game name and rank pickers once a game is selected and ranks have loaded", () => {
    const ranks = [buildGameRank({ id: "r1", name: "Gold" }), buildGameRank({ id: "r2", name: "Plat" })];
    render(
      <UserGameEditor
        value={{ ...emptyValue, game_id: "game-1", in_game_name: "Tag#123" }}
        ranks={ranks}
        onChange={vi.fn()}
      />,
    );
    expect(screen.getByDisplayValue("Tag#123")).toBeInTheDocument();
    expect(screen.getByText("Current rank *")).toBeInTheDocument();
    expect(screen.getByText("Peak rank *")).toBeInTheDocument();
  });

  it("calls onChange with a cleared rank pair when a different game is picked", async () => {
    const onChange = vi.fn();
    const user = userEvent.setup();
    const games = [buildGame({ id: "game-1", name: "Valorant" }), buildGame({ id: "game-2", name: "League" })];
    render(
      <UserGameEditor
        value={{ ...emptyValue, game_id: "game-1", current_rank: "r1", peak_rank: "r2" }}
        allGames={games}
        ranks={[]}
        onChange={onChange}
      />,
    );

    await user.click(screen.getByRole("combobox"));
    await user.click(screen.getByText("League"));

    expect(onChange).toHaveBeenCalledWith(
      expect.objectContaining({ game_id: "game-2", current_rank: "", peak_rank: "" }),
    );
  });

  it("hides the game selector and shows the provided label when hideGameSelector is set", () => {
    render(
      <UserGameEditor
        value={{ ...emptyValue, game_id: "game-1" }}
        ranks={[]}
        hideGameSelector
        gameLabel="Valorant"
        onChange={vi.fn()}
      />,
    );
    expect(screen.getByText("Valorant")).toBeInTheDocument();
    expect(screen.queryByRole("combobox")).not.toBeInTheDocument();
  });

  it("calls onRemove from the remove-game button", async () => {
    const onRemove = vi.fn();
    const user = userEvent.setup();
    render(<UserGameEditor value={emptyValue} ranks={[]} onChange={vi.fn()} onRemove={onRemove} />);

    await user.click(screen.getByRole("button", { name: "Remove game" }));

    expect(onRemove).toHaveBeenCalledTimes(1);
  });

  it("renders field-level error copy", () => {
    render(
      <UserGameEditor
        value={{ ...emptyValue, game_id: "game-1" }}
        ranks={[buildGameRank({ id: "r1", name: "Gold" })]}
        errors={{ in_game_name: "In-game name is required", current_rank: "Current rank is required" }}
        onChange={vi.fn()}
      />,
    );
    expect(screen.getByText("In-game name is required")).toBeInTheDocument();
    expect(screen.getByText("Current rank is required")).toBeInTheDocument();
  });

  it("has no accessibility violations before a game is selected", async () => {
    const { container } = render(<UserGameEditor value={emptyValue} ranks={[]} onChange={vi.fn()} />);
    expect(await axe(container)).toHaveNoViolations();
  });

  it("has no accessibility violations with name and rank fields", async () => {
    const { container } = render(
      <UserGameEditor
        value={{ ...emptyValue, game_id: "game-1", in_game_name: "Tag#123" }}
        ranks={[buildGameRank({ id: "r1", name: "Gold" })]}
        onChange={vi.fn()}
      />,
    );
    expect(await axe(container)).toHaveNoViolations();
  });
});
