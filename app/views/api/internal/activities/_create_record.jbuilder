# frozen_string_literal: true

json.work do
  json.call(activity.recipient.work, :id, :title)
end

json.item do
  json.image_url annict_image_url(activity.recipient.work.item, :tombo_image, size: "140x140")
end

json.episode do
  json.call(activity.recipient, :id, :number, :title)
end

json.record do
  json.call(activity.trackable, :id, :comment, :comments_count, :likes_count, :rating)
  json.hide_comment user_signed_in? && current_user.hide_checkin_comment?(activity.trackable)
end
