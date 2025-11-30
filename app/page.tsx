//import Link from "next/link";
import { Box, Container, Section, Heading, Button, Text } from "@radix-ui/themes";
import { cookies } from "next/headers"
import LoginButton from './_components/LoginButton/LoginButton';

export default function Home() {
    return (
        <div>
            <Box py="5">
                <Section>
                    <Heading align="center" size="8">Hello there!</Heading>
                    <Container py="5">
                        <Text as="p" align="center">
                           {process.env.NEXT_PUBLIC_LIMIT_BY_DISCORD_SERVER === 1 ? (
                               <>
                                    Welcome to the {process.env.NEXT_PUBLIC_LIMITING_DISCORD_NAME} In-House Games signup site. This system allows members of the {process.env.NEXT_PUBLIC_LIMITING_DISCORD_NAME} Discord server to register for in-house custom 5v5 games. To get started, please log in via Discord.
                                </>
                            ) : (
                                <>
                                    Welcome to Matchmaker, an application that helps make custom competitive games fairer for the competitors and easier to organize for admins. To get started, please log in via Discord.
                                </>
                            )}
                        </Text>
                        <Text as="p" align="center" className="pt-8">
                            Please note per regulations in the EU and elsewhere we are required to inform you that this site uses cookies, but only for necessary, functional things. It does NOT store cookies for analytics, ads, etc, and it won't store anything until you click the Login button below. If your browser has cookies entirely disabled, the site will not work as expected.
                        </Text>
                        <div className="text-center pt-8">
                            <LoginButton />
                        </div>
                    </Container>
                </Section>
            </Box>
        </div>
    );
}
