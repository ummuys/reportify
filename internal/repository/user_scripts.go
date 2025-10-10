package repository

// NEW USER
const NewUser = `INSERT INTO users.data (username, password) VALUES ($1, $2)`

// GET PASSWORD
const GetPass = `SELECT password FROM users.data  WHERE username = $1`

// CHECK USER
const CheckUset = `SELECT 1 FROM users.data WHERE username = $1`
