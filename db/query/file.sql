-- name: CreateFile :execresult
INSERT INTO tbl_file (
    file_sha1,
    file_name,
    file_size,
    file_addr,
    create_at,
    update_at,
    status
) VALUES (
  ?,?,?,?,?,?,?
);


-- name: GetFile :one
SELECT * FROM tbl_file
WHERE file_sha1 = ? LIMIT 1;

-- name: UpdateFile :exec
update tbl_file 
SET file_name = ?
WHERE file_sha1 = ?;


-- name: DeleteFile :exec
DELETE FROM tbl_file
WHERE file_sha1 = ?;

-- name: UpdateFileLocation :exec
update tbl_file 
SET file_addr = ?
WHERE file_sha1 = ?;

