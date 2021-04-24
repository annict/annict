# frozen_string_literal: true

module Dropdowns
  class RecordOptionsDropdownComponent2 < ApplicationComponent2
    def initialize(view_context, current_user:, record:)
      super view_context
      @current_user = current_user
      @record = record
    end

    def render
      build_html do |h|
        h.tag :div, class: "c-record-options-dropdown d-inline-block dropdown" do
          h.tag :div, class: "dropdown-toggle", data_toggle: "dropdown" do
            h.tag :i, class: "far fa-ellipsis-h"
          end

          h.tag :div, class: "dropdown-menu dropdown-menu-right" do
            if @current_user.id != @record.user_id
              h.tag :a, href: "#", class: "dropdown-item" do
                h.tag :div,
                  data_controller: "mute-user",
                  data_mute_user_user_id: @record.user_id do
                    h.text t("verb.mute")
                  end
              end
            end

            if RecordPolicy.new(@current_user, @record).update?
              h.tag :a, href: view_context.my_edit_record_path(@record.user.username, @record.id), class: "dropdown-item" do
                h.text t("noun.edit")
              end

              h.tag :a, href: view_context.record_path(@record.user.username, @record.id), class: "dropdown-item", data: { confirm: t("messages._common.are_you_sure") }, method: :delete do
                h.text t("noun.delete")
              end
            end
          end
        end
      end
    end
  end
end
