# frozen_string_literal: true

class SeriesAnimeEntity < ApplicationEntity
  local_attributes :summary, :title

  attribute? :database_id, Types::Integer
  attribute? :summary, Types::String.optional
  attribute? :summary_en, Types::String.optional
  attribute? :title, Types::String
  attribute? :title_en, Types::String.optional
  attribute? :image_url_1x, Types::String.optional
  attribute? :image_url_2x, Types::String.optional

  def self.from_edges(series_work_edges)
    series_work_edges.map do |series_work_edge|
      from_edge(series_work_edge)
    end
  end

  def self.from_edge(series_work_edge)
    attrs = {}

    if database_id = series_work_edge.dig("node", "databaseId")
      attrs[:database_id] = database_id
    end

    if summary = series_work_edge["summary"]
      attrs[:summary] = summary
    end

    if summary_en = series_work_edge["summaryEn"]
      attrs[:summary_en] = summary_en
    end

    if title = series_work_edge.dig("node", "title")
      attrs[:title] = title
    end

    if title_en = series_work_edge.dig("node", "titleEn")
      attrs[:title_en] = title_en
    end

    if image_url_1x = series_work_edge.dig("node", "image", "internalUrl1x")
      attrs[:image_url_1x] = image_url_1x
    end

    if image_url_2x = series_work_edge.dig("node", "image", "internalUrl2x")
      attrs[:image_url_2x] = image_url_2x
    end

    new attrs
  end
end
