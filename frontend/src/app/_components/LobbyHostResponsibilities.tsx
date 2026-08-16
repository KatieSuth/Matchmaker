/** Shared copy for lobby host duty reminders (registration hint and event page banner). */
export function LobbyHostResponsibilitiesList({ className = "" }: { className?: string }) {
  return (
    <ul className={["list-disc space-y-1 pl-4", className].filter(Boolean).join(" ")}>
      <li>Approximately 10 minutes before game start time, create the lobby and add the join code/URL via the &quot;Join Lobby&quot; link next to your lobby.</li>
      <li>Your event host should tell you if they require special in-game lobby settings. Update the lobby settings accordingly.</li>
      <li>Choose the most balanced server for the majority of players, if applicable to the game.</li>
      <li>Attempt to contact missing players to confirm they are ready to play. If they do not respond in a reasonable time, alert the event host so they can find a substitute.</li>
      <li>Alert the event host of any toxicity occurring during games.</li>
      <li>Contact the event host if you need help or have questions.</li>
    </ul>
  );
}
