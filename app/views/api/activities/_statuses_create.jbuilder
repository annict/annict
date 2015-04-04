json.work do
  json.(activity.recipient, :id, :title)
end

json.main_item do
  json.image_url tombo_thumb_url(activity.recipient.main_item, :tombo_image, 'w:140,h:140')
end

json.status do
  json.(activity.trackable, :id, :likes_count)
  json.kind activity.trackable.kind.text
end
