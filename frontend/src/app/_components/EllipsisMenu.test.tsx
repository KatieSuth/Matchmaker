// Tests for the reusable vertical-ellipsis contextual-action dropdown.
import { describe, expect, it, vi } from "vitest";
import { axe } from "@/test/axe";
import { render, screen, userEvent } from "@/test/render";
import { EllipsisMenu } from "./EllipsisMenu";

describe("EllipsisMenu", () => {
  it("renders nothing when there are no enabled options", () => {
    const { container } = render(<EllipsisMenu options={[{ label: "Delete", onSelect: vi.fn(), disabled: true }]} />);
    expect(container).toBeEmptyDOMElement();
  });

  it("hides the menu until the trigger is clicked", async () => {
    const user = userEvent.setup();
    render(<EllipsisMenu options={[{ label: "Delete", onSelect: vi.fn() }]} />);

    expect(screen.queryByRole("menu")).not.toBeInTheDocument();
    expect(screen.queryByText("Delete")).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "More actions" }));

    expect(screen.getByText("Delete")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "More actions" })).toHaveAttribute("aria-expanded", "true");
  });

  it("filters out disabled options but still shows enabled ones", async () => {
    const user = userEvent.setup();
    render(
      <EllipsisMenu
        options={[
          { label: "Edit", onSelect: vi.fn() },
          { label: "Delete", onSelect: vi.fn(), disabled: true },
        ]}
      />,
    );

    await user.click(screen.getByRole("button", { name: "More actions" }));

    expect(screen.getByText("Edit")).toBeInTheDocument();
    expect(screen.queryByText("Delete")).not.toBeInTheDocument();
  });

  it("calls onSelect and closes the menu when an option is clicked", async () => {
    const onSelect = vi.fn();
    const user = userEvent.setup();
    render(<EllipsisMenu options={[{ label: "Delete", onSelect }]} />);
    await user.click(screen.getByRole("button", { name: "More actions" }));

    await user.click(screen.getByText("Delete"));

    expect(onSelect).toHaveBeenCalledTimes(1);
    expect(screen.queryByText("Delete")).not.toBeInTheDocument();
  });

  it("closes when clicking outside the menu", async () => {
    const user = userEvent.setup();
    render(
      <div>
        <EllipsisMenu options={[{ label: "Delete", onSelect: vi.fn() }]} />
        <button>Outside</button>
      </div>,
    );
    await user.click(screen.getByRole("button", { name: "More actions" }));
    expect(screen.getByText("Delete")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Outside" }));

    expect(screen.queryByText("Delete")).not.toBeInTheDocument();
  });

  it("closes on Escape", async () => {
    const user = userEvent.setup();
    render(<EllipsisMenu options={[{ label: "Delete", onSelect: vi.fn() }]} />);
    await user.click(screen.getByRole("button", { name: "More actions" }));
    expect(screen.getByText("Delete")).toBeInTheDocument();

    await user.keyboard("{Escape}");

    expect(screen.queryByText("Delete")).not.toBeInTheDocument();
  });

  it("applies a danger tone class to danger-flagged options", async () => {
    const user = userEvent.setup();
    render(<EllipsisMenu options={[{ label: "Delete", onSelect: vi.fn(), tone: "danger" }]} />);
    await user.click(screen.getByRole("button", { name: "More actions" }));

    expect(screen.getByText("Delete")).toHaveClass("text-[var(--color-text-danger)]");
  });

  it("uses a custom ariaLabel for the trigger button", () => {
    render(<EllipsisMenu options={[{ label: "Delete", onSelect: vi.fn() }]} ariaLabel="Lobby actions" />);
    expect(screen.getByRole("button", { name: "Lobby actions" })).toBeInTheDocument();
  });

  it("has no accessibility violations when open", async () => {
    const user = userEvent.setup();
    const { container } = render(<EllipsisMenu options={[{ label: "Delete", onSelect: vi.fn() }]} />);
    await user.click(screen.getByRole("button", { name: "More actions" }));

    expect(await axe(container)).toHaveNoViolations();
  });
});
