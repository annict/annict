# typed: false
# frozen_string_literal: true

module Deprecated::ButtonGroups
  class WorkButtonGroupComponent < Deprecated::ApplicationV6Component
    def initialize(view_context, work:, class_name: "", show_option_button: true)
      super view_context
      @work = work
      @class_name = class_name
      @show_option_button = show_option_button
    end

    def render
      build_html do |h|
        h.tag :div, class: "btn-group c-work-button-group #{@class_name}" do
          h.html Deprecated::Dropdowns::StatusSelectDropdownComponent.new(view_context, work: @work).render

          # TODO: アニメをお気に入りできるようにする
          # h.tag :button, type: "button", class: "btn btn-outline-warning" do
          #   h.tag :i, class: "fa-solid fa-star"
          # end

          if @show_option_button
            h.tag :button, {
              class: "btn btn-outline-secondary",
              data_controller: "tracking-offcanvas-button",
              data_tracking_offcanvas_button_work_id_value: @work.id,
              data_tracking_offcanvas_button_frame_path: view_context.fragment_trackable_work_path(@work.id),
              data_action: "click->tracking-offcanvas-button#open",
              type: "button"
            } do
              h.tag :i, class: "fa-solid fa-ellipsis-h"
            end
          end
        end
      end
    end
  end
end
