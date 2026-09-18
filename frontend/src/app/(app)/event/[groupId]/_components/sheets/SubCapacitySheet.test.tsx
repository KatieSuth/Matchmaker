// Tests for the informational "sub-capacity adjusted" notice shown after lock-in.
import { describe, expect, it, vi } from "vitest";
import { axe } from "@/test/axe";
import { render, screen, userEvent } from "@/test/render";
import { SubCapacitySheet } from "./SubCapacitySheet";

describe("SubCapacitySheet", () => {
  it("does not render when closed", () => {
    render(<SubCapacitySheet isOpen={false} onClose={vi.fn()} />);
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("shows the explanation and calls onClose from the Okay button", async () => {
    const onClose = vi.fn();
    const user = userEvent.setup();
    render(<SubCapacitySheet isOpen={true} onClose={onClose} />);

    expect(screen.getByText(/not enough players willing to substitute/i)).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Okay" }));

    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("has no accessibility violations", async () => {
    render(<SubCapacitySheet isOpen={true} onClose={vi.fn()} />);
    // The sheet portals to document.body, so axe needs to scan the whole document.
    expect(await axe(document.body)).toHaveNoViolations();
  });
});
