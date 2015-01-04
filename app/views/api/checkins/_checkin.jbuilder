json.(checkin, :id, :comment, :spoil, :twitter_click_count, :comments_count,
  :likes_count, :created_at)

json.user do
  json.username checkin.user.username
end

json.profile do
  json.avatar_url checkin.user.profile.avatar.thumb('100x100').url
  json.name checkin.user.profile.name
end

json.meta do
  json.liked (user_signed_in? && current_user.like_r?(checkin))
end
