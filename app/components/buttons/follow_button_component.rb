# frozen_string_literal: true

module Buttons
  class FollowButtonComponent < ApplicationV6Component
    def initialize(view_context, user:, page_category:, class_name: "")
      super view_context
      @user = user
      @page_category = page_category
      @class_name = class_name
    end

    def render
      build_html do |h|
        h.tag :button, {
          class: "btn c-follow-button #{@class_name}",
          data_action: "follow-button#toggle",
          data_controller: "follow-button",
          data_follow_button_default_class: "btn-outline-primary",
          data_follow_button_default_text_value: t("noun.follow"),
          data_follow_button_following_class: "btn-primary",
          data_follow_button_following_text_value: t("noun.following"),
          data_follow_button_user_id_value: @user.id,
          type: "button"
        } do
          h.text "Follow"
        end
      end
    end
  end
end
