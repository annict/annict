# frozen_string_literal: true

class EpisodeEntity < ApplicationEntity
  local_attributes :number, :title

  attribute? :id, Types::String
  attribute? :database_id, Types::Integer
  attribute? :raw_number, Types::Float.optional
  attribute? :number, Types::String.optional
  attribute? :number_en, Types::String.optional
  attribute? :title, Types::String.optional
  attribute? :title_en, Types::String.optional
  attribute? :viewer_did_track_in_current_status, Types::Bool
  attribute? :episode_records_count, Types::Integer
  attribute? :commented_episode_records_count, Types::Integer
  attribute? :prev_episode, EpisodeEntity
  attribute? :next_episode, EpisodeEntity
  attribute? :anime, AnimeEntity
  attribute? :records, Types::Array.of(RecordEntity)

  def self.from_nodes(nodes)
    nodes.map do |node|
      from_node(node)
    end
  end

  def self.from_node(node)
    attrs = {}

    if id = node["id"]
      attrs[:id] = id
    end

    if database_id = node["databaseId"]
      attrs[:database_id] = database_id
    end

    if raw_number = node["rawNumber"]
      attrs[:raw_number] = raw_number
    end

    if number = node["number"]
      attrs[:number] = number
    end

    if number_en = node["numberEn"]
      attrs[:number_en] = number_en
    end

    if title = node["title"]
      attrs[:title] = title
    end

    if title_en = node["titleEn"]
      attrs[:title_en] = title_en
    end

    if viewer_did_track_in_current_status = node["viewerDidTrackInCurrentStatus"]
      attrs[:viewer_did_track_in_current_status] = viewer_did_track_in_current_status
    end

    if episode_records_count = node["episodeRecordsCount"]
      attrs[:episode_records_count] = episode_records_count
    end

    if commented_episode_records_count = node["commentedEpisodeRecordsCount"]
      attrs[:commented_episode_records_count] = commented_episode_records_count
    end

    if prev_episode_node = node["prevEpisode"]
      attrs[:prev_episode] = from_node(prev_episode_node)
    end

    if next_episode_node = node["nextEpisode"]
      attrs[:next_episode] = from_node(next_episode_node)
    end

    if anime_node = node["anime"]
      attrs[:anime] = AnimeEntity.from_node(anime_node)
    end

    record_nodes = node.dig("records", "nodes")
    attrs[:records] = RecordEntity.from_nodes(record_nodes || [])

    new attrs
  end

  def title_with_number
    if local_number.present? && local_title.present?
      return "#{local_number} #{local_title}"
    end

    if local_number.blank?
      return local_title
    end

    local_number
  end
end
