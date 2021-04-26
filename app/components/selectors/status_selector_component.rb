# frozen_string_literal: true

module Selectors
  class StatusSelectorComponent < ApplicationComponent2
    def initialize(view_context, anime:, page_category:, small: false, class_name: "")
      super view_context
      @anime = anime
      @page_category = page_category
      @small = small
      @class_name = class_name
      @init_kind = @anime.status_kind.presence || "no_select"
    end

    def render
      build_html do |h|
        h.tag :div,
          class: "c-status-selector #{status_selector_class_name}",
          data_controller: "status-selector",
          data_status_selector_anime_id_value: @anime.id,
          data_status_selector_init_kind_value: @init_kind,
          data_status_selector_page_category_value: @page_category do
            h.tag :div, class: "c-status-selector__object" do
              h.tag :select, class: "w-100", data_status_selector_target: "kind", data_action: "status-selector#change" do
                status_options.each do |status_option|
                  h.tag :option, value: status_option[1] do
                    h.text status_option[0]
                  end
                end
              end

              h.tag :i, class: "fas fa-caret-down"
            end
        end
      end
    end

    private

    def status_selector_class_name
      classes = []
      classes += @class_name.split(" ")
      classes << "c-status-selector--small" if @small
      classes.uniq.join(" ")
    end

    def status_options
      Status.kind.options.insert(0, [t("messages.components.status_selector.select_status"), "no_select"])
    end
  end
end
