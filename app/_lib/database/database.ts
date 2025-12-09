export const runtime = "nodejs"

import sqlite3 from "sqlite3"
const dbPath = process.env.DB_PATH

export const db = new sqlite3.Database(
    dbPath,
    sqlite3.OPEN_READWRITE,
    //sqlite3.OPEN_READWRITE | sqlite3.OPEN_CREATE, //TODO: consider
    (err) => {
        if (err) {
            console.error(err.message);
        } else {
            console.log("Connected to the app database.");
        }
    }
);

export const dbSelect = async (query: string, values: string[]) => {
    query = db.prepare(query)

    return await new Promise((resolve, reject) => {
        query.all(values, (err: Error, rows) => {
            if (err) {
                console.log(err);
                return reject(err);
            }
            return resolve(rows);
        });
    });
};

export const dbUpdate = async (query: string, values: string[]) => {
    query = db.prepare(query)

    return await new Promise((resolve, reject) => {
        query.run(values, (err) => {
            if (err) {
                console.log(err);
                reject(err);
            }
            resolve(null);
        });
    });
};
