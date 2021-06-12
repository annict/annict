# frozen_string_literal: true

class LibraryEntryEntity < ApplicationEntity
  attribute? :tracked_episodes_count_in_current_status, Types::Integer
  attribute? :anime, AnimeEntity

  def self.from_nodes(nodes)
    nodes.map do |node|
      from_node(node)
    end
  end

  def self.from_node(node)
    attrs = {}

    if (tracked_episodes_count_in_current_status = node["trackedEpisodesCountInCurrentStatus"])
      attrs[:tracked_episodes_count_in_current_status] = tracked_episodes_count_in_current_status
    end

    if (anime_node = node["anime"])
      attrs[:anime] = V4::AnimeEntity.from_node(anime_node)
    end

    new attrs
  end
end
