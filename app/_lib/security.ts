export async function generateState() {
    const state = crypto.randomUUID()
    const encrypted = await encryptData(state)

    return {
        'state': state,
        'encryptedState': encrypted.encryptedData,
        'initVector': encrypted.initVector,
        'cookie': encrypted.encryptedData + ":" + encrypted.initVector
    }
}

export async function validateState(state: string, storedState: string) {
    //split the state to get the encrypted version of state & the initialization vector
    const stateArr = storedState.split(":");
    const encryptedState = stateArr[0];
    const initVector = stateArr[1];

    //decrypt the state & make sure it matches what we got back from Discord
    const decryptedState = await decryptData(encryptedState, initVector);

    return decryptedState == state;
}

export async function encryptData(plainData: string) {
    //random 96-bit initialization vector
    const initVector = crypto.getRandomValues(new Uint8Array(12));

    //encode data
    const encodedData = new TextEncoder().encode(plainData);

    //prepare the key
    const cryptoKey = await crypto.subtle.importKey(
        "raw",
        Buffer.from(process.env.AES_ENCRYPTION_KEY, "base64"),
        {
            name: "AES-GCM",
            length: 256,
        },
        true,
        ["encrypt", "decrypt"]
    );

    //encrypt the data with the key
    const encryptedData = await crypto.subtle.encrypt(
        {
            name: "AES-GCM",
            iv: initVector
        },
        cryptoKey,
        encodedData
    );

    //Return the encrypted data and the IV
    return {
        encryptedData: Buffer.from(encryptedData).toString("base64"),
        initVector: Buffer.from(initVector).toString("base64")
    }
}

export async function decryptData(encryptedData: string, initVector: string) {
    //prep decryption key
    const cryptoKey = await crypto.subtle.importKey(
        "raw",
        Buffer.from(process.env.AES_ENCRYPTION_KEY, "base64"),
        {
            name: "AES-GCM",
            length: 256
        },
        true,
        ["encrypt","decrypt"]
    );

    try {
        const decodedData = await crypto.subtle.decrypt(
            {
                name: "AES-GCM",
                iv: Buffer.from(initVector, "base64")
            },
            cryptoKey,
            Buffer.from(encryptedData, "base64")
        );

        return new TextDecoder().decode(decodedData);
    } catch(error) {
        return JSON.stringify({ payload: null });
    }
}
