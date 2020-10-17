# frozen_string_literal: true

class LibraryEntryEntity < ApplicationEntity
  attribute? :tracked_episodes_count, Types::Integer
  attribute? :anime, AnimeEntity

  def self.from_nodes(nodes)
    nodes.map do |node|
      from_node(node)
    end
  end

  def self.from_node(node)
    attrs = {}

    if tracked_episodes_count = node["trackedEpisodesCount"]
      attrs[:tracked_episodes_count] = tracked_episodes_count
    end

    if anime_node = node["anime"]
      attrs[:anime] = AnimeEntity.from_node(anime_node)
    end

    new attrs
  end
end
