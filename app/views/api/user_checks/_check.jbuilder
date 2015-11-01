json.(check, :id)

json.work do
  json.(check.work, :id, :title)
  json.image_url annict_image_url(check.work.item, :tombo_image, size: "40x40")
end

if check.episode.present?
  json.episode do
    json.(check.episode, :id, :number, :title)
  end
end
