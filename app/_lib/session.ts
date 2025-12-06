import { dbSelect, dbUpdate } from "./database/database.ts"
import { cookies } from 'next/headers'
import { generateState, decryptData } from "./security.ts"

export async function storeSessionData(userId: integer) {
    if (userId == null) {
        throw new Error('missing data')
    }

    const sessionId = await generateState()
    const createdAt = new Date();
    const nowEpoch = parseInt(Math.floor(createdAt.getTime() / 1000))

    const sessionAdd = 'INSERT INTO session (userId, sessionValue, sessionIv, createdAt, expires, lastUpdate) VALUES (?, ?, ?, ?, ?, ?)'
    const values = [
        userId,
        sessionId.state,
        sessionId.initVector,
        nowEpoch,
        nowEpoch + parseInt(process.env.SESSION_EXPIRATION),
        nowEpoch
    ];

    //store the session
    const updateResult = await dbUpdate(sessionAdd, values)
        .then(data => {
            return true
        }).catch(error => {
            console.error('Error storing session: ', error.message);
            throw error;
        });

    return {
        'userId': userId,
        'sessionId': sessionId.cookie,
    }
}

export async function getSession(userId: integer, sessionValue: string) {
    if (userId == null || sessionValue == null) {
        throw new Error('missing data')
    }

    //make sure we have both a valid session and valid user
    const sessionSelect = `
        SELECT S.* FROM session AS S
        JOIN user AS U ON (S.userId = U.userId)
        WHERE S.userId = ? AND S.sessionValue = ?
    `;

    const session = await dbSelect(sessionSelect, [userId, sessionValue])
        .then(data => {
            return data;
        }).catch(error => {
            console.error('Error fetching user: ', error.message);
            return []
        });

    return session
}

export async function deleteSessionById(sessionId: integer) {
    if (sessionId == null) {
        throw new Error('missing data')
    }

    const sessionDelete = 'DELETE FROM session WHERE sessionId = ?'
    const session = await dbUpdate(sessionDelete, [sessionId])
        .then(data => {
            return true
        }).catch(error => {
            console.error('Error deleting session: ', error.message);
            throw error;
        });
}

export async function updateSessionTime(session: object, skipLastUpdateCheck: boolean = false) {
    if (session == null) {
        throw new Error('missing data')
    }

    const currentTime = new Date();
    const nowEpoch = parseInt(Math.floor(currentTime.getTime() / 1000))
    const newExpires = nowEpoch + parseInt(process.env.SESSION_EXPIRATION)

    let doUpdate = true
    /*
     * the expiration is set by default to a week, we don't need to update every single time
     * the user interacts with a page (unless we purposely skip this check)--only do this
     * update if it hasn't been done in the last (default value) 5 minutes
     */
    if (!skipLastUpdateCheck && nowEpoch - session.lastUpdate < process.env.SESSION_REFRESH) {
        doUpdate = false
    }

    if (doUpdate) {
        const sessionUpdate = 'UPDATE session SET expires = ?, lastUpdate = ? WHERE sessionId = ?'
        const updateSession = await dbUpdate(sessionUpdate, [newExpires, nowEpoch, session.sessionId])
            .then(data => {
                return true
            }).catch(error => {
                console.error('Error storing session update: ', error.message);
                throw error;
            });
    }
}

export async function getSessionCookieData() {
    //decrypt the session from the cookie
    const cookie = (await cookies()).get('matchmaker_session')?.value
    
    let userId = null,
        sessionValue = null,
        sessionKey = ""

    if (cookie != null) {
        const cookieArr = cookie.split(":");
        userId = cookieArr[0]
        const encryptedSession = cookieArr[1],
              initVector = cookieArr[2]

        sessionValue = await decryptData(encryptedSession, initVector);
    }

    return {
        'userId': userId,
        'sessionValue': sessionValue
    }
}

export async function logoutUserSession(userId: integer, sessionValue: string) {
    if (userId == null || sessionValue == null) {
        throw new Error('missing data')
    }

    const sessionDelete = 'DELETE FROM session WHERE userId = ? AND sessionValue = ?';

    //store the session
    const deleteResult = await dbUpdate(sessionDelete, [userId, sessionValue])
        .then(data => {
            return true
        }).catch(error => {
            console.error('Error deleting session: ', error.message);
            throw error;
        });

    return true
}
