# frozen_string_literal: true

module Deprecated
  class SlotEntity < Deprecated::ApplicationEntity
    attribute? :started_at, Types::Params::Time
    attribute? :rebroadcast, Types::Bool
    attribute? :episode, Deprecated::EpisodeEntity
    attribute? :anime, Deprecated::AnimeEntity
    attribute? :channel, Deprecated::ChannelEntity

    def self.from_nodes(nodes)
      nodes.map do |node|
        from_node(node)
      end
    end

    def self.from_node(node)
      attrs = {}

      if started_at = node["startedAt"]
        attrs[:started_at] = started_at
      end

      if rebroadcast = node["rebroadcast"]
        attrs[:rebroadcast] = rebroadcast
      end

      if episode_node = node["episode"]
        attrs[:episode] = Deprecated::EpisodeEntity.from_node(episode_node)
      end

      if anime_node = node["anime"]
        attrs[:anime] = Deprecated::AnimeEntity.from_node(anime_node)
      end

      if channel_node = node["channel"]
        attrs[:channel] = Deprecated::ChannelEntity.from_node(channel_node)
      end

      new attrs
    end
  end
end
