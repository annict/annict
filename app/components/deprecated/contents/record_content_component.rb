# typed: false
# frozen_string_literal: true

module Deprecated::Contents
  class RecordContentComponent < Deprecated::ApplicationV6Component
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
          if rating_or_comment?
            h.html(Deprecated::SpoilerGuardComponent.new(view_context, record: @record).render { |h|
              h.tag :div, class: "c-record-content__wrapper mb-3" do
                if @record.episode_record?
                  if @record.rating
                    h.html Deprecated::Labels::RatingLabelComponent.new(view_context, rating: @record.rating, advanced_rating: @record.advanced_rating, class_name: "mb-1").render
                  end

                  h.html Deprecated::BodyV6Component.new(view_context, content: @record.comment, format: :markdown, height: 300).render
                else
                  h.tag :div, class: "row g-3" do
                    h.tag :div, class: "col-12 col-md-4 order-1 order-md-2" do
                      h.html Deprecated::LabelGroups::WorkRecordRatingLabelGroupComponent.new(view_context, work_record: @record.work_record).render
                    end

                    h.tag :div, class: "col-12 col-md-8 order-2 order-md-1" do
                      h.html Deprecated::BodyV6Component.new(view_context, content: @record.comment, format: :markdown, height: 300).render
                    end
                  end
                end
              end
            })
          end

          if @show_box
            h.tag :hr

            h.html Deprecated::Boxes::WorkBoxComponent.new(view_context, work: @work, episode: @episode).render

            h.tag :hr
          end

          h.tag :div, class: "mt-1" do
            h.html Deprecated::Footers::RecordFooterComponent.new(view_context, record: @record).render
          end
        end
      end
    end

    private

    def rating_or_comment?
      @record.rating || @record.comment.present?
    end
  end
end
