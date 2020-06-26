# frozen_string_literal: true

class PageInfoEntity < ApplicationEntity
  attribute? :start_cursor, Types::String.optional
  attribute? :end_cursor, Types::String.optional
  attribute? :has_next_page, Types::Bool
  attribute? :has_previous_page, Types::Bool

  def self.from_node(page_info_node)
    attrs = {}

    if start_cursor = page_info_node["startCursor"]
      attrs[:start_cursor] = start_cursor
    end

    if end_cursor = page_info_node["endCursor"]
      attrs[:end_cursor] = end_cursor
    end

    if has_next_page = page_info_node["hasNextPage"]
      attrs[:has_next_page] = has_next_page
    end

    if has_previous_page = page_info_node["hasPreviousPage"]
      attrs[:has_previous_page] = has_previous_page
    end

    new attrs
  end
end
