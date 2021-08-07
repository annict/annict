# frozen_string_literal: true

module Modals
  class TrackingModalComponent < ApplicationV6Component
    def initialize(view_context)
      super view_context
    end

    def render
      build_html do |h|
        h.tag :div, class: "c-tracking-modal modal" do
          h.tag :div, class: "modal-dialog modal-lg" do
            h.tag :div, class: "modal-content" do
              h.tag :turbo_frame,
                data_controller: "reloadable",
                data_reloadable_event_name_value: "tracking-modal",
                id: "c-tracking-modal-frame",
                src: "",
                tabindex: "-1"
            end
          end
        end
      end
    end
  end
end
