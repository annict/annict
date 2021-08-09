# frozen_string_literal: true

module Radios
  class ProgramSelectRadioComponent < ApplicationV6Component
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
          data_program_select_radio_anime_id_value: @library_entry.work_id,
          data_program_select_radio_init_program_id_value: @library_entry.program_id
        } do
          @programs.each do |program|
            h.tag :div, class: "col-6" do
              h.tag :div, class: "form-check" do
                id = "program-#{program.id}"

                input_attrs = {
                  class: "form-check-input",
                  data_action: "program-select-radio#change",
                  id: id,
                  name: "program-select",
                  type: "radio",
                  value: program.id
                }
                if @library_entry.program_id == program.id
                  input_attrs[:checked] = true
                end
                h.tag :input, input_attrs

                h.tag :label, class: "form-check-label", for: id do
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
