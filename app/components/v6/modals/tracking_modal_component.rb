# frozen_string_literal: true

module V6::Modals
  class TrackingModalComponent < V6::ApplicationComponent
    def initialize(view_context)
      super view_context
    end

    def render
      build_html do |h|
        h.tag :div, class: "c-tracking-modal" do
          h.tag :turbo_frame,
            class: "modal",
            data_controller: "reloadable",
            data_reloadable_event_name_value: "tracking-modal",
            id: "c-tracking-modal__frame",
            src: "",
            tabindex: "-1"
        end
      end
    end
  end
end
