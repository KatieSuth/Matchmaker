'use server'
import { cookies } from 'next/headers';
import { revalidateTag } from 'next/cache';
import { apiPOST } from '../_lib/client-utilities.ts';

export async function submitUserPreference(formData: FormData) {
    const showPronouns  = formData.get('showPronouns') as string;
    const cookie = (await cookies()).get('matchmaker_session')?.value

    let obj = {
        'showPronouns': showPronouns
    }

    const userUpdate = await apiPOST('update_user_preference', cookie, JSON.stringify(obj))
    console.log('userUpdate', userUpdate)

    /*
    const userUpdate = fetch(process.env.NEXT_PUBLIC_URL + '/api/update_user_preference', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Cookie': 'matchmaker_session='+cookie
        },
        body: JSON.stringify(obj)
    }).then((res) => {
        const resstatus = res.status
        if (resstatus == 401) {
            setHas401(true)
        } else if (resstatus !== 200) {
            throw new Error(res.status)
        }
        return res.json()
    }).then((data) => {
        if (!has401) {
          document.cookie = "matchmaker_session=" + data.session + ";" //TODO: add secure option before deploy

          if (data.newUser) {
              console.log('LOGIN DONE! redirect new user')
              const newUserUrl = process.env.NEXT_PUBLIC_URL + "/users/new";
              window.location.replace(newUserUrl)
          } else {
              console.log('LOGIN DONE! redirect to events')
              const eventsUrl = process.env.NEXT_PUBLIC_URL + "/events";
              window.location.replace(eventsUrl)
          }
        }

        //revalidateTag('showPronouns')
    }).catch((error) => {
        console.error(error)
        //TODO: handle this
    })
   */

    //revalidateTag('showPronouns')
}

export default submitUserPreference
