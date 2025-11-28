-- name: CreateEmailNotification :one
INSERT INTO email_notifications (
    user_id,
    unsubscription_key,
    event_followed_user,
    event_liked_episode_record,
    event_friends_joined,
    event_next_season_came,
    event_favorite_works_added,
    event_related_works_added,
    created_at,
    updated_at
)
VALUES (
    $1,    -- user_id
    $2,    -- unsubscription_key (UUID)
    true,  -- event_followed_user (デフォルト)
    true,  -- event_liked_episode_record (デフォルト)
    true,  -- event_friends_joined (デフォルト)
    true,  -- event_next_season_came (デフォルト)
    true,  -- event_favorite_works_added (デフォルト)
    true,  -- event_related_works_added (デフォルト)
    NOW(), -- created_at
    NOW()  -- updated_at
)
RETURNING id, user_id, unsubscription_key, created_at, updated_at;

-- name: GetEmailNotificationByUserID :one
SELECT id, user_id, unsubscription_key, event_followed_user, event_liked_episode_record, created_at, updated_at
FROM email_notifications
WHERE user_id = $1
LIMIT 1;
