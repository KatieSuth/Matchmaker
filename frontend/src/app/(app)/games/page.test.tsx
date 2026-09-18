import { describe, expect, it } from "vitest";
import { render, screen } from "@/test/render";
import Page from "./page";

describe("Games placeholder page", () => {
  it("renders the placeholder copy", () => {
    render(<Page />);
    expect(screen.getByText("Games")).toBeInTheDocument();
  });
});
