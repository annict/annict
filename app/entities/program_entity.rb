# frozen_string_literal: true

class ProgramEntity < ApplicationEntity
  attribute? :id, Types::String
  attribute? :vod_title_name, Types::String.optional
  attribute? :vod_title_url, Types::String.optional
  attribute? :started_at, Types::Params::Time
  attribute? :channel, ChannelEntity

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

    if vod_title_name = node["vodTitleName"]
      attrs[:vod_title_name] = vod_title_name
    end

    if vod_title_url = node["vodTitleUrl"]
      attrs[:vod_title_url] = vod_title_url
    end

    if started_at = node["startedAt"]
      attrs[:started_at] = started_at
    end

    if channel_node = node["channel"]
      attrs[:channel] = ChannelEntity.from_node(channel_node)
    end

    new attrs
  end
end
