// Tests for the register/edit-registration flow's toggle button, form, and save footer.
import { Dispatch, SetStateAction } from "react";
import { fireEvent } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { axe } from "@/test/axe";
import { buildEventGroupDetail, buildEventGroupEvent, buildGameRank } from "@/test/fixtures";
import { RegistrationDraft } from "../_types";
import { render, screen, userEvent } from "@/test/render";
import {
  RegistrationEditorForm,
  RegistrationSaveFooter,
  RegistrationToggleButton,
} from "./RegistrationEditorPanel";

describe("RegistrationToggleButton", () => {
  function baseProps(overrides: Partial<Parameters<typeof RegistrationToggleButton>[0]> = {}) {
    return {
      group: buildEventGroupDetail({ registration_open: true }),
      registrationEditorOpen: false,
      working: false,
      myRegistrationsCount: 0,
      onOpen: vi.fn(),
      onClose: vi.fn(),
      ...overrides,
    };
  }

  it("shows 'Register Now' when the viewer has no existing registrations", () => {
    render(<RegistrationToggleButton {...baseProps({ myRegistrationsCount: 0 })} />);
    expect(screen.getByRole("button", { name: "Register Now" })).toBeInTheDocument();
  });

  it("shows 'Edit My Registration' when the viewer already has registrations", () => {
    render(<RegistrationToggleButton {...baseProps({ myRegistrationsCount: 1 })} />);
    expect(screen.getByRole("button", { name: "Edit My Registration" })).toBeInTheDocument();
  });

  it("disables opening when registration is closed", () => {
    const group = buildEventGroupDetail({ registration_open: false });
    render(<RegistrationToggleButton {...baseProps({ group })} />);
    expect(screen.getByRole("button", { name: "Register Now" })).toBeDisabled();
  });

  it("calls onOpen when clicked while registration is open", async () => {
    const onOpen = vi.fn();
    const user = userEvent.setup();
    render(<RegistrationToggleButton {...baseProps({ onOpen })} />);

    await user.click(screen.getByRole("button", { name: "Register Now" }));

    expect(onOpen).toHaveBeenCalledTimes(1);
  });

  it("shows 'Cancel Registration' and calls onClose when the editor is open", async () => {
    const onClose = vi.fn();
    const user = userEvent.setup();
    render(<RegistrationToggleButton {...baseProps({ registrationEditorOpen: true, onClose })} />);

    await user.click(screen.getByRole("button", { name: "Cancel Registration" }));

    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("disables Cancel Registration while working", () => {
    render(<RegistrationToggleButton {...baseProps({ registrationEditorOpen: true, working: true })} />);
    expect(screen.getByRole("button", { name: "Cancel Registration" })).toBeDisabled();
  });

  it("has no accessibility violations", async () => {
    const { container } = render(<RegistrationToggleButton {...baseProps()} />);
    expect(await axe(container)).toHaveNoViolations();
  });
});

describe("RegistrationEditorForm", () => {
  function baseDraft(overrides: Partial<RegistrationDraft> = {}): RegistrationDraft {
    return {
      selected_event_ids: [],
      per_event: {},
      duo_request: "",
      ...overrides,
    };
  }

  function baseProps(overrides: Partial<Parameters<typeof RegistrationEditorForm>[0]> = {}) {
    return {
      group: buildEventGroupDetail({ events: [buildEventGroupEvent({ id: "event-1" })] }),
      regionMismatchWarning: null,
      myRegistrationsCount: 0,
      userGameDraft: { game_id: "game-1", in_game_name: "", current_rank: "", peak_rank: "", show_rank: false },
      setUserGameDraft: vi.fn(),
      userGameRanks: [buildGameRank({ name: "Gold" })],
      registrationLoading: false,
      userGameErrors: {},
      registrationDraft: baseDraft(),
      setRegistrationDraft: vi.fn(),
      selectedValidEventIds: [],
      canDeleteAllViaSave: false,
      registrationError: null,
      ...overrides,
    };
  }

  it("shows 'Register' as the heading when the viewer has no registrations, 'Edit registration' otherwise", () => {
    const { rerender } = render(<RegistrationEditorForm {...baseProps({ myRegistrationsCount: 0 })} />);
    expect(screen.getByRole("heading", { name: "Register" })).toBeInTheDocument();

    rerender(<RegistrationEditorForm {...baseProps({ myRegistrationsCount: 1 })} />);
    expect(screen.getByRole("heading", { name: "Edit registration" })).toBeInTheDocument();
  });

  it("shows the region mismatch warning when provided", () => {
    render(<RegistrationEditorForm {...baseProps({ regionMismatchWarning: "Your region is EU but this event is AMER." })} />);
    expect(screen.getByText("Your region is EU but this event is AMER.")).toBeInTheDocument();
  });

  it("lists a toggle per event, checked according to the draft's selected_event_ids", () => {
    const events = [buildEventGroupEvent({ id: "event-1", game_mode_name: "5v5" }), buildEventGroupEvent({ id: "event-2", game_mode_name: "3v3" })];
    const group = buildEventGroupDetail({ events });
    render(<RegistrationEditorForm {...baseProps({ group, registrationDraft: baseDraft({ selected_event_ids: ["event-1"] }) })} />);

    expect(screen.getByRole("switch", { name: /Register for Game 1/ })).toBeChecked();
    expect(screen.getByRole("switch", { name: /Register for Game 2/ })).not.toBeChecked();
  });

  it("toggling an event on adds it to selected_event_ids with default per-event settings", async () => {
    const setRegistrationDraft = vi.fn();
    const events = [buildEventGroupEvent({ id: "event-1" })];
    const group = buildEventGroupDetail({ events });
    const user = userEvent.setup();
    render(<RegistrationEditorForm {...baseProps({ group, setRegistrationDraft })} />);

    await user.click(screen.getByRole("switch", { name: /Register for Game 1/ }));

    const updater = setRegistrationDraft.mock.calls[0][0];
    const result = updater(baseDraft());
    expect(result.selected_event_ids).toEqual(["event-1"]);
    expect(result.per_event["event-1"]).toEqual({ can_substitute: true, can_lobby_host: true });
  });

  it("shows Can lobby host on for a selected event using new-registration defaults", () => {
    const events = [buildEventGroupEvent({ id: "event-1" })];
    const group = buildEventGroupDetail({ events });
    render(
      <RegistrationEditorForm
        {...baseProps({
          group,
          registrationDraft: baseDraft({
            selected_event_ids: ["event-1"],
            per_event: { "event-1": { can_substitute: true, can_lobby_host: true } },
          }),
        })}
      />,
    );

    expect(screen.getByRole("switch", { name: "Can lobby host" })).toBeChecked();
  });

  it("disables the per-event Can substitute / Can lobby host toggles until the event is selected", () => {
    const events = [buildEventGroupEvent({ id: "event-1" })];
    const group = buildEventGroupDetail({ events });
    render(<RegistrationEditorForm {...baseProps({ group, registrationDraft: baseDraft() })} />);

    expect(screen.getByRole("switch", { name: "Can substitute" })).toBeDisabled();
    expect(screen.getByRole("switch", { name: "Can lobby host" })).toBeDisabled();
  });

  it("updates duo_request as the input changes", () => {
    // The onChange handler reads `event.target.value` lazily inside its updater closure, and React
    // restores a controlled input's DOM value back to its (here, static/mocked) `value` prop right
    // after the event finishes dispatching — so extracting the updater from `mock.calls` and
    // invoking it *after* `fireEvent.change` returns would read the already-reverted value instead.
    // Applying the updater synchronously inside the mock (i.e. still within the event dispatch)
    // captures the real intended value instead.
    let captured: string | undefined;
    const setRegistrationDraft: Dispatch<SetStateAction<RegistrationDraft>> = vi.fn((updater) => {
      captured =
        typeof updater === "function"
          ? (updater as (prev: RegistrationDraft) => RegistrationDraft)(baseDraft()).duo_request
          : updater.duo_request;
    });
    render(<RegistrationEditorForm {...baseProps({ setRegistrationDraft })} />);

    fireEvent.change(screen.getByPlaceholderText("Discord Name"), { target: { value: "x" } });

    expect(captured).toBe("x");
  });

  it("shows a validation message when no events are selected and it's not a delete-all save", () => {
    render(<RegistrationEditorForm {...baseProps({ selectedValidEventIds: [], canDeleteAllViaSave: false })} />);
    expect(screen.getByText("Select at least one event to save your registration.")).toBeInTheDocument();
  });

  it("hides the validation message when canDeleteAllViaSave is true even with no selected events", () => {
    render(<RegistrationEditorForm {...baseProps({ selectedValidEventIds: [], canDeleteAllViaSave: true })} />);
    expect(screen.queryByText("Select at least one event to save your registration.")).not.toBeInTheDocument();
  });

  it("shows the registrationError message when present", () => {
    render(<RegistrationEditorForm {...baseProps({ registrationError: "Event full." })} />);
    expect(screen.getByText("Event full.")).toBeInTheDocument();
  });

  it("has no accessibility violations", async () => {
    const events = [buildEventGroupEvent({ id: "event-1", game_mode_name: "5v5" })];
    const group = buildEventGroupDetail({ events });
    const { container } = render(
      <RegistrationEditorForm
        {...baseProps({
          group,
          registrationDraft: baseDraft({ selected_event_ids: ["event-1"] }),
        })}
      />,
    );
    expect(await axe(container)).toHaveNoViolations();
  });
});

describe("RegistrationSaveFooter", () => {
  function baseProps(overrides: Partial<Parameters<typeof RegistrationSaveFooter>[0]> = {}) {
    return {
      working: false,
      canDeleteAllViaSave: false,
      hasUserGameErrors: false,
      canSubmitRegistration: true,
      onSave: vi.fn(),
      ...overrides,
    };
  }

  it("shows 'Save Registration' by default", () => {
    render(<RegistrationSaveFooter {...baseProps()} />);
    expect(screen.getByRole("button", { name: "Save Registration" })).toBeInTheDocument();
  });

  it("shows 'Delete Registration' when canDeleteAllViaSave is true", () => {
    render(<RegistrationSaveFooter {...baseProps({ canDeleteAllViaSave: true })} />);
    expect(screen.getByRole("button", { name: "Delete Registration" })).toBeInTheDocument();
  });

  it("shows 'Saving...' / 'Deleting...' while working", () => {
    const { rerender } = render(<RegistrationSaveFooter {...baseProps({ working: true })} />);
    expect(screen.getByRole("button", { name: "Saving..." })).toBeInTheDocument();

    rerender(<RegistrationSaveFooter {...baseProps({ working: true, canDeleteAllViaSave: true })} />);
    expect(screen.getByRole("button", { name: "Deleting..." })).toBeInTheDocument();
  });

  it("disables the button and shows a hint when there are unresolved game-profile errors", () => {
    render(<RegistrationSaveFooter {...baseProps({ canSubmitRegistration: false, hasUserGameErrors: true })} />);
    expect(screen.getByRole("button", { name: "Save Registration" })).toBeDisabled();
    expect(screen.getByText(/Save is disabled until your in-game name/)).toBeInTheDocument();
  });

  it("calls onSave when clicked while submittable", async () => {
    const onSave = vi.fn();
    const user = userEvent.setup();
    render(<RegistrationSaveFooter {...baseProps({ onSave })} />);

    await user.click(screen.getByRole("button", { name: "Save Registration" }));

    expect(onSave).toHaveBeenCalledTimes(1);
  });

  it("does not call onSave when disabled", async () => {
    const onSave = vi.fn();
    const user = userEvent.setup();
    render(<RegistrationSaveFooter {...baseProps({ onSave, canSubmitRegistration: false })} />);

    await user.click(screen.getByRole("button", { name: "Save Registration" }));

    expect(onSave).not.toHaveBeenCalled();
  });

  it("has no accessibility violations", async () => {
    const { container } = render(<RegistrationSaveFooter {...baseProps()} />);
    expect(await axe(container)).toHaveNoViolations();
  });
});
