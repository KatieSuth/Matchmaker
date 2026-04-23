// Root layout: wraps the app with AuthProvider so any subtree can use useAuth and axios auth.
import type { Metadata } from "next";
import { AuthProvider } from "@/app/_context/AuthContext"
import "./globals.css";

export const metadata: Metadata = {
  title: "Matchmaker",
  description: "Matchmaker Custom Games",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <AuthProvider>
          {children}
        </AuthProvider>
      </body>
    </html>
  );
}
