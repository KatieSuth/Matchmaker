// Public landing: Discord login. With auth_session set, the route guard (see `proxy.ts`) can redirect to the app.
import Image from "next/image";
import LoginButton from "@/app/_components/LoginButton";

export default function Page() {
  return (
    <main className="bg-page relative overflow-hidden min-h-screen flex flex-col items-center justify-center px-5 py-8">

      {/* Animated glow orb — top-left */}
      <div
        className="animate-drift glow-orb-blue pointer-events-none fixed rounded-full w-[600px] h-[600px] -top-[200px] -left-[200px]"
        aria-hidden
      />

      {/* Animated glow orb — bottom-right */}
      <div
        className="animate-drift-reverse glow-orb-orange pointer-events-none fixed rounded-full w-[500px] h-[500px] -bottom-[150px] -right-[150px]"
        aria-hidden
      />

      {/* Card */}
      <div className="animate-rise card relative z-10 flex flex-col items-center w-full max-w-[540px] px-12 py-14 max-sm:px-6 max-sm:py-10">

        {/* Top edge gradient */}
        <div className="bg-top-edge absolute top-0 left-0 right-0 h-0.5 rounded-t-sm" aria-hidden />

        {/* Logo */}
        <div className="animate-rise-1 logo-glow-lg relative flex items-center justify-center w-[300px] h-[300px] mb-7 max-sm:w-[220px] max-sm:h-[220px]">
          <Image
            src="/logo-full.png"
            className="object-contain"
            alt="Matchmaker"
            fill
            sizes="(max-width: 480px) 220px, 300px"
            priority
          />
        </div>

        {/* Decorative divider */}
        <div className="animate-rise-2 flex items-center gap-2 w-full mb-6">
          <span className="bg-divider-blue flex-1 h-px" />
          <span className="divider-gem w-1.5 h-1.5 flex-shrink-0 rotate-45" />
          <span className="bg-divider-orange flex-1 h-px" />
        </div>

        {/* Body */}
        <p
          className="animate-rise-3 font-light leading-[1.8] text-center mb-9 max-w-[36ch]"
          style={{ fontSize: "clamp(1.1rem, 2.5vw, 1.12rem)", color: "var(--color-text-body)" }}
        >
          Welcome to Matchmaker, built to ensure fair play and effortless match organization. To get started,
          please log in with Discord.
        </p>

        <LoginButton />
      </div>

      {/* Footer */}
      <footer className="animate-rise-5 relative z-10 mt-8 text-xs text-center" style={{ letterSpacing: "0.06em" }}>
        <a
          href="https://github.com/KatieSuth/Matchmaker"
          target="_blank"
          rel="noopener noreferrer"
          className="footer-link inline-flex items-center gap-1.5 transition-[color,border-color] duration-200"
        >
          <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor" aria-hidden>
            <path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0 0 24 12c0-6.63-5.37-12-12-12z"/>
          </svg>
          Source on GitHub
        </a>
      </footer>
    </main>
  );
}
