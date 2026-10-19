// Tests for the post-registration "add to calendar" modal.
import { describe, expect, it, vi } from "vitest";
import { axe } from "@/test/axe";
import { render, screen, userEvent } from "@/test/render";
import { CalendarInviteSheet } from "./CalendarInviteSheet";

describe("CalendarInviteSheet", () => {
  it("does not render when closed", () => {
    render(<CalendarInviteSheet isOpen={false} onClose={vi.fn()} onDownload={vi.fn()} />);
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("explains the reminders and offers download and dismiss", async () => {
    const onClose = vi.fn();
    const onDownload = vi.fn();
    const user = userEvent.setup();
    render(<CalendarInviteSheet isOpen onClose={onClose} onDownload={onDownload} />);

    expect(screen.getByRole("dialog", { name: "You're registered!" })).toHaveTextContent(
      "24 hours, 90 minutes, and 10 minutes",
    );
    await user.click(screen.getByRole("button", { name: "Download calendar invite" }));
    expect(onDownload).toHaveBeenCalledTimes(1);
    await user.click(screen.getByRole("button", { name: "Not now" }));
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("closes on Escape", async () => {
    const onClose = vi.fn();
    const user = userEvent.setup();
    render(<CalendarInviteSheet isOpen onClose={onClose} onDownload={vi.fn()} />);
    await user.keyboard("{Escape}");
    expect(onClose).toHaveBeenCalled();
  });

  it("has no accessibility violations", async () => {
    const { container } = render(<CalendarInviteSheet isOpen onClose={vi.fn()} onDownload={vi.fn()} />);
    expect(await axe(container)).toHaveNoViolations();
  });
});
