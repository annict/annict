# frozen_string_literal: true

module V4
  class ProgramEntity < V4::ApplicationEntity
    attribute? :id, Types::String
    attribute? :vod_title_name, Types::String.optional
    attribute? :vod_title_url, Types::String.optional
    attribute? :viewer_did_check, Types::Bool
    attribute? :started_at, Types::Params::Time
    attribute? :channel, V4::ChannelEntity

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

      if viewer_did_check = node["viewerDidCheck"]
        attrs[:viewer_did_check] = viewer_did_check
      end

      if started_at = node["startedAt"]
        attrs[:started_at] = started_at
      end

      if channel_node = node["channel"]
        attrs[:channel] = V4::ChannelEntity.from_node(channel_node)
      end

      new attrs
    end
  end
end
