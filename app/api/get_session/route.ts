import { getSession } from '../../_lib/session.ts'
import { validateSession } from '../../_lib/api-utilities.ts'
import { decryptData } from "../../_lib/security.ts"
import { cookies } from 'next/headers'

export async function POST(request: Request, response: Response) {
    const req = await request.json();

    const validSession = await validateSession()
    if (!validSession) {
        const session = {
            userId: null
        }
        return new Response(JSON.stringify(session), {status: 401})
    }

    /*
     * this is a little repetitive considering we've already validated it,
     * but we need to have it in order to get the actual session for the
     * purpose of the endpoint--consider refactoring later
     */
    const cookie = (await cookies()).get('matchmaker_session')?.value
    const cookieArr = cookie.split(":");
    const userId = cookieArr[0]
    const encryptedSession = cookieArr[1],
          initVector = cookieArr[2]

    const sessionValue = await decryptData(encryptedSession, initVector);

    const session = await getSession(userId, sessionValue)

    return new Response(JSON.stringify(session[0]), {status: 200})
}
