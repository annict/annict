# frozen_string_literal: true

class PaginationComponent < ApplicationComponent
  def initialize(page_info_entity:, resources_path:, position: "center")
    @page_info_entity = page_info_entity
    @resources_path = resources_path
    @position = position
  end

  private

  attr_reader :page_info_entity, :position, :resources_path

  def pagination_class_name
    ["pagination", "justify-content-#{position}"].join(" ")
  end

  def prev_page_item_class_name
    classes = %w(page-item)

    unless page_info_entity.has_previous_page
      classes << "disabled"
    end

    classes.join(" ")
  end

  def next_page_item_class_name
    classes = %w(page-item)

    unless page_info_entity.has_next_page
      classes << "disabled"
    end

    classes.join(" ")
  end

  def next_resources_path
    query_values.
      except("before").
      merge(after: page_info_entity.end_cursor)
  end

  def prev_resources_path
    query_values.
      except("after").
      merge(before: page_info_entity.start_cursor)
  end

  def query_values
    @query_values ||= (Addressable::URI.parse(resources_path).query_values || {})
  end
end
