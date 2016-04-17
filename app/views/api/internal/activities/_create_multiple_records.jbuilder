# frozen_string_literal: true

json.work do
  json.call(activity.recipient, :id, :title)
end

json.item do
  json.image_url annict_image_url(activity.recipient.item, :tombo_image, size: "140x140")
end

json.multiple_record do
  json.call(activity.trackable, :id, :likes_count)
end

records = activity.trackable.checkins.joins(:episode).order("episodes.sort_number ASC")
json.episodes records do |record|
  json.call(record.episode, :id, :number, :title)
end
