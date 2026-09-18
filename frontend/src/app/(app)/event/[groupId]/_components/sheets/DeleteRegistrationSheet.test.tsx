// Tests for the delete-registration confirmation sheet's self/other and single/all copy variants.
import { describe, expect, it, vi } from "vitest";
import { axe } from "@/test/axe";
import { render, screen, userEvent } from "@/test/render";
import { PendingDeleteAction } from "../../_types";
import { DeleteRegistrationSheet } from "./DeleteRegistrationSheet";

function baseProps(overrides: Partial<Parameters<typeof DeleteRegistrationSheet>[0]> = {}) {
  return {
    isOpen: true,
    onClose: vi.fn(),
    pendingDeleteAction: {
      mode: "single",
      userId: "user-1",
      userName: "Alice#0001",
      eventId: "event-1",
      gameNumber: 1,
      registrationsInGroup: 1,
    } as PendingDeleteAction,
    deletingSelf: false,
    working: false,
    onConfirm: vi.fn(),
    ...overrides,
  };
}

describe("DeleteRegistrationSheet", () => {
  it("titles a single-game deletion by game number", () => {
    render(<DeleteRegistrationSheet {...baseProps()} />);
    expect(screen.getByRole("dialog", { name: "Delete Registration for Game 1" })).toBeInTheDocument();
  });

  it("titles an all-games deletion by the target user's name", () => {
    render(<DeleteRegistrationSheet {...baseProps({ pendingDeleteAction: { mode: "all", userId: "user-1", userName: "Alice#0001", eventId: "event-1", gameNumber: 1, registrationsInGroup: 2 } })} />);
    expect(screen.getByRole("dialog", { name: "Delete All Registrations From Alice#0001" })).toBeInTheDocument();
  });

  it("uses first-person copy when deleting the viewer's own single-game registration", () => {
    render(<DeleteRegistrationSheet {...baseProps({ deletingSelf: true })} />);
    expect(screen.getByText(/you will need to register again to play this game/)).toBeInTheDocument();
  });

  it("uses third-person copy when a host deletes another player's registration", () => {
    render(<DeleteRegistrationSheet {...baseProps({ deletingSelf: false })} />);
    expect(screen.getByText(/they will need to register again to play this game/)).toBeInTheDocument();
  });

  it("mentions other games in the series when the user has multiple registrations", () => {
    render(
      <DeleteRegistrationSheet
        {...baseProps({
          deletingSelf: true,
          pendingDeleteAction: { mode: "single", userId: "user-1", userName: "Alice#0001", eventId: "event-1", gameNumber: 2, registrationsInGroup: 3 },
        })}
      />,
    );
    expect(screen.getByText(/registered for other games in this series/)).toBeInTheDocument();
  });

  it("shows 'Delete All Registrations' as the confirm label for mode 'all'", () => {
    render(<DeleteRegistrationSheet {...baseProps({ pendingDeleteAction: { mode: "all", userId: "user-1", userName: "Alice#0001", eventId: "event-1", gameNumber: 1, registrationsInGroup: 1 } })} />);
    expect(screen.getByRole("button", { name: "Delete All Registrations" })).toBeInTheDocument();
  });

  it("disables the confirm button while working, and relabels it", () => {
    render(<DeleteRegistrationSheet {...baseProps({ working: true })} />);
    expect(screen.getByRole("button", { name: "Deleting..." })).toBeDisabled();
  });

  it("calls onConfirm and onClose from their buttons", async () => {
    const onConfirm = vi.fn();
    const onClose = vi.fn();
    const user = userEvent.setup();
    render(<DeleteRegistrationSheet {...baseProps({ onConfirm, onClose })} />);

    await user.click(screen.getByRole("button", { name: "Delete Registration" }));
    expect(onConfirm).toHaveBeenCalledTimes(1);

    await user.click(screen.getByRole("button", { name: "Cancel" }));
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("has no accessibility violations", async () => {
    render(<DeleteRegistrationSheet {...baseProps()} />);
    // The sheet portals to document.body, so axe needs to scan the whole document.
    expect(await axe(document.body)).toHaveNoViolations();
  });
});
