// Tests for the form layout primitive: label + optional hint/error + control slot.
import { describe, expect, it } from "vitest";
import { axe } from "@/test/axe";
import { render, screen } from "@/test/render";
import { Field } from "./Field";

describe("Field", () => {
  it("renders the label and children", () => {
    render(
      <Field label="Display name">
        <input type="text" />
      </Field>,
    );
    expect(screen.getByText("Display name")).toBeInTheDocument();
    expect(screen.getByRole("textbox")).toBeInTheDocument();
  });

  it("renders an optional hint", () => {
    render(
      <Field label="Display name" hint="Shown to other players">
        <input type="text" />
      </Field>,
    );
    expect(screen.getByText("Shown to other players")).toBeInTheDocument();
  });

  it("renders an optional error message", () => {
    render(
      <Field label="Display name" error="Required">
        <input type="text" />
      </Field>,
    );
    expect(screen.getByText("Required")).toBeInTheDocument();
  });

  it("omits hint/error paragraphs when not given", () => {
    render(
      <Field label="Display name">
        <input type="text" />
      </Field>,
    );
    expect(screen.queryByText("Required")).not.toBeInTheDocument();
  });

  it("associates the label with the control via htmlFor", () => {
    render(
      <Field label="Display name" htmlFor="display-name">
        <input id="display-name" type="text" />
      </Field>,
    );
    expect(screen.getByRole("textbox", { name: "Display name" })).toBeInTheDocument();
  });

  it("has no accessibility violations", async () => {
    const { container } = render(
      <Field label="Display name" htmlFor="display-name" hint="Shown to other players">
        <input id="display-name" type="text" />
      </Field>,
    );
    expect(await axe(container)).toHaveNoViolations();
  });
});
