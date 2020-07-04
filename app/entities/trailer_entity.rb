# frozen_string_literal: true

class TrailerEntity < ApplicationEntity
  attribute? :title, Types::String
  attribute? :url, Types::String
  attribute? :image_url, Types::String

  def self.from_nodes(trailer_nodes)
    trailer_nodes.map do |trailer_node|
      from_node(trailer_node)
    end
  end

  def self.from_node(trailer_node)
    attrs = {}

    if title = trailer_node["title"]
      attrs[:title] = title
    end

    if url = trailer_node["url"]
      attrs[:url] = url
    end

    if image_url = trailer_node["internalImageUrl"]
      attrs[:image_url] = image_url
    end

    new attrs
  end
end
