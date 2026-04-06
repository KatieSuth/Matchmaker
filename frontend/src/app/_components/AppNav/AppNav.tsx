"use client";
// frontend/src/app/_components/AppNav/AppNav.tsx

import { useState, useRef, useEffect, useCallback } from "react";
import Image from "next/image";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useAuth } from "@/app/_context/AuthContext";
import styles from "./AppNav.module.css";

// Discord avatar URL helper
function avatarUrl(discordId: string | null, avatarHash: string | null): string {
  if (!avatarHash || !discordId) {
    return `https://cdn.discordapp.com/embed/avatars/0.png`;
  }
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

  // Close dropdown on outside click
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
    <header className={styles.header}>
      {/* Top-edge gradient accent — matches card in login page */}
      <div className={styles.topEdge} aria-hidden />

      <nav className={styles.nav}>
        {/* ── Logo ── */}
        <Link href="/events" className={styles.logoLink}>
          <div className={styles.logoWrap}>
            <Image
              src="/logo-small.png"
              alt="Matchmaker"
              fill
              sizes="120px"
              className={styles.logoImg}
              priority
            />
          </div>
        </Link>

        {/* ── Nav links ── */}
        <ul className={styles.links}>
          {NAV_LINKS.map(({ label, href }) => (
            <li key={href}>
              <Link
                href={href}
                className={`${styles.link} ${pathname.startsWith(href) ? styles.linkActive : ""}`}
              >
                {label}
                <span className={styles.linkUnderline} aria-hidden />
              </Link>
            </li>
          ))}
        </ul>

        {/* ── User menu ── */}
        {isAuthenticated && (
          <div className={styles.userArea} ref={dropdownRef}>
            <button
              className={styles.avatarBtn}
              onClick={() => setDropdownOpen(v => !v)}
              aria-expanded={dropdownOpen}
              aria-haspopup="menu"
            >
              <div className={styles.avatarRing}>
                {user?.image_url ? (
                  <Image
                    src={avatarUrl(user.discord_id, user.image_url)}
                    alt={user.discord_name ?? "discord image"}
                    width={36}
                    height={36}
                    className={styles.avatar}
                  />
                ) : (
                  <div className={styles.avatarPlaceholder} />
                )}
              </div>
              <span className={styles.username}>
               {user?.discord_name ?? "Account"} 
              </span>
              <svg
                className={`${styles.chevron} ${dropdownOpen ? styles.chevronOpen : ""}`}
                width="12" height="12" viewBox="0 0 12 12" fill="none"
                aria-hidden
              >
                <path d="M2 4l4 4 4-4" stroke="currentColor" strokeWidth="1.5"
                  strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
            </button>

            {dropdownOpen && (
              <div className={styles.dropdown} role="menu">
                <div className={styles.dropdownArrow} aria-hidden />
                <Link href="/my_account" className={styles.dropdownItem} role="menuitem"
                  onClick={() => setDropdownOpen(false)}>
                  My Account
                </Link>
                <div className={styles.dropdownDivider} />
                <button className={`${styles.dropdownItem} ${styles.dropdownItemDanger}`}
                  role="menuitem" onClick={handleLogout}>
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
