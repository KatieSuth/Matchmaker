// Local-datetime helpers shared by the create/edit event form's date pickers.

export function toDateTimeLocalValue(date: Date): string {
  const pad = (n: number) => n.toString().padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(
    date.getHours(),
  )}:${pad(date.getMinutes())}`;
}

export function roundUpToQuarterHour(date: Date): Date {
  const rounded = new Date(date);
  rounded.setSeconds(0, 0);
  const minutes = rounded.getMinutes();
  const remainder = minutes % 15;
  if (remainder !== 0) rounded.setMinutes(minutes + (15 - remainder));
  return rounded;
}

export function getInitialStartTimeLocal(mode: "create" | "edit"): string {
  if (mode !== "create") return "";
  const now = roundUpToQuarterHour(new Date(Date.now() + 30 * 60 * 1000));
  return toDateTimeLocalValue(now);
}

export function parseLocalDateTimeString(s: string): Date | null {
  if (!s?.trim()) return null;
  const d = new Date(s);
  return Number.isNaN(d.getTime()) ? null : d;
}

export function startOfToday(): Date {
  const d = new Date();
  d.setHours(0, 0, 0, 0);
  return d;
}
