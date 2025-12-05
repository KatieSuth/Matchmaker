'use client'

import React, { useEffect, useState } from 'react'
import { useSearchParams } from 'next/navigation'
import { Box, Container, Section, Heading, Button, Text } from "@radix-ui/themes";
import { getCookie } from "../_lib/client-utilities.ts";
import LoginButton from '../_components/LoginButton/LoginButton';

const ErrorResponse = () => {
    return (
        <Section>
            <Heading align="center" size="8">Something Went Wrong...</Heading>
            <Container py="5">
                <Text as="p" align="center">
                    Unfortunately, we couldn't log you in. If you canceled the login request from Discord by mistake, no worries! You can try to log in again below. If you keep getting this error message or if you have concerns about using this application for custom game registration, please reach out to the server mods and/or custom game event organizer.
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
                   {process.env.NEXT_PUBLIC_LIMIT_BY_DISCORD_SERVER === 1 ? (
                       <>
                            Sorry; as far as we can tell, you're not a member of the {process.env.NEXT_PUBLIC_LIMITING_DISCORD_NAME} Discord server. If you think you're seeing this message in error, you probably are! Not only is this a new system that may still have a few bugs hanging around, but also sometimes Discord services are just down and we can't do our verification process. Please wait a couple minutes and try again, and if you keep getting this message, reach out to the server mods and/or in-house event organizer and they'll help figure out what's going wrong. If you really aren't a member of the {process.env.NEXT_PUBLIC_LIMITING_DISCORD_NAME} server and would like to join, please visit us at <a href={process.env.NEXT_PUBLIC_DISCORD_LINK}>{process.env.NEXT_PUBLIC_DISCORD_LINK}</a>.
                        </>
                    ) : (
                        <>
                            Sorry; we were unable to log you in. Please contact a site admin or submit an issue on our <a href="https://github.com/KatieSuth/Matchmaker">Github</a>.
                        </>
                    )}
                </Text>
                <div className="text-center pt-8">
                    <LoginButton />
                </div>
            </Container>
        </Section>
    )
}

const SuccessResponse = ({code, state}) => {
    const [hasError, setHasError] = useState(false),
          [has401, setHas401] = useState(false)

    useEffect(() => {
        let storedState = getCookie("matchmaker_state");

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
              document.cookie = "matchmaker_session=" + data.session + ";" //TODO: add secure option before deploy

              if (data.newUser) {
                  console.log('LOGIN DONE! redirect new user')
                  const newUserUrl = process.env.NEXT_PUBLIC_URL + "/my_account";
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
                   {process.env.NEXT_PUBLIC_LIMIT_BY_DISCORD_SERVER === 1 ? (
                       <>
                            Please wait while we make sure you're a member of the {process.env.NEXT_PUBLIC_LIMITING_DISCORD_NAME} Discord server...
                        </>
                    ) : (
                        <>
                            Please wait while we log you in...
                        </>
                    )}
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
        <div>
            <Box py="5">
                {(error != null || code == null || state == null) && (
                    <ErrorResponse />
                )}
                {(code != null && state != null) && (
                    <SuccessResponse code={code} state={state} />
                )}
            </Box>
        </div>
    )
}

export default LoginRedirect
