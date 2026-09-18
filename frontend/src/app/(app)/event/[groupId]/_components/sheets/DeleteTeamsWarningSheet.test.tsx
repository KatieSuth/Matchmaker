// Tests for the host-only "delete teams" confirmation sheet.
import { describe, expect, it, vi } from "vitest";
import { axe } from "@/test/axe";
import { render, screen, userEvent } from "@/test/render";
import { DeleteTeamsWarningSheet } from "./DeleteTeamsWarningSheet";

describe("DeleteTeamsWarningSheet", () => {
  it("does not render when closed", () => {
    render(<DeleteTeamsWarningSheet isOpen={false} onClose={vi.fn()} working={false} onConfirm={vi.fn()} />);
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("calls onClose from Cancel and onConfirm from Delete Teams", async () => {
    const onClose = vi.fn();
    const onConfirm = vi.fn();
    const user = userEvent.setup();
    render(<DeleteTeamsWarningSheet isOpen={true} onClose={onClose} working={false} onConfirm={onConfirm} />);

    await user.click(screen.getByRole("button", { name: "Cancel" }));
    expect(onClose).toHaveBeenCalledTimes(1);

    await user.click(screen.getByRole("button", { name: "Delete Teams" }));
    expect(onConfirm).toHaveBeenCalledTimes(1);
  });

  it("disables and relabels the confirm button while working", () => {
    render(<DeleteTeamsWarningSheet isOpen={true} onClose={vi.fn()} working={true} onConfirm={vi.fn()} />);
    expect(screen.getByRole("button", { name: "Deleting..." })).toBeDisabled();
  });

  it("has no accessibility violations", async () => {
    render(<DeleteTeamsWarningSheet isOpen={true} onClose={vi.fn()} working={false} onConfirm={vi.fn()} />);
    // The sheet portals to document.body, so axe needs to scan the whole document.
    expect(await axe(document.body)).toHaveNoViolations();
  });
});
