-- name: ListNumberFormats :many
SELECT id, name, sort_number
FROM number_formats
ORDER BY sort_number;

-- name: ExistsNumberFormatByID :one
SELECT EXISTS (
    SELECT 1
    FROM number_formats
    WHERE id = sqlc.arg('id')
);
