'use client'
import React, { useEffect, useState } from 'react'
import { getCookie } from "../_lib/client-utilities.ts";

const Logout = () => {
    const [hasError, setHasError] = useState(false)

    useEffect(() => {
        let cookie = getCookie("matchmaker_session");

        let data = fetch(process.env.NEXT_PUBLIC_URL + '/api/logout', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Cookie': 'matchmaker_session='+cookie
            },
        }).then((res) => {
            const resstatus = res.status
            if (resstatus !== 200) {
                throw new Error(res.status)
            }
            return res.json()
        }).then((data) => {
            document.cookie = "matchmaker_session=;expires=Thu, 01 Jan 1970 00:00:00 UTC;" //TODO: add secure option before deploy
            window.location.replace("/")
        }).catch((error) => {
            console.log(error)
            setHasError(true)
        })
    }, [])

    return (
        <div>
            {hasError === 1 ? (
                <>There was an error while trying to log out. Please clear your cookies and cache and try again.</>
            ) : (
                <>Please wait while you're logged out...</>
            )}
        </div>
    )
}

export default Logout
