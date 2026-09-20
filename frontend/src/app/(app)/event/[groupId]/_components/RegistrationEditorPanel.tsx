"use client";

// Register / edit-registration flow, split into the three pieces the page renders around its
// error banner and games list: the open/close toggle button, the form itself, and the save footer.
import { Dispatch, SetStateAction } from "react";
import { LobbyHostInfoHint } from "@/app/_components/LobbyHostInfoHint";
import { ToggleRow } from "@/app/_components/ToggleRow";
import { ToggleSwitch } from "@/app/_components/ToggleSwitch";
import { UserGameEditor, UserGameEditorValue } from "@/app/_components/forms/UserGameEditor";
import { inputCls } from "@/app/_lib/styles";
import { EventGroupDetail, GameRank } from "@/app/_types/types";
import { formatDateTime } from "../_lib/formatters";
import { DEFAULT_EVENT_REGISTRATION_DRAFT, RegistrationDraft } from "../_types";

export function RegistrationToggleButton({
  group,
  registrationEditorOpen,
  working,
  myRegistrationsCount,
  onOpen,
  onClose,
}: {
  group: EventGroupDetail;
  registrationEditorOpen: boolean;
  working: boolean;
  myRegistrationsCount: number;
  onOpen: () => void;
  onClose: () => void;
}) {
  return (
    <div className="flex justify-center">
      {registrationEditorOpen ? (
        <button
          type="button"
          onClick={onClose}
          disabled={working}
          className="rounded-lg border border-[var(--color-text-danger)]/40 bg-[var(--color-text-danger)]/10 px-5 py-2.5 text-sm font-medium text-[var(--color-text-danger)] hover:bg-[var(--color-text-danger)]/20 disabled:opacity-40"
        >
          Cancel Registration
        </button>
      ) : (
        <span
          title={!group.registration_open ? "Registration is closed" : undefined}
          className={[
            "inline-flex rounded-lg",
            !group.registration_open ? "cursor-not-allowed" : "",
          ]
            .filter(Boolean)
            .join(" ")}
        >
          <button
            type="button"
            disabled={!group.registration_open}
            onClick={onOpen}
            className={[
              "rounded-lg border px-5 py-2.5 text-sm font-medium",
              group.registration_open
                ? "border-[var(--color-accent-blue)]/35 bg-[var(--color-accent-blue)]/10 text-[var(--color-accent-blue)] hover:bg-[var(--color-accent-blue)]/20"
                : "pointer-events-none border-white/10 bg-white/[0.03] text-[var(--color-text-muted)] opacity-50",
            ].join(" ")}
          >
            {myRegistrationsCount > 0 ? "Edit My Registration" : "Register Now"}
          </button>
        </span>
      )}
    </div>
  );
}

export function RegistrationEditorForm({
  group,
  regionMismatchWarning,
  myRegistrationsCount,
  userGameDraft,
  setUserGameDraft,
  userGameRanks,
  registrationLoading,
  userGameErrors,
  registrationDraft,
  setRegistrationDraft,
  selectedValidEventIds,
  canDeleteAllViaSave,
  registrationError,
}: {
  group: EventGroupDetail;
  regionMismatchWarning: string | null;
  myRegistrationsCount: number;
  userGameDraft: UserGameEditorValue;
  setUserGameDraft: (value: UserGameEditorValue) => void;
  userGameRanks: GameRank[];
  registrationLoading: boolean;
  userGameErrors: { in_game_name?: string; current_rank?: string; peak_rank?: string };
  registrationDraft: RegistrationDraft;
  setRegistrationDraft: Dispatch<SetStateAction<RegistrationDraft>>;
  selectedValidEventIds: string[];
  canDeleteAllViaSave: boolean;
  registrationError: string | null;
}) {
  return (
    <div className="card rounded-xl p-4 sm:p-5 flex flex-col gap-4 relative overflow-hidden">
      <div className="absolute top-0 left-4 right-4 h-px bg-top-edge opacity-20 rounded-full" />
      <h2 className="text-sm font-semibold text-[var(--color-text)]">
        {myRegistrationsCount > 0 ? "Edit registration" : "Register"}
      </h2>
      {regionMismatchWarning && (
        <p className="text-xs text-amber-300">{regionMismatchWarning}</p>
      )}
      <UserGameEditor
        hideGameSelector
        gameLabel={group.game_name}
        value={userGameDraft}
        ranks={userGameRanks}
        ranksLoading={registrationLoading}
        errors={userGameErrors}
        onChange={(next) => setUserGameDraft(next)}
      />

      <div className="flex flex-col gap-4">
        <p className="text-sm text-[var(--color-text-soft)]">Choose games to register for</p>
        {group.events.map((event, index) => {
          const checked = registrationDraft.selected_event_ids.includes(event.id);
          const settings = registrationDraft.per_event[event.id] ?? DEFAULT_EVENT_REGISTRATION_DRAFT;
          return (
            <div key={event.id} className="flex flex-col gap-3">
              <div className="flex items-start gap-3 select-none">
                <ToggleSwitch
                  checked={checked}
                  ariaLabel={`Register for Game ${index + 1} · ${event.game_mode_name}`}
                  onChange={(nextChecked) => {
                    setRegistrationDraft((prev) => {
                      if (nextChecked) {
                        const nextIds = prev.selected_event_ids.includes(event.id)
                          ? prev.selected_event_ids
                          : [...prev.selected_event_ids, event.id];
                        return {
                          ...prev,
                          selected_event_ids: nextIds,
                          per_event: {
                            ...prev.per_event,
                            [event.id]: prev.per_event[event.id] ?? DEFAULT_EVENT_REGISTRATION_DRAFT,
                          },
                        };
                      }
                      return {
                        ...prev,
                        selected_event_ids: prev.selected_event_ids.filter((id) => id !== event.id),
                      };
                    });
                  }}
                  className="mt-0.5"
                />
                <span className="text-sm font-semibold text-[var(--color-text)] leading-snug pt-0.5">
                  Game {index + 1} · {event.game_mode_name} · {formatDateTime(event.start_time)}
                </span>
              </div>
              <div className="ml-14 rounded-lg border border-white/[0.06] bg-white/[0.02] p-3 flex flex-col gap-2">
                <ToggleRow
                  label="Can substitute"
                  checked={settings.can_substitute}
                  disabled={!checked}
                  onChange={(val) =>
                    setRegistrationDraft((prev) => ({
                      ...prev,
                      per_event: {
                        ...prev.per_event,
                        [event.id]: {
                          ...(prev.per_event[event.id] ?? DEFAULT_EVENT_REGISTRATION_DRAFT),
                          can_substitute: val,
                        },
                      },
                    }))
                  }
                />
                <ToggleRow
                  label="Can lobby host"
                  labelAccessory={<LobbyHostInfoHint />}
                  checked={settings.can_lobby_host}
                  disabled={!checked}
                  onChange={(val) =>
                    setRegistrationDraft((prev) => ({
                      ...prev,
                      per_event: {
                        ...prev.per_event,
                        [event.id]: {
                          ...(prev.per_event[event.id] ?? DEFAULT_EVENT_REGISTRATION_DRAFT),
                          can_lobby_host: val,
                        },
                      },
                    }))
                  }
                />
              </div>
            </div>
          );
        })}
      </div>

      <div className="flex flex-col gap-1.5">
        <label htmlFor="registration-duo-request" className="text-sm text-[var(--color-text-soft)]">Duo Request (They must list you here too. Applies to each selected event. Cannot be guaranteed)</label>
        <input
          id="registration-duo-request"
          className={inputCls}
          placeholder="Discord Name"
          value={registrationDraft.duo_request}
          onChange={(event) =>
            setRegistrationDraft((prev) => ({ ...prev, duo_request: event.target.value }))
          }
        />
      </div>
      {selectedValidEventIds.length === 0 && !canDeleteAllViaSave && (
        <p className="text-xs text-[var(--color-text-danger)]">Select at least one event to save your registration.</p>
      )}
      {registrationError && (
        <p className="text-xs text-[var(--color-text-danger)]">{registrationError}</p>
      )}
    </div>
  );
}

export function RegistrationSaveFooter({
  working,
  canDeleteAllViaSave,
  hasUserGameErrors,
  canSubmitRegistration,
  onSave,
}: {
  working: boolean;
  canDeleteAllViaSave: boolean;
  hasUserGameErrors: boolean;
  canSubmitRegistration: boolean;
  onSave: () => void;
}) {
  return (
    <div className="w-full flex justify-end pt-2 pb-1 border-t border-white/[0.08]">
      <div className="flex flex-col items-end gap-2">
        {!canDeleteAllViaSave && hasUserGameErrors && (
          <p className="text-xs text-[var(--color-text-muted)]">
            Save is disabled until your in-game name, current rank, and peak rank are filled out.
          </p>
        )}
        <button
          type="button"
          onClick={() => {
            if (!canSubmitRegistration) return;
            onSave();
          }}
          disabled={!canSubmitRegistration}
          aria-disabled={!canSubmitRegistration}
          className={[
            "rounded-lg border px-5 py-2.5 text-sm font-medium",
            canSubmitRegistration
              ? canDeleteAllViaSave
                ? "border-[var(--color-text-danger)]/40 bg-[var(--color-text-danger)]/10 text-[var(--color-text-danger)] hover:bg-[var(--color-text-danger)]/20"
                : "border-[var(--color-accent-blue)]/35 bg-[var(--color-accent-blue)]/10 text-[var(--color-accent-blue)] hover:bg-[var(--color-accent-blue)]/20"
              : "cursor-not-allowed border-white/10 bg-white/[0.03] text-[var(--color-text-muted)]",
          ].join(" ")}
        >
          {working ? (canDeleteAllViaSave ? "Deleting..." : "Saving...") : canDeleteAllViaSave ? "Delete Registration" : "Save Registration"}
        </button>
      </div>
    </div>
  );
}
