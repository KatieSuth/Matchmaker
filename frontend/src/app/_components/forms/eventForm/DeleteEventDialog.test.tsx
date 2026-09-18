// Tests for the full-screen "delete event group" confirmation overlay.
import { describe, expect, it, vi } from "vitest";
import { axe } from "@/test/axe";
import { render, screen, userEvent } from "@/test/render";
import { DeleteEventDialog } from "./DeleteEventDialog";

function baseProps(overrides: Partial<Parameters<typeof DeleteEventDialog>[0]> = {}) {
  return {
    isOpen: true,
    deleteError: null,
    isDeleting: false,
    onCancel: vi.fn(),
    onConfirm: vi.fn(),
    ...overrides,
  };
}

describe("DeleteEventDialog", () => {
  it("renders nothing when closed", () => {
    const { container } = render(<DeleteEventDialog {...baseProps({ isOpen: false })} />);
    expect(container).toBeEmptyDOMElement();
  });

  it("shows the delete error message when present", () => {
    render(<DeleteEventDialog {...baseProps({ deleteError: "Could not delete this event group." })} />);
    expect(screen.getByText("Could not delete this event group.")).toBeInTheDocument();
  });

  it("calls onCancel / onConfirm from their buttons", async () => {
    const onCancel = vi.fn();
    const onConfirm = vi.fn();
    const user = userEvent.setup();
    render(<DeleteEventDialog {...baseProps({ onCancel, onConfirm })} />);

    await user.click(screen.getByRole("button", { name: "Cancel" }));
    expect(onCancel).toHaveBeenCalledTimes(1);

    await user.click(screen.getByRole("button", { name: "Delete Permanently" }));
    expect(onConfirm).toHaveBeenCalledTimes(1);
  });

  it("disables both buttons and relabels confirm while deleting", () => {
    render(<DeleteEventDialog {...baseProps({ isDeleting: true })} />);
    expect(screen.getByRole("button", { name: "Cancel" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Deleting..." })).toBeDisabled();
  });

  it("has no accessibility violations", async () => {
    const { container } = render(<DeleteEventDialog {...baseProps()} />);
    expect(await axe(container)).toHaveNoViolations();
  });
});
