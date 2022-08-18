
-- name: GetUserFileMeteList :many
SELECT * FROM tbl_user_file
WHERE user_name = ? LIMIT ?;

-- name: CreateUserFile :execresult
INSERT INTO tbl_user_file (
  user_name,
  file_sha1,
  file_size,
  file_name,
  upload_at,
  last_update,
  status
) VALUES (
  ?,?,?,?,?,?,?
);
