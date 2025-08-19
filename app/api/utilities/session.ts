import { dbSelect, dbUpdate } from "../database/database.ts"
import { generateState } from "./security.ts"

export async function storeSessionData(userId: integer) {
    if (userId == null) {
        throw new Error('missing data')
    }

    const sessionId = await generateState()
    const createdAt = new Date();

    const sessionAdd = 'INSERT INTO session (userId, sessionValue, sessionIv, createdAt, device) VALUES (?, ?, ?, ?, ?)'
    const values = [
        userId,
        sessionId.state,
        sessionId.initVector,
        Math.floor(createdAt.getTime() / 1000),
        '' //TODO
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

    const sessionSelect = 'SELECT * FROM session WHERE userId = ? AND sessionValue = ?';

    const session = await dbSelect(sessionSelect, [userId, sessionValue])
        .then(data => {
            return data;
        }).catch(error => {
            console.error('Error fetching user: ', error.message);
            return []
        });

    return session
}
