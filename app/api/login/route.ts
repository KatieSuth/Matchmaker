import { getAccessTokens, verifyValidUser, verifyValidGalorantsMember, storeUserData } from '../utilities/user.ts'
import { storeSessionData } from '../utilities/session.ts'
import { validateState } from '../utilities/security.ts'

export async function POST(request: Request, response: Response) {
    const req = await request.json();

    //verify the state value we got back from Discord matches what we expect to help prevent cross-site scripting attacks
    if (!validateState(req.state, req.storedState)) {
        return new Response(JSON.stringify({error: 'OAuth2.0 authentication error'}), {status: 401})
    }
    try {
        //try to get user's access tokens
        var data = await getAccessTokens(req.code);
    } catch (error) {
        console.error('data (error)', error) //TODO: implement better logging
        return new Response(JSON.stringify({error: 'error getting user access token'}), {status: 500})
    }

    //verify we've got a valid user
    const validity = await Promise.all([
        verifyValidUser(data.access_token),
        verifyValidGalorantsMember(data.access_token)
    ]);

    if (validity[0]["id"] === -1 || validity[1] === false) {
        //we found an impostor; eject
        return new Response(JSON.stringify({error: 'not allowed'}), {status: 401})
    }

    //we have a user data object & they're a valid Galorants user; store/update their user info
    let userData = validity[0]
    userData.access_token = data.access_token
    userData.refresh_token = data.refresh_token

    const user = await storeUserData(userData)

    if (user.userId == -1) {
        return new Response(JSON.stringify({error: 'error storing up-to-date user info'}), {status: 500})
    }

    let newUser = false

    if (user.riotId == null ||
        user.currentRank == null ||
        user.peakRank == null ||
        user.region == null) {
        newUser = true
    }

    //store the session & return it as a cookie
    try {
        var sessionResult = await storeSessionData(user.userId)
    } catch (error) {
        console.log('error', error)
        return new Response(JSON.stringify({error: 'error storing session'}), {status: 500})
    }

    return new Response(JSON.stringify({newUser: newUser, session: sessionResult.userId + ":" + sessionResult.sessionId}), {status: 200})
}
