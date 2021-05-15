# frozen_string_literal: true

module V4
  class SlotEntity < V4::ApplicationEntity
    attribute? :started_at, Types::Params::Time
    attribute? :rebroadcast, Types::Bool
    attribute? :episode, V4::EpisodeEntity
    attribute? :anime, V4::AnimeEntity
    attribute? :channel, V4::ChannelEntity

    def self.from_nodes(nodes)
      nodes.map do |node|
        from_node(node)
      end
    end

    def self.from_node(node)
      attrs = {}

      if (started_at = node["startedAt"])
        attrs[:started_at] = started_at
      end

      if (rebroadcast = node["rebroadcast"])
        attrs[:rebroadcast] = rebroadcast
      end

      if (episode_node = node["episode"])
        attrs[:episode] = V4::EpisodeEntity.from_node(episode_node)
      end

      if (anime_node = node["anime"])
        attrs[:anime] = V4::AnimeEntity.from_node(anime_node)
      end

      if (channel_node = node["channel"])
        attrs[:channel] = V4::ChannelEntity.from_node(channel_node)
      end

      new attrs
    end
  end
end
