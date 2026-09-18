// Tests for the floating "Copied"/"Copy failed" bubble driven by useCopyStatus.
import { describe, expect, it } from "vitest";
import { axe } from "@/test/axe";
import { render, screen } from "@/test/render";
import { CopyStatusTooltip } from "./CopyStatusTooltip";

describe("CopyStatusTooltip", () => {
  it("renders nothing when idle", () => {
    const { container } = render(<CopyStatusTooltip status="idle" />);
    expect(container).toBeEmptyDOMElement();
  });

  it("renders the default success label", () => {
    render(<CopyStatusTooltip status="success" />);
    expect(screen.getByText("Copied")).toBeInTheDocument();
  });

  it("renders the default error label", () => {
    render(<CopyStatusTooltip status="error" />);
    expect(screen.getByText("Copy failed")).toBeInTheDocument();
  });

  it("renders custom success/error labels when given", () => {
    const { rerender } = render(<CopyStatusTooltip status="success" successLabel="Link copied" />);
    expect(screen.getByText("Link copied")).toBeInTheDocument();

    rerender(<CopyStatusTooltip status="error" errorLabel="Could not copy" />);
    expect(screen.getByText("Could not copy")).toBeInTheDocument();
  });

  it("applies the shadow class by default and omits it when shadow=false", () => {
    const { container, rerender } = render(<CopyStatusTooltip status="success" />);
    expect(container.firstChild).toHaveClass("shadow-[0_10px_24px_rgba(0,0,0,0.45)]");

    rerender(<CopyStatusTooltip status="success" shadow={false} />);
    expect(container.firstChild).not.toHaveClass("shadow-[0_10px_24px_rgba(0,0,0,0.45)]");
  });

  it("has no accessibility violations", async () => {
    const { container } = render(<CopyStatusTooltip status="success" />);
    expect(await axe(container)).toHaveNoViolations();
  });
});
