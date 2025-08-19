'use client'

import React, { useEffect, useState } from 'react'
import { useSearchParams } from 'next/navigation'
import { Box, Container, Section, Heading, Button, Text } from "@radix-ui/themes";
import { getCookie } from "../client-components/utilities.ts";
import LoginButton from '../components/LoginButton/LoginButton';

const ErrorResponse = () => {
    return (
        <Section>
            <Heading align="center" size="8">Something Went Wrong...</Heading>
            <Container py="5">
                <Text as="p" align="center">
                    Unfortunately, we couldn't log you in. If you canceled the login request from Discord by mistake, no worries! You can try to log in again below. If you keep getting this error message or if you have concerns about using this application for in-house registration, please reach out to the server mods and/or in-house event organizer.
                </Text>
                <div className="text-center pt-8">
                    <LoginButton />
                </div>
            </Container>
        </Section>
    )
}

const NotAllowedResponse = () => {
    return (
        <Section>
            <Heading align="center" size="8">Something Went Wrong...</Heading>
            <Container py="5">
                <Text as="p" align="center">
                    Sorry; as far as we can tell, you're not a member of the Galorants server. If you think you're seeing this message in error, you probably are! Not only is this a new system, but also sometimes Discord services are just down and we can't do our verification process. Please reach out to the server mods and/or in-house event organizer and they'll take a look at our connections to Discord to see what went wrong. If you really aren't a member of the Galorants server and would like to join, please visit us at <a href="http://discord.gg/galorants">http://discord.gg/galorants</a>.
                </Text>
            </Container>
        </Section>
    )
}

const SuccessResponse = ({code, state}) => {
    const [hasError, setHasError] = useState(false),
          [has401, setHas401] = useState(false)

    useEffect(() => {
        let storedState = getCookie("galorants_state");

        let obj = {
            'code': code,
            'state': state,
            'storedState': storedState
        }

        let data = fetch(process.env.NEXT_PUBLIC_URL + '/api/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(obj)
        }).then((res) => {
            const resstatus = res.status
            if (resstatus == 401) {
                setHas401(true)
            } else if (resstatus !== 200) {
                throw new Error(res.status)
            }
            return res.json()
        }).then((data) => {
            if (!has401) {
              document.cookie = "galorants_session=" + data.session + ";" //TODO: add secure option before deploy

              if (data.newUser) {
                  console.log('LOGIN DONE! redirect new user')
                  const newUserUrl = process.env.NEXT_PUBLIC_URL + "/users/new";
                  window.location.replace(newUserUrl)
              } else {
                  console.log('LOGIN DONE! redirect to events')
                  const eventsUrl = process.env.NEXT_PUBLIC_URL + "/events";
                  window.location.replace(eventsUrl)
              }
            }
        }).catch((error) => {
            setHasError(true)
        })
    }, [])

    if (has401) {
        return (
            <NotAllowedResponse />
        )
    }

    if (hasError) {
        return (
            <ErrorResponse />
        )
    }

    return (
        <Section>
            <Heading align="center" size="8">Please wait...</Heading>
            <Container py="5">
                <Text as="p" align="center">
                    Please wait while we make sure you're a member of the Galorants Discord server...
                </Text>
            </Container>
        </Section>
    )
}

const LoginRedirect = () => {
    const params = useSearchParams()
    const error = params.get('error')
    const code = params.get('code')
    const state = params.get('state')

    return (
        <>
            <Box>
                <Container size="4">
                    <img src="/GalorantsBanner.jpg" alt="Galorants Banner" style={{borderRadius:"0px 0px 15px 15px"}} />
                </Container>
            </Box>
            <Box py="5">
                {(error != null || code == null || state == null) && (
                    <ErrorResponse />
                )}
                {(code != null && state != null) && (
                    <SuccessResponse code={code} state={state} />
                )}
            </Box>
        </>
    )
}

export default LoginRedirect
