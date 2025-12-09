export function getCookie(cname: string) {
    let name = cname + "=";
    let decodedCookie = decodeURIComponent(document.cookie);
    let ca = decodedCookie.split(';');
    for(let i = 0; i <ca.length; i++) {
        let c = ca[i];
        while (c.charAt(0) == ' ') {
            c = c.substring(1);
        }
        if (c.indexOf(name) == 0) {
            return c.substring(name.length, c.length);
        }
    }
    return "";
}

/*
export async function apiPOST(path: string, sessionCookie: string, body?: string, headers?: obj, opts?: obj) {
    if (body == null) {
        body = JSON.stringify({})
    }

    try {
        var response = await fetch(process.env.NEXT_PUBLIC_URL + '/api/' + path, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Cookie': 'matchmaker_session='+sessionCookie,
                ...headers
            },
            body: body,
            ...opts
        }).then((res) => {
            return res.json()
        }).then((data) => {
            return data
        }).catch((error) => {
            throw error
        })
    } catch (error) {
        throw error
    }

    return response;
}

export async function apiGET(path: string, sessionCookie: string, headers?: obj, opts?: obj) {
    try {
        var response = await fetch(process.env.NEXT_PUBLIC_URL + '/api/' + path, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
                'Cookie': 'matchmaker_session='+sessionCookie,
                ...headers
            },
            ...opts
        }).then((res) => {
            return res.json()
        }).then((data) => {
            return data
        }).catch((error) => {
            throw error
        })
    } catch (error) {
        throw error
    }

    return response;
}
*/
