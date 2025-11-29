import React from 'react'
import Link from "next/link";
import { Box, Container, Text } from "@radix-ui/themes";

const Footer = () => {
    return (
        <footer>
            <Box pt="9">
            {/*TODO: figure out how to do this better with radix ui theme*/}
                <hr style={{color:"#262626"}} />
                <Container pt="6">
                    <Text as="p" size="2" align="center">
                        <Link href="/faq">Frequently Asked Questions</Link>
                    </Text>
                </Container>
            </Box>
        </footer>
    )
}

export default Footer
