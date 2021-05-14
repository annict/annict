# frozen_string_literal: true

module Deprecated
  class CharacterEntity < Deprecated::ApplicationEntity
    local_attributes :name

    attribute? :database_id, Types::Integer
    attribute? :name, Types::String
    attribute? :name_en, Types::String.optional
    attribute? :series, Deprecated::SeriesEntity.optional

    def self.from_node(character_node)
      attrs = {}

      if database_id = character_node["databaseId"]
        attrs[:database_id] = database_id
      end

      if name = character_node["name"]
        attrs[:name] = name
      end

      if name_en = character_node["nameEn"]
        attrs[:name_en] = name_en
      end

      if series_node = character_node["series"]
        attrs[:series] = Deprecated::SeriesEntity.from_node(series_node)
      end

      new attrs
    end
  end
end
