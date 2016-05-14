# frozen_string_literal: true

json.programs @programs do |program|
  json.partial!("/api/v1/programs/program", program: program, params: @params, field_prefix: "")

  json.channel do
    json.partial!("/api/v1/channels/channel", channel: program.channel, params: @params, field_prefix: "channel.")
  end

  json.work do
    json.partial!("/api/v1/works/work", work: program.work, params: @params, field_prefix: "work.")
  end

  json.episode do
    json.partial!("/api/v1/episodes/episode", episode: program.episode, params: @params, field_prefix: "episode.")
  end
end

json.partial!("/api/v1/application/pagination", collection: @programs, params: @params)
