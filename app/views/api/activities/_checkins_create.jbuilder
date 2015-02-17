json.work do
  json.(activity.recipient.work, :id, :title)
end

json.main_item do
  json.image_url thumb_url(activity.recipient.work.main_item, :image, '140x140e')
end

json.episode do
  json.(activity.recipient, :id, :number, :title)
end

json.checkin do
  json.(activity.trackable, :id, :comment, :comments_count, :likes_count)
  json.hide_comment user_signed_in? && current_user.hide_checkin_comment?(activity.trackable)
end
