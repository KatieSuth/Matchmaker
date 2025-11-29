'use client'
import React from 'react'
import { Button } from "@radix-ui/themes";
import { getCookie } from "../../_lib/client-utilities.ts";

const LoginButton = () => {
    const generateLoginState = () => {
        fetch(process.env.NEXT_PUBLIC_URL + '/api/generate_state', {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
            }
        }).then((res) => res.json())
          .then((data) => {
              document.cookie = "matchmaker_state=" + data.cookie + "; max-age=300" //TODO: add secure option before deploy
              const loginUrl = process.env.NEXT_PUBLIC_DISCORD_OAUTH + "&state=" + data.state;
              window.location.replace(loginUrl)
          })
    }

    return (
        <div>
            <Button asChild>
                <a onClick={generateLoginState}>Login</a>
            </Button>
        </div>
    )
}

export default LoginButton
