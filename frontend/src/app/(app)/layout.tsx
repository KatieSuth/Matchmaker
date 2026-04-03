// frontend/src/app/(app)/layout.tsx
//
// Authenticated shell layout. All pages that require login live under
// frontend/src/app/(app)/ — e.g. (app)/events/page.tsx, (app)/my_account/page.tsx

import AppNav from "@/app/_components/AppNav/AppNav";
import styles from "./layout.module.css";

export default function AppLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className={styles.shell}>
      <AppNav />
      <main className={styles.main}>{children}</main>
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
    </div>
  );
}
