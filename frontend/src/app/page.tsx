// Public landing: Discord login. With auth_session set, the route guard (see `proxy.ts`) can redirect to the app.
import Image from "next/image";
import LoginButton from "@/app/_components/LoginButton";
import PageBackgroundOrbs from "@/app/_components/PageBackgroundOrbs";
import { PostLoginRedirectCapture } from "@/app/_components/PostLoginRedirectCapture";
import SiteFooterLinks from "@/app/_components/SiteFooterLinks";

export default function Page() {
  return (
    <main className="bg-page relative overflow-hidden min-h-screen flex flex-col items-center justify-center px-5 py-8">
      <PostLoginRedirectCapture />
      <PageBackgroundOrbs />

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
        <SiteFooterLinks />
      </footer>
    </main>
  );
}
