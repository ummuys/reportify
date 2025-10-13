package repository

// NEW USER
const NewUser = `INSERT INTO users.data (username, password) VALUES ($1, $2)`

// GET PASSWORD
const GetPass = `SELECT password FROM users.data  WHERE username = $1`

// CHECK USER
const CheckUset = `SELECT 1 FROM users.data WHERE username = $1`

// CACHE
const (
	qSetCacheQuery = `
  INSERT INTO users.user_queries VALUES($1, $2)
  ON CONFLICT (username) DO UPDATE SET queries = EXCLUDED.queries`
	qGetCacheQuery = `SELECT * FROM users.user_queries`
)
