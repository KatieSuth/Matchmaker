"use client";

import { ResponsiveSheet } from "@/app/_components/ResponsiveSheet";
import { EventForm } from "@/app/_components/forms/EventForm";
import { EventGroupDetail } from "@/app/_types/types";

export function EditEventSheet({
  isOpen,
  onClose,
  isHost,
  group,
  onSubmitted,
}: {
  isOpen: boolean;
  onClose: () => void;
  isHost: boolean;
  group: EventGroupDetail;
  onSubmitted: () => void;
}) {
  return (
    <ResponsiveSheet isOpen={isOpen} onClose={onClose} title={isHost ? "Edit event settings" : "Event settings"}>
      <EventForm
        mode="edit"
        readOnly={!isHost}
        eventGroupId={group.id}
        editSchedule={group.events.map((e) => ({
          id: e.id,
          start_time: e.start_time,
          game_mode_id: e.game_mode_id,
        }))}
        initialValues={{
          name: group.name,
          game_id: group.game_id,
          region: group.region,
          sub_min: group.sub_min,
          registration_open: group.registration_open,
          sort_logic: group.sort_logic,
          discord_lock: (group.discord_guilds?.length ?? 0) > 0,
          discord_guild_ids: (group.discord_guilds ?? []).map((g) => g.id),
        }}
        onCancel={onClose}
        onSubmitted={isHost ? onSubmitted : undefined}
      />
    </ResponsiveSheet>
  );
}
