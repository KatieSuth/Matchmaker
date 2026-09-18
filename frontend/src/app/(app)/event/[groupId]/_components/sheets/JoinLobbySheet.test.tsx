// Tests for the "Join Lobby" sheet's editable (host) and read-only (player) views.
import { describe, expect, it, vi } from "vitest";
import { buildEventLobby } from "@/test/fixtures";
import { axe } from "@/test/axe";
import { render, screen, userEvent } from "@/test/render";
import { PendingJoinLobby } from "../../_types";
import { JoinLobbySheet } from "./JoinLobbySheet";

const pendingJoinLobby: PendingJoinLobby = {
  lobby: buildEventLobby(),
  lobbyIndex: 0,
  gameNumber: 1,
  startTime: "2026-09-20T18:00:00Z",
};

function baseProps(overrides: Partial<Parameters<typeof JoinLobbySheet>[0]> = {}) {
  return {
    isOpen: true,
    onClose: vi.fn(),
    pendingJoinLobby,
    canEdit: true,
    joinLobbyDraft: "",
    onChangeDraft: vi.fn(),
    joinLobbyError: null,
    copyStatus: "idle" as const,
    onCopy: vi.fn(),
    onSave: vi.fn(),
    pendingJoinDisplay: null,
    joinLinkBase: null,
    working: false,
    ...overrides,
  };
}

describe("JoinLobbySheet", () => {
  it("titles the sheet with lobby number, game number, and start time", () => {
    render(<JoinLobbySheet {...baseProps()} />);
    expect(screen.getByRole("dialog", { name: /Join Lobby 1 \(Game 1/ })).toBeInTheDocument();
  });

  describe("editable (host) view", () => {
    it("shows an input for the join code/link", () => {
      render(<JoinLobbySheet {...baseProps({ canEdit: true })} />);
      expect(screen.getByPlaceholderText(/Code or https/)).toBeInTheDocument();
    });

    it("shows a join-lobby validation error when present", () => {
      render(<JoinLobbySheet {...baseProps({ joinLobbyError: "Invalid lobby code." })} />);
      expect(screen.getByText("Invalid lobby code.")).toBeInTheDocument();
    });

    it("disables Save when the draft fails validation (a link with no join_link_base)", () => {
      render(<JoinLobbySheet {...baseProps({ joinLobbyDraft: "https://gg.example.com/abc", joinLinkBase: null })} />);
      expect(screen.getByRole("button", { name: "Save" })).toBeDisabled();
    });

    it("enables Save for a plain code", () => {
      render(<JoinLobbySheet {...baseProps({ joinLobbyDraft: "ABC-123" })} />);
      expect(screen.getByRole("button", { name: "Save" })).toBeEnabled();
    });

    it("hides the copy button until there is a draft or existing display value", () => {
      const { rerender } = render(<JoinLobbySheet {...baseProps({ joinLobbyDraft: "" })} />);
      expect(screen.queryByRole("button", { name: "Copy lobby join info" })).not.toBeInTheDocument();

      rerender(<JoinLobbySheet {...baseProps({ joinLobbyDraft: "ABC-123" })} />);
      expect(screen.getByRole("button", { name: "Copy lobby join info" })).toBeInTheDocument();
    });

    it("calls onChangeDraft, onSave, onClose from their respective controls", async () => {
      const onChangeDraft = vi.fn();
      const onSave = vi.fn();
      const onClose = vi.fn();
      const user = userEvent.setup();
      render(<JoinLobbySheet {...baseProps({ joinLobbyDraft: "ABC-123", onChangeDraft, onSave, onClose })} />);

      await user.type(screen.getByPlaceholderText(/Code or https/), "X");
      expect(onChangeDraft).toHaveBeenCalled();

      await user.click(screen.getByRole("button", { name: "Save" }));
      expect(onSave).toHaveBeenCalledTimes(1);

      // The sheet's own backdrop is also an (aria-labeled, textless) "Close" button, so disambiguate
      // by visible text content to target the footer button specifically.
      await user.click(screen.getByText("Close"));
      expect(onClose).toHaveBeenCalledTimes(1);
    });

    it("has no accessibility violations", async () => {
      render(<JoinLobbySheet {...baseProps({ joinLobbyDraft: "ABC-123" })} />);
      // The sheet portals to document.body, so axe needs to scan the whole document.
      expect(await axe(document.body)).toHaveNoViolations();
    });
  });

  describe("read-only (player) view", () => {
    it("renders the code as plain text when the display kind is 'code'", () => {
      render(<JoinLobbySheet {...baseProps({ canEdit: false, pendingJoinDisplay: { kind: "code", value: "ABC-123" } })} />);
      expect(screen.getByText("ABC-123")).toBeInTheDocument();
      expect(screen.queryByRole("link")).not.toBeInTheDocument();
    });

    it("renders the value as a link when the display kind is 'link'", () => {
      render(
        <JoinLobbySheet
          {...baseProps({ canEdit: false, pendingJoinDisplay: { kind: "link", value: "https://gg.example.com/abc" } })}
        />,
      );
      expect(screen.getByRole("link", { name: "https://gg.example.com/abc" })).toHaveAttribute(
        "href",
        "https://gg.example.com/abc",
      );
    });

    it("shows a not-yet-added message when there is nothing to display", () => {
      render(<JoinLobbySheet {...baseProps({ canEdit: false, pendingJoinDisplay: null })} />);
      expect(screen.getByText(/has not been added yet/)).toBeInTheDocument();
    });

    it("calls onClose from the Close button", async () => {
      const onClose = vi.fn();
      const user = userEvent.setup();
      render(<JoinLobbySheet {...baseProps({ canEdit: false, pendingJoinDisplay: null, onClose })} />);

      await user.click(screen.getByText("Close"));

      expect(onClose).toHaveBeenCalledTimes(1);
    });

    it("has no accessibility violations", async () => {
      render(
        <JoinLobbySheet
          {...baseProps({ canEdit: false, pendingJoinDisplay: { kind: "code", value: "ABC-123" } })}
        />,
      );
      expect(await axe(document.body)).toHaveNoViolations();
    });
  });
});
