'use client'
import React, { useActionState, useEffect } from 'react'
import { useFormStatus } from 'react-dom'
import Select from 'react-select'
//import { Flex, Box, DropdownMenu, TextField } from '@radix-ui/themes'
import { Box, Container, Section, Heading, Button, Text } from "@radix-ui/themes";
import { riotRegionOpts, gameTypeOpts, pronounOpts, valorantRankOpts } from '../../_constants/options.ts'
import { colourStyles } from '../../_constants/selectStyles.ts'
import styles from './UserForm.module.css';
import submitUserPreference from '../../_actions/submitUserPreferenceForm.ts'
import useMyUser from '../../_hooks/useMyUser.ts'


export default function UserForm() {
    //const [state, formAction, pending] = useActionState(submitUserPreference, null);
    const { user, isLoading, isError } = useMyUser()

    console.log('data', user)
    console.log('loading', isLoading)
    console.log('error', isError)
    if (isLoading) {
        return (
            <div role="status">
                <svg aria-hidden="true" className="inline w-10 h-10 w-8 h-8 text-neutral-tertiary animate-spin fill-brand" viewBox="0 0 100 101" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path d="M100 50.5908C100 78.2051 77.6142 100.591 50 100.591C22.3858 100.591 0 78.2051 0 50.5908C0 22.9766 22.3858 0.59082 50 0.59082C77.6142 0.59082 100 22.9766 100 50.5908ZM9.08144 50.5908C9.08144 73.1895 27.4013 91.5094 50 91.5094C72.5987 91.5094 90.9186 73.1895 90.9186 50.5908C90.9186 27.9921 72.5987 9.67226 50 9.67226C27.4013 9.67226 9.08144 27.9921 9.08144 50.5908Z" fill="currentColor"/>
                    <path d="M93.9676 39.0409C96.393 38.4038 97.8624 35.9116 97.0079 33.5539C95.2932 28.8227 92.871 24.3692 89.8167 20.348C85.8452 15.1192 80.8826 10.7238 75.2124 7.41289C69.5422 4.10194 63.2754 1.94025 56.7698 1.05124C51.7666 0.367541 46.6976 0.446843 41.7345 1.27873C39.2613 1.69328 37.813 4.19778 38.4501 6.62326C39.0873 9.04874 41.5694 10.4717 44.0505 10.1071C47.8511 9.54855 51.7191 9.52689 55.5402 10.0491C60.8642 10.7766 65.9928 12.5457 70.6331 15.2552C75.2735 17.9648 79.3347 21.5619 82.5849 25.841C84.9175 28.9121 86.7997 32.2913 88.1811 35.8758C89.083 38.2158 91.5421 39.6781 93.9676 39.0409Z" fill="currentFill"/>
                </svg>
                <span className="sr-only">Loading...</span>
            </div>
        )
    }

    return (
        <form action={submitUserPreference}>
            <table>
                <tbody>
                    <tr>
                        <td colSpan="2">
                            <Heading align="left" size="6">General</Heading>
                            <hr />
                        </td>
                    </tr>
                    <tr>
                        <td><label>Discord Name:</label></td>
                        <td><label>{user.userObj.discordName}</label></td>
                    </tr>
                    <tr>
                        <td><label htmlFor="showPronouns">Publicly display my pronouns: </label></td>
                        <td>
                            <label className={styles.switch}>
                                <input type="checkbox" name="showPronouns" />
                                <span className={`${styles.slider} ${styles.round}`} />
                            </label>
                        </td>
                    </tr>
                    <tr>
                        <td><label htmlFor="pronouns">My pronouns: </label></td>
                        <td>
                            <Select
                                className="my-react-select-container"
                                classNamePrefix="my-react-select"
                                name="pronouns"
                                options={pronounOpts}
                            />
                            <br />
                            <input type="text" name="otherPronouns" className="bg-neutral-secondary-medium border border-default-medium text-heading text-sm rounded-base focus:ring-brand focus:border-brand block w-full px-3 py-2.5 shadow-xs placeholder:text-body" placeholder="Enter your preferred pronouns" />
                        </td>
                    </tr>
                    <tr>
                        <td colSpan="2">
                            <Heading align="left" size="6">Games</Heading>
                            <hr />
                        </td>
                    </tr>
                    <tr>
                        <td><label htmlFor="gameTypes">Games I play: </label></td>
                        <td>
                            <Select
                                className="my-react-select-container"
                                classNamePrefix="my-react-select"
                                name="gameTypes"
                                isMulti
                                options={gameTypeOpts}
                            />
                        </td>
                    </tr>

                    <tr>
                        <td colSpan="2">
                            {/*TODO: repeat this per game*/}
                            <Heading align="left" size="4">Valorant</Heading>
                        </td>
                    </tr>
                    <tr>
                        <td><label htmlFor="region-val">Region: </label></td>
                        <td>
                            <Select
                                className="my-react-select-container"
                                classNamePrefix="my-react-select"
                                name="region-val"
                                options={riotRegionOpts}
                            />
                        </td>
                    </tr>
                    <tr>
                        <td><label htmlFor="accountName">Account Name: </label></td>
                        <td>
                            <input type="text" name="accountName" className="bg-neutral-secondary-medium border border-default-medium text-heading text-sm rounded-base focus:ring-brand focus:border-brand block w-full px-3 py-2.5 shadow-xs placeholder:text-body" placeholder="myaccount#123" />
                        </td>
                    </tr>
                    <tr>
                        <td><label htmlFor="currentRank-val">Current Rank: </label></td>
                        <td>
                            <Select
                                className="my-react-select-container"
                                classNamePrefix="my-react-select"
                                name="currentRank-val"
                                options={valorantRankOpts}
                            />
                        </td>
                    </tr>
                    <tr>
                        <td><label htmlFor="peakRank-val">Peak Rank: </label></td>
                        <td>
                            <Select
                                className="my-react-select-container"
                                classNamePrefix="my-react-select"
                                name="peakRank-val"
                                options={valorantRankOpts}
                            />
                        </td>
                    </tr>
                    <tr>
                        <td><label htmlFor="showRank-val">Publicly display my rank average: </label></td>
                        <td>
                            <label className={styles.switch}>
                                <input type="checkbox" name="showRank-val" />
                                <span className={`${styles.slider} ${styles.round}`} />
                            </label>
                        </td>
                    </tr>
                </tbody>
            </table>
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
