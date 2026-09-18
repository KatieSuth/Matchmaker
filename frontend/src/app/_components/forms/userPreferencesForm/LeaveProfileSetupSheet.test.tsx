import { describe, expect, it, vi } from "vitest";
import { axe } from "@/test/axe";
import { render, screen, userEvent } from "@/test/render";
import { LeaveProfileSetupSheet } from "./LeaveProfileSetupSheet";

describe("LeaveProfileSetupSheet", () => {
  it("renders nothing when closed", () => {
    const { container } = render(
      <LeaveProfileSetupSheet isOpen={false} onClose={vi.fn()} onConfirmLeave={vi.fn()} />,
    );
    expect(container).toBeEmptyDOMElement();
  });

  it("calls onConfirmLeave and onClose from their buttons", async () => {
    const onClose = vi.fn();
    const onConfirmLeave = vi.fn();
    const user = userEvent.setup();
    render(<LeaveProfileSetupSheet isOpen onClose={onClose} onConfirmLeave={onConfirmLeave} />);

    expect(screen.getByText("Leave profile setup?")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Leave" }));
    expect(onConfirmLeave).toHaveBeenCalledTimes(1);

    await user.click(screen.getByRole("button", { name: "Stay" }));
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("has no detectable accessibility violations when open", async () => {
    render(
      <LeaveProfileSetupSheet isOpen onClose={vi.fn()} onConfirmLeave={vi.fn()} />,
    );
    // The sheet portals to document.body, so axe needs to scan the whole document.
    expect(await axe(document.body)).toHaveNoViolations();
  });
});
