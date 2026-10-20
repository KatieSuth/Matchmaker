// Tests for the "Deprioritize no-shows" toggle + absence-limit stepper. Wrapped in a minimal
// react-hook-form harness since the component is driven by RHF `Control`/`setValue`/`getValues`.
import { useForm, useWatch } from "react-hook-form";
import { describe, expect, it } from "vitest";
import { axe } from "@/test/axe";
import { render, screen, userEvent } from "@/test/render";
import { EventFormValues } from "./schema";
import { DEFAULT_MAX_NOSHOWS, NoShowDeprioritizationFields } from "./NoShowDeprioritizationFields";

function Harness({
  deprioritize = false,
  maxNoshows = 0,
  maxNoshowsError,
  readOnly = false,
}: {
  deprioritize?: boolean;
  maxNoshows?: number;
  maxNoshowsError?: string;
  readOnly?: boolean;
}) {
  const { control, setValue, getValues } = useForm<EventFormValues>({
    defaultValues: {
      name: "",
      game_id: "game-1",
      game_mode_id: "mode-1",
      region: "AMER",
      start_time_local: "",
      sub_min: 0,
      games_to_run: 1,
      registration_open: true,
      sort_logic: "balanced",
      discord_lock: false,
      discord_guild_ids: [],
      deprioritize_noshows: deprioritize,
      max_noshows: maxNoshows,
    },
  });
  const watchedDeprioritize = useWatch({ control, name: "deprioritize_noshows" });

  return (
    <NoShowDeprioritizationFields
      control={control}
      setValue={setValue}
      getValues={getValues}
      watchedDeprioritize={watchedDeprioritize}
      maxNoshowsError={maxNoshowsError}
      readOnly={readOnly}
    />
  );
}

const LIMIT_LABEL = "Absences before deprioritizing";

describe("NoShowDeprioritizationFields", () => {
  it("hides the absence limit and explains the off state until toggled on", () => {
    render(<Harness />);
    expect(screen.getByText("Deprioritize no-shows")).toBeInTheDocument();
    expect(screen.getByText("Absences don't affect who gets picked for teams.")).toBeInTheDocument();
    expect(screen.queryByLabelText(LIMIT_LABEL)).not.toBeInTheDocument();
  });

  it("turning it on reveals the limit seeded with the default and the on-state blurb", async () => {
    const user = userEvent.setup();
    render(<Harness />);

    await user.click(screen.getByRole("switch"));

    expect(screen.getByLabelText(LIMIT_LABEL)).toHaveValue(DEFAULT_MAX_NOSHOWS);
    expect(screen.getByText(/last 6 months are picked last when teams are created/)).toBeInTheDocument();
  });

  it("keeps a previously stored limit when turned back on", async () => {
    const user = userEvent.setup();
    render(<Harness maxNoshows={5} />);

    await user.click(screen.getByRole("switch"));

    expect(screen.getByLabelText(LIMIT_LABEL)).toHaveValue(5);
  });

  it("does not let the limit go below 1", async () => {
    const user = userEvent.setup();
    render(<Harness deprioritize maxNoshows={1} />);
    expect(screen.getByRole("button", { name: `Decrease ${LIMIT_LABEL}` })).toBeDisabled();

    await user.click(screen.getByRole("button", { name: `Increase ${LIMIT_LABEL}` }));
    expect(screen.getByLabelText(LIMIT_LABEL)).toHaveValue(2);
  });

  it("hides the limit again when toggled off", async () => {
    const user = userEvent.setup();
    render(<Harness deprioritize maxNoshows={3} />);

    await user.click(screen.getByRole("switch"));

    expect(screen.queryByLabelText(LIMIT_LABEL)).not.toBeInTheDocument();
  });

  it("shows a validation error for the limit", () => {
    render(<Harness deprioritize maxNoshows={1} maxNoshowsError="Absence limit must be a whole number greater than 0." />);
    expect(screen.getByText("Absence limit must be a whole number greater than 0.")).toBeInTheDocument();
  });

  it("disables the controls in read-only mode", () => {
    render(<Harness deprioritize maxNoshows={3} readOnly />);
    expect(screen.getByRole("switch")).toBeDisabled();
    expect(screen.getByLabelText(LIMIT_LABEL)).toBeDisabled();
  });

  it("has no accessibility violations when enabled", async () => {
    const { container } = render(<Harness deprioritize maxNoshows={3} />);
    expect(await axe(container)).toHaveNoViolations();
  });
});
