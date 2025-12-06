'use client'
import React, { useActionState } from 'react'
import { useFormStatus } from 'react-dom'
import { Flex, Box, DropdownMenu, TextField } from '@radix-ui/themes'
import { regionOpts, pronounOpts, rankOpts } from '../../_constants/options.ts'
import { colourStyles } from '../../_constants/selectStyles.ts'
import submitUserPreference from '../../_actions/submitUserPreferenceForm.ts'
import Select from 'react-select'

export default function UserForm() {
    //const [state, formAction, pending] = useActionState(submitUserPreference, null);

    return (
        <form action={submitUserPreference}>
            <label htmlFor="showPronouns">Publicly display my pronouns: </label>
            <input type="checkbox" name="showPronouns" />
            <br />
            <button type="submit">Submit</button>
        </form>
    )
}


/*
const initialState = {
    message: "",
};

function SubmitButton() {
    const { pending } = useFormStatus();

    return (
        <button type="submit" className="bg-pink-600 hover:bg-pink-800 text-white font-bold py-2 px-4 rounded" aria-disabled={pending}>
            Update Profile
        </button>
    );
}

//TODO: move out of client component
async function updateUser(formData: FormData) {
*/
    /*
    'use server'
0|userId|INTEGER|0||1
1|discordId|TEXT|1||0
2|discordName|TEXT|1||0
3|imageUrl|TEXT|1||0
4|riotId|TEXT|0||0
5|pronouns|TEXT|0||0
6|currentRank|TEXT|0||0
7|peakRank|TEXT|0||0
8|region|TEXT|0||0
9|accessToken|TEXT|0||0
10|accessIv|TEXT|0||0
11|refreshToken|TEXT|0||0
12|refreshIv|TEXT|0||0
13|isAdmin|INTEGER|1|0|0
     */

    /*
    const rawFormData = {
        region: formData.get('region'),
        riotId: formData.get('riotId'),
        pronouns: formData.get('pronouns'),
        currentRank: formData.get('currentRank'),
        peakRank: formData.get('peakRank')
    }

    console.log('rawFormData', rawFormData)
}

export default function UserForm() {

    const [state, formAction] = useActionState(updateUser, initialState)

    return (
        <form action={formAction}>
            <Flex direction="column" gap="3">
                <Box>
                    <label htmlFor="region">Region</label>
                    <Select
                        className="my-react-select-container"
                        classNamePrefix="my-react-select"
                        defaultValue={regionOpts[0]}
                        id="region"
                        name="region"
                        options={regionOpts}
                    />
                </Box>
                <Box>
                    <label htmlFor="region">Preferred Pronouns</label>
                    {/*styles={colourStyles}*/
    /*
}
                   /*
                    <Select
                        className="my-react-select-container"
                        classNamePrefix="my-react-select"
                        isMulti
                        name="pronouns"
                        options={pronounOpts}
                    />
                </Box>
                <Box>
                    <label htmlFor="region">Riot ID</label>
                    {/*<input type="text" name="riotId" />*/
       /*
}
                    <TextField.Root size="3" placeholder="Riot ID (ex: matchmaker#tag)" name="riotId" />
                </Box>
                <Box>
                    <label htmlFor="region">Current Competitive Rank</label>
                    <Select
                        className="my-react-select-container"
                        classNamePrefix="my-react-select"
                        defaultValue={rankOpts[0]}
                        name="currentRank"
                        options={rankOpts}
                    />
                </Box>
                <Box>
                    <label htmlFor="region">Peak Competitive Rank (within past 4 acts)</label>
                    <Select
                        className="my-react-select-container"
                        classNamePrefix="my-react-select"
                        defaultValue={rankOpts[0]}
                        name="peakRank"
                        options={rankOpts}
                    />
                </Box>
                <Box>
                    <SubmitButton />
                </Box>
            </Flex>
            <p aria-live="polite" className="sr-only" role="status">
                {state?.message}
            </p>
        </form>
    )
}
*/
