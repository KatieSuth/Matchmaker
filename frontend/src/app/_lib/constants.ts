// Allowed event region values; keep in sync with how hosts pick regions in forms.
export const REGIONS = ["NA", "EMEA", "APAC"] as const;

export type Region = (typeof REGIONS)[number];
