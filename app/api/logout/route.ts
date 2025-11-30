import { logoutUserSession, getSessionCookieData } from '../../_lib/session.ts'

export async function POST(request: Request, response: Response) {
    const session = await getSessionCookieData()

    try {
        var sessionResult = await logoutUserSession(session.userId, session.sessionValue)
    } catch (error) {
        console.log('error', error)
        return new Response(JSON.stringify({error: 'error logging out'}), {status: 500})
    }

    return new Response(JSON.stringify({success: "success"}), {status: 200})
}
