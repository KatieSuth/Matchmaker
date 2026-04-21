export const REGIONS = ["NA", "EMEA", "APAC"] as const;

export type Region = (typeof REGIONS)[number];
