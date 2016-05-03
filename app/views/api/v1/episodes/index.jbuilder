# frozen_string_literal: true

json.episodes @episodes do |episode|
  json.partial!("/api/v1/episodes/episode", episode: episode, params: @params, field_prefix: "")

  json.work do
    json.partial!("/api/v1/works/work", work: episode.work, params: @params, field_prefix: "work.")
  end

  if episode.prev_episode.present?
    json.prev_episode do
      json.partial!("/api/v1/episodes/episode", episode: episode.prev_episode, params: @params, field_prefix: "prev_episode.")
    end
  else
    json.prev_episode nil if @params.fields.blank? || @params.fields&.include?("prev_episode")
  end

  if episode.next_episode.present?
    json.next_episode do
      json.partial!("/api/v1/episodes/episode", episode: episode.next_episode, params: @params, field_prefix: "next_episode.")
    end
  else
    json.next_episode nil if @params.fields.blank? || @params.fields.include?("next_episode")
  end
end
