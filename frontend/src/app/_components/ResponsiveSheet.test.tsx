// Tests for the portaled modal "sheet" used for registration/event-edit flows.
import { afterEach, describe, expect, it, vi } from "vitest";
import { axe } from "@/test/axe";
import { render, screen, userEvent } from "@/test/render";
import { ResponsiveSheet } from "./ResponsiveSheet";

describe("ResponsiveSheet", () => {
  afterEach(() => {
    document.body.style.overflow = "";
  });

  it("renders nothing when closed", () => {
    render(
      <ResponsiveSheet isOpen={false} onClose={vi.fn()}>
        <p>Sheet content</p>
      </ResponsiveSheet>,
    );
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("renders the title and children when open", () => {
    render(
      <ResponsiveSheet isOpen={true} onClose={vi.fn()} title="Edit Event">
        <p>Sheet content</p>
      </ResponsiveSheet>,
    );
    expect(screen.getByRole("dialog", { name: "Edit Event" })).toBeInTheDocument();
    expect(screen.getByText("Sheet content")).toBeInTheDocument();
  });

  it("locks body scroll while open and restores it on close/unmount", () => {
    const { rerender, unmount } = render(
      <ResponsiveSheet isOpen={true} onClose={vi.fn()}>
        <p>Sheet content</p>
      </ResponsiveSheet>,
    );
    expect(document.body).toHaveStyle({ overflow: "hidden" });

    rerender(
      <ResponsiveSheet isOpen={false} onClose={vi.fn()}>
        <p>Sheet content</p>
      </ResponsiveSheet>,
    );
    expect(document.body).toHaveStyle({ overflow: "" });

    rerender(
      <ResponsiveSheet isOpen={true} onClose={vi.fn()}>
        <p>Sheet content</p>
      </ResponsiveSheet>,
    );
    expect(document.body).toHaveStyle({ overflow: "hidden" });
    unmount();
    expect(document.body).toHaveStyle({ overflow: "" });
  });

  it("calls onClose when the backdrop is clicked", async () => {
    const onClose = vi.fn();
    const user = userEvent.setup();
    render(
      <ResponsiveSheet isOpen={true} onClose={onClose}>
        <p>Sheet content</p>
      </ResponsiveSheet>,
    );

    await user.click(screen.getByRole("button", { name: "Close" }));

    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("calls onClose when the header close button is clicked", async () => {
    const onClose = vi.fn();
    const user = userEvent.setup();
    render(
      <ResponsiveSheet isOpen={true} onClose={onClose} title="Edit Event">
        <p>Sheet content</p>
      </ResponsiveSheet>,
    );

    await user.click(screen.getByRole("button", { name: "Close sheet" }));

    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("calls onClose on Escape", async () => {
    const onClose = vi.fn();
    const user = userEvent.setup();
    render(
      <ResponsiveSheet isOpen={true} onClose={onClose}>
        <p>Sheet content</p>
      </ResponsiveSheet>,
    );

    await user.keyboard("{Escape}");

    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("does not listen for Escape while closed", async () => {
    const onClose = vi.fn();
    const user = userEvent.setup();
    render(
      <ResponsiveSheet isOpen={false} onClose={onClose}>
        <p>Sheet content</p>
      </ResponsiveSheet>,
    );

    await user.keyboard("{Escape}");

    expect(onClose).not.toHaveBeenCalled();
  });

  it("has no accessibility violations when open", async () => {
    render(
      <ResponsiveSheet isOpen={true} onClose={vi.fn()} title="Edit Event">
        <p>Sheet content</p>
      </ResponsiveSheet>,
    );
    // The sheet portals to document.body, so axe needs to scan the whole document.
    expect(await axe(document.body)).toHaveNoViolations();
  });
});
