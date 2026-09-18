"use client";

// Single event row in the My Events list: name/game/mode/region, host, player count, and status.
import Link from "next/link";
import { formatUserDisplayLabel } from "@/app/_lib/userDisplayName";
import { Event } from "@/app/_types/types";
import { formatMyEventsDate } from "../_lib/dateFmt";
import { Badge } from "./Badge";

interface EventCardProps {
  event: Event;
  currentUserId?: string;
  isHostingList: boolean;
  hostingIds: Set<string>;
}

export function EventCard({ event, currentUserId, isHostingList, hostingIds }: EventCardProps) {
  const isAlsoHosting = !isHostingList && hostingIds.has(event.id);
  const isCurrentUserHost = event.host_id === currentUserId;

  return (
    <Link href={`/event/${event.id}`} className="block group focus:outline-none">
      <div
        className={[
          "card rounded-xl p-4 flex flex-col gap-3 relative overflow-hidden",
          "transition-all duration-200",
          "group-hover:border-white/15 group-hover:shadow-[0_0_0_1px_rgba(255,255,255,0.04),0_30px_90px_rgba(0,0,0,0.7),0_0_40px_rgba(30,80,200,0.12)]",
          "group-focus-visible:ring-2 group-focus-visible:ring-[var(--color-accent-blue)]/50",
        ].join(" ")}
      >
        <div className="absolute top-0 left-4 right-4 h-px bg-top-edge opacity-0 group-hover:opacity-20 transition-opacity duration-200 rounded-full" />

        {/* Header row */}
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0 flex-1">
            <p className="font-semibold text-[var(--color-text)] text-sm leading-snug truncate group-hover:text-white transition-colors">
              {event.name.trim() ? event.name : event.game_name}
            </p>
            <p className="text-xs text-[var(--color-text-muted)] mt-0.5 truncate">
              {event.name.trim()
                ? [event.game_name, event.game_mode, event.region].filter(Boolean).join(" · ")
                : [event.game_mode, event.region].filter(Boolean).join(" · ")}
            </p>
          </div>
          <div className="flex items-center gap-1.5 flex-shrink-0">
            {isAlsoHosting && (
              <span className="inline-flex items-center px-2 py-0.5 rounded-full text-[0.65rem] font-semibold tracking-wide bg-[var(--color-accent-blue)]/10 text-[var(--color-accent-blue)] border border-[var(--color-accent-blue)]/20">
                Hosting
              </span>
            )}
            <Badge open={event.registration_open} />
          </div>
        </div>

        {/* Divider */}
        <div className="h-px bg-white/[0.06]" />

        {/* Details grid */}
        <div className="grid grid-cols-3 gap-x-3 gap-y-2">
          <EventDetail
            icon={
              <svg width="11" height="11" viewBox="0 0 16 16" fill="none">
                <rect x="1" y="3" width="14" height="12" rx="2" stroke="currentColor" strokeWidth="1.5" />
                <path d="M1 7h14" stroke="currentColor" strokeWidth="1.5" />
                <path d="M5 1v4M11 1v4" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
              </svg>
            }
            label="Date"
            value={formatMyEventsDate(new Date(event.event_date))}
          />
          <EventDetail
            icon={
              <svg width="11" height="11" viewBox="0 0 16 16" fill="none">
                <circle cx="8" cy="5" r="3" stroke="currentColor" strokeWidth="1.5" />
                <path d="M2 14c0-3.314 2.686-6 6-6s6 2.686 6 6" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
              </svg>
            }
            label="Host"
            value={
              isCurrentUserHost
                ? "You"
                : formatUserDisplayLabel(event.host_display_name, event.host_name)
            }
          />
          <EventDetail
            icon={
              <svg width="11" height="11" viewBox="0 0 16 16" fill="none">
                <circle cx="5" cy="5" r="2.5" stroke="currentColor" strokeWidth="1.5" />
                <circle cx="11" cy="5" r="2.5" stroke="currentColor" strokeWidth="1.5" />
                <path d="M1 14c0-2.761 1.79-5 4-5M15 14c0-2.761-1.79-5-4-5M8 14c0-2.761 1.343-5 3-5s3 2.239 3 5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
              </svg>
            }
            label="Players"
            value={`${event.registered_count}`}
          />
        </div>

        {/* Chevron */}
        <div className="absolute right-4 bottom-4 text-[var(--color-text-faint)] group-hover:text-[var(--color-text-muted)] group-hover:translate-x-0.5 transition-all duration-150">
          <svg width="12" height="12" viewBox="0 0 12 12" fill="none">
            <path d="M4.5 2.5L8 6l-3.5 3.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
          </svg>
        </div>
      </div>
    </Link>
  );
}

function EventDetail({
  icon,
  label,
  value,
}: {
  icon: React.ReactNode;
  label: string;
  value: string;
}) {
  return (
    <div className="flex flex-col gap-0.5 min-w-0">
      <div className="flex items-center gap-1 text-[var(--color-text-faint)]">
        {icon}
        <span className="text-[0.6rem] font-medium tracking-wider uppercase">{label}</span>
      </div>
      <span className="text-xs text-[var(--color-text-soft)] truncate">{value}</span>
    </div>
  );
}
