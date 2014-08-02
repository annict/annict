json.work do
  json.(activity.recipient, :id, :title)
end

json.main_item do
  json.image_url activity.recipient.main_item.image.thumb('80x80').url
end

json.status do
  json.(activity.trackable, :id, :likes_count)
  json.kind activity.trackable.kind.text
end