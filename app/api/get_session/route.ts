import { getSession } from '../utilities/session.ts'

export async function POST(request: Request, response: Response) {
    const req = await request.json();

    if (req.sessionKey != process.env.SESSION_FETCH_KEY) {
        const session = {
            userId: null
        }
        return new Response(JSON.stringify(session), {status: 401})
    }

    const session = await getSession(req.userId, req.sessionValue)

    return new Response(JSON.stringify(session[0]), {status: 200})
}
