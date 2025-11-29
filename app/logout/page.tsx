'use client'
import React, { useEffect } from 'react'

const Logout = () => {

    useEffect(() => {
        let data = fetch(process.env.NEXT_PUBLIC_URL + '/api/logout', {
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
              document.cookie = "matchmaker_session=;expires=Thu, 01 Jan 1970 00:00:00 UTC;" //TODO: add secure option before deploy

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

    return (
        <div>
            Upcoming events will be shown here!
        </div>
    )
}

export default Logout
