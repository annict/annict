# frozen_string_literal: true

module V6::Contents
  class EpisodeRecordContentComponent < V6::ApplicationComponent
    def initialize(view_context, record:, page_category:, show_box: true)
      super view_context
      @record = record
      @page_category = page_category
      @episode_record = @record.episode_record
      @show_box = show_box
      @anime = @record.anime
      @episode = @episode_record.episode
    end

    def render
      build_html do |h|
        h.tag :div, class: "" do
          if rating_or_comment?
            h.html(V6::SpoilerGuardComponent.new(view_context, record: @record).render { |h|
              h.tag :div, class: "c-record-content__wrapper mb-3" do
                if @record.episode_record.rating_state
                  h.html V6::RatingLabelComponent.new(view_context, rating: @episode_record.rating_state, advanced_rating: @episode_record.rating, class_name: "mb-1").render
                end

                h.html(V6::BodyComponent.new(view_context, height: 300, format: :html).render { |h|
                  h.html render_markdown(@episode_record.comment)
                })
              end
            })
          end

          if @show_box
            if rating_or_comment?
              h.tag :hr
            end

            h.html V6::Boxes::AnimeBoxComponent.new(view_context,
              anime: @anime,
              page_category: @page_category,
              episode: @episode).render

            h.tag :hr
          end

          h.tag :div, class: "mt-1" do
            h.html V6::Footers::RecordFooterComponent.new(view_context, record: @record).render
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
