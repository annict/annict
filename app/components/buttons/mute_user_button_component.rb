# frozen_string_literal: true

module Buttons
  class MuteUserButtonComponent < ApplicationV6Component
    def initialize(view_context, user:, class_name: "")
      super view_context
      @user = user
      @class_name = class_name
    end

    def render
      build_html do |h|
        h.tag :button, {
          class: "btn #{@class_name}",
          data_action: "mute-user-button#toggle",
          data_controller: "mute-user-button",
          data_mute_user_button_default_class: "btn-outline-primary",
          data_mute_user_button_default_text_value: t("verb.mute"),
          data_mute_user_button_muted_class: "btn-primary",
          data_mute_user_button_muted_text_value: t("verb.unmute"),
          data_mute_user_button_user_id_value: @user.id,
          type: "button"
        } do
          h.text "Mute"
        end
      end
    end
  end
end
