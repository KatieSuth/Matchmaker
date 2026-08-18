"use client";

// Top nav for the authenticated shell: app links, Discord avatar, and logout.
import Image from "next/image";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { NAV_LINKS } from "@/app/_lib/constants";
import UserAccountMenu from "@/app/_components/UserAccountMenu";

export default function AppNav() {
  const pathname = usePathname();

  return (
    <header className="nav-header sticky top-0 z-[100] w-full md:backdrop-blur-[18px]">

      {/* Top-edge gradient accent */}
      <div className="bg-top-edge h-0.5" aria-hidden />

      <nav className="relative flex items-center h-[60px] px-6 max-sm:px-4 max-sm:gap-4">

        {/* Logo */}
        <Link href="/my_events" className="flex-shrink-0 flex items-center no-underline mr-2">
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

        {/* Nav links — centered absolutely */}
        <ul className="absolute left-1/2 -translate-x-1/2 flex items-center gap-1 list-none max-sm:gap-0">
          {NAV_LINKS.map(({ label, href }) => {
            const isActive = pathname.startsWith(href);
            return (
              <li key={href}>
                <Link
                  href={href}
                  className={`nav-link ${isActive ? "nav-link-active" : ""} relative inline-flex flex-col items-center rounded px-[0.85rem] py-[0.4rem] text-sm font-semibold tracking-[0.13em] uppercase no-underline transition-[color,background] duration-200 max-sm:px-[0.6rem] max-sm:text-[0.7rem]`}
                >
                  {label}
                  <span className="link-underline" aria-hidden />
                </Link>
              </li>
            );
          })}
        </ul>

        <UserAccountMenu />
      </nav>
    </header>
  );
}
