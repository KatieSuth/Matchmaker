import { validateSession } from '../../../_lib/api-utilities.ts'
import { getUserById } from '../../../_lib/user.ts'
import { cookies } from 'next/headers'

export async function GET(request: Request, response: Response) {
    const validSession = await validateSession()
    if (!validSession) {
        const response = {
            message: 'Unauthorized'
        }
        return new Response(JSON.stringify(response), {status: 401})
    }

    const cookie = (await cookies()).get('matchmaker_session')?.value
    const cookieArr = cookie.split(":");
    const userId = cookieArr[0]

    let userObj = await getUserById(userId);

    return new Response(JSON.stringify({userObj}), {status: 200})
}
