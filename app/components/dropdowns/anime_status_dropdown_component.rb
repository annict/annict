# frozen_string_literal: true

module Dropdowns
  class AnimeStatusDropdownComponent < ApplicationV6Component
    def initialize(view_context, anime:, class_name: "")
      super view_context
      @anime = anime
      @class_name = class_name
    end

    def render
      build_html do |h|
        h.tag :div, class: "btn-group c-anime-status-dropdown" do
          h.tag :button, class: "btn dropdown-toggle u-btn-outline-status", type: "button", id: "anime#{@anime.id}Dropdown", data_bs_toggle: "dropdown" do
            h.tag :i, class: "far fa-bars"
          end

          h.tag :ul, class: "dropdown-menu" do
            h.tag :li do
              h.tag :a, class: "dropdown-item", href: "#" do
                h.text "watching"
              end
            end
          end
        end
        # h.tag :div, {
        #   class: "c-anime-status-button #{status_selector_class_name}",
        #   data_controller: "status-selector",
        #   data_status_selector_anime_id_value: @anime.id,
        #   data_status_selector_selected_class: "c-status-selector--selected",
        #   data_status_selector_page_category_value: page_category,
        #   style: @style
        # } do
        #   h.tag :select, {
        #     class: "form-select",
        #     data_action: "status-selector#change",
        #     data_status_selector_target: "kind"
        #   } do
        #     status_options.each do |status_option|
        #       h.tag :option, value: status_option[1] do
        #         h.text status_option[0]
        #       end
        #     end
        #   end

        #   h.tag :i, class: "fas fa-caret-down"
        # end
      end
    end

    private

    def status_selector_class_name
      classes = []
      classes += @class_name.split(" ")
      classes.uniq.join(" ")
    end

    def status_options
      Status.kind.options.insert(0, [t("messages.components.status_selector.select_status"), "no_select"])
    end
  end
end
