json.programs @programs do |program|
  json.(program, :id, :started_at)

  json.channel do
    json.name program.channel.name
  end

  json.episode do
    json.id program.episode.id
    json.number program.episode.number
    json.title program.episode.title
  end

  json.work do
    json.id program.work.id
    json.title program.episode.work.title
    json.image_url program.work.main_item.image.thumb('140x140').url
  end
end
