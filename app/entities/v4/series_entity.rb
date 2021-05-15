# frozen_string_literal: true

module Deprecated
  class SeriesEntity < Deprecated::ApplicationEntity
    local_attributes :name

    attribute? :name, Types::String
    attribute? :name_en, Types::String.optional
    attribute? :series_anime_list, Types::Array.of(SeriesAnimeEntity)

    def self.from_nodes(series_nodes)
      series_nodes.map do |series_node|
        from_node(series_node)
      end
    end

    def self.from_node(series_node)
      attrs = {}

      if name = series_node["name"]
        attrs[:name] = name
      end

      if name_en = series_node["nameEn"]
        attrs[:name_en] = name_en
      end

      series_anime_edges = series_node.dig("animeList", "edges")
      attrs[:series_anime_list] = Deprecated::SeriesAnimeEntity.from_edges(series_anime_edges || [])

      new attrs
    end
  end
end
