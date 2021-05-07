# frozen_string_literal: true

class RecordFooterComponent < ApplicationComponent
  def initialize(view_context, record:)
    super view_context
    @record = record
  end

  def render
    build_html do |h|
      h.tag :div, class: "c-record-footer" do
        h.html LikeButtonComponent2.new(view_context,
          resource_name: "Record",
          resource_id: @record.id,
          likes_count: @record.likes_count,
          page_category: "user-home",
          class_name: "me-3",
          init_is_liked: @record.is_liked).render

        if @record.episode_record?
          h.tag :a, href: view_context.record_path(@record.user.username, @record.id), class: "me-3", data_turbo_frame: "_top" do
            h.tag :i, class: "far fa-comment"
            h.text @record.episode_record.comments_count
          end
        end
      end
    end
  end
end
