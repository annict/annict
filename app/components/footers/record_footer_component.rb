# frozen_string_literal: true

module Footers
  class RecordFooterComponent < ApplicationV6Component
    def initialize(view_context, record:)
      super view_context
      @record = record
    end

    def render
      build_html do |h|
        h.tag :div, class: "c-record-footer" do
          h.html Buttons::LikeButtonComponent.new(view_context,
            resource_name: "Record",
            resource_id: @record.id,
            likes_count: @record.likes_count,
            page_category: @page_category).render

          h.tag :a, href: view_context.record_path(@record.user.username, @record.id), class: "ms-3", data_turbo_frame: "_top" do
            h.tag :i, class: "far fa-comment me-1"
            h.text @record.comments_count
          end
        end
      end
    end
  end
end
