# frozen_string_literal: true

module Contents
  class RecordContentComponent < ApplicationV6Component
    def initialize(view_context, record:, show_box: true)
      super view_context
      @record = record
      @work = @record.work
      @episode = @record.episode
      @show_box = show_box
    end

    def render
      build_html do |h|
        h.tag :div do
          unless @record.instant?
            h.html(SpoilerGuardComponent.new(view_context, record: @record).render { |h|
              h.tag :div, class: "c-record-content__wrapper mb-3" do
                if @record.rating
                  h.html Labels::RatingLabelComponent.new(view_context, rating: @record.rating, advanced_rating: @record.advanced_rating, class_name: "mb-1").render
                end

                if @record.deprecated_rating_exists?
                  h.html Collapses::DeprecatedRatingsCollapseComponent.new(view_context, record: @record, class_name: "ms-2 text-muted").render
                end

                h.html BodyV6Component.new(view_context, content: @record.body, format: :markdown, height: 300).render
              end
            })
          end

          if @show_box
            h.tag :hr

            h.html Boxes::WorkBoxComponent.new(view_context, work: @work, episode: @episode).render

            h.tag :hr
          end

          h.tag :div, class: "mt-1" do
            h.html Footers::RecordFooterComponent.new(view_context, record: @record).render
          end
        end
      end
    end
  end
end
