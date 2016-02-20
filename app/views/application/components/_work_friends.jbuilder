json.array! users do |user|
  json.call(user, :username)
  json.avatar_url annict_image_url(user.profile, :tombo_avatar, size: "30x30")
end
