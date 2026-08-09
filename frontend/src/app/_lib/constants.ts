// Allowed event region values; keep in sync with how hosts pick regions in forms.
export const REGIONS = ["AMER", "EMEA", "APAC"] as const;
export const EMPTY_VALUE = "—";
export const NO_SUBSTITUTES_MESSAGE = "There are no substitutes available";

export type Region = (typeof REGIONS)[number];
