"use client";

// "Deprioritize no-shows" toggle + the absence-limit stepper that appears when enabled.
import { Control, Controller, UseFormGetValues, UseFormSetValue } from "react-hook-form";
import { NumberStepper } from "@/app/_components/NumberStepper";
import { ToggleRow } from "@/app/_components/ToggleRow";
import { EventFormValues } from "./schema";

/** Suggested absence limit the first time a host turns deprioritization on. */
export const DEFAULT_MAX_NOSHOWS = 3;

interface Props {
  control: Control<EventFormValues>;
  setValue: UseFormSetValue<EventFormValues>;
  getValues: UseFormGetValues<EventFormValues>;
  watchedDeprioritize: boolean;
  maxNoshowsError?: string;
  readOnly: boolean;
}

export function NoShowDeprioritizationFields({
  control,
  setValue,
  getValues,
  watchedDeprioritize,
  maxNoshowsError,
  readOnly,
}: Props) {
  return (
    <div className="pt-1 border-t border-white/[0.06] flex flex-col gap-3">
      <Controller
        name="deprioritize_noshows"
        control={control}
        render={({ field }) => (
          <ToggleRow
            label="Deprioritize no-shows"
            description={
              field.value
                ? "Players marked absent this many times at your events in the last 6 months are picked last when teams are created."
                : "Absences don't affect who gets picked for teams."
            }
            checked={field.value}
            onChange={(next) => {
              field.onChange(next);
              // Off keeps the last limit so turning it back on restores it; only seed a blank one.
              if (next && getValues("max_noshows") < 1) {
                setValue("max_noshows", DEFAULT_MAX_NOSHOWS);
              }
            }}
            disabled={readOnly}
          />
        )}
      />
      {watchedDeprioritize && (
        // Same half-width column as "Minimum subs per lobby" so the stepper does not span the sheet.
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <div className="flex flex-col gap-1">
            <Controller
              name="max_noshows"
              control={control}
              render={({ field }) => (
                <NumberStepper
                  label="Absences before deprioritizing"
                  value={field.value}
                  min={1}
                  onChange={field.onChange}
                  disabled={readOnly}
                />
              )}
            />
            {maxNoshowsError && (
              <p className="text-xs text-[var(--color-text-danger)]">{maxNoshowsError}</p>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
