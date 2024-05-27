# typed: false
# frozen_string_literal: true

module Deprecated::Textareas
  class RecordTextareaComponent < Deprecated::ApplicationV6Component
    def initialize(view_context, form:, textarea_name:, optional_textarea_classname: "", autofocus: true)
      super view_context
      @form = form
      @textarea_name = textarea_name
      @optional_textarea_classname = optional_textarea_classname
      @autofocus = autofocus
    end

    def render
      build_html do |h|
        h.html text_area_tag(
          @textarea_name,
          @form.object.comment,
          autofocus: @autofocus,
          class: textarea_classname,
          "data-action": "keyup->record-textarea#updateCharactersCount",
          "data-controller": "record-textarea",
          "data-record-textarea-characters-counter-id": @form.object.unique_id,
          placeholder: t("messages.episode_records.new.write_your_comment"),
          rows: "5"
        )

        h.tag :div, {
          class: "small text-muted text-end",
          data_controller: "characters-counter",
          data_characters_counter_characters_counter_id: @form.object.unique_id
        } do
          h.tag :span, data_characters_counter_target: "value"
          h.text t("noun.characters2")
        end
      end
    end

    private

    def textarea_classname
      @textarea_classname = %w[]
      @textarea_classname << @optional_textarea_classname
      @textarea_classname.join(" ")
    end
  end
end
