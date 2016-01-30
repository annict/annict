json.programs @programs do |program|
  json.(program, :started_at, :rebroadcast)

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
      json.image_url annict_image_url(program.work.item, :tombo_image, size: "70x70")
    end
  end
end
