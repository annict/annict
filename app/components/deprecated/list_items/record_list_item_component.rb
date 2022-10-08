# frozen_string_literal: true

module Deprecated::ListItems
  class RecordListItemComponent < Deprecated::ApplicationV6Component
    def initialize(view_context, record:, show_box: true, show_options: true)
      super view_context
      @record = record
      @show_box = show_box
      @show_options = show_options
    end

    def render
      build_html do |h|
        h.tag :turbo_frame, id: dom_id(@record) do
          h.html Deprecated::Headers::RecordHeaderComponent.new(view_context, record: @record, show_box: @show_box, show_options: @show_options).render

          h.tag :div, class: "mt-3" do
            h.html Deprecated::Contents::RecordContentComponent.new(view_context, record: @record, show_box: @show_box).render
          end
        end
      end
    end
  end
end
