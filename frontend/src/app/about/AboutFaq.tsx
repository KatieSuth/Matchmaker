import { GITHUB_REPO_URL } from "@/app/_lib/constants";

const bodyText = "font-light leading-[1.8] text-[var(--color-text-body)]";
const questionCls =
  "text-sm font-semibold tracking-wide text-[var(--color-text-soft)]";
const listCls = `${bodyText} list-disc pl-5 flex flex-col gap-2`;

const MATCHMAKING_DOC_URL = `${GITHUB_REPO_URL}/blob/main/backend/internal/matchmaking/MATCHMAKING.md`;

export default function AboutFaq() {
  return (
    <div className="flex flex-col gap-8">
      <div className="flex flex-col gap-2">
        <h2 className={questionCls}>
          Why do I need this when Valorant custom games have an autobalance button?
        </h2>
        <p className={bodyText}>
          The built-in team balancer is a great resource if you have exactly 10 players and just
          don&apos;t know how to configure the teams. If you have a pool of more than 10 players
          and could potentially be running multiple games at once though, it cannot be used to
          fairly determine who should be in which lobby. That said, this isn&apos;t just for
          enormous Discord servers that will have hundreds of players joining at once; it can also be
          used to determine good 2v2 or 3v3 matches from a pool of available competitors for
          Skirmish or other modes.
        </p>
      </div>

      <div className="flex flex-col gap-3">
        <h2 className={questionCls}>How does matchmaking work?</h2>
        <p className={bodyText}>
          When the host clicks <strong>Lock In &amp; Create Teams</strong>, Matchmaker looks at
          everyone signed up for each game and builds teams automatically using the rank
          information players provided (current rank and peak rank). For each player, it calculates
          a single skill number by averaging those two ranks. That number is what gets used for
          sorting and balancing, not just current rank alone.
        </p>
        <p className={bodyText}>
          The host chooses a <strong>matchmaking mode</strong> when creating the event:
        </p>
        <ul className={listCls}>
          <li>
            <strong>Balanced</strong> (default). Best for casual games. Matchmaker splits the rank
            ladder into as many bands as there are players per team, then puts one player from each
            band on each side so the two teams stay even. If more people signed up than can play at
            once, it still fills those bands (typical ranks of each band first) rather than only
            taking the highest-ranked or lowest-ranked players.
          </li>
          <li>
            <strong>Rank Grouping</strong>. Best for serious practice. Matchmaker groups players
            of similar rank into the same lobby. If there are too many players and not everyone
            can play, it keeps whichever skill level most players belong to (for example, if most
            signups are lower rank, lower-ranked players are prioritized for roster spots).
          </li>
        </ul>
        <p className={bodyText}>After teams are formed, a few other things happen automatically:</p>
        <ul className={listCls}>
          <li>
            Players who marked <strong>Can substitute</strong> during sign-up may be placed in a
            sub pool instead of a team if there are more signups than spots.
          </li>
          <li>
            Players who did <strong>not</strong> mark <strong>Can substitute</strong> and did not
            make a team stay signed up but are listed as unplaced for that game, and will not be
            included in the Discord Pings list for that game.
          </li>
          <li>
            If enough players signed up, multiple lobbies can be created so more than one game can
            run at once. Matchmaker opens as many as the substitute minimum allows. It only skips
            an extra lobby when filling it would bench every substitute volunteer and the remaining
            players still cannot form even teams. If those teams would already be even, the extra
            lobby stays.
          </li>
          <li>
            One player per lobby is assigned as lobby host (whoever volunteered first, or the
            earliest sign-up if no one volunteered).
          </li>
          <li>
            <strong>Duo requests</strong>. If two players list each other&apos;s Discord username as
            their duo partner, Matchmaker tries to put them in the same lobby and on the same
            team. Balance comes first; duos are not guaranteed if fairness requires otherwise.
            Only one duo pair per team is attempted. Both players must enter each other&apos;s{" "}
            <i>exact</i> Discord <strong>username </strong>(not a Discord display name, Matchmaker display name, or
            in-game name) for the request to match.
          </li>
        </ul>
        <p className={bodyText}>
          The host can still review the results and make manual changes afterward, and they will
          see a warning if the teams are not balanced.
        </p>
        <p className={bodyText}>
          For more detailed technical information, see{" "}
          <a
            href={MATCHMAKING_DOC_URL}
            className="body-link"
            target="_blank"
            rel="noopener noreferrer"
          >
            the matchmaking notes on GitHub
          </a>
          .
        </p>
      </div>

      <div className="flex flex-col gap-3">
        <h2 className={questionCls}>How do I know the matches are fair?</h2>
        <p className={bodyText}>
          Matchmaker does its best to create even teams, but perfect balance is not always
          possible, especially when signups span a very wide range of ranks.
        </p>
        <p className={bodyText}>
          After teams are created, each lobby is checked for balance. If Matchmaker could not get
          a good result, you will see a warning on that game or lobby. The message explains that
          teams were formed with the best available balance, but the rank spread was too wide for
          fully fair teams in that lobby.
        </p>
        <p className={bodyText}>A warning usually means one of two things:</p>
        <ul className={listCls}>
          <li>
            One player is much higher or lower ranked than everyone else in the lobby (a large
            skill gap).
          </li>
          <li>
            The two teams&apos; average skill levels are still noticeably different after
            balancing.
          </li>
        </ul>
        <p className={bodyText}>
          If you are the event host, you can also see each team&apos;s average rank on the team
          headers to judge balance yourself, even if the balance difference has not triggered a
          warning.
        </p>
        <p className={bodyText}>
          Fairness warnings are a heads-up, not a failure. The teams are still playable. The host
          can always adjust rosters manually if something looks off.
        </p>
      </div>

      <div className="flex flex-col gap-2">
        <h2 className={questionCls}>
          Why is an event blocked, and how does the Discord server lock work?
        </h2>
        <p className={bodyText}>
          Hosts can lock an event to one or more Discord servers they belong to. Anyone else,
          including lobby hosts, must be in at least one of those servers to open the event or
          register. The event owner is never locked out. Saving the lock with no servers, or
          turning it off, makes the event open to anyone with the link again.
        </p>
      </div>

      <div className="flex flex-col gap-2">
        <h2 className={questionCls}>
          Discord is having trouble and users can&apos;t register for events. What do I do?
        </h2>
        <p className={bodyText}>
          Generally wait, and check{" "}
          <a
            href="https://discordstatus.com"
            className="body-link"
            target="_blank"
            rel="noopener noreferrer"
          >
            discordstatus.com
          </a>
          . If the Discord API is having an outage, it can affect Matchmaker&apos;s ability to check
          server membership. Hosts are never locked out of their own events, so they can temporarily
          turn off the Discord server lock and save until Discord&apos;s API recovers, then lock it again.
        </p>
      </div>

      <div className="flex flex-col gap-2">
        <h2 className={questionCls}>
          Why does my display name and event name hit the 50-character limit so fast when I use emojis?
        </h2>
        <p className={bodyText}>
          Display names and event names are capped at 50 characters. Ordinary letters and numbers
          usually count as one character each, but many emojis (and especially ones made of several
          parts, like some family or flag emojis) count as more than one toward that limit even though
          they look like a single symbol on screen. That&apos;s just how computers store those
          characters; it isn&apos;t a bug in the counter. The counter on the form shows how much
          of the limit you have left; if it turns red, shorten the name or use fewer emojis.
        </p>
      </div>

      <div className="flex flex-col gap-3">
        <h2 className={questionCls}>How can teams be edited?</h2>
        <p className={bodyText}>
          After you click <strong>Lock In &amp; Create Teams</strong>, the host can adjust rosters
          without deleting and recreating everything. Edits are done <strong>one game at a
          time</strong> (each scheduled match in the group has its own teams/lobbies).
        </p>
        <p className={bodyText}>To move players around:</p>
        <ol className={`${bodyText} list-decimal pl-5 flex flex-col gap-2`}>
          <li>Open the event group page and select the game you want to change.</li>
          <li>
            On any player card in a <strong>team</strong>, <strong>subs</strong>, or{" "}
            <strong>unplaced</strong> list, open the <strong>⋯</strong> menu.
          </li>
          <li>
            Choose <strong>Swap</strong>.
          </li>
          <li>
            Pick another player from the dropdown: anyone on a different team, in another lobby,
            in the sub pool, or on the unplaced list, and click <strong>Submit</strong>.
          </li>
        </ol>
        <p className={bodyText}>
          A swap exchanges the two players&apos; spots. Each player takes the other&apos;s exact
          placement (lobby, team, sub slot, or unplaced status). You can use this to rebalance
          teams, move someone into or out of subs, pull an unplaced player onto a roster, or send
          a rostered player to unplaced.
        </p>
        <p className={bodyText}>A few rules to be aware of:</p>
        <ul className={listCls}>
          <li>
            You <strong>cannot</strong> swap two players already in the <strong>same group</strong>.
            You can&apos;t swap two players on the same team; you can&apos;t swap two substitutes
            in the same lobby with each other; you can&apos;t swap two unplaced players with each
            other. They won&apos;t appear as options for your swap.
          </li>
          <li>
            If a rostered player who did <strong>not</strong> sign up as a substitute is swapped
            into a sub slot, they become <strong>unplaced</strong> instead. Only substitute
            volunteers can fill sub spots.
          </li>
          <li>
            If your event requires a minimum number of substitutes <strong>and</strong> has more
            than one lobby, a swap is blocked when it would leave any lobby below that minimum.
          </li>
        </ul>
        <p className={bodyText}>
          After each swap, Matchmaker refreshes lobby hosts and fairness warnings for the affected
          lobbies so you can see whether balance changed. The original warning from lock-in is
          kept separately so you can still tell what the auto-generated teams looked like.
        </p>
        <p className={bodyText}>
          If you want to undo many changes at once, see the next FAQ entry about{" "}
          <strong>Delete teams</strong>.
        </p>
      </div>

      <div className="flex flex-col gap-2">
        <h2 className={questionCls}>
          I&apos;ve edited a team, but I don&apos;t like what I&apos;ve done. How do I get back to
          where it started?
        </h2>
        <p className={bodyText}>
          To get back to the default teams that were created, simply click &quot;Delete teams&quot;
          in the event header. Doing this will reset the event back to registrations-only from
          before teams were created, but registrations will remain closed. To regenerate the
          teams, click &quot;Create Teams&quot; in the event header and the teams will be created
          again the same way they were before.
        </p>
      </div>

      <div className="flex flex-col gap-2">
        <h2 className={questionCls}>
          I&apos;ve closed registrations and created teams, but a new player wants to join. How do
          I add them?
        </h2>
        <p className={bodyText}>
          As the host, you can&apos;t manually add players; all players must register themselves.
          To let the new player in, click &quot;Delete teams&quot; in the event header. This will
          reset the event back to the registrations, but registrations will remain closed. To
          reopen them, click &quot;Edit&quot; in the event header. At the bottom of the modal,
          there is a &quot;Registration Status&quot; toggle. Toggle it on and click &quot;Save
          Settings&quot; to allow new players to register.
        </p>
      </div>
    </div>
  );
}
