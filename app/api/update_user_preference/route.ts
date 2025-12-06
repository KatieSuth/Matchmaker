import { validateSession } from '../../_lib/api-utilities.ts'

export async function POST(request: Request, response: Response) {
    const validSession = await validateSession()
    if (!validSession) {
        const response = {
            message: 'Unauthorized'
        }
        return new Response(JSON.stringify(response), {status: 401})
    }

    const req = await request.json();

    //console.log('submit route request', req)

    return new Response(JSON.stringify({
        message: 'fixme'
    }), {status: 200})
}
