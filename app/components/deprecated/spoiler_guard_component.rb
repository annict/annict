# frozen_string_literal: true

class Deprecated::SpoilerGuardComponent < Deprecated::ApplicationV6Component
  def initialize(view_context, record:)
    super view_context
    @record = record
  end

  def render
    build_html do |h|
      h.tag :div, {
        class: "c-spoiler-guard is-spoiler",
        data_controller: "spoiler-guard",
        data_action: "click->spoiler-guard#hide",
        data_spoiler_guard_work_id_value: @record.work_id,
        data_spoiler_guard_episode_id_value: @record.episode_record&.episode_id
      } do
        yield h
      end
    end
  end
end
