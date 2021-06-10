# frozen_string_literal: true

module Buttons
  class LikeButtonComponent < ApplicationV6Component
    def initialize(view_context, resource_name:, resource_id:, likes_count:, page_category:, class_name: "", init_is_liked: false)
      super view_context
      @resource_name = resource_name
      @resource_id = resource_id
      @likes_count = likes_count
      @page_category = page_category
      @class_name = class_name
      @init_is_liked = init_is_liked == true
    end

    def render
      build_html do |h|
        h.tag :div,
          class: "c-like-button #{like_button_class_name}",
          data_controller: "like-button",
          data_like_button_resource_name: @resource_name,
          data_like_button_resource_id: @resource_id,
          data_like_button_page_category: @page_category,
          data_like_button_init_is_liked_value: @init_is_liked,
          data_action: "click->like-button#toggle" do
          h.tag :i, class: "c-like-button__icon"
          h.tag :span, class: "c-like-button__count ms-1", data_like_button_target: "count" do
            h.text @likes_count
          end
        end
      end
    end

    private

    def like_button_class_name
      classes = %w[d-inline-block u-fake-link]
      classes += @class_name.split(" ")
      classes.uniq.join(" ")
    end
  end
end
