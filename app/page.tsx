//import Link from "next/link";
import { Box, Container, Section, Heading, Button, Text } from "@radix-ui/themes";
import { cookies } from "next/headers"
import LoginButton from './components/LoginButton/LoginButton';

export default function Home() {
    return (
        <main>
            <Box>
                <Container size="4">
                    <img src="/GalorantsBanner.jpg" alt="Galorants Banner" style={{borderRadius:"0px 0px 15px 15px"}} />
                </Container>
            </Box>
            <Box py="5">
                <Section>
                    <Heading align="center" size="8">Hello there!</Heading>
                    <Container py="5">
                        <Text as="p" align="center">
                            Welcome to the Galorants In-House Games signup site. This system allows members of the Galorants Discord server to register for in-house custom 5v5 games. To get started, please log in via Discord.
                        </Text>
                        <Text as="p" align="center" className="pt-8">
                            Please note per regulations in the EU and elsewhere we are required to inform you that this site uses cookies, but only for necessary, functional things. It does NOT store cookies for analytics, ads, etc, and it won't store anything until you click the Login button below.
                        </Text>
                        <div className="text-center pt-8">
                            <LoginButton />
                        </div>
                    </Container>
                </Section>
            </Box>
        </main>
    );
}
