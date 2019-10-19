class InitSchema < ActiveRecord::Migration[5.1]
  def up
    # These are extensions that must be enabled in order to support this database
    enable_extension "plpgsql"
    create_table "activities", id: :serial, force: :cascade do |t|
      t.integer "user_id", null: false
      t.integer "recipient_id", null: false
      t.string "recipient_type", null: false
      t.integer "trackable_id", null: false
      t.string "trackable_type", null: false
      t.string "action", null: false
      t.datetime "created_at"
      t.datetime "updated_at"
      t.index ["recipient_id", "recipient_type"], name: "index_activities_on_recipient_id_and_recipient_type"
      t.index ["trackable_id", "trackable_type"], name: "index_activities_on_trackable_id_and_trackable_type"
    end
    create_table "casts", id: :serial, force: :cascade do |t|
      t.integer "person_id", null: false
      t.integer "work_id", null: false
      t.string "name", null: false
      t.string "part", null: false
      t.string "aasm_state", default: "published", null: false
      t.integer "sort_number", default: 0, null: false
      t.datetime "created_at", null: false
      t.datetime "updated_at", null: false
      t.integer "character_id"
      t.string "name_en", default: "", null: false
      t.index ["aasm_state"], name: "index_casts_on_aasm_state"
      t.index ["character_id"], name: "index_casts_on_character_id"
      t.index ["person_id"], name: "index_casts_on_person_id"
      t.index ["sort_number"], name: "index_casts_on_sort_number"
      t.index ["work_id"], name: "index_casts_on_work_id"
    end
    create_table "channel_groups", id: :serial, force: :cascade do |t|
      t.string "sc_chgid", null: false
      t.string "name", null: false
      t.integer "sort_number"
      t.datetime "created_at"
      t.datetime "updated_at"
      t.index ["sc_chgid"], name: "index_channel_groups_on_sc_chgid", unique: true
    end
    create_table "channel_works", id: :serial, force: :cascade do |t|
      t.integer "user_id", null: false
      t.integer "work_id", null: false
      t.integer "channel_id", null: false
      t.datetime "created_at"
      t.datetime "updated_at"
      t.index ["user_id", "work_id", "channel_id"], name: "index_channel_works_on_user_id_and_work_id_and_channel_id", unique: true
      t.index ["user_id", "work_id"], name: "index_channel_works_on_user_id_and_work_id"
    end
    create_table "channels", id: :serial, force: :cascade do |t|
      t.integer "channel_group_id", null: false
      t.integer "sc_chid", null: false
      t.string "name", null: false, collation: "C"
      t.datetime "created_at"
      t.datetime "updated_at"
      t.boolean "published", default: true, null: false
      t.index ["published"], name: "index_channels_on_published"
      t.index ["sc_chid"], name: "index_channels_on_sc_chid", unique: true
    end
    create_table "character_images", id: :serial, force: :cascade do |t|
      t.integer "character_id", null: false
      t.integer "user_id", null: false
      t.string "attachment_file_name", null: false
      t.integer "attachment_file_size", null: false
      t.string "attachment_content_type", null: false
      t.datetime "attachment_updated_at", null: false
      t.string "copyright", default: "", null: false
      t.string "asin", default: "", null: false
      t.datetime "created_at", null: false
      t.datetime "updated_at", null: false
      t.index ["character_id"], name: "index_character_images_on_character_id"
      t.index ["user_id"], name: "index_character_images_on_user_id"
    end
    create_table "characters", id: :serial, force: :cascade do |t|
      t.string "name", null: false
      t.string "name_kana", default: "", null: false
      t.string "name_en", default: "", null: false
      t.string "kind", default: "", null: false
      t.string "kind_en", default: "", null: false
      t.string "nickname", default: "", null: false
      t.string "nickname_en", default: "", null: false
      t.string "birthday", default: "", null: false
      t.string "birthday_en", default: "", null: false
      t.string "age", default: "", null: false
      t.string "age_en", default: "", null: false
      t.string "blood_type", default: "", null: false
      t.string "blood_type_en", default: "", null: false
      t.string "height", default: "", null: false
      t.string "height_en", default: "", null: false
      t.string "weight", default: "", null: false
      t.string "weight_en", default: "", null: false
      t.string "nationality", default: "", null: false
      t.string "nationality_en", default: "", null: false
      t.string "occupation", default: "", null: false
      t.string "occupation_en", default: "", null: false
      t.text "description", default: "", null: false
      t.text "description_en", default: "", null: false
      t.string "aasm_state", default: "published", null: false
      t.datetime "created_at", null: false
      t.datetime "updated_at", null: false
      t.index ["name", "kind"], name: "index_characters_on_name_and_kind", unique: true
    end
    create_table "checkins", id: :serial, force: :cascade do |t|
      t.integer "user_id", null: false
      t.integer "episode_id", null: false
      t.text "comment"
      t.string "twitter_url_hash"
      t.integer "twitter_click_count", default: 0, null: false
      t.datetime "created_at"
      t.datetime "updated_at"
      t.string "facebook_url_hash"
      t.integer "facebook_click_count", default: 0, null: false
      t.integer "comments_count", default: 0, null: false
      t.integer "likes_count", default: 0, null: false
      t.boolean "modify_comment", default: false, null: false
      t.boolean "shared_twitter", default: false, null: false
      t.boolean "shared_facebook", default: false, null: false
      t.integer "work_id", null: false
      t.float "rating"
      t.integer "multiple_record_id"
      t.integer "oauth_application_id"
      t.index ["facebook_url_hash"], name: "index_checkins_on_facebook_url_hash", unique: true
      t.index ["multiple_record_id"], name: "index_checkins_on_multiple_record_id"
      t.index ["oauth_application_id"], name: "index_checkins_on_oauth_application_id"
      t.index ["twitter_url_hash"], name: "index_checkins_on_twitter_url_hash", unique: true
      t.index ["work_id"], name: "index_checkins_on_work_id"
    end
    create_table "comments", id: :serial, force: :cascade do |t|
      t.integer "user_id", null: false
      t.integer "checkin_id", null: false
      t.text "body", null: false
      t.datetime "created_at"
      t.datetime "updated_at"
      t.integer "likes_count", default: 0, null: false
      t.integer "work_id"
      t.index ["work_id"], name: "index_comments_on_work_id"
    end
    create_table "cover_images", id: :serial, force: :cascade do |t|
      t.integer "work_id", null: false
      t.string "file_name", null: false
      t.string "location", null: false
      t.datetime "created_at"
      t.datetime "updated_at"
    end
    create_table "db_activities", id: :serial, force: :cascade do |t|
      t.integer "user_id", null: false
      t.integer "trackable_id", null: false
      t.string "trackable_type", null: false
      t.string "action", null: false
      t.json "parameters"
      t.datetime "created_at", null: false
      t.datetime "updated_at", null: false
      t.integer "root_resource_id"
      t.string "root_resource_type"
      t.integer "object_id"
      t.string "object_type"
      t.index ["object_id", "object_type"], name: "index_db_activities_on_object_id_and_object_type"
      t.index ["root_resource_id", "root_resource_type"], name: "index_db_activities_on_root_resource_id_and_root_resource_type"
      t.index ["trackable_id", "trackable_type"], name: "index_db_activities_on_trackable_id_and_trackable_type"
    end
    create_table "db_comments", id: :serial, force: :cascade do |t|
      t.integer "user_id", null: false
      t.integer "resource_id", null: false
      t.string "resource_type", null: false
      t.text "body", null: false
      t.datetime "created_at", null: false
      t.datetime "updated_at", null: false
      t.index ["resource_id", "resource_type"], name: "index_db_comments_on_resource_id_and_resource_type"
      t.index ["user_id"], name: "index_db_comments_on_user_id"
    end
    create_table "delayed_jobs", id: :serial, force: :cascade do |t|
      t.integer "priority", default: 0, null: false
      t.integer "attempts", default: 0, null: false
      t.text "handler", null: false
      t.text "last_error"
      t.datetime "run_at"
      t.datetime "locked_at"
      t.datetime "failed_at"
      t.string "locked_by"
      t.string "queue"
      t.datetime "created_at"
      t.datetime "updated_at"
      t.index ["priority", "run_at"], name: "delayed_jobs_priority"
    end
    create_table "draft_casts", id: :serial, force: :cascade do |t|
      t.integer "cast_id"
      t.integer "person_id", null: false
      t.integer "work_id", null: false
      t.string "name", null: false
      t.string "part", null: false
      t.integer "sort_number", default: 0, null: false
      t.datetime "created_at", null: false
      t.datetime "updated_at", null: false
      t.index ["cast_id"], name: "index_draft_casts_on_cast_id"
      t.index ["person_id"], name: "index_draft_casts_on_person_id"
      t.index ["sort_number"], name: "index_draft_casts_on_sort_number"
      t.index ["work_id"], name: "index_draft_casts_on_work_id"
    end
    create_table "draft_episodes", id: :serial, force: :cascade do |t|
      t.integer "episode_id", null: false
      t.integer "work_id", null: false
      t.string "number"
      t.integer "sort_number", default: 0, null: false
      t.string "title"
      t.datetime "created_at", null: false
      t.datetime "updated_at", null: false
      t.integer "prev_episode_id"
      t.boolean "fetch_syobocal", default: false, null: false
      t.string "raw_number"
      t.integer "sc_count"
      t.index ["episode_id"], name: "index_draft_episodes_on_episode_id"
      t.index ["prev_episode_id"], name: "index_draft_episodes_on_prev_episode_id"
      t.index ["work_id"], name: "index_draft_episodes_on_work_id"
    end
    create_table "draft_items", id: :serial, force: :cascade do |t|
      t.integer "item_id"
      t.integer "work_id", null: false
      t.string "name", null: false
      t.string "url", null: false
      t.boolean "main", default: false, null: false
      t.string "tombo_image_file_name", null: false
      t.string "tombo_image_content_type", null: false
      t.integer "tombo_image_file_size", null: false
      t.datetime "tombo_image_updated_at", null: false
      t.datetime "created_at", null: false
      t.datetime "updated_at", null: false
      t.index ["item_id"], name: "index_draft_items_on_item_id"
      t.index ["work_id"], name: "index_draft_items_on_work_id"
    end
    create_table "draft_multiple_episodes", id: :serial, force: :cascade do |t|
      t.integer "work_id", null: false
      t.text "body", null: false
      t.datetime "created_at", null: false
      t.datetime "updated_at", null: false
      t.index ["work_id"], name: "index_draft_multiple_episodes_on_work_id"
    end
    create_table "draft_organizations", id: :serial, force: :cascade do |t|
      t.integer "organization_id"
      t.string "name", null: false
      t.string "url"
      t.string "wikipedia_url"
      t.string "twitter_username"
      t.datetime "created_at", null: false
      t.datetime "updated_at", null: false
      t.string "name_kana", default: "", null: false
      t.index ["name"], name: "index_draft_organizations_on_name"
      t.index ["organization_id"], name: "index_draft_organizations_on_organization_id"
    end
    create_table "draft_people", id: :serial, force: :cascade do |t|
      t.integer "person_id"
      t.integer "prefecture_id"
      t.string "name", null: false
      t.string "name_kana"
      t.string "nickname"
      t.string "gender"
      t.string "url"
      t.string "wikipedia_url"
      t.string "twitter_username"
      t.date "birthday"
      t.string "blood_type"
      t.integer "height"
      t.datetime "created_at", null: false
      t.datetime "updated_at", null: false
      t.index ["name"], name: "index_draft_people_on_name"
      t.index ["person_id"], name: "index_draft_people_on_person_id"
      t.index ["prefecture_id"], name: "index_draft_people_on_prefecture_id"
    end
    create_table "draft_programs", id: :serial, force: :cascade do |t|
      t.integer "program_id"
      t.integer "channel_id", null: false
      t.integer "episode_id", null: false
      t.integer "work_id", null: false
      t.datetime "started_at", null: false
      t.datetime "created_at", null: false
      t.datetime "updated_at", null: false
      t.boolean "rebroadcast", default: false, null: false
      t.index ["channel_id"], name: "index_draft_programs_on_channel_id"
      t.index ["episode_id"], name: "index_draft_programs_on_episode_id"
      t.index ["program_id"], name: "index_draft_programs_on_program_id"
      t.index ["work_id"], name: "index_draft_programs_on_work_id"
    end
    create_table "draft_staffs", id: :serial, force: :cascade do |t|
      t.integer "staff_id"
      t.integer "work_id", null: false
      t.string "name", null: false
      t.string "role", null: false
      t.string "role_other"
      t.integer "sort_number", default: 0, null: false
      t.datetime "created_at", null: false
      t.datetime "updated_at", null: false
      t.integer "resource_id"
      t.string "resource_type"
      t.index ["resource_id", "resource_type"], name: "index_draft_staffs_on_resource_id_and_resource_type"
      t.index ["sort_number"], name: "index_draft_staffs_on_sort_number"
      t.index ["staff_id"], name: "index_draft_staffs_on_staff_id"
      t.index ["work_id"], name: "index_draft_staffs_on_work_id"
    end
    create_table "draft_works", id: :serial, force: :cascade do |t|
      t.integer "work_id"
      t.integer "season_id"
      t.integer "sc_tid"
      t.string "title", null: false
      t.integer "media", null: false
      t.string "official_site_url", default: "", null: false
      t.string "wikipedia_url", default: "", null: false
      t.date "released_at"
      t.string "twitter_username"
      t.string "twitter_hashtag"
      t.string "released_at_about"
      t.datetime "created_at", null: false
      t.datetime "updated_at", null: false
      t.integer "number_format_id"
      t.string "title_kana", default: "", null: false
      t.index ["number_format_id"], name: "index_draft_works_on_number_format_id"
      t.index ["season_id"], name: "index_draft_works_on_season_id"
      t.index ["work_id"], name: "index_draft_works_on_work_id"
    end
    create_table "edit_request_comments", id: :serial, force: :cascade do |t|
      t.integer "edit_request_id", null: false
      t.integer "user_id", null: false
      t.text "body", null: false
      t.datetime "created_at", null: false
      t.datetime "updated_at", null: false
      t.index ["edit_request_id"], name: "index_edit_request_comments_on_edit_request_id"
      t.index ["user_id"], name: "index_edit_request_comments_on_user_id"
    end
    create_table "edit_request_participants", id: :serial, force: :cascade do |t|
      t.integer "edit_request_id", null: false
      t.integer "user_id", null: false
      t.datetime "created_at", null: false
      t.datetime "updated_at", null: false
      t.index ["edit_request_id", "user_id"], name: "index_edit_request_participants_on_edit_request_id_and_user_id", unique: true
      t.index ["edit_request_id"], name: "index_edit_request_participants_on_edit_request_id"
      t.index ["user_id"], name: "index_edit_request_participants_on_user_id"
    end
    create_table "edit_requests", id: :serial, force: :cascade do |t|
      t.integer "user_id", null: false
      t.string "draft_resource_type", null: false
      t.integer "draft_resource_id", null: false
      t.string "title", null: false
      t.text "body"
      t.string "aasm_state", default: "opened", null: false
      t.datetime "published_at"
      t.datetime "closed_at"
      t.datetime "created_at", null: false
      t.datetime "updated_at", null: false
      t.index ["draft_resource_id", "draft_resource_type"], name: "index_er_on_drid_and_drtype"
      t.index ["user_id"], name: "index_edit_requests_on_user_id"
    end
    create_table "episodes", id: :serial, force: :cascade do |t|
      t.integer "work_id", null: false
      t.string "number"
      t.integer "sort_number", default: 0, null: false
      t.string "title"
      t.datetime "created_at"
      t.datetime "updated_at"
      t.integer "checkins_count", default: 0, null: false
      t.integer "sc_count"
      t.integer "prev_episode_id"
      t.string "aasm_state", default: "published", null: false
      t.boolean "fetch_syobocal", default: false, null: false
      t.string "raw_number"
      t.float "avg_rating"
      t.string "title_ro", default: "", null: false
      t.string "title_en", default: "", null: false
      t.index ["aasm_state"], name: "index_episodes_on_aasm_state"
      t.index ["checkins_count"], name: "index_episodes_on_checkins_count"
      t.index ["prev_episode_id"], name: "index_episodes_on_prev_episode_id"
      t.index ["work_id", "sc_count"], name: "index_episodes_on_work_id_and_sc_count", unique: true
    end
    create_table "finished_tips", id: :serial, force: :cascade do |t|
      t.integer "user_id", null: false
      t.integer "tip_id", null: false
      t.datetime "created_at", null: false
      t.datetime "updated_at", null: false
      t.index ["user_id", "tip_id"], name: "index_finished_tips_on_user_id_and_tip_id", unique: true
    end
    create_table "follows", id: :serial, force: :cascade do |t|
      t.integer "user_id", null: false
      t.integer "following_id", null: false
      t.datetime "created_at"
      t.datetime "updated_at"
      t.index ["user_id", "following_id"], name: "index_follows_on_user_id_and_following_id", unique: true
    end
    create_table "items", id: :serial, force: :cascade do |t|
      t.integer "work_id"
      t.string "name", null: false
      t.string "url", null: false
      t.datetime "created_at"
      t.datetime "updated_at"
      t.string "tombo_image_file_name"
      t.string "tombo_image_content_type"
      t.integer "tombo_image_file_size"
      t.datetime "tombo_image_updated_at"
      t.index ["work_id"], name: "index_items_on_work_id", unique: true
    end
    create_table "latest_statuses", id: :serial, force: :cascade do |t|
      t.integer "user_id", null: false
      t.integer "work_id", null: false
      t.integer "next_episode_id"
      t.integer "kind", null: false
      t.integer "watched_episode_ids", default: [], null: false, array: true
      t.integer "position", default: 0, null: false
      t.datetime "created_at", null: false
      t.datetime "updated_at", null: false
      t.index ["next_episode_id"], name: "index_latest_statuses_on_next_episode_id"
      t.index ["user_id", "position"], name: "index_latest_statuses_on_user_id_and_position"
      t.index ["user_id", "work_id"], name: "index_latest_statuses_on_user_id_and_work_id", unique: true
      t.index ["user_id"], name: "index_latest_statuses_on_user_id"
      t.index ["work_id"], name: "index_latest_statuses_on_work_id"
    end
    create_table "likes", id: :serial, force: :cascade do |t|
      t.integer "user_id", null: false
      t.integer "recipient_id", null: false
      t.string "recipient_type", null: false
      t.datetime "created_at"
      t.datetime "updated_at"
      t.index ["recipient_id", "recipient_type"], name: "index_likes_on_recipient_id_and_recipient_type"
    end
    create_table "multiple_records", id: :serial, force: :cascade do |t|
      t.integer "user_id", null: false
      t.integer "work_id", null: false
      t.integer "likes_count", default: 0, null: false
      t.datetime "created_at", null: false
      t.datetime "updated_at", null: false
      t.index ["user_id"], name: "index_multiple_records_on_user_id"
      t.index ["work_id"], name: "index_multiple_records_on_work_id"
    end
    create_table "mute_users", id: :serial, force: :cascade do |t|
      t.integer "user_id", null: false
      t.integer "muted_user_id", null: false
      t.datetime "created_at", null: false
      t.datetime "updated_at", null: false
      t.index ["muted_user_id"], name: "index_mute_users_on_muted_user_id"
      t.index ["user_id", "muted_user_id"], name: "index_mute_users_on_user_id_and_muted_user_id", unique: true
      t.index ["user_id"], name: "index_mute_users_on_user_id"
    end
    create_table "notifications", id: :serial, force: :cascade do |t|
      t.integer "user_id", null: false
      t.integer "action_user_id", null: false
      t.integer "trackable_id", null: false
      t.string "trackable_type", null: false
      t.string "action", null: false
      t.boolean "read", default: false, null: false
      t.datetime "created_at"
      t.datetime "updated_at"
      t.index ["read"], name: "index_notifications_on_read"
      t.index ["trackable_id", "trackable_type"], name: "index_notifications_on_trackable_id_and_trackable_type"
    end
    create_table "number_formats", id: :serial, force: :cascade do |t|
      t.string "name", null: false
      t.string "data", default: [], null: false, array: true
      t.integer "sort_number", default: 0, null: false
      t.datetime "created_at", null: false
      t.datetime "updated_at", null: false
      t.string "format", default: "", null: false
      t.index ["name"], name: "index_number_formats_on_name", unique: true
    end
    create_table "oauth_access_grants", id: :serial, force: :cascade do |t|
      t.integer "resource_owner_id", null: false
      t.integer "application_id", null: false
      t.string "token", null: false
      t.integer "expires_in", null: false
      t.text "redirect_uri", null: false
      t.datetime "created_at", null: false
      t.datetime "revoked_at"
      t.string "scopes"
      t.index ["token"], name: "index_oauth_access_grants_on_token", unique: true
    end
    create_table "oauth_access_tokens", id: :serial, force: :cascade do |t|
      t.integer "resource_owner_id", null: false
      t.integer "application_id", null: false
      t.string "token", null: false
      t.string "refresh_token"
      t.integer "expires_in"
      t.datetime "revoked_at"
      t.datetime "created_at", null: false
      t.string "scopes"
      t.string "previous_refresh_token", default: "", null: false
      t.index ["refresh_token"], name: "index_oauth_access_tokens_on_refresh_token", unique: true
      t.index ["resource_owner_id"], name: "index_oauth_access_tokens_on_resource_owner_id"
      t.index ["token"], name: "index_oauth_access_tokens_on_token", unique: true
    end
    create_table "oauth_applications", id: :serial, force: :cascade do |t|
      t.string "name", null: false
      t.string "uid", null: false
      t.string "secret", null: false
      t.text "redirect_uri", null: false
      t.string "scopes", default: "", null: false
      t.string "aasm_state", default: "published", null: false
      t.datetime "created_at"
      t.datetime "updated_at"
      t.integer "owner_id"
      t.string "owner_type"
      t.index ["owner_id", "owner_type"], name: "index_oauth_applications_on_owner_id_and_owner_type"
      t.index ["uid"], name: "index_oauth_applications_on_uid", unique: true
    end
    create_table "organizations", id: :serial, force: :cascade do |t|
      t.string "name", null: false
      t.string "url"
      t.string "wikipedia_url"
      t.string "twitter_username"
      t.string "aasm_state", default: "published", null: false
      t.datetime "created_at", null: false
      t.datetime "updated_at", null: false
      t.string "name_kana", default: "", null: false
      t.string "name_en", default: "", null: false
      t.string "url_en", default: "", null: false
      t.string "wikipedia_url_en", default: "", null: false
      t.string "twitter_username_en", default: "", null: false
      t.index ["aasm_state"], name: "index_organizations_on_aasm_state"
      t.index ["name"], name: "index_organizations_on_name", unique: true
    end
    create_table "people", id: :serial, force: :cascade do |t|
      t.integer "prefecture_id"
      t.string "name", null: false
      t.string "name_kana", default: "", null: false
      t.string "nickname"
      t.string "gender"
      t.string "url"
      t.string "wikipedia_url"
      t.string "twitter_username"
      t.date "birthday"
      t.string "blood_type"
      t.integer "height"
      t.string "aasm_state", default: "published", null: false
      t.datetime "created_at", null: false
      t.datetime "updated_at", null: false
      t.string "name_en", default: "", null: false
      t.string "nickname_en", default: "", null: false
      t.string "url_en", default: "", null: false
      t.string "wikipedia_url_en", default: "", null: false
      t.string "twitter_username_en", default: "", null: false
      t.index ["aasm_state"], name: "index_people_on_aasm_state"
      t.index ["name"], name: "index_people_on_name", unique: true
      t.index ["prefecture_id"], name: "index_people_on_prefecture_id"
    end
    create_table "prefectures", id: :serial, force: :cascade do |t|
      t.string "name", null: false
      t.datetime "created_at", null: false
      t.datetime "updated_at", null: false
      t.index ["name"], name: "index_prefectures_on_name", unique: true
    end
    create_table "profiles", id: :serial, force: :cascade do |t|
      t.integer "user_id", null: false
      t.string "name", default: "", null: false
      t.string "description", default: "", null: false
      t.datetime "created_at"
      t.datetime "updated_at"
      t.boolean "background_image_animated", default: false, null: false
      t.string "tombo_avatar_file_name"
      t.string "tombo_avatar_content_type"
      t.integer "tombo_avatar_file_size"
      t.datetime "tombo_avatar_updated_at"
      t.string "tombo_background_image_file_name"
      t.string "tombo_background_image_content_type"
      t.integer "tombo_background_image_file_size"
      t.datetime "tombo_background_image_updated_at"
      t.string "url"
      t.index ["user_id"], name: "index_profiles_on_user_id", unique: true
    end
    create_table "programs", id: :serial, force: :cascade do |t|
      t.integer "channel_id", null: false
      t.integer "episode_id", null: false
      t.integer "work_id", null: false
      t.datetime "started_at", null: false
      t.datetime "created_at"
      t.datetime "updated_at"
      t.datetime "sc_last_update"
      t.integer "sc_pid"
      t.boolean "rebroadcast", default: false, null: false
      t.index ["sc_pid"], name: "index_programs_on_sc_pid", unique: true
    end
    create_table "providers", id: :serial, force: :cascade do |t|
      t.integer "user_id", null: false
      t.string "name", null: false
      t.string "uid", null: false
      t.string "token", null: false
      t.integer "token_expires_at"
      t.string "token_secret"
      t.datetime "created_at"
      t.datetime "updated_at"
      t.index ["name", "uid"], name: "index_providers_on_name_and_uid", unique: true
    end
    create_table "receptions", id: :serial, force: :cascade do |t|
      t.integer "user_id", null: false
      t.integer "channel_id", null: false
      t.datetime "created_at"
      t.datetime "updated_at"
      t.index ["user_id", "channel_id"], name: "index_receptions_on_user_id_and_channel_id", unique: true
    end
    create_table "seasons", id: :serial, force: :cascade do |t|
      t.string "name", null: false
      t.datetime "created_at"
      t.datetime "updated_at"
      t.integer "sort_number", null: false
      t.integer "year", null: false
      t.index ["sort_number"], name: "index_seasons_on_sort_number", unique: true
      t.index ["year", "name"], name: "index_seasons_on_year_and_name", unique: true
      t.index ["year"], name: "index_seasons_on_year"
    end
    create_table "sessions", id: :serial, force: :cascade do |t|
      t.string "session_id", null: false
      t.text "data"
      t.datetime "created_at"
      t.datetime "updated_at"
      t.index ["session_id"], name: "index_sessions_on_session_id", unique: true
      t.index ["updated_at"], name: "index_sessions_on_updated_at"
    end
    create_table "settings", id: :serial, force: :cascade do |t|
      t.integer "user_id", null: false
      t.boolean "hide_checkin_comment", default: true, null: false
      t.datetime "created_at", null: false
      t.datetime "updated_at", null: false
      t.boolean "share_record_to_twitter", default: false
      t.boolean "share_record_to_facebook", default: false
      t.string "programs_sort_type", default: "", null: false
      t.string "display_option_work_list", default: "list", null: false
      t.string "display_option_user_work_list", default: "list", null: false
      t.index ["user_id"], name: "index_settings_on_user_id"
    end
    create_table "staffs", id: :serial, force: :cascade do |t|
      t.integer "work_id", null: false
      t.string "name", null: false
      t.string "role", null: false
      t.string "role_other"
      t.string "aasm_state", default: "published", null: false
      t.integer "sort_number", default: 0, null: false
      t.datetime "created_at", null: false
      t.datetime "updated_at", null: false
      t.integer "resource_id", null: false
      t.string "resource_type", null: false
      t.string "name_en", default: "", null: false
      t.string "role_other_en", default: "", null: false
      t.index ["aasm_state"], name: "index_staffs_on_aasm_state"
      t.index ["resource_id", "resource_type"], name: "index_staffs_on_resource_id_and_resource_type"
      t.index ["sort_number"], name: "index_staffs_on_sort_number"
      t.index ["work_id"], name: "index_staffs_on_work_id"
    end
    create_table "statuses", id: :serial, force: :cascade do |t|
      t.integer "user_id", null: false
      t.integer "work_id", null: false
      t.integer "kind", null: false
      t.datetime "created_at"
      t.datetime "updated_at"
      t.integer "likes_count", default: 0, null: false
      t.integer "oauth_application_id"
      t.index ["oauth_application_id"], name: "index_statuses_on_oauth_application_id"
    end
    create_table "syobocal_alerts", id: :serial, force: :cascade do |t|
      t.integer "work_id"
      t.integer "kind", null: false
      t.integer "sc_prog_item_id"
      t.string "sc_sub_title"
      t.string "sc_prog_comment"
      t.datetime "created_at"
      t.datetime "updated_at"
      t.index ["kind"], name: "index_syobocal_alerts_on_kind"
      t.index ["sc_prog_item_id"], name: "index_syobocal_alerts_on_sc_prog_item_id"
    end
    create_table "tips", id: :serial, force: :cascade do |t|
      t.integer "target", null: false
      t.string "slug", null: false
      t.string "title", null: false
      t.string "icon_name", null: false
      t.datetime "created_at", null: false
      t.datetime "updated_at", null: false
      t.string "title_en", default: "", null: false
      t.index ["slug"], name: "index_tips_on_slug", unique: true
    end
    create_table "twitter_bots", id: :serial, force: :cascade do |t|
      t.string "name", null: false
      t.datetime "created_at"
      t.datetime "updated_at"
      t.index ["name"], name: "index_twitter_bots_on_name", unique: true
    end
    create_table "users", id: :serial, force: :cascade do |t|
      t.string "username", null: false
      t.string "email", null: false
      t.string "encrypted_password", default: "", null: false
      t.datetime "remember_created_at"
      t.integer "sign_in_count", default: 0, null: false
      t.datetime "current_sign_in_at"
      t.datetime "last_sign_in_at"
      t.string "current_sign_in_ip"
      t.string "last_sign_in_ip"
      t.string "confirmation_token"
      t.datetime "confirmed_at"
      t.datetime "confirmation_sent_at"
      t.datetime "created_at"
      t.datetime "updated_at"
      t.string "unconfirmed_email"
      t.integer "role", null: false
      t.integer "checkins_count", default: 0, null: false
      t.integer "notifications_count", default: 0, null: false
      t.string "time_zone", default: "", null: false
      t.string "locale", default: "", null: false
      t.string "reset_password_token"
      t.datetime "reset_password_sent_at"
      t.index ["confirmation_token"], name: "index_users_on_confirmation_token", unique: true
      t.index ["email"], name: "index_users_on_email", unique: true
      t.index ["role"], name: "index_users_on_role"
      t.index ["username"], name: "index_users_on_username", unique: true
    end
    create_table "versions", id: :serial, force: :cascade do |t|
      t.string "item_type", null: false
      t.integer "item_id", null: false
      t.string "event", null: false
      t.string "whodunnit"
      t.text "object"
      t.datetime "created_at"
      t.index ["item_type", "item_id"], name: "index_versions_on_item_type_and_item_id"
    end
    create_table "work_images", id: :serial, force: :cascade do |t|
      t.integer "work_id", null: false
      t.integer "user_id", null: false
      t.string "attachment_file_name", null: false
      t.integer "attachment_file_size", null: false
      t.string "attachment_content_type", null: false
      t.datetime "attachment_updated_at", null: false
      t.string "copyright", default: "", null: false
      t.string "asin", default: "", null: false
      t.datetime "created_at", null: false
      t.datetime "updated_at", null: false
      t.index ["user_id"], name: "index_work_images_on_user_id"
      t.index ["work_id"], name: "index_work_images_on_work_id"
    end
    create_table "works", id: :serial, force: :cascade do |t|
      t.string "title", null: false
      t.integer "media", null: false
      t.string "official_site_url", default: "", null: false
      t.string "wikipedia_url", default: "", null: false
      t.date "released_at"
      t.datetime "created_at"
      t.datetime "updated_at"
      t.integer "episodes_count", default: 0, null: false
      t.integer "season_id"
      t.string "twitter_username"
      t.string "twitter_hashtag"
      t.integer "watchers_count", default: 0, null: false
      t.integer "sc_tid"
      t.string "released_at_about"
      t.string "aasm_state", default: "published", null: false
      t.integer "number_format_id"
      t.string "title_kana", default: "", null: false
      t.string "title_ro", default: "", null: false
      t.string "title_en", default: "", null: false
      t.string "official_site_url_en", default: "", null: false
      t.string "wikipedia_url_en", default: "", null: false
      t.text "synopsis", default: "", null: false
      t.text "synopsis_en", default: "", null: false
      t.string "synopsis_source", default: "", null: false
      t.string "synopsis_source_en", default: "", null: false
      t.integer "mal_anime_id"
      t.index ["aasm_state"], name: "index_works_on_aasm_state"
      t.index ["episodes_count"], name: "index_works_on_episodes_count"
      t.index ["media"], name: "index_works_on_media"
      t.index ["number_format_id"], name: "index_works_on_number_format_id"
      t.index ["released_at"], name: "index_works_on_released_at"
      t.index ["watchers_count"], name: "index_works_on_watchers_count"
    end
    add_foreign_key "activities", "users"
    add_foreign_key "casts", "characters"
    add_foreign_key "casts", "people"
    add_foreign_key "casts", "works"
    add_foreign_key "channel_works", "channels"
    add_foreign_key "channel_works", "users"
    add_foreign_key "channel_works", "works"
    add_foreign_key "channels", "channel_groups"
    add_foreign_key "character_images", "characters"
    add_foreign_key "character_images", "users"
    add_foreign_key "checkins", "episodes"
    add_foreign_key "checkins", "multiple_records"
    add_foreign_key "checkins", "oauth_applications"
    add_foreign_key "checkins", "users"
    add_foreign_key "checkins", "works"
    add_foreign_key "comments", "checkins"
    add_foreign_key "comments", "users"
    add_foreign_key "comments", "works"
    add_foreign_key "cover_images", "works"
    add_foreign_key "db_activities", "users"
    add_foreign_key "draft_casts", "casts"
    add_foreign_key "draft_casts", "people"
    add_foreign_key "draft_casts", "works"
    add_foreign_key "draft_episodes", "episodes"
    add_foreign_key "draft_episodes", "episodes", column: "prev_episode_id"
    add_foreign_key "draft_episodes", "works"
    add_foreign_key "draft_items", "items"
    add_foreign_key "draft_items", "works"
    add_foreign_key "draft_multiple_episodes", "works"
    add_foreign_key "draft_organizations", "organizations"
    add_foreign_key "draft_people", "people"
    add_foreign_key "draft_people", "prefectures"
    add_foreign_key "draft_programs", "channels"
    add_foreign_key "draft_programs", "episodes"
    add_foreign_key "draft_programs", "programs"
    add_foreign_key "draft_programs", "works"
    add_foreign_key "draft_staffs", "staffs"
    add_foreign_key "draft_staffs", "works"
    add_foreign_key "draft_works", "number_formats"
    add_foreign_key "draft_works", "seasons"
    add_foreign_key "draft_works", "works"
    add_foreign_key "edit_request_comments", "edit_requests", on_delete: :cascade
    add_foreign_key "edit_request_comments", "users", on_delete: :cascade
    add_foreign_key "edit_request_participants", "edit_requests"
    add_foreign_key "edit_request_participants", "users"
    add_foreign_key "edit_requests", "users", on_delete: :cascade
    add_foreign_key "episodes", "episodes", column: "prev_episode_id"
    add_foreign_key "episodes", "works"
    add_foreign_key "finished_tips", "tips"
    add_foreign_key "finished_tips", "users"
    add_foreign_key "follows", "users"
    add_foreign_key "follows", "users", column: "following_id"
    add_foreign_key "items", "works"
    add_foreign_key "latest_statuses", "episodes", column: "next_episode_id"
    add_foreign_key "latest_statuses", "users"
    add_foreign_key "latest_statuses", "works"
    add_foreign_key "likes", "users"
    add_foreign_key "multiple_records", "users"
    add_foreign_key "multiple_records", "works"
    add_foreign_key "mute_users", "users"
    add_foreign_key "mute_users", "users", column: "muted_user_id"
    add_foreign_key "notifications", "users"
    add_foreign_key "notifications", "users", column: "action_user_id"
    add_foreign_key "oauth_access_grants", "oauth_applications", column: "application_id"
    add_foreign_key "oauth_access_grants", "users", column: "resource_owner_id"
    add_foreign_key "oauth_access_tokens", "oauth_applications", column: "application_id"
    add_foreign_key "oauth_access_tokens", "users", column: "resource_owner_id"
    add_foreign_key "people", "prefectures"
    add_foreign_key "profiles", "users"
    add_foreign_key "programs", "channels"
    add_foreign_key "programs", "episodes"
    add_foreign_key "programs", "works"
    add_foreign_key "providers", "users"
    add_foreign_key "receptions", "channels"
    add_foreign_key "receptions", "users"
    add_foreign_key "settings", "users"
    add_foreign_key "staffs", "works"
    add_foreign_key "statuses", "oauth_applications"
    add_foreign_key "statuses", "users"
    add_foreign_key "statuses", "works"
    add_foreign_key "syobocal_alerts", "works"
    add_foreign_key "work_images", "users"
    add_foreign_key "work_images", "works"
    add_foreign_key "works", "number_formats"
    add_foreign_key "works", "seasons"
  end

  def down
    raise ActiveRecord::IrreversibleMigration, "The initial migration is not revertable"
  end
end
