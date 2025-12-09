import React from 'react'
import { Box, Container, Section, Heading, Button, Text } from "@radix-ui/themes";
import UserForm from '../_components/UserForm/UserForm.tsx'

const MyAccount = () => {
    return (
        <div>
            <Box>
                <Section>
                    <Heading align="center" size="8">My Account Preferences</Heading>
                    <Container py="5">
                        <UserForm />
                    </Container>
                </Section>
            </Box>
        </div>
    )
}

export default MyAccount
