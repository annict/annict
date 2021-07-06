# frozen_string_literal: true

module Contents
  class EpisodeRecordContentComponent < ApplicationV6Component
    def initialize(view_context, record:, show_box: true)
      super view_context
      @record = record
      @episode_record = @record.episode_record
      @show_box = show_box
      @anime = @record.anime
      @episode = @episode_record.episode
    end

    def render
      build_html do |h|
        h.tag :div do
          if rating_or_comment?
            h.html(SpoilerGuardV6Component.new(view_context, record: @record).render { |h|
              h.tag :div, class: "c-record-content__wrapper mb-3" do
                if @record.episode_record.rating_state
                  h.html Labels::RatingLabelComponent.new(view_context, rating: @episode_record.rating_state, advanced_rating: @episode_record.rating, class_name: "mb-1").render
                end

                h.html BodyV6Component.new(view_context, content: @episode_record.comment, format: :markdown, height: 300).render
              end
            })
          end

          if @show_box
            h.tag :hr

            h.html Boxes::AnimeBoxComponent.new(view_context,
              anime: @anime,
              episode: @episode).render

            h.tag :hr
          end

          h.tag :div, class: "mt-1" do
            h.html Footers::RecordFooterComponent.new(view_context, record: @record).render
          end
        end
      end
    end

    private

    def rating_or_comment?
      @record.episode_record.rating_state || @episode_record.body.present?
    end
  end
end
