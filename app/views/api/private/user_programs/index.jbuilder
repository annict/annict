json.programs @programs do |program|
  json.(program, :started_at)

  json.links do
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
      json.image_url program.work.decorate.item_image_url("w:140,h:140")
    end
  end
end
