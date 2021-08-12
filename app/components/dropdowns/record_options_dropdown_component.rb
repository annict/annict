# frozen_string_literal: true

module Dropdowns
  class RecordOptionsDropdownComponent < ApplicationV6Component
    def initialize(view_context, record:, show_box: true, show_options: true)
      super view_context
      @record = record
      @show_box = show_box
      @show_options = show_options
    end

    def render
      build_html do |h|
        h.tag :div, class: "c-record-options-dropdown d-inline-block dropdown u-cursor-pointer" do
          h.tag :div, class: "dropdown-toggle", data_bs_toggle: "dropdown" do
            h.tag :i, class: "far fa-ellipsis-h"
          end

          h.tag :div, class: "dropdown-menu" do
            if RecordPolicy.new(current_user, @record).update?
              h.tag :a, href: view_context.fragment_edit_record_path(@record.user.username, @record.id, show_box: @show_box, show_options: @show_options), class: "dropdown-item" do
                h.text t("noun.edit")
              end

              h.tag :a, href: view_context.record_path(@record.user.username, @record.id), class: "dropdown-item", data_confirm: t("messages._common.are_you_sure"), data_method: :delete do
                h.text t("noun.delete")
              end
            end
          end
        end
      end
    end
  end
end
