import Link from "next/link";
import { DEFAULT_FEEDBACK_URL, GITHUB_REPO_URL } from "@/app/_lib/constants";

export default function SiteFooterLinks() {
  const feedbackUrl =
    process.env.NEXT_PUBLIC_FEEDBACK_URL || DEFAULT_FEEDBACK_URL;

  return (
    <span className="inline-flex flex-wrap items-center justify-center gap-2">
      <Link
        href="/about"
        className="footer-link transition-[color,border-color] duration-200"
      >
        About &amp; Privacy
      </Link>
      <span className="text-[var(--color-text-faint)]" aria-hidden>·</span>
      <a
        href={GITHUB_REPO_URL}
        target="_blank"
        rel="noopener noreferrer"
        className="footer-link inline-flex items-center gap-1.5 transition-[color,border-color] duration-200"
      >
        <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor" aria-hidden>
          <path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0 0 24 12c0-6.63-5.37-12-12-12z" />
        </svg>
        Source on GitHub
      </a>
      <span className="text-[var(--color-text-faint)]" aria-hidden>·</span>
      <a
        href={feedbackUrl}
        target="_blank"
        rel="noopener noreferrer"
        className="footer-link transition-[color,border-color] duration-200"
      >
        Submit Issues/Feedback
      </a>
    </span>
  );
}
