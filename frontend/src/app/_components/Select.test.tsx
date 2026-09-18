// Tests for the themed react-select wrapper (single Select + MultiSelect).
import { describe, expect, it, vi } from "vitest";
import { axe } from "@/test/axe";
import { render, screen, userEvent } from "@/test/render";
import { MultiSelect, Select, SelectOption } from "./Select";

const options: SelectOption[] = [
  { value: "na", label: "North America" },
  { value: "eu", label: "Europe" },
  { value: "apac", label: "Asia-Pacific" },
];

describe("Select", () => {
  it("shows the placeholder when no value is selected", () => {
    render(<Select value="" onChange={vi.fn()} options={options} placeholder="Choose a region" />);
    expect(screen.getByText("Choose a region")).toBeInTheDocument();
  });

  it("shows the label for the currently selected value", () => {
    render(<Select value="eu" onChange={vi.fn()} options={options} />);
    expect(screen.getByText("Europe")).toBeInTheDocument();
  });

  it("lists every option and calls onChange with the picked value", async () => {
    const onChange = vi.fn();
    const user = userEvent.setup();
    render(<Select value="" onChange={onChange} options={options} />);

    await user.click(screen.getByRole("combobox"));
    expect(screen.getByText("North America")).toBeInTheDocument();
    expect(screen.getByText("Asia-Pacific")).toBeInTheDocument();
    await user.click(screen.getByText("Europe"));

    expect(onChange).toHaveBeenCalledWith("eu");
  });

  it("disables the control when disabled is set", () => {
    render(<Select value="na" onChange={vi.fn()} options={options} disabled />);
    // react-select removes the input from the tab order and marks the control read-only when disabled.
    expect(screen.getByRole("combobox")).toHaveAttribute("aria-readonly", "true");
  });

  it("associates a visible label via inputId", () => {
    render(
      <>
        <label htmlFor="region-select">Region</label>
        <Select inputId="region-select" value="" onChange={vi.fn()} options={options} />
      </>,
    );
    expect(screen.getByRole("combobox", { name: "Region" })).toBeInTheDocument();
  });

  it("has no accessibility violations when labeled", async () => {
    const { container } = render(
      <>
        <label htmlFor="region-select">Region</label>
        <Select inputId="region-select" value="" onChange={vi.fn()} options={options} />
      </>,
    );
    expect(await axe(container)).toHaveNoViolations();
  });
});

describe("MultiSelect", () => {
  it("renders the labels for every selected value", () => {
    render(<MultiSelect value={["na", "eu"]} onChange={vi.fn()} options={options} />);
    expect(screen.getByText("North America")).toBeInTheDocument();
    expect(screen.getByText("Europe")).toBeInTheDocument();
    expect(screen.queryByText("Asia-Pacific")).not.toBeInTheDocument();
  });

  it("shows the placeholder when nothing is selected", () => {
    render(<MultiSelect value={[]} onChange={vi.fn()} options={options} placeholder="Pick regions" />);
    expect(screen.getByText("Pick regions")).toBeInTheDocument();
  });

  it("adds a value to the selection when an option is picked", async () => {
    const onChange = vi.fn();
    const user = userEvent.setup();
    render(<MultiSelect value={["na"]} onChange={onChange} options={options} />);

    await user.click(screen.getByRole("combobox"));
    await user.click(screen.getByText("Europe"));

    expect(onChange).toHaveBeenCalledWith(["na", "eu"]);
  });

  it("removes a value from the selection when its chip's remove button is clicked", async () => {
    const onChange = vi.fn();
    const user = userEvent.setup();
    render(<MultiSelect value={["na", "eu"]} onChange={onChange} options={options} />);

    await user.click(screen.getByRole("button", { name: "Remove North America" }));

    expect(onChange).toHaveBeenCalledWith(["eu"]);
  });
});
