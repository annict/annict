# frozen_string_literal: true

json.call(activity, :action, :created_at)

json.user do
  json.username activity.user.username
end

json.profile do
  json.avatar_url annict_image_url(activity.user.profile, :tombo_avatar, size: "25x25")
  json.name activity.user.profile.name
end

json.meta do
  json.liked user_signed_in? && current_user.like_r?(activity.trackable)
end

if activity.action.create_record?
  json.partial!("/api/internal/activities/create_record", activity: activity)
elsif activity.action.create_status?
  json.partial!("/api/internal/activities/create_status", activity: activity)
end
