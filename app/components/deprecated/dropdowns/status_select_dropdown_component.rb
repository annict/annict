# typed: false
# frozen_string_literal: true

module Deprecated::Dropdowns
  class StatusSelectDropdownComponent < Deprecated::ApplicationV6Component
    def initialize(view_context, work:, class_name: "")
      super view_context
      @work = work
      @class_name = class_name
    end

    def render
      build_html do |h|
        h.tag :div, {
          class: "btn-group c-status-select-dropdown",
          data_controller: "status-select-dropdown",
          data_status_select_dropdown_work_id_value: @work.id,
          data_status_select_dropdown_page_category_value: page_category,
          data_status_select_dropdown_kind_icons_value: Status::KIND_ICONS.to_json
        } do
          h.tag :button, {
            class: "btn dropdown-toggle u-btn-outline-status",
            data_status_select_dropdown_target: "button",
            type: "button",
            id: "work#{@work.id}Dropdown",
            data_bs_toggle: "dropdown"
          } do
            h.tag :i, class: "fas fa-bars"
          end

          h.tag :ul, class: "dropdown-menu" do
            status_options.each do |(text, kind)|
              kind_v3 = Status.kind_v2_to_v3(kind).to_s

              h.tag :li do
                h.tag :button, {
                  class: "dropdown-item py-2",
                  data_action: "status-select-dropdown#change",
                  data_status_kind: kind_v3
                } do
                  if kind == "no_status"
                    h.tag :i, class: "fas fa-bars me-3"
                  else
                    h.tag :i, class: "fas fa-#{Status.kind_icon(kind_v3.to_sym)} me-3 text-status-#{kind_v3.dasherize}"
                  end

                  h.text text
                end
              end
            end
          end
        end
      end
    end

    private

    def status_options
      Status.kind.options.insert(0, [t("noun.no_select"), "no_select"])
    end
  end
end
