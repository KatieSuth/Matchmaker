// Public About & Privacy page. Reachable logged-in and logged-out (see `proxy.ts`).
import type { Metadata } from "next";
import Image from "next/image";
import Link from "next/link";
import AboutCta from "@/app/about/AboutCta";
import AboutFaq from "@/app/about/AboutFaq";
import BackToTop from "@/app/_components/BackToTop";
import PageBackgroundOrbs from "@/app/_components/PageBackgroundOrbs";
import { SectionDivider } from "@/app/_components/SectionDivider";
import SiteFooterLinks from "@/app/_components/SiteFooterLinks";
import UserAccountMenu from "@/app/_components/UserAccountMenu";
import { DEFAULT_FEEDBACK_URL, GITHUB_REPO_URL } from "@/app/_lib/constants";

export const metadata: Metadata = {
  title: "About · Matchmaker",
  description:
    "What Matchmaker does, how events and fair lobbies work, how your data is handled, and answers to common questions.",
};

const bodyText =
  "font-light leading-[1.8] text-[var(--color-text-body)]";

const PRIVACY_UPDATED = "August 17, 2026";

export default function AboutPage() {
  const feedbackUrl =
    process.env.NEXT_PUBLIC_FEEDBACK_URL || DEFAULT_FEEDBACK_URL;

  return (
    <div className="bg-page relative overflow-hidden min-h-screen flex flex-col">
      <PageBackgroundOrbs />

      <div className="relative z-10 flex min-h-0 flex-1 flex-col">
        <header
          id="top"
          className="relative flex items-center px-6 py-4 max-sm:px-4"
        >
          <Link href="/" className="flex-shrink-0 flex items-center no-underline">
            <div className="logo-glow-sm relative w-[120px] h-[42px] max-sm:w-[90px] max-sm:h-[34px]">
              <Image
                src="/logo-small.png"
                alt="Matchmaker"
                fill
                sizes="120px"
                className="object-contain object-left-center"
                priority
              />
            </div>
          </Link>
          <nav
            aria-label="On this page"
            className="absolute left-1/2 -translate-x-1/2 flex items-center gap-2 text-sm"
          >
            <a href="#about" className="body-link">
              About
            </a>
            <span className="text-[var(--color-text-faint)]" aria-hidden>
              ·
            </span>
            <a href="#faq" className="body-link">
              FAQ &amp; Troubleshooting
            </a>
            <span className="text-[var(--color-text-faint)]" aria-hidden>
              ·
            </span>
            <a href="#privacy" className="body-link">
              Privacy
            </a>
          </nav>
          <UserAccountMenu />
        </header>

        <main className="flex flex-1 flex-col items-center gap-8 px-5 pb-12">
          <article
            id="about"
            className="card relative w-full max-w-3xl scroll-mt-6 px-10 py-12 max-sm:px-6 max-sm:py-8"
          >
            <div className="bg-top-edge absolute top-0 left-0 right-0 h-0.5 rounded-t-sm" aria-hidden />

            <h1 className="text-2xl font-semibold tracking-wide mb-4 max-sm:text-xl">
              About Matchmaker
            </h1>
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
                substitute minimum, and whether registration is open. The first lobby can be created
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
                details and register. Before registering, you can review the in-game name and
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
          </article>

          <BackToTop />

          <article
            id="faq"
            className="card relative w-full max-w-3xl scroll-mt-6 px-10 py-12 max-sm:px-6 max-sm:py-8"
          >
            <div className="bg-top-edge absolute top-0 left-0 right-0 h-0.5 rounded-t-sm" aria-hidden />

            <h1 className="text-2xl font-semibold tracking-wide mb-8 max-sm:text-xl">
              FAQ &amp; Troubleshooting
            </h1>

            <AboutFaq />
          </article>

          <BackToTop />

          <article
            id="privacy"
            className="card relative w-full max-w-3xl scroll-mt-6 px-10 py-12 max-sm:px-6 max-sm:py-8"
          >
            <div className="bg-top-edge absolute top-0 left-0 right-0 h-0.5 rounded-t-sm" aria-hidden />

            <h1 className="text-2xl font-semibold tracking-wide mb-8 max-sm:text-xl">
              Privacy
            </h1>

            <div className="flex flex-col gap-6">
              <div className="flex flex-col gap-2">
                <h3 className="text-sm font-semibold tracking-wide text-[var(--color-text-soft)]">
                  What we store
                </h3>
                <p className={bodyText}>
                  When you log in with Discord, we keep the identifiers we need to recognize you:
                  your Discord account ID, username, and the ID used to show your avatar. We copy
                  your Discord display name when you first sign in. You can change or clear that
                  name in Matchmaker; doing so only updates Matchmaker and does not change your
                  Discord profile. Later logins may refresh your Discord username and avatar, but
                  they do not overwrite a display name you already have in Matchmaker.
                </p>
                <p className={bodyText}>
                  If you fill in a profile, we store optional pronouns and region, plus per-game
                  in-game names and ranks. For events, we store the event itself, who signed up,
                  team and lobby assignments, duo player requests, and any join codes a host enters.
                </p>
              </div>

              <div className="flex flex-col gap-2">
                <h3 className="text-sm font-semibold tracking-wide text-[var(--color-text-soft)]">
                  Cookies
                </h3>
                <p className={bodyText}>
                  We use cookies to keep you logged in, not for advertising or analytics. After you
                  sign in, the browser keeps a secure login cookie. Using the site keeps that cookie
                  current. If about a week goes by without you using Matchmaker, you will need to sign
                  in again. You can log out at any time, which ends the session right away and deletes
                  the cookie.
                </p>
                <p className={bodyText}>
                  A second, smaller cookie only remembers whether to show the login screen or the
                  app. It does not prove who you are or let anyone use your account, and it is also
                  deleted when you log out. Additionally, while Discord sign-in is in progress, there
                  is also a short-lived cookie that is deleted when that process finishes.
                </p>
              </div>

              <div className="flex flex-col gap-2">
                <h3 className="text-sm font-semibold tracking-wide text-[var(--color-text-soft)]">
                  Server logs
                </h3>
                <p className={bodyText}>
                  Each request to the API is logged with technical details such as the path, whether
                  it succeeded, how long it took, and the visitor&apos;s IP address. Those logs are
                  for running and debugging the site. They are not a user profile we browse.
                </p>
              </div>

              <div className="flex flex-col gap-2">
                <h3 className="text-sm font-semibold tracking-wide text-[var(--color-text-soft)]">
                  How data is protected
                </h3>
                <p className={bodyText}>
                  You sign in through Discord. Matchmaker never sees a password. Traffic uses
                  HTTPS. We store a hashed version of your session refresh token, not the raw
                  token. We also store an encrypted Discord refresh token so we can later check
                  your Discord server memberships without asking you to authorize again. That
                  token is only used to request the Discord data needed for that feature. Ranks
                  and pronouns stay visible only to event hosts unless you choose otherwise.
                </p>
              </div>

              <div className="flex flex-col gap-2">
                <h3 className="text-sm font-semibold tracking-wide text-[var(--color-text-soft)]">
                  Who we share with
                </h3>
                <p className={bodyText}>
                  Matchmaker does not sell your data, use it for advertising, or share your
                  profile, ranks, or event data with other companies for product or marketing.
                  Discord is used so you can sign in; avatar images are loaded from Discord&apos;s
                  network in your browser.
                </p>
                <p className={bodyText}>
                  The live site uses cloud providers for hosting and traffic routing. Those
                  providers process request details, including IP addresses. Cloudflare sits in
                  front of the live site and shows aggregate traffic stats, such as unique visitors
                  and total requests. Matchmaker does not set analytics cookies or run its own
                  visitor trackers.{" "}
                  <a
                    href="https://www.cloudflare.com/privacypolicy/"
                    className="body-link"
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    Cloudflare&apos;s privacy policy
                  </a>{" "}
                  applies to that edge service.
                </p>
                <p className={bodyText}>
                  The live site sends technical error reports to a crash/error service so problems
                  can be fixed. Those reports can include technical request details, such as what
                  path failed. They are not used as a user database. This technical error processing
                  is always on in production. It is not a setting you can turn off, and it is not
                  advertising or analytics tracking.
                </p>
                <p className={bodyText}>
                  Some footer links take you off this site.{" "}
                  <a
                    href={feedbackUrl}
                    className="body-link"
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    Issues/Feedback
                  </a>{" "}
                  on{" "}
                  <a
                    href="https://matchmaker.games"
                    className="body-link"
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    matchmaker.games
                  </a>{" "}
                  opens a Google Form hosted by Matchmaker&apos;s creator. What you submit there is
                  handled by that form, not stored in Matchmaker, and{" "}
                  <a
                    href="https://policies.google.com/privacy"
                    className="body-link"
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    Google&apos;s privacy policy
                  </a>{" "}
                  applies from that point.{" "}
                  <a
                    href={GITHUB_REPO_URL}
                    className="body-link"
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    Source on GitHub
                  </a>{" "}
                  is also off this site, and{" "}
                  <a
                    href="https://docs.github.com/en/site-policy/privacy-policies/github-privacy-statement"
                    className="body-link"
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    GitHub&apos;s privacy policy
                  </a>{" "}
                  applies there.
                </p>
              </div>

              <div className="flex flex-col gap-2">
                <h3 className="text-sm font-semibold tracking-wide text-[var(--color-text-soft)]">
                  Who can see you inside the app
                </h3>
                <p className={bodyText}>
                  People in an event you&apos;ve registered to join can see your Discord username and your
                  Matchmaker display name (if provided). Hosts always see ranks and pronouns for players
                  in their events. Other players do not see ranks or pronouns unless you turn that on in{" "}
                  <Link href="/my_account" className="body-link">
                    My Account
                  </Link>
                  . If you host an event, players can see your display name and Discord username.
                </p>
              </div>

              <div className="flex flex-col gap-2">
                <h3 className="text-sm font-semibold tracking-wide text-[var(--color-text-soft)]">
                  What we don&apos;t store
                </h3>
                <p className={bodyText}>
                  No email. No password. No Discord chat, DMs, or message history. Discord&apos;s
                  login screen asks for permission to see your servers, but Matchmaker does not
                  currently save a list of your Discord servers. A future server-membership feature
                  may periodically request the current list from Discord using the stored encrypted
                  authorization token, use it to check membership, and discard the list rather than
                  storing it. Matchmaker does not set analytics cookies or run its own advertising
                  or visitor trackers. Your avatar image stays on Discord&apos;s network; we only
                  store the identifier used to display it.
                </p>
              </div>

              <div className="flex flex-col gap-2">
                <h3 className="text-sm font-semibold tracking-wide text-[var(--color-text-soft)]">
                  Your controls
                </h3>
                <p className={bodyText}>
                  In{" "}
                  <Link href="/my_account" className="body-link">
                    My Account
                  </Link>{" "}
                  you can edit your profile and game accounts, including changing or clearing the
                  display name copied from Discord (that change stays in Matchmaker). You can hide
                  or show your ranks and pronouns, remove a game profile, leave an event, or delete
                  an event you host. Logging out ends your session.
                </p>
                <p className={bodyText}>
                  You cannot delete the whole account in the app today. Use{" "}
                  <a
                    href={feedbackUrl}
                    className="body-link"
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    Issues/Feedback
                  </a>{" "}
                  if you want data removed. To delete data, you will be asked to provide proof that
                  you are the account owner.
                </p>
              </div>

              <div className="flex flex-col gap-2">
                <h3 className="text-sm font-semibold tracking-wide text-[var(--color-text-soft)]">
                  How long we keep data
                </h3>
                <p className={bodyText}>
                  Your account and related event data stay until you remove what the app allows
                  (game profiles, leaving or deleting events) or until you ask for the account to
                  be removed. Login sessions stay active while you use the site, end when you log
                  out, and otherwise lapse after about a week without use. Crash reports
                  and server logs are kept by those systems only as needed to run the site, not as
                  a directory of users.
                </p>
              </div>

              <div className="flex flex-col gap-2">
                <h3 className="text-sm font-semibold tracking-wide text-[var(--color-text-soft)]">
                  Children
                </h3>
                <p className={bodyText}>
                  Matchmaker is not directed at children under 13. Do not use the site if you are
                  under 13.
                </p>
              </div>

              <div className="flex flex-col gap-2">
                <h3 className="text-sm font-semibold tracking-wide text-[var(--color-text-soft)]">
                  How to reach me
                </h3>
                <p className={bodyText}>
                  Matchmaker is a hobby project. To reach me, use{" "}
                  <a
                    href={feedbackUrl}
                    className="body-link"
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    Issues/Feedback
                  </a>{" "}
                  or the{" "}
                  <a
                    href={GITHUB_REPO_URL}
                    className="body-link"
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    GitHub project
                  </a>
                  .
                </p>
              </div>

              <p className="text-xs text-[var(--color-text-muted)]">
                Last updated {PRIVACY_UPDATED}.
              </p>
            </div>
          </article>

          <BackToTop />
        </main>

        <footer className="py-5 text-center text-xs" style={{ letterSpacing: "0.06em" }}>
          <SiteFooterLinks />
        </footer>
      </div>
    </div>
  );
}
