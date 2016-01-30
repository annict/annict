json.(latest_status, :id)

json.work do
  json.(latest_status.work, :id, :title)
  json.image_url annict_image_url(latest_status.work.item, :tombo_image, size: "40x40")
end

if latest_status.next_episode.present?
  json.next_episode do
    json.(latest_status.next_episode, :id, :number, :title)
  end
end
