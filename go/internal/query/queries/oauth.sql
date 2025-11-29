-- name: CreateOAuthApplication :one
INSERT INTO oauth_applications (
	name, uid, secret, redirect_uri, scopes,
	aasm_state, created_at, updated_at,
	owner_id, owner_type, confidential,
	hide_social_login
) VALUES (
	$1, $2, $3, $4, $5,
	$6, $7, $8,
	$9, $10, $11,
	$12
) RETURNING id;

-- name: CreateOAuthAccessToken :one
INSERT INTO oauth_access_tokens (
	resource_owner_id, application_id, token,
	refresh_token, expires_in, revoked_at,
	created_at, scopes, previous_refresh_token,
	description
) VALUES (
	$1, $2, $3,
	$4, $5, $6,
	$7, $8, $9,
	$10
) RETURNING id;

-- name: GetOAuthApplicationByUID :one
SELECT id, name, uid, secret, redirect_uri, scopes,
	aasm_state, created_at, updated_at,
	owner_id, owner_type, confidential,
	hide_social_login, deleted_at
FROM oauth_applications
WHERE uid = $1
LIMIT 1;
