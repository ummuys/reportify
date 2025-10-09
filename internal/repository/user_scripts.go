package repository

// NEW USER
const NewUser = `INSERT INTO auth.users VALUES (username, password) ($1, $2)`

// GET PASSWORD
const GetPass = `SELECT password FROM auth.users WHERE username = $1`

// CHECK USER
const CheckUset = `SELECT 1 FROM auth.users WHERE username = $1`
