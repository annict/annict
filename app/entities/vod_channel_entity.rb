# frozen_string_literal: true

class VodChannelEntity < ApplicationEntity
  attribute? :database_id, Types::Integer
  attribute? :name, Types::String
  attribute? :programs, Types::Array.of(ProgramEntity)

  def self.from_nodes(channel_nodes, anime_entity:)
    channel_nodes.map do |channel_node|
      from_node(channel_node, anime_entity: anime_entity)
    end
  end

  def self.from_node(channel_node, anime_entity:)
    attrs = {}

    if database_id = channel_node["databaseId"]
      attrs[:database_id] = database_id
    end

    if name = channel_node["name"]
      attrs[:name] = name
    end

    attrs[:programs] = anime_entity.programs.select do |program_entity|
      program_entity.channel.database_id == channel_node["databaseId"]
    end

    new attrs
  end
end
