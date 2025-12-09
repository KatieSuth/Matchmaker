import { NextRequest, NextResponse } from 'next/server'
import { apiPOST } from './app/_lib/api-utilities.ts'
import { cookies } from 'next/headers'

//public routes
const publicRoutes = ['/', '/login_redirect']
const adminRoutes = ['/admin_events', '/admin_users']

export default async function proxy(req: NextRequest) {
    //check if the current route is protected or public
    const path = req.nextUrl.pathname

    const cookie = (await cookies()).get('matchmaker_session')?.value

    const isPublicRoute = publicRoutes.includes(path),
          isAdminRoute = adminRoutes.includes(path)

    try {
        var session = await apiPOST('get_session')
    } catch (error) {
        console.error(error)
    }

    if (!isPublicRoute && session?.userId == null) {
        //redirect to home if we don't have a valid session for the user & the route isn't public
        return NextResponse.redirect(new URL('/', req.nextUrl))
    } else if (isPublicRoute && session?.userId != null) {
        //if the route is public & we have a user ID with valid session, just redirect them to the events dashboard
        return NextResponse.redirect(new URL('/events', req.nextUrl))
    }
    //TODO: implement admin
    //redirect to events if we have a valid session but the user isn't admin

    const requestHeaders = new Headers(req.headers)
    requestHeaders.set('x-next-pathname', req.nextUrl.pathname)
    return NextResponse.next({ request: { headers: requestHeaders} })
}

//routes middleware shouldn't run on
export const config = {
    matcher: ['/((?!api|_next/static|_next/image|.*\\.png$|.*\\.jpg$|.*\\.ico$).*)'],
}
