# frozen_string_literal: true

module Radios
  class ProgramSelectRadioComponent < ApplicationV6Component
    NO_SELECT = 0

    def initialize(view_context, library_entry:, programs:, class_name: "")
      super view_context
      @class_name = class_name
      @library_entry = library_entry
      @programs = programs
    end

    def render
      build_html do |h|
        h.tag :div, {
          class: "gx-3 row #{@class_name}",
          data_controller: "program-select-radio",
          data_program_select_radio_work_id_value: @library_entry.work_id,
          data_program_select_radio_init_program_id_value: @library_entry.program_id.presence || NO_SELECT
        } do
          h.tag :div, class: "col-6" do
            h.tag :div, class: "form-check" do
              input_id = "program-#{NO_SELECT}"

              input_attrs = {
                class: "form-check-input",
                data_action: "program-select-radio#change",
                id: input_id,
                name: "program-select",
                type: "radio",
                value: NO_SELECT
              }
              if @library_entry.program_id.nil?
                input_attrs[:checked] = true
              end
              h.tag :input, input_attrs

              h.tag :label, class: "form-check-label", for: input_id do
                h.tag :div do
                  h.text t("noun.no_select")
                end
              end
            end
          end

          @programs.each do |program|
            h.tag :div, class: "col-6" do
              h.tag :div, class: "form-check" do
                input_id = "program-#{program.id}"

                input_attrs = {
                  class: "form-check-input",
                  data_action: "program-select-radio#change",
                  id: input_id,
                  name: "program-select",
                  type: "radio",
                  value: program.id
                }
                if @library_entry.program_id == program.id
                  input_attrs[:checked] = true
                end
                h.tag :input, input_attrs

                h.tag :label, class: "form-check-label", for: input_id do
                  h.tag :div do
                    h.text program.channel.name
                  end

                  if program.started_at.present?
                    h.tag :div, class: "small text-muted" do
                      h.text "#{display_time(program.started_at)}~"
                    end
                  end
                end
              end
            end
          end
        end
      end
    end
  end
end
