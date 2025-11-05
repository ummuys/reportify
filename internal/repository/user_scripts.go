package repository

// NEW USER
const NewUserStep1 = `
INSERT INTO identity.users(username, password) VALUES
($1,$2);
`

const NewUserStep2 = `
INSERT INTO identity.user_roles (user_id, role_id)
VALUES (
    (SELECT user_id FROM identity.users WHERE username = $1),
    (SELECT role_id FROM identity.roles WHERE name = $2)
);`

// CheckCredentials
const GetPass = `
SELECT 
    u.user_id,
    u.password,
    r.name AS role
FROM identity.users AS u
JOIN identity.user_roles AS ur ON ur.user_id = u.user_id
JOIN identity.roles AS r ON r.role_id = ur.role_id
WHERE u.username = $1;
`

// CHECK USER
const CheckUset = `SELECT 1 FROM identity.users WHERE username = $1`
