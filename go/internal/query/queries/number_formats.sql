-- name: ListNumberFormats :many
SELECT id, name, sort_number
FROM number_formats
ORDER BY sort_number;
