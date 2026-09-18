// Shared local types for the My Events dashboard.
import { Event } from "@/app/_types/types";

export interface BucketState {
  events: Event[];
  nextCursor: string | null;
  hasMore: boolean;
  isLoadingMore: boolean;
  loaded: boolean;
}

export type Tab = "hosting" | "registered";
export type TimeFilter = "upcoming" | "past";
export type BucketKey = `${Tab}:${TimeFilter}`;

export const EMPTY_BUCKET: BucketState = {
  events: [],
  nextCursor: null,
  hasMore: false,
  isLoadingMore: false,
  loaded: false,
};
