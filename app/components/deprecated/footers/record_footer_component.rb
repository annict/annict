# typed: false
# frozen_string_literal: true

module Deprecated::Footers
  class RecordFooterComponent < Deprecated::ApplicationV6Component
    def initialize(view_context, record:)
      super view_context
      @record = record
    end

    def render
      build_html do |h|
        h.tag :div, class: "c-record-footer" do
          h.html Deprecated::Buttons::LikeButtonComponent.new(view_context,
            resource_name: "Record",
            resource_id: @record.id,
            likes_count: @record.likes_count,
            page_category: @page_category).render

          if @record.episode_record?
            h.tag :a, href: view_context.record_path(@record.user.username, @record.id), class: "ms-3", data_turbo_frame: "_top" do
              h.tag :i, class: "fa-solid fa-comment me-1"
              h.text @record.episode_record.comments_count
            end
          end
        end
      end
    end
  end
end
