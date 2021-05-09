# frozen_string_literal: true

module Contents
  class EpisodeRecordContentComponent < ApplicationComponent
    def initialize(view_context, record:, show_card: true)
      super view_context
      @record = record
      @episode_record = @record.episode_record
      @show_card = show_card
    end

    def render
      build_html do |h|
        h.tag :div, class: "" do
          if @record.episode_record.rating_state || @episode_record.body.present?
            h.html(SpoilerGuardComponent.new(view_context, record: @record).render { |h|
              h.tag :div, class: "c-record-content__wrapper mb-3" do
                if @record.episode_record.rating_state
                  h.html RatingLabelComponent.new(view_context, rating: @episode_record.rating_state, advanced_rating: @episode_record.rating, class_name: "mb-1").render
                end

                h.html(BodyComponent.new(view_context, height: 300, format: :html).render { |h|
                  h.html render_markdown(@episode_record.comment)
                })
              end
            })
          end

          if @show_card
            h.html Cards::EpisodeRecordCardComponent.new(view_context, episode_record: @episode_record).render
          end

          h.tag :div, class: "mt-1" do
            h.html Footers::RecordFooterComponent.new(view_context, record: @record).render
          end
        end
      end
    end
  end
end
