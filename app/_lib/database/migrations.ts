import { dbUpdate } from "./database";

async function migrate() {
    //discordId stored as text because it's too big for int maybe? (value got rounded up when it was an int)
    const createUser = `CREATE TABLE IF NOT EXISTS user (
        userId INTEGER PRIMARY KEY AUTOINCREMENT,
        discordId TEXT NOT NULL UNIQUE,
        discordName TEXT NOT NULL UNIQUE,
        imageUrl TEXT NOT NULL,
        riotId TEXT NULL,
        pronouns TEXT NULL,
        currentRank TEXT NULL,
        peakRank TEXT NULL,
        region TEXT NULL,
        accessToken TEXT NULL,
        accessIv TEXT NULL,
        refreshToken TEXT NULL,
        refreshIv TEXT NULL,
        isAdmin INTEGER NOT NULL DEFAULT 0
    );`;

    const createSession = `CREATE TABLE IF NOT EXISTS session (
        sessionId INTEGER PRIMARY KEY AUTOINCREMENT,
        sessionValue TEXT NOT NULL,
        sessionIv TEXT NOT NULL,
        userId INTEGER NOT NULL,
        createdAt INTEGER NOT NULL,
        device TEXT NULL
    );`;

    const createSessionIndex = `CREATE INDEX IF NOT EXISTS idx_session_userId ON session (userId);`;

    await Promise.all([
        dbUpdate(createUser, [])
            .then(data => {
                console.log("user table created successfully.");
            })
            .catch(error => {
                console.error(error)
            }),

        dbUpdate(createSession, [])
            .then(data => {
                console.log("session table created successfully.");
            })
            .catch(error => {
                console.error(error)
            }),

    ]);

    await Promise.all([
        dbUpdate(createSessionIndex, [])
            .then(data => {
                console.log("session table indexes created successfully.");
            })
            .catch(error => {
                console.error(error)
            })
    ]);

        //TODO: consider index on discordName, discordId

        /*
        db.run(
            `CREATE TABLE IF NOT EXISTS event (
                eventId INTEGER PRIMARY KEY AUTOINCREMENT,
                name TEXT NOT NULL,
                date TEXT NOT NULL,
                region TEXT NOT NULL
            );
            `,
            (err: Error) => {
                if (err) {
                    console.error(err.message);
                }
                console.log("event table created successfully.");
            }
        );

        db.run(
            `CREATE TABLE IF NOT EXISTS timeslot (
                timeslotId INTEGER PRIMARY KEY AUTOINCREMENT,
                eventId TEXT NOT NULL,
                time INTEGER NOT NULL
            );
            `,
            (err: Error) => {
                if (err) {
                    console.error(err.message);
                }
                console.log("timeslot table created successfully.");
            }
        );

        db.run(
            `CREATE TABLE IF NOT EXISTS lobby (
                lobbyId INTEGER PRIMARY KEY AUTOINCREMENT,
                timeslotId TEXT NOT NULL,
                host INTEGER NULL
                averageRank INTEGER NULL
            );
            `,
            (err: Error) => {
                if (err) {
                    console.error(err.message);
                }
                console.log("lobby table created successfully.");
            }
        );

        db.run(
            `CREATE TABLE IF NOT EXISTS registration (
                registerId INTEGER PRIMARY KEY AUTOINCREMENT,
                userId INTEGER NOT NULL,
                eventId INTEGER NOT NULL,
                timestamp INTEGER NOT NULL,
                playMultiple INTEGER NOT NULL DEFAULT 0,
                canSub INTEGER NOT NULL DEFAULT 0,
                canHost INTEGER NOT NULL DEFAULT 0,
                duoRequest TEXT NULL
            );
            `,
            (err: Error) => {
                if (err) {
                    console.error(err.message);
                }
                console.log("registration table created successfully.");
            }
        );

        db.run(
            `CREATE TABLE IF NOT EXISTS registrationToTimeslot (
                registerId INTEGER PRIMARY KEY AUTOINCREMENT,
                timeslotId INTEGER NOT NULL
            );
            `,
            (err: Error) => {
                if (err) {
                    console.error(err.message);
                }
                console.log("registrationToTimeslot table created successfully.");
            }
        );
 
        db.run(
            `CREATE TABLE IF NOT EXISTS team (
                teamId INTEGER PRIMARY KEY AUTOINCREMENT,
                lobbyId INTEGER NOT NULL,
                averageRank INTEGER NULL,
                meetsDuoRequest INTEGER NOT NULL DEFAULT 0,
                wonGame INTEGER NOT NULL DEFAULT 0
            );
            `,
            (err: Error) => {
                if (err) {
                    console.error(err.message);
                }
                console.log("team table created successfully.");
            }
        );
 
        db.run(
            `CREATE TABLE IF NOT EXISTS teamPlayers (
                playerId INTEGER PRIMARY KEY AUTOINCREMENT,
                teamId INTEGER NOT NULL,
                userId INTEGER NOT NULL,
                isSub INTEGER NOT NULL DEFAULT 0,
                noShow INTEGER NOT NULL DEFAULT 0
            );
            `,
            (err: Error) => {
                if (err) {
                    console.error(err.message);
                }
                console.log("team table created successfully.");
            }
        );
       */

}

migrate()
