json.work do
  json.(activity.recipient.work, :id, :title)
end

json.main_item do
  json.image_url thumb_url(activity.recipient.work.main_item, :image, '80x80')
end

json.episode do
  json.(activity.recipient, :id, :number, :title)
end

json.checkin do
  json.(activity.trackable, :id, :comment, :spoil, :comments_count, :likes_count)
end
