# frozen_string_literal: true

class ProgramEntity < ApplicationEntity
  attribute? :vod_title_name, Types::String.optional
  attribute? :vod_title_url, Types::String.optional
  attribute? :channel, ChannelEntity

  def self.from_nodes(program_nodes)
    program_nodes.map do |program_node|
      from_node(program_node)
    end
  end

  def self.from_node(program_node)
    attrs = {}

    if vod_title_name = program_node["vodTitleName"]
      attrs[:vod_title_name] = vod_title_name
    end

    if vod_title_url = program_node["vodTitleUrl"]
      attrs[:vod_title_url] = vod_title_url
    end

    if channel_node = program_node["channel"]
      attrs[:channel] = ChannelEntity.from_node(channel_node)
    end

    new attrs
  end
end
