# frozen_string_literal: true

json.work do
  json.call(activity.recipient, :id, :title)
end

json.item do
  json.image_url annict_image_url(activity.recipient.item, :tombo_image, size: "140x140")
end

json.status do
  json.call(activity.trackable, :id, :likes_count)
  json.kind activity.trackable.kind.text
end
