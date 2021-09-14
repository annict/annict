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
          h.tag :a, aria_expanded: "false", class: "text-body u-collapse-with-icon", data_bs_toggle: "collapse", href: "#recordFormOptions" do
            h.text t("noun.options")
            h.tag :i, class: "far fa-angle-down ms-1"
          end

          h.tag :div, class: "collapse mt-2", id: "recordFormOptions" do
            h.html @form.label(:watched_at)

            h.tag :div, class: "row" do
              h.tag :div, class: "col-12 col-xl-4" do
                h.html @form.datetime_field(:watched_at, class: "form-control", disabled: !current_user&.supporter?)
              end
            end

            h.tag :div, class: "form-text text-muted" do
              if current_user&.supporter?
                h.text t("messages._components.episode_record_form.hint_on_watched_at_for_supporter")
              else
                h.html t("messages._components.episode_record_form.hint_on_watched_at_html")
              end
            end
          end
        end
      end
    end
  end
end