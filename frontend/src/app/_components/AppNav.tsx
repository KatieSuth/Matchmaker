"use client";

import { useState, useRef, useEffect, useCallback } from "react";
import Image from "next/image";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useAuth } from "@/app/_context/AuthContext";

function avatarUrl(discordId: string | null, avatarHash: string | null): string {
  if (!avatarHash || !discordId) return `https://cdn.discordapp.com/embed/avatars/0.png`;
  return `https://cdn.discordapp.com/avatars/${discordId}/${avatarHash}.webp?size=64`;
}

const NAV_LINKS = [
  { label: "Events", href: "/events" },
  { label: "Games",  href: "/games"  },
];

export default function AppNav() {
  const pathname = usePathname();
  const router   = useRouter();
  const { user, isAuthenticated, logout } = useAuth();

  const [dropdownOpen, setDropdownOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setDropdownOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, []);

  const handleLogout = useCallback(async () => {
    setDropdownOpen(false);
    await logout();
    router.push("/");
  }, [logout, router]);

  return (
    <header className="nav-header sticky top-0 z-[100] w-full backdrop-blur-[18px]">

      {/* Top-edge gradient accent */}
      <div className="bg-top-edge h-0.5" aria-hidden />

      <nav className="relative flex items-center h-[60px] px-6 max-sm:px-4 max-sm:gap-4">

        {/* Logo */}
        <Link href="/events" className="flex-shrink-0 flex items-center no-underline mr-2">
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

        {/* User area */}
        {isAuthenticated && (
          <div className="relative ml-auto flex-shrink-0" ref={dropdownRef}>
            <button
              className="avatar-btn flex items-center gap-[0.6rem] py-[0.3rem] pr-[0.6rem] pl-[0.3rem] rounded-full cursor-pointer transition-[background,border-color,color] duration-200"
              onClick={() => setDropdownOpen(v => !v)}
              aria-expanded={dropdownOpen}
              aria-haspopup="menu"
            >
              {/* Avatar ring */}
              <div className="bg-brand-gradient w-9 h-9 rounded-full flex-shrink-0 p-[1.5px]">
                {user?.image_url ? (
                  <Image
                    src={avatarUrl(user.discord_id, user.image_url)}
                    alt={user.discord_name ?? "discord image"}
                    width={36}
                    height={36}
                    className="block w-full h-full rounded-full object-cover border-[1.5px] border-[--color-bg]"
                  />
                ) : (
                  <div className="w-full h-full rounded-full bg-white/[0.08] border-[1.5px] border-[--color-bg]" />
                )}
              </div>

              <span className="text-sm font-semibold tracking-[0.04em] max-w-[120px] whitespace-nowrap overflow-hidden text-ellipsis max-sm:hidden">
                {user?.discord_name ?? "Account"}
              </span>

              <svg
                className={`flex-shrink-0 transition-transform duration-200 text-[rgba(180,200,235,0.5)] ${dropdownOpen ? "rotate-180" : "rotate-0"}`}
                width="12" height="12" viewBox="0 0 12 12" fill="none"
                aria-hidden
              >
                <path d="M2 4l4 4 4-4" stroke="currentColor" strokeWidth="1.5"
                  strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
            </button>

            {dropdownOpen && (
              <div
                className="animate-drop-in dropdown-panel absolute top-[calc(100%+10px)] right-0 min-w-[180px] rounded-lg overflow-hidden"
                role="menu"
              >
                {/* Arrow */}
                <div className="dropdown-arrow absolute -top-[5px] right-[18px] w-2.5 h-2.5 rotate-45" aria-hidden />

                <Link
                  href="/my_account"
                  className="dropdown-item w-full flex items-center gap-[0.65rem] px-4 py-3 text-[0.8rem] font-medium tracking-[0.03em] transition-[background,color] duration-150"
                  role="menuitem"
                  onClick={() => setDropdownOpen(false)}
                >
                  My Account
                </Link>

                <div className="h-px mx-3 bg-white/[0.06]" />

                <button
                  className="dropdown-item dropdown-item-danger w-full flex items-center gap-[0.65rem] px-4 py-3 text-[0.8rem] font-medium tracking-[0.03em] bg-transparent border-none cursor-pointer font-[inherit] text-left transition-[background,color] duration-150"
                  role="menuitem"
                  onClick={handleLogout}
                >
                  Logout
                </button>
              </div>
            )}
          </div>
        )}
      </nav>
    </header>
  );
}
