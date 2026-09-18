import { describe, expect, it, vi } from "vitest";
import { axe } from "@/test/axe";
import { buildGameMode } from "@/test/fixtures";
import { render, screen, userEvent } from "@/test/render";
import { PerGameScheduleEditor } from "./PerGameScheduleEditor";
import { PerGameDraftRow } from "./schema";

describe("PerGameScheduleEditor", () => {
  it("renders nothing when there are no scheduled games", () => {
    const { container } = render(
      <PerGameScheduleEditor
        perGameDraft={[]}
        setPerGameDraft={vi.fn()}
        modes={[]}
        modesLoading={false}
        modesError={null}
        watchedGameId="game-1"
        readOnly={false}
        userTz="UTC"
      />,
    );
    expect(container).toBeEmptyDOMElement();
  });

  it("renders a numbered row per scheduled game", () => {
    const rows: PerGameDraftRow[] = [
      { eventId: "e1", startLocal: "2026-09-20T12:00", modeId: "mode-1" },
      { eventId: "e2", startLocal: "2026-09-20T13:00", modeId: "mode-1" },
    ];
    render(
      <PerGameScheduleEditor
        perGameDraft={rows}
        setPerGameDraft={vi.fn()}
        modes={[buildGameMode({ id: "mode-1", name: "5v5" })]}
        modesLoading={false}
        modesError={null}
        watchedGameId="game-1"
        readOnly={false}
        userTz="UTC"
      />,
    );
    expect(screen.getByText("Game 1")).toBeInTheDocument();
    expect(screen.getByText("Game 2")).toBeInTheDocument();
    expect(screen.getByText(/Times use your local timezone \(UTC\)/)).toBeInTheDocument();
  });

  it("disables the mode picker in read-only mode", () => {
    const rows: PerGameDraftRow[] = [{ eventId: "e1", startLocal: "2026-09-20T12:00", modeId: "mode-1" }];
    render(
      <PerGameScheduleEditor
        perGameDraft={rows}
        setPerGameDraft={vi.fn()}
        modes={[buildGameMode({ id: "mode-1", name: "5v5" })]}
        modesLoading={false}
        modesError={null}
        watchedGameId="game-1"
        readOnly
        userTz="UTC"
      />,
    );
    expect(screen.getByRole("combobox")).toHaveAttribute("aria-readonly", "true");
    expect(screen.getByRole("textbox")).toBeDisabled();
  });

  it("updates the draft when a different game mode is picked", async () => {
    const setPerGameDraft = vi.fn();
    const user = userEvent.setup();
    const rows: PerGameDraftRow[] = [{ eventId: "e1", startLocal: "2026-09-20T12:00", modeId: "mode-1" }];
    render(
      <PerGameScheduleEditor
        perGameDraft={rows}
        setPerGameDraft={setPerGameDraft}
        modes={[buildGameMode({ id: "mode-1", name: "5v5" }), buildGameMode({ id: "mode-2", name: "3v3" })]}
        modesLoading={false}
        modesError={null}
        watchedGameId="game-1"
        readOnly={false}
        userTz="UTC"
      />,
    );

    await user.click(screen.getByRole("combobox"));
    await user.click(screen.getByText("3v3"));

    expect(setPerGameDraft).toHaveBeenCalledTimes(1);
    const updater = setPerGameDraft.mock.calls[0][0] as (prev: PerGameDraftRow[]) => PerGameDraftRow[];
    expect(updater(rows)).toEqual([{ eventId: "e1", startLocal: "2026-09-20T12:00", modeId: "mode-2" }]);
  });

  it("has no accessibility violations", async () => {
    const rows: PerGameDraftRow[] = [{ eventId: "e1", startLocal: "2026-09-20T12:00", modeId: "mode-1" }];
    const { container } = render(
      <PerGameScheduleEditor
        perGameDraft={rows}
        setPerGameDraft={vi.fn()}
        modes={[buildGameMode({ id: "mode-1", name: "5v5" })]}
        modesLoading={false}
        modesError={null}
        watchedGameId="game-1"
        readOnly={false}
        userTz="UTC"
      />,
    );
    expect(await axe(container)).toHaveNoViolations();
  });
});
