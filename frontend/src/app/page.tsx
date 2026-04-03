import Image from "next/image";
import styles from "./page.module.css";
import LoginButton from "@/app/_components/LoginButton/LoginButton";

export default function Page() {
  return (
    <main className={styles.page}>
      <div className={styles.card}>

        {/* Logo */}
        <div className={styles.logo_wrap}>
          <Image
            className={styles.logo_img}
            src="/logo.png"
            alt="Matchmaker"
            fill
            priority
          />
        </div>

        {/* Decorative divider */}
        <div className={styles.divider}>
          <span className={styles.divider_line_blue} />
          <span className={styles.divider_gem} />
          <span className={styles.divider_line_orange} />
        </div>

        {/* Body */}
        <p className={styles.body}>
          Welcome to Matchmaker, built to ensure fair play and effortless match organization. To get started,
          please log in with Discord.
        </p>

        {/* Client-side login button */}
        <LoginButton />

      </div>

      <footer className={styles.footer}>
        <a
          href="https://github.com/KatieSuth/Matchmaker"
          target="_blank"
          rel="noopener noreferrer"
          className={styles.footerLink}
        >
          Source on GitHub
        </a>
      </footer>
    </main>
  );
}
