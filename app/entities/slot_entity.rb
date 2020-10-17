# frozen_string_literal: true

class SlotEntity < ApplicationEntity
  attribute? :started_at, Types::Params::Time
  attribute? :rebroadcast, Types::Bool
  attribute? :episode, EpisodeEntity
  attribute? :anime, AnimeEntity
  attribute? :channel, ChannelEntity

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
      attrs[:episode] = EpisodeEntity.from_node(episode_node)
    end

    if anime_node = node["anime"]
      attrs[:anime] = AnimeEntity.from_node(anime_node)
    end

    if channel_node = node["channel"]
      attrs[:channel] = ChannelEntity.from_node(channel_node)
    end

    new attrs
  end
end
