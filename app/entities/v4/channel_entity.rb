# frozen_string_literal: true

module V4
  class ChannelEntity < V4::ApplicationEntity
    attribute? :database_id, Types::Integer
    attribute? :name, Types::String

    def self.from_node(channel_node)
      attrs = {}

      if (database_id = channel_node["databaseId"])
        attrs[:database_id] = database_id
      end

      if (name = channel_node["name"])
        attrs[:name] = name
      end

      new attrs
    end
  end
end
