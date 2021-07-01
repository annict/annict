# frozen_string_literal: true

module Contents
  class RecordContentComponent < ApplicationV6Component
    def initialize(view_context, record:, show_box: true)
      super view_context
      @record = record
      @show_box = show_box
    end

    def render
      build_html do |h|
        h.tag :div, class: "c-record-content" do
          h.tag :div, class: "mb-3" do
            h.html Headers::RecordHeaderComponent.new(view_context, record: @record).render
          end

          if @record.episode_record?
            h.html Contents::EpisodeRecordContentComponent.new(
              view_context,
              record: @record,
              show_box: @show_box
            ).render
          elsif @record.anime_record?
            h.html Contents::AnimeRecordContentComponent.new(
              view_context,
              record: @record,
              show_box: @show_box
            ).render
          end
        end
      end
    end
  end
end
