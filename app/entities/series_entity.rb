# frozen_string_literal: true

class SeriesEntity < ApplicationEntity
  local_attributes :name

  attribute? :name, Types::String
  attribute? :name_en, Types::String.optional
  attribute? :series_works, Types::Array.of(SeriesWorkEntity)

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

    series_work_edges = series_node.dig("works", "edges")
    attrs[:series_works] = SeriesWorkEntity.from_edges(series_work_edges || [])

    new attrs
  end
end
