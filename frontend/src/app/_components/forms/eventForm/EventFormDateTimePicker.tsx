"use client";

// Local datetime string (same shape as `datetime-local`) for API + zod; 15-minute steps.
import DatePicker from "react-datepicker";
import { DATEPICKER_PORTAL_ID, inputCls } from "@/app/_lib/styles";
import { parseLocalDateTimeString, startOfToday, toDateTimeLocalValue } from "./dateTime";

export function EventFormDateTimePicker({
  value,
  onChange,
  onBlur,
  name,
  id,
  disallowPast,
  placeholderText = "Select date & time",
  disabled = false,
}: {
  value: string;
  onChange: (v: string) => void;
  onBlur?: () => void;
  name?: string;
  id?: string;
  disallowPast?: boolean;
  placeholderText?: string;
  disabled?: boolean;
}) {
  const selected = parseLocalDateTimeString(value);
  const filterTime = (time: Date) => {
    if (!disallowPast) return true;
    return time.getTime() > Date.now();
  };

  return (
    <DatePicker
      id={id}
      name={name}
      onBlur={onBlur}
      portalId={DATEPICKER_PORTAL_ID}
      selected={selected}
      onChange={(date: Date | null) => {
        onChange(date ? toDateTimeLocalValue(date) : "");
      }}
      showTimeSelect
      timeIntervals={15}
      timeCaption="Time"
      showMonthDropdown
      showYearDropdown
      dropdownMode="select"
      dateFormat="MMM d, yyyy h:mm aa"
      placeholderText={placeholderText}
      className={inputCls}
      minDate={disallowPast ? startOfToday() : undefined}
      filterTime={disallowPast ? filterTime : undefined}
      popperPlacement="bottom-start"
      popperProps={{ strategy: "fixed" }}
      disabled={disabled}
    />
  );
}
