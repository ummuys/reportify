package repository

// NEW USER
const NewUser = `INSERT INTO identity.users (username, password) VALUES ($1, $2)`

// GET PASSWORD
const GetPass = `SELECT user_id, password FROM identity.users  WHERE username = $1`

// CHECK USER
const CheckUset = `SELECT 1 FROM identity.users WHERE username = $1`

const (
	qSetCacheQuery = `
  INSERT INTO identity.queries VALUES($1, $2)
  ON CONFLICT (user_id) DO UPDATE SET list = EXCLUDED.list`
	qGetCacheQuery = `SELECT * FROM identity.queries`
)
