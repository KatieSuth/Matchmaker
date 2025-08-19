console.log('starting db')

const sqlite3 = require('sqlite3').verbose();
const db = new sqlite3.Database('./app.db');
console.log('db', db)

console.log('preparing fetchuser')

const fetchUser = db.prepare(
    'SELECT * FROM user;'
)

console.log('fetching user')

const user = fetchUser.all([], (err, rows) => {
    console.log('user', rows)
});
