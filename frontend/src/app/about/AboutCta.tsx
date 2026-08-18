"use client";

// About page exit link: login for guests, Events when signed in.
import Link from "next/link";
import { useAuth } from "@/app/_context/AuthContext";

export default function AboutCta() {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    return <span className="body-link invisible" aria-hidden>Back to Login</span>;
  }

  if (isAuthenticated) {
    return (
      <Link href="/my_events" className="body-link">
        Back to Events
      </Link>
    );
  }

  return (
    <Link href="/" className="body-link">
      Back to Login
    </Link>
  );
}
