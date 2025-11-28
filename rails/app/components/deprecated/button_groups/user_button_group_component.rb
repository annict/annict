# typed: false
# frozen_string_literal: true

module Deprecated::ButtonGroups
  class UserButtonGroupComponent < Deprecated::ApplicationV6Component
    def initialize(view_context, user:, class_name: "")
      super view_context
      @user = user
      @class_name = class_name
    end

    def render
      build_html do |h|
        h.tag :div, class: "btn-group c-user-button-group #{@class_name}" do
          h.html Deprecated::Buttons::FollowButtonComponent.new(view_context, user: @user, page_category: page_category).render

          h.tag :div, class: "btn-group" do
            h.tag :button, {
              class: "btn btn-outline-primary dropdown-toggle",
              data_bs_toggle: "dropdown",
              type: "button"
            } do
              h.tag :i, class: "fa-solid fa-ellipsis-h"
            end

            h.tag :ul, class: "dropdown-menu" do
              h.tag :li do
                h.html Deprecated::Buttons::MuteUserButtonComponent.new(view_context, user: @user, class_name: "dropdown-item").render
              end
            end
          end
        end
      end
    end
  end
end
