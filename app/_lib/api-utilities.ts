'use server'
import { decryptData } from "./security.ts"
import { getSession, deleteSessionById, updateSessionTime } from "./session.ts"
import { cookies } from 'next/headers'

export async function validateSession(): boolean {
    const cookie = (await cookies()).get('matchmaker_session')?.value
    if (cookie == "") {
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
