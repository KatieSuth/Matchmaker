export interface RegionOption {
    readonly value: string;
    readonly label: string;
    readonly isDisabled?: boolean;
}

export const regionOpts: readonly RegionOption[] = [
    { value: '', label: 'Choose Region', isDisabled: true },
    { value: 'EU', label: 'EU' },
    { value: 'NA', label: 'NA' }
];

export interface PronounOption {
    readonly value: string;
    readonly label: string;
    readonly isDisabled?: boolean;
}

export const pronounOpts: readonly RegionOption[] = [
    { value: '', label: 'Choose Preferred Pronouns', isDisabled: true },
    { value: 'She/Her', label: 'She/Her' },
    { value: 'He/Him', label: 'He/Him' },
    { value: 'She/They', label: 'She/They' },
    { value: 'He/They', label: 'He/They' },
    { value: 'They/Them', label: 'They/Them' },
    { value: 'Any', label: 'Any' },
    { value: 'Other', label: 'Other' }
];

export interface RankOption {
    readonly value: integer;
    readonly label: string;
    readonly isDisabled?: boolean;
    readonly rankValue: integer;
}

export const rankOpts: readonly RankOption[] = [
    { value: 0, label: 'Choose Rank', rankValue: 0, isDisabled: true },
    { value: 1, label: 'Iron 1', rankValue: 1 },
    { value: 2, label: 'Iron 2', rankValue: 3 },
    { value: 3, label: 'Iron 3', rankValue: 6 },
    { value: 4, label: 'Bronze 1', rankValue: 10 },
    { value: 5, label: 'Bronze 2', rankValue: 12 },
    { value: 6, label: 'Bronze 3', rankValue: 15 },
    { value: 7, label: 'Silver 1', rankValue: 20 },
    { value: 8, label: 'Silver 2', rankValue: 22 },
    { value: 9, label: 'Silver 3', rankValue: 25 },
    { value: 10, label: 'Gold 1', rankValue: 30 },
    { value: 11, label: 'Gold 2', rankValue: 32 },
    { value: 12, label: 'Gold 3', rankValue: 35 },
    { value: 13, label: 'Platinum 1', rankValue: 40 },
    { value: 14, label: 'Platinum 2', rankValue: 42 },
    { value: 15, label: 'Platinum 3', rankValue: 45 },
    { value: 16, label: 'Diamond 1', rankValue: 50 },
    { value: 17, label: 'Diamond 2', rankValue: 52 },
    { value: 18, label: 'Diamond 3', rankValue: 55 },
    { value: 19, label: 'Ascendant 1', rankValue: 60 },
    { value: 20, label: 'Ascendant 2', rankValue: 65 },
    { value: 21, label: 'Ascendant 3', rankValue: 70 },
    { value: 22, label: 'Immortal 1', rankValue: 80 },
    { value: 23, label: 'Immortal 2', rankValue: 85 },
    { value: 24, label: 'Immortal 3', rankValue: 95 },
    { value: 25, label: 'Radiant', rankValue: 110 }
];
