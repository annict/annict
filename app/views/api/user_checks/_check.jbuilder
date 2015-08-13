json.(check, :id)

json.work do
  json.(check.work, :id, :title)
  json.image_url check.work.decorate.item_image_url("w:80,h:80")
end

if check.episode.present?
  json.episode do
    json.(check.episode, :id, :number, :title)
  end
end
