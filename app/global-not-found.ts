import { Box, Container, Section, Heading, Text } from "@radix-ui/themes";
import Link from "next/link"
import { cookies } from "next/headers"
import LoginButton from './_components/LoginButton/LoginButton';

export default function NotFound() {
    return (
        <div>
            <Box py="5">
                <Section>
                    <Heading align="center" size="8">Oops!</Heading>
                    <Container py="5">
                        <Text as="p" align="center">
                            We're not sure what you're looking for. The link may have moved, or you may have a typo.
                        </Text>
                        <div className="text-center pt-8">
                            <Link href="/">Return Home</Link>
                        </div>
                    </Container>
                </Section>
            </Box>
        </div>
    );
}
