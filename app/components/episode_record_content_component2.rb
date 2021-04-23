# frozen_string_literal: true

class EpisodeRecordContentComponent2 < ApplicationComponent2
  def initialize(view_context, current_user:, record:, show_card: true)
    super view_context
    @current_user = current_user
    @record = record
    @episode_record = @record.episode_record
    @show_card = show_card
  end

  def render
    build_html do |h|
      h.tag :div, class: "c-episode-record-content c-record-content" do
        if @record.episode_record.rating_state || @episode_record.body.present?
          h.html(SpoilerGuardComponent2.new(view_context, record: @record, current_user: @current_user).render do |h|
            h.tag :div, class: "c-record-content__wrapper mb-3" do
              if @record.episode_record.rating_state
                h.html RatingLabelComponent2.new(view_context, rating: @episode_record.rating_state, advanced_rating: @episode_record.rating, class_name: "mb-1").render
              end

              h.html(BodyComponent2.new(view_context, height: 300, format: :html).render do |h|
                h.html render_markdown(@episode_record.comment)
              end)
            end
          end)
        end

        if @show_card
          render EpisodeRecordCardComponent2.new(episode_record: @episode_record)
        end

        h.html RecordFooterComponent2.new(view_context, record: @record).render
      end
    end
  end
end
