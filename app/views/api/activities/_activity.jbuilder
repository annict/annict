json.(activity, :id, :action, :created_at)

json.links do
  json.user do
    json.username activity.user.username
  end

  json.profile do
    json.avatar_url tombo_thumb_url(activity.user.profile, :tombo_avatar, 'w:70,h:70')
    json.name activity.user.profile.name
  end

  json.meta do
    json.liked (user_signed_in? && current_user.like_r?(activity.trackable))
  end

  if activity.action == 'checkins.create'
    json.partial!('/api/activities/checkins_create', activity: activity)
  elsif activity.action == 'statuses.create'
    json.partial!('/api/activities/statuses_create', activity: activity)
  end
end
