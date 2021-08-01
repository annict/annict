# frozen_string_literal: true

module ListItems
  class RecordListItemComponent < ApplicationV6Component
    def initialize(view_context, record:, show_box: true)
      super view_context
      @record = record
      @show_box = show_box
    end

    def render
      build_html do |h|
        h.tag :turbo_frame, id: dom_id(@record) do
          h.html Headers::RecordHeaderComponent.new(view_context, record: @record).render

          h.tag :div, class: "mt-3" do
            h.html Contents::RecordContentComponent.new(view_context, record: @record, show_box: @show_box).render
          end
        end
      end
    end
  end
end
