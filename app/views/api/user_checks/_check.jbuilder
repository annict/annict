json.(check, :id)

json.work do
  json.(check.work, :id, :title)
  json.image_url thumb_url(check.work.main_item, :image, "80x80e")
end

if check.episode.present?
  json.episode do
    json.(check.episode, :id, :number, :title)
  end
end
