// Allowed event region values; keep in sync with how hosts pick regions in forms.
export const REGIONS = ["AMER", "EMEA", "APAC"] as const;
export const EMPTY_VALUE = "—";

export type Region = (typeof REGIONS)[number];
