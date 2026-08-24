"use client";

// Discord OAuth: full-page navigation to GET /auth/login on the API (not XHR).
import { POST_LOGIN_REDIRECT_STORAGE_KEY } from "@/app/_lib/postLoginRedirect";

export default function DiscordLoginButton() {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL ?? "";
  const loginHref = `${apiUrl}/auth/login`;

  const handleClick = (e: React.MouseEvent<HTMLAnchorElement>) => {
    const onLanding = window.location.pathname === "/";
    const hasNext = new URLSearchParams(window.location.search).get("next");
    if (onLanding && !hasNext) {
      sessionStorage.removeItem(POST_LOGIN_REDIRECT_STORAGE_KEY);
    }

    // Same-origin /api is proxied by Caddy, not a Next.js route. Force a document
    // navigation so the App Router cannot treat this click as a client-side transition.
    e.preventDefault();
    window.location.assign(loginHref);
  };

  return (
    <a
      href={loginHref}
      onClick={handleClick}
      className="login-btn animate-rise-btn relative flex items-center gap-3 overflow-hidden rounded-[3px] px-[2.1rem] py-[0.9rem] text-[0.78rem] font-semibold tracking-[0.12em] uppercase no-underline cursor-pointer transition-[transform,box-shadow,opacity] duration-200 max-sm:w-full max-sm:justify-center"
      aria-label="Log in with Discord"
    >
      <svg className="w-5 h-5 flex-shrink-0" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
        <path d="M20.317 4.492c-1.53-.69-3.17-1.2-4.885-1.49a.075.075 0 0 0-.079.036c-.21.369-.444.85-.608 1.23a18.566 18.566 0 0 0-5.487 0 12.36 12.36 0 0 0-.617-1.23A.077.077 0 0 0 8.562 3c-1.714.29-3.354.8-4.885 1.491a.07.07 0 0 0-.032.027C.533 9.093-.32 13.555.099 17.961a.08.08 0 0 0 .031.055 20.03 20.03 0 0 0 5.993 2.98.078.078 0 0 0 .084-.026c.462-.62.874-1.275 1.226-1.963a.074.074 0 0 0-.041-.104 13.201 13.201 0 0 1-1.872-.878.075.075 0 0 1-.008-.125c.126-.093.252-.19.372-.287a.075.075 0 0 1 .078-.01c3.927 1.764 8.18 1.764 12.061 0a.075.075 0 0 1 .079.009c.12.098.245.195.372.288a.075.075 0 0 1-.006.125c-.598.344-1.22.635-1.873.877a.075.075 0 0 0-.041.105c.36.687.772 1.341 1.225 1.962a.077.077 0 0 0 .084.028 19.963 19.963 0 0 0 6.002-2.981.076.076 0 0 0 .032-.054c.5-5.094-.838-9.52-3.549-13.442a.06.06 0 0 0-.031-.028zM8.02 15.278c-1.182 0-2.157-1.069-2.157-2.38 0-1.312.956-2.38 2.157-2.38 1.21 0 2.176 1.077 2.157 2.38 0 1.312-.956 2.38-2.157 2.38zm7.975 0c-1.183 0-2.157-1.069-2.157-2.38 0-1.312.955-2.38 2.157-2.38 1.21 0 2.176 1.077 2.157 2.38 0 1.312-.946 2.38-2.157 2.38z"/>
      </svg>
      <span>Log in with Discord</span>
    </a>
  );
}
