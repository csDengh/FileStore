-- name: CreateToken :execresult
INSERT INTO tbl_user_token (
  user_name,
  user_token,
  expire_at
) VALUES (
 ?,?,?
);

-- name: GetToken :one
SELECT * FROM tbl_user_token
WHERE user_name = ? LIMIT 1;