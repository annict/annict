# frozen_string_literal: true

module Collapses
  class RecordFormOptionsCollapseComponent < ApplicationV6Component
    def initialize(view_context, form:)
      super view_context
      @form = form
    end

    def render
      build_html do |h|
        h.tag :div do
          h.tag :a, aria_expanded: "false", class: "text-body u-collapse-with-icon", data_bs_toggle: "collapse", href: "##{collapse_id}" do
            h.text t("noun.options")
            h.tag :i, class: "far fa-angle-down ms-1"
          end

          h.tag :div, class: "collapse mt-2", id: collapse_id do
            h.tag :div do
              h.html @form.label(:watched_at, class: "form-label")
              h.html @form.datetime_field(
                :watched_at,
                class: "form-control",
                disabled: !current_user&.supporter?,
                style: "width: 300px;",
                value: @form.object.watched_at&.strftime("%Y-%m-%dT%H:%M")
              )
            end

            h.tag :div, class: "form-text text-muted" do
              h.tag :p, class: "mb-0" do
                h.html t("messages._components.record_form_options_collapse.hint_on_watched_at_html")
              end

              unless current_user&.supporter?
                h.tag :p, class: "mb-0" do
                  h.html t("messages._components.record_form_options_collapse.hint_on_watched_at_for_non_supporters_html")
                end
              end
            end
          end
        end
      end
    end

    private

    def collapse_id
      "recordFormOptions#{@form.object_id}"
    end
  end
end
