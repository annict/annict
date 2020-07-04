# frozen_string_literal: true

class EpisodeEntity < ApplicationEntity
  local_attributes :title

  attribute? :database_id, Types::Integer
  attribute? :number_text, Types::String.optional
  attribute? :title, Types::String.optional
  attribute? :title_en, Types::String.optional

  def self.from_nodes(episode_nodes)
    episode_nodes.map do |episode_node|
      from_node(episode_node)
    end
  end

  def self.from_node(episode_node)
    attrs = {}

    if database_id = episode_node["databaseId"]
      attrs[:database_id] = database_id
    end

    if number_text = episode_node["numberText"]
      attrs[:number_text] = number_text
    end

    if title = episode_node["title"]
      attrs[:title] = title
    end

    if title_en = episode_node["titleEn"]
      attrs[:title_en] = title_en
    end

    new attrs
  end
end
