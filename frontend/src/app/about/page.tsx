// Public About & Privacy page. Reachable logged-in and logged-out (see `proxy.ts`).
import type { Metadata } from "next";
import Image from "next/image";
import Link from "next/link";
import AboutFaq from "@/app/about/AboutFaq";
import AboutIntro from "@/app/about/AboutIntro";
import PrivacyContent from "@/app/about/PrivacyContent";
import BackToTop from "@/app/_components/BackToTop";
import PageBackgroundOrbs from "@/app/_components/PageBackgroundOrbs";
import SiteFooterLinks from "@/app/_components/SiteFooterLinks";
import UserAccountMenu from "@/app/_components/UserAccountMenu";

export const metadata: Metadata = {
  title: "About · Matchmaker",
  description:
    "What Matchmaker does, how events and fair lobbies work, how your data is handled, and answers to common questions.",
};

export default function AboutPage() {
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
            <AboutIntro />
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

            <PrivacyContent />
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
