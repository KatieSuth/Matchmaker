import { NextRequest, NextResponse } from 'next/server'
import { cookies } from 'next/headers'
import { decryptData } from "./app/api/utilities/security.ts"

//public routes
const publicRoutes = ['/', '/login_redirect']
const adminRoutes = []

export default async function middleware(req: NextRequest) {
    //check if the current route is protected or public
    const path = req.nextUrl.pathname

    const isPublicRoute = publicRoutes.includes(path),
          isAdminRoute = adminRoutes.includes(path)

    //decrypt the session from the cookie
    const cookie = (await cookies()).get('galorants_session')?.value
    
    let userId = null,
        session = {}

    if (cookie != null) {
        const cookieArr = cookie.split(":");
        userId = cookieArr[0]
        const encryptedSession = cookieArr[1],
              initVector = cookieArr[2]

        const sessionValue = await decryptData(encryptedSession, initVector);

        const body = {
            'userId': userId,
            'sessionValue': sessionValue,
            'sessionKey': process.env.SESSION_FETCH_KEY
        }

        session = await fetch(process.env.NEXT_PUBLIC_URL + '/api/get_session', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(body),
            cache: 'no-store'
        }).then((res) => {
            return res.json()
        }).then((data) => {
            return data
        }).catch((error) => {
            return {
                userId: null
            }
        })
    }

    if (!isPublicRoute && (userId == null || session?.userId == null)) {
        //redirect to home if we don't have a valid session for the user & the route isn't public
        return NextResponse.redirect(new URL('/', req.nextUrl))
    } else if (isPublicRoute && session?.userId != null) {
        //if the route is public & we have a user ID with valid session, just redirect them to the events dashboard
        return NextResponse.redirect(new URL('/events', req.nextUrl))
    }
    //TODO: implement admin
    //redirect to events if we have a valid session but the user isn't admin

    return NextResponse.next()
}

//routes middleware shouldn't run on
export const config = {
    matcher: ['/((?!api|_next/static|_next/image|.*\\.png$|.*\\.jpg$|.*\\.ico$).*)'],
}
