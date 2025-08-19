import { dbSelect, dbUpdate } from "../database/database.ts"
import { encryptData } from "./security.ts"

export async function getAccessTokens(code: string) {
    return fetch(process.env.DISCORD_API + '/oauth2/token', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded'
        },
        body: new URLSearchParams({
            'grant_type': 'authorization_code',
            'code': code,
            'client_id': process.env.DISCORD_CLIENT_ID,
            'client_secret': process.env.DISCORD_CLIENT_SECRET,
            'scope': 'identify',
            'redirect_uri': 'http://localhost:3000/login_redirect'
        })
    }).then((res) => {
        return res.json()
    }).then((data) => {
        if (data.error != null) {
            throw new Error(data.error + ": " + data.error_description)
        }

        return data
    }).catch((error) => {
        throw new Error(error)
    })
}

export async function refreshAccess(refresh_token: string) {
    //TODO: test
    return fetch(process.env.DISCORD_API + '/oauth2/token', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded'
        },
        body: new URLSearchParams({
            'grant_type': 'refresh_token',
            'refresh_token': refresh_token,
            'client_id': process.env.DISCORD_CLIENT_ID,
            'client_secret': process.env.DISCORD_CLIENT_SECRET,
        })
    }).then((res) => res.json())
      .then((data) => {
          console.log(data)
    })
}

export async function verifyValidUser(access_token: string) {
    const failResponse = {
        "id": -1,
        "username": "",
        "avatar": "",
    }

    return fetch(process.env.DISCORD_API + '/users/@me', {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer ' + access_token
        },
    }).then((res) => res.json())
      .then((data) => {
          if (data.id == null) {
              console.error('failed users/@me')
              return failResponse
          } else {
              return {
                  "id": data.id,
                  "username": data.username,
                  "avatar": data.avatar,
              }
          }
    }).catch((error) => {
        console.error(error)
        return failResponse
    })
}

export async function verifyValidGalorantsMember(access_token: string) {
    return fetch(process.env.DISCORD_API + '/users/@me/guilds', {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer ' + access_token
        },
    }).then((res) => res.json())
      .then((data) => {
          const result = Object.keys(data).find(i => data[i].id == process.env.GALORANTS_DISCORD_ID)

          if (result != null) {
              return true
          }

          return false
      })
}

export async function storeUserData(userData: object) {
    const tokens = await Promise.all([
        encryptData(userData.access_token),
        encryptData(userData.refresh_token)
    ]);

    const accessToken = tokens[0];
    const refreshToken = tokens[1];

    //check if we have the user already first
    const userSelect = 'SELECT * FROM user WHERE discordId = ?';

    const user = await dbSelect(userSelect, [userData.id])
        .then(data => {
            return data;
        }).catch(error => {
            console.error('Error fetching user: ', error.message);
            return []
        });

    //set up the statement to insert or update as necessary
    let userUpdate = '';
    let values = []
    if (user.length > 0) { //found a user
        userUpdate = 'UPDATE user SET discordId = ?, discordName = ?, imageUrl = ?, accessToken = ?, accessIv = ?, refreshToken = ?, refreshIv = ? WHERE userId = ?'
        values.push(
            userData.id,
            userData.username,
            userData.avatar,
            accessToken.encryptedData,
            accessToken.initVector,
            refreshToken.encryptedData,
            refreshToken.initVector,
            user[0].userId
       );
    } else {
        userUpdate = 'INSERT INTO user (discordId, discordName, imageUrl, accessToken, accessIv, refreshToken, refreshIv) VALUES (?, ?, ?, ?, ?, ?, ?)'
        values.push(
            userData.id,
            userData.username,
            userData.avatar,
            accessToken.encryptedData,
            accessToken.initVector,
            refreshToken.encryptedData,
            refreshToken.initVector
        );
    }

    //perform the update
    const updateResult = await dbUpdate(userUpdate, values)
        .then(data => {
            return true
        }).catch(error => {
            console.error('Error storing user: ', error.message);
            return false
        });

    if (!updateResult) {
        console.log('update failed', updateResult)
        return {userId: -1};
    }

    //make sure we actually did store the user & get their user ID to return to the api endpoint
    const validateUser = await dbSelect(userSelect, [userData.id])
        .then(data => {
            return data;
        }).catch(error => {
            console.error('Error fetching user: ', error.message);
            return []
        });

    if (user.length > 0) {
        return user[0];
    } else {
        console.log('user length 0: ', user)
        return {userId: -1};
    }
}
