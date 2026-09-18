// Static "About" article body: what Matchmaker does, hosting, joining, and fair lobbies.
import Link from "next/link";
import AboutCta from "@/app/about/AboutCta";
import { SectionDivider } from "@/app/_components/SectionDivider";
import { bodyText } from "@/app/about/_lib/text";

export default function AboutIntro() {
  return (
    <>
      <p className={`${bodyText} mb-8`}>
        Matchmaker is a free, open-source way to run custom competitive games over Discord.
        Hosts set up an event, players sign up with their in-game ranks, and the app builds
        fair two-team lobbies, so organizing a custom with a large pool of players takes seconds
        instead of hours. Valorant and League of Legends are supported today, and user-defined
        games are coming soon. Matchmaker isn&apos;t endorsed by Riot Games and doesn&apos;t reflect the
        views or opinions of Riot Games or anyone officially involved in producing or managing
        Riot Games properties. Riot Games, and all associated properties are trademarks or
        registered trademarks of Riot Games, Inc. Matchmaker isn&apos;t endorsed by or affiliated with 
        any other game developer or publisher, for that matter.
      </p>

      <section className="flex flex-col gap-3 mb-8">
        <SectionDivider title="Host an event" />
        <p className={bodyText}>
          From{" "}
          <Link href="/my_events" className="body-link">
            Events
          </Link>
          , a host creates an event with a name, game and mode, region, schedule, a
          substitute minimum, and whether registration is open. Hosts can optionally lock
          an event to one or more Discord servers they belong to; only members of at least
          one selected server can open the event or register. Turning the lock off or
          saving with no servers selected removes the lock. The first lobby can be created
          as soon as there are enough players to fill its teams. Before Matchmaker creates a
          second lobby, there must also be enough players who volunteered to substitute to
          meet the host&apos;s minimum, with at least one volunteer still able to play so
          extra lobbies do not lock every sub on the bench. Volunteering to sub is optional and does
          not affect a player&apos;s chance of being placed on a team. The host also chooses
          how Matchmaker should build lobbies:
        </p>
        <div className="grid gap-3 sm:grid-cols-2">
          <div className="rounded-lg border border-white/10 bg-white/[0.02] p-4 flex flex-col gap-1">
            <h3 className="text-sm font-medium text-[var(--color-text-soft)]">Balanced</h3>
            <p className="text-xs leading-relaxed text-[var(--color-text-muted)]">
              The default. Puts similar ranks on opposite teams using one rank band per
              player slot, so each team gets a matching mix. Best for casual games.
            </p>
          </div>
          <div className="rounded-lg border border-white/10 bg-white/[0.02] p-4 flex flex-col gap-1">
            <h3 className="text-sm font-medium text-[var(--color-text-soft)]">Rank Grouping</h3>
            <p className="text-xs leading-relaxed text-[var(--color-text-muted)]">
              Keeps players of similar rank in the same lobby. Best for serious practice.
            </p>
          </div>
        </div>
      </section>

      <section className="flex flex-col gap-3 mb-8">
        <SectionDivider title="Join an event" />
        <p className={bodyText}>
          When a host sends you an event link, you can open the event page to review the
          details and register. If the host locked the event to selected Discord servers,
          Matchmakerchecks your Discord server membership when you open the page and again
          when you register. Leaving every required server while the tab stays open means
          you cannot register. Before registering, you can review the in-game name and
          competitive rank saved for that game and update them there or in{" "}
          <Link href="/my_account" className="body-link">
            My Account
          </Link>
          . Hosts always see ranks so they can run a fair custom. Other players only see your
          rank if you choose to show it. Pronouns follow the same host-only or public setting.
          You can also volunteer to be a substitute or lobby host.
        </p>
        <p className={bodyText}>
          Each lobby needs a lobby host to create the custom game and invite the players assigned
          to each team. Matchmaker gives preference to volunteers, but if no one in your lobby
          volunteered, you may be assigned anyway. If you cannot take that role, ask the event
          creator to choose someone else.
        </p>
        <p className={bodyText}>
          You can return to the event page and change your registration while registration
          remains open. If it has closed, contact the event creator about any changes you
          need. Note that if you are registered for multiple events and your game rank changes,
          you can update it in{" "}
          <Link href="/my_account" className="body-link">
            My Account
          </Link>
          {" "}and the change will be reflected in all your events; however, if the event host has
          already closed registration and created teams, you will need to contact them to ask
          them to re-run the lobby builder to create new teams with your updated rank.
        </p>
      </section>

      <section className="flex flex-col gap-3 mb-10">
        <SectionDivider title="Fair lobbies" />
        <p className={bodyText}>
          Hosts can lock in whenever they are ready, whether that is the day of the event or
          weeks in advance. Matchmaker then builds lobbies using the mode they chose: two
          teams per lobby, plus substitutes if the event asked for them. If two players
          request each other as a duo, Matchmaker tries to keep them together. Fairness
          comes first, so duos are not guaranteed. After lock-in, hosts can still swap
          players, assign lobby hosts, and set join codes.
        </p>
        <p className={bodyText}>
          Once teams exist, the host can copy a Discord ping message that mentions every player
          on a team and in the substitute pool, ready to paste into Discord as a reminder.
        </p>
      </section>

      <div className="flex justify-center">
        <AboutCta />
      </div>
    </>
  );
}
