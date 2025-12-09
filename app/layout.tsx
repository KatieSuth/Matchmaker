import type { Metadata } from "next";
import { headers } from "next/headers";
import { Geist, Geist_Mono } from "next/font/google";
import { Theme } from "@radix-ui/themes";
import "@radix-ui/themes/styles.css";
import "./globals.css";
import { Box, Container } from "@radix-ui/themes";
import Header from "./_components/Header/Header"
import Footer from "./_components/Footer"
import { getCurrentUser } from "./_lib/user"
//import { getSessionData } from "./_lib/security.ts"

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Matchmaker",
  description: "Matchmaker Custom Games",
};

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
    const header = await headers()
    const pathName = header.get('x-next-pathname') as string
    const loginRoutes = ['/', '/login_redirect']

    const isLoginRoute = loginRoutes.includes(pathName)

    let userIsAdmin = false,
        discordName = "Me",
        discordId = "",
        image = ""

    try {
        //note: this could be changed to use the useMyUser hook from SWR but I'm not doing that
        //because the data that's being used here are the discord name & profile image, which
        //will never be updated via this application, so it doesn't need revalidation like
        //other things that use that hook
        let currentUser = await getCurrentUser()

        if (currentUser?.isAdmin == 1) {
            userIsAdmin = true
        }
        discordName = currentUser.discordName
        discordId = currentUser.discordId
        image = currentUser.imageUrl
    } catch (e) {
        //console.log('e', e)
        //TODO: log the user out (if logged in)
        //for now, do nothing
    }

    return (
        <html lang="en">
            <body
                className={`${geistSans.variable} ${geistMono.variable} antialiased p-5`}
            >
                <Theme appearance="dark">
                    {isLoginRoute && (
                        <header>
                            <Box>
                                <Container size="4">
                                    <img src={process.env.NEXT_PUBLIC_SITE_BANNER} alt="Welcome Banner" style={{borderRadius:"0px 0px 15px 15px"}} />
                                </Container>
                            </Box>
                        </header>
                    )}

                    {!isLoginRoute && (
                        <Header
                            discordId={discordId}
                            discordName={discordName}
                            image={image}
                            isAdmin={userIsAdmin}
                        />
                    )}

                    <main style={{padding:"1rem"}}>
                        {children}
                    </main>

                    <Footer />
                </Theme>
            </body>
        </html>
    );
}
