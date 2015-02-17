json.work do
  json.(activity.recipient, :id, :title)
end

json.main_item do
  json.image_url thumb_url(activity.recipient.main_item, :image, '140x140e')
end

json.status do
  json.(activity.trackable, :id, :likes_count)
  json.kind activity.trackable.kind.text
end
