# frozen_string_literal: true

module LabelGroups
  class WorkRecordRatingLabelGroupComponent < ApplicationV6Component
    def initialize(view_context, record:, class_name: "")
      super view_context
      @record = record
      @work_record = @record.work_record
      @rating = @record.rating
      @animation_rating = @work_record.animation_rating
      @music_rating = @work_record.music_rating
      @story_rating = @work_record.story_rating
      @character_rating = @work_record.character_rating
      @class_name = class_name
    end

    def render
      return "" unless @work_record

      build_html do |h|
        h.tag :div, class: @class_name do
          if @rating
            h.tag :div, class: "gx-3 justify-content-center justify-content-md-start row" do
              h.tag :div, class: "col-4 col-md-6" do
                h.tag :span, class: "me-1 small" do
                  h.text t("noun.overall")
                end
              end

              h.tag :div, class: "col-4" do
                h.html Labels::RatingLabelComponent.new(view_context, rating: @rating).render
              end
            end
          end

          if @animation_rating
            h.tag :div, class: "gx-3 justify-content-center justify-content-md-start row" do
              h.tag :div, class: "col-4 col-md-6" do
                h.tag :span, class: "me-1 small" do
                  h.text t("noun.animation")
                end
              end

              h.tag :div, class: "col-4 col-md-6" do
                h.html Labels::RatingLabelComponent.new(view_context, rating: @animation_rating).render
              end
            end
          end

          if @character_rating
            h.tag :div, class: "gx-3 justify-content-center justify-content-md-start row" do
              h.tag :div, class: "col-4 col-md-6" do
                h.tag :span, class: "me-1 small" do
                  h.text t("noun.character")
                end
              end

              h.tag :div, class: "col-4 col-md-6" do
                h.html Labels::RatingLabelComponent.new(view_context, rating: @character_rating).render
              end
            end
          end

          if @story_rating
            h.tag :div, class: "gx-3 justify-content-center justify-content-md-start row" do
              h.tag :div, class: "col-4 col-md-6" do
                h.tag :span, class: "me-1 small" do
                  h.text t("noun.story")
                end
              end

              h.tag :div, class: "col-4 col-md-6" do
                h.html Labels::RatingLabelComponent.new(view_context, rating: @story_rating).render
              end
            end
          end

          if @music_rating
            h.tag :div, class: "gx-3 justify-content-center justify-content-md-start row" do
              h.tag :div, class: "col-4 col-md-6" do
                h.tag :span, class: "me-1 small" do
                  h.text t("noun.music")
                end
              end

              h.tag :div, class: "col-4 col-md-6" do
                h.html Labels::RatingLabelComponent.new(view_context, rating: @music_rating).render
              end
            end
          end
        end
      end
    end
  end
end
