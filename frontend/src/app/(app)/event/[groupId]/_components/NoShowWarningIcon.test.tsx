// Tests for the host-only no-show warning icon and its hover/focus tooltip.
import { describe, expect, it } from "vitest";
import { axe } from "@/test/axe";
import { render, screen, userEvent } from "@/test/render";
import { NoShowWarningIcon } from "./NoShowWarningIcon";

describe("NoShowWarningIcon", () => {
  it("shows the no-show count tooltip on hover and hides it on leave", async () => {
    const user = userEvent.setup();
    render(<NoShowWarningIcon count={4} />);
    expect(screen.queryByRole("tooltip")).not.toBeInTheDocument();

    await user.hover(screen.getByRole("button"));
    expect(screen.getByRole("tooltip")).toHaveTextContent("No-show 4 times at your events in the last 6 months");

    await user.unhover(screen.getByRole("button"));
    expect(screen.queryByRole("tooltip")).not.toBeInTheDocument();
  });

  it("shows the tooltip on keyboard focus", async () => {
    const user = userEvent.setup();
    render(<NoShowWarningIcon count={3} />);

    await user.tab();
    expect(screen.getByRole("button")).toHaveFocus();
    expect(screen.getByRole("tooltip")).toBeInTheDocument();
    expect(screen.getByRole("button")).toHaveAttribute("aria-describedby", screen.getByRole("tooltip").id);

    await user.tab();
    expect(screen.queryByRole("tooltip")).not.toBeInTheDocument();
  });

  it("opens on tap or click and stays open until focus leaves", async () => {
    const user = userEvent.setup();
    render(
      <>
        <NoShowWarningIcon count={2} />
        <button type="button">elsewhere</button>
      </>,
    );
    const icon = screen.getByRole("button", { name: /No-show 2 times/ });

    await user.click(icon);
    expect(screen.getByRole("tooltip")).toBeInTheDocument();
    await user.click(icon);
    expect(screen.getByRole("tooltip")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "elsewhere" }));
    expect(screen.queryByRole("tooltip")).not.toBeInTheDocument();
  });

  it("uses singular wording for one no-show", () => {
    render(<NoShowWarningIcon count={1} />);
    expect(screen.getByRole("button", { name: "No-show 1 time at your events in the last 6 months" })).toBeInTheDocument();
  });

  it("is orange", () => {
    render(<NoShowWarningIcon count={2} />);
    expect(screen.getByRole("button")).toHaveClass("text-orange-400");
  });

  it("has no accessibility violations with the tooltip open", async () => {
    const user = userEvent.setup();
    const { container } = render(<NoShowWarningIcon count={2} />);
    await user.hover(screen.getByRole("button"));
    expect(await axe(container)).toHaveNoViolations();
  });
});
