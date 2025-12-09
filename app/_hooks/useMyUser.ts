import { fetcher } from "./fetcher.ts"
import useSWR from 'swr'

export default function useMyUser() {
    const { data, error, isLoading } = useSWR(`api/users/get_my_info`, fetcher)

    return {
        user: data,
        isLoading,
        isError: error
    }
}
