'use server'
import { decryptData } from "./security.ts"
import { getSession, deleteSessionById, updateSessionTime } from "./session.ts"
import { cookies } from 'next/headers'


export async function validateSession(): boolean {
    const cookie = (await cookies()).get('matchmaker_session')?.value
    if (cookie == null || cookie == "") {
        return false
    }

    //decrypt the session from the cookie
    const cookieArr = cookie.split(":");
    const userId = cookieArr[0]
    const encryptedSession = cookieArr[1],
          initVector = cookieArr[2]

    const sessionValue = await decryptData(encryptedSession, initVector);
    const currentTime = new Date();

    try {
        let session = await getSession(userId, sessionValue)
        if (session.length == 0) {
            //didn't find a session
            return false
        } else if (session[0].expires < Math.floor(currentTime.getTime() / 1000)) {
            //session has expired; delete it & return false
            deleteSessionById(session[0].sessionId)
            return false
        } else {
            //it's a valid session, make sure it stays that way & return true
            updateSessionTime(session[0])
            return true
        }
    } catch (error) {
        return false
        console.error('Could not validate the session: ' + error)
    }
}

/*
 * utility function for calling an authenticated POST request. It's pretty
 * flexible in terms of usage, but per request do evaluate whether it would
 * be best to use this or just write out the fetch request (maybe you don't
 * always need to wait for the response, for example?)
 */
export async function apiPOST(path: string, body?: string, headers?: obj, opts?: obj) {
    if (body == null) {
        body = JSON.stringify({})
    }

    const cookie = (await cookies()).get('matchmaker_session')?.value

    try {
        var response = await fetch(process.env.NEXT_PUBLIC_URL + '/api/' + path, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Cookie': 'matchmaker_session='+cookie,
                ...headers
            },
            body: body,
            ...opts
        }).then((res) => {
            return res.json()
        }).then((data) => {
            return data
        }).catch((error) => {
            throw error
        })
    } catch (error) {
        throw error
    }

    return response;
}

/*
 * utility function for calling an authenticated GET request. It's pretty
 * flexible in terms of usage, but per request do evaluate whether it would
 * be best to use this or just write out the fetch request (maybe you don't
 * always need to wait for the response, for example?)
 */
export async function apiGET(path: string, headers?: obj, opts?: obj) {
    const cookie = (await cookies()).get('matchmaker_session')?.value

    try {
        var response = await fetch(process.env.NEXT_PUBLIC_URL + '/api/' + path, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
                'Cookie': 'matchmaker_session='+cookie,
                ...headers
            },
            ...opts
        }).then((res) => {
            return res.json()
        }).then((data) => {
            return data
        }).catch((error) => {
            throw error
        })
    } catch (error) {
        throw error
    }

    return response;
}
