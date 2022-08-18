-- name: CreateUser :execresult
INSERT INTO tbl_user (
  user_name,
  user_pwd,
  email,
  phone,
  email_validated,
  phone_validated,
  signup_at,
  last_active,
  profile,
  status
) VALUES (
  ?,?,?,?,?,?,?,?,?,?
);

-- name: GetUser :one
SELECT * FROM tbl_user
WHERE user_name = ? LIMIT 1;

-- name: UpdateUser :exec
update tbl_user 
SET email = ?
WHERE user_name = ?;


-- name: DeleteUser :exec
DELETE FROM tbl_user
WHERE user_name = ?;