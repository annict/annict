json.(check, :id)

json.work do
  json.(check.work, :id, :title)
  json.image_url tombo_thumb_url(check.work.main_item, :tombo_image, "w:80,h:80")
end

if check.episode.present?
  json.episode do
    json.(check.episode, :id, :number, :title)
  end
end
