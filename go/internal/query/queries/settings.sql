-- name: CreateSetting :one
INSERT INTO settings (
    user_id,
    privacy_policy_agreed,
    hide_record_body,
    slots_sort_type,
    display_option_work_list,
    display_option_user_work_list,
    records_sort_type,
    display_option_record_list,
    share_record_to_twitter,
    share_record_to_facebook,
    share_review_to_twitter,
    share_review_to_facebook,
    hide_supporter_badge,
    share_status_to_twitter,
    share_status_to_facebook,
    timeline_mode,
    created_at,
    updated_at
)
VALUES (
    $1,                      -- user_id
    true,                    -- privacy_policy_agreed (新規登録時はtrue)
    true,                    -- hide_record_body (デフォルト)
    '',                      -- slots_sort_type (デフォルト)
    'list_detailed',         -- display_option_work_list (デフォルト)
    'grid_detailed',         -- display_option_user_work_list (デフォルト)
    'created_at_desc',       -- records_sort_type (デフォルト)
    'all_comments',          -- display_option_record_list (デフォルト)
    false,                   -- share_record_to_twitter
    false,                   -- share_record_to_facebook
    false,                   -- share_review_to_twitter
    false,                   -- share_review_to_facebook
    false,                   -- hide_supporter_badge
    false,                   -- share_status_to_twitter
    false,                   -- share_status_to_facebook
    'following',             -- timeline_mode (デフォルト)
    NOW(),                   -- created_at
    NOW()                    -- updated_at
)
RETURNING id, user_id, privacy_policy_agreed, created_at, updated_at;

-- name: GetSettingByUserID :one
SELECT id, user_id, privacy_policy_agreed, hide_record_body, created_at, updated_at
FROM settings
WHERE user_id = $1
LIMIT 1;
