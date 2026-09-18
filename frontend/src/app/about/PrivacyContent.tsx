// Static "Privacy" article body: what's stored, cookies, logs, sharing, retention, and controls.
import Link from "next/link";
import { DEFAULT_FEEDBACK_URL, GITHUB_REPO_URL } from "@/app/_lib/constants";
import { bodyText } from "@/app/about/_lib/text";

const PRIVACY_UPDATED = "August 30, 2026";

export default function PrivacyContent() {
  const feedbackUrl = process.env.NEXT_PUBLIC_FEEDBACK_URL || DEFAULT_FEEDBACK_URL;

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-2">
        <h2 className="text-sm font-semibold tracking-wide text-[var(--color-text-soft)]">
          What we store
        </h2>
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
        <h2 className="text-sm font-semibold tracking-wide text-[var(--color-text-soft)]">
          Cookies
        </h2>
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
        <h2 className="text-sm font-semibold tracking-wide text-[var(--color-text-soft)]">
          Server logs
        </h2>
        <p className={bodyText}>
          Each request to the API is logged with technical details such as the path, whether
          it succeeded, how long it took, and the visitor&apos;s IP address. Those logs are
          for running and debugging the site. They are not a user profile we browse.
        </p>
      </div>

      <div className="flex flex-col gap-2">
        <h2 className="text-sm font-semibold tracking-wide text-[var(--color-text-soft)]">
          How data is protected
        </h2>
        <p className={bodyText}>
          You sign in through Discord. Matchmaker never sees a password. Traffic uses
          HTTPS. We store a hashed version of your session refresh token, not the raw
          token. We also store an encrypted Discord refresh token so we can request your
          current server list from Discord when a host picks servers or when a locked
          event is opened or registered. Ranks and pronouns stay visible only to event
          hosts unless you choose otherwise.
        </p>
      </div>

      <div className="flex flex-col gap-2">
        <h2 className="text-sm font-semibold tracking-wide text-[var(--color-text-soft)]">
          Who we share with
        </h2>
        <p className={bodyText}>
          Matchmaker does not sell your data, use it for advertising, or share your
          profile, ranks, or event data with other companies for product or marketing.
          Discord is used so you can sign in; avatar images are loaded from Discord&apos;s
          network in your browser. Discord also receives guild-list API calls (using your
          OAuth user token) when a host picks servers or when a locked event is opened or
          registered. We do not send Discord your Matchmaker event roster.
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
        <h2 className="text-sm font-semibold tracking-wide text-[var(--color-text-soft)]">
          Who can see you inside the app
        </h2>
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
        <h2 className="text-sm font-semibold tracking-wide text-[var(--color-text-soft)]">
          What we don&apos;t store
        </h2>
        <p className={bodyText}>
          No email. No password. No Discord chat, DMs, or message history. We do not
          persist your full Discord server list in the database. We snapshot only the
          host-selected server id and name on an event. Membership checks ask Discord for
          your current server list and compare it; that full list is not written to disk.
          The API may keep it in memory for up to 60 seconds so nearby checks (for example
          opening a locked event) do not each call Discord again. Matchmaker does not set
          analytics cookies or run its own advertising or visitor trackers. Your avatar
          image stays on Discord&apos;s network; we only store the identifier used to
          display it.
        </p>
      </div>

      <div className="flex flex-col gap-2">
        <h2 className="text-sm font-semibold tracking-wide text-[var(--color-text-soft)]">
          Your controls
        </h2>
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
        <h2 className="text-sm font-semibold tracking-wide text-[var(--color-text-soft)]">
          How long we keep data
        </h2>
        <p className={bodyText}>
          Your account and related event data stay until you remove what the app allows
          (game profiles, leaving or deleting events) or until you ask for the account to
          be removed. Login sessions stay active while you use the site, end when you log
          out, and otherwise lapse after about a week without use. Event Discord-server
          snapshots last until the host clears the lock or deletes the event. Your full
          Discord server list, when fetched for a membership check or the host server
          picker, may stay in the API process&apos;s memory for up to 60 seconds and is
          not stored in the database. The encrypted Discord refresh token is kept until
          Discord login is revoked or the account is removed. Crash reports and server
          logs are kept by those systems only as needed to run the site, not as a
          directory of users.
        </p>
      </div>

      <div className="flex flex-col gap-2">
        <h2 className="text-sm font-semibold tracking-wide text-[var(--color-text-soft)]">
          Children
        </h2>
        <p className={bodyText}>
          Matchmaker is not directed at children under 13. Do not use the site if you are
          under 13.
        </p>
      </div>

      <div className="flex flex-col gap-2">
        <h2 className="text-sm font-semibold tracking-wide text-[var(--color-text-soft)]">
          How to reach me
        </h2>
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
  );
}
