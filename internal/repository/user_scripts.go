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

// UPDATE USER
const UpdateUsername = `
UPDATE identity.users SET username = $2 WHERE user_id = $1;
`

const UpdateUserPassword = `
UPDATE identity.users SET password = $2 WHERE user_id = $1;
`

// #nosec G101 -- SQL query, not hardcoded password
const UpdateUserRole = `
UPDATE 
    identity.user_roles 
SET 
    role_id = (SELECT role_id FROM identity.roles WHERE name = $2)
WHERE 
    user_id = $1
`

// #nosec G101 -- SQL query, not hardcoded password
const GetCredentials = `
SELECT 
    u.user_id,
    u.password,
    r.name AS role
FROM identity.users AS u
JOIN identity.user_roles AS ur ON ur.user_id = u.user_id
JOIN identity.roles AS r ON r.role_id = ur.role_id
WHERE u.username = $1;
`

// DELETE USER
const DeleteUser = `DELETE FROM identity.users WHERE username = $1`

// CHECK ROLE
const CheckRole = `SELECT 1 FROM identity.roles WHERE name = $1`

// GET USERS
const GetUsers = `
SELECT 
    u.user_id,
    u.username,
    r.name AS role
FROM identity.users AS u
JOIN identity.user_roles AS ur ON ur.user_id = u.user_id
JOIN identity.roles AS r ON r.role_id = ur.role_id
`
