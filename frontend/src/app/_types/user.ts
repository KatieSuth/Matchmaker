export interface User {
    discord_id: string;
    discord_name: string | null;
    image_url: string | null;
    pronouns: string | null
    show_pronouns: boolean
    region: string | null
    new_user: boolean
}