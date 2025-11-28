-- name: CreateWorkImage :one
INSERT INTO work_images (
    work_id,
    user_id,
    image_data,
    copyright,
    asin,
    color_rgb,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING id;
