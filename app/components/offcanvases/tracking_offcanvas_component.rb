# frozen_string_literal: true

module Offcanvases
  class TrackingOffcanvasComponent < ApplicationV6Component
    def initialize(view_context)
      super view_context
    end

    def render
      build_html do |h|
        h.tag :div, class: "c-tracking-offcanvas offcanvas offcanvas-end" do
          h.tag :div, class: "justify-content-end offcanvas-header" do
            h.tag :button, class: "btn-close text-reset", data_bs_dismiss: "offcanvas", type: "button"
          end

          h.tag :turbo_frame, {
            class: "h-100",
            data_controller: "reloadable",
            data_reloadable_event_name_value: "tracking-offcanvas",
            id: "c-tracking-offcanvas-frame",
            src: "",
            tabindex: "-1"
          } do
            h.tag :div, class: "align-items-center d-flex h-100 justify-content-center" do
              h.tag :div, class: "spinner-border" do
                h.tag :span, class: "visually-hidden" do
                  h.text "Loading..."
                end
              end
            end
          end
        end
      end
    end
  end
end
