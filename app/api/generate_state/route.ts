import { generateState } from '../utilities/security.ts'

export async function GET(request: Request, response: Response) {
    const state = await generateState()
    return new Response(JSON.stringify({state: state.state, cookie: state.cookie}), {status: 200})
}
