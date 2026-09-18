// Tests for the empty-list placeholder, worded per active tab/time/filter combination.
import { describe, expect, it } from "vitest";
import { axe } from "@/test/axe";
import { render, screen } from "@/test/render";
import { EmptyState } from "./EmptyState";

describe("EmptyState", () => {
  it("shows a filter-specific message when filters are applied, regardless of tab/time", () => {
    render(<EmptyState tab="hosting" time="upcoming" hasFilters={true} />);
    expect(screen.getByText("No events match your filters")).toBeInTheDocument();
  });

  it("shows a generic past-events message for the past time filter", () => {
    render(<EmptyState tab="hosting" time="past" hasFilters={false} />);
    expect(screen.getByText("No past events")).toBeInTheDocument();
  });

  it("shows a hosting-specific message for the upcoming hosting tab", () => {
    render(<EmptyState tab="hosting" time="upcoming" hasFilters={false} />);
    expect(screen.getByText("You're not hosting any upcoming events")).toBeInTheDocument();
  });

  it("shows a registered-specific message for the upcoming registered tab", () => {
    render(<EmptyState tab="registered" time="upcoming" hasFilters={false} />);
    expect(screen.getByText("You're not registered for any upcoming events")).toBeInTheDocument();
  });

  it("has no accessibility violations", async () => {
    const { container } = render(<EmptyState tab="hosting" time="upcoming" hasFilters={false} />);
    expect(await axe(container)).toHaveNoViolations();
  });
});
