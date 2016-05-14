# frozen_string_literal: true

json.records @records do |record|
  json.partial!("/api/v1/records/record", record: record, params: @params, field_prefix: "")

  json.user do
    json.partial!("/api/v1/users/user", user: record.user, params: @params, field_prefix: "user.", show_all: false)
  end

  json.work do
    json.partial!("/api/v1/works/work", work: record.work, params: @params, field_prefix: "work.")
  end

  json.episode do
    json.partial!("/api/v1/episodes/episode", episode: record.episode, params: @params, field_prefix: "episode.")
  end
end

json.partial!("/api/v1/application/pagination", collection: @records, params: @params)
