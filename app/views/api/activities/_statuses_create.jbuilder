json.work do
  json.(activity.recipient, :id, :title)
end

json.main_item do
  json.image_url activity.recipient.decorate.item_image_url('w:140,h:140')
end

json.status do
  json.(activity.trackable, :id, :likes_count)
  json.kind activity.trackable.kind.text
end
