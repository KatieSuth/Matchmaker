// Authenticated layout. All pages that require login live under
// frontend/src/app/(app)/ — e.g. (app)/events/page.tsx, (app)/my_account/page.tsx

import AppNav from "@/app/_components/AppNav";
import PageBackgroundOrbs from "@/app/_components/PageBackgroundOrbs";
import SiteFooterLinks from "@/app/_components/SiteFooterLinks";

export default function AppLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="bg-page relative overflow-hidden min-h-screen flex flex-col">
      <PageBackgroundOrbs />
      <div className="relative z-10 flex min-h-0 flex-1 flex-col">
        <AppNav />
        <main className="flex flex-1 flex-col">{children}</main>
        <footer className="py-5 text-center text-xs" style={{ letterSpacing: "0.06em" }}>
          <SiteFooterLinks />
        </footer>
      </div>
    </div>
  );
}
