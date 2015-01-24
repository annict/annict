json.(activity, :id, :action, :created_at)

json.links do
  json.user do
    json.username activity.user.username
  end

  json.profile do
    json.avatar_url thumb_url(activity.user.profile, :avatar, '100x100e')
    json.name activity.user.profile.name
  end

  json.meta do
    json.liked (user_signed_in? && current_user.like_r?(activity.trackable))
  end

  if 'checkins.create' == activity.action
    json.partial!('/api/activities/checkins_create', activity: activity)
  elsif 'statuses.create' == activity.action
    json.partial!('/api/activities/statuses_create', activity: activity)
  end
end
