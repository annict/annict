# frozen_string_literal: true

module LabelGroups
  class WorkRecordRatingLabelGroupComponent < ApplicationV6Component
    def initialize(view_context, work_record:, class_name: "")
      super view_context
      @work_record = work_record
      @class_name = class_name
    end

    def render
      build_html do |h|
        h.tag :div, class: @class_name do
          if @work_record.rating_overall_state
            h.tag :div, class: "gx-3 justify-content-center justify-content-md-start row" do
              h.tag :div, class: "col-4 col-md-6" do
                h.tag :span, class: "me-1 small" do
                  h.text t("noun.overall")
                end
              end

              h.tag :div, class: "col-4" do
                h.html Labels::RatingLabelComponent.new(view_context, rating: @work_record.rating_overall_state).render
              end
            end
          end

          if @work_record.rating_animation_state
            h.tag :div, class: "gx-3 justify-content-center justify-content-md-start row" do
              h.tag :div, class: "col-4 col-md-6" do
                h.tag :span, class: "me-1 small" do
                  h.text t("noun.animation")
                end
              end

              h.tag :div, class: "col-4 col-md-6" do
                h.html Labels::RatingLabelComponent.new(view_context, rating: @work_record.rating_animation_state).render
              end
            end
          end

          if @work_record.rating_character_state
            h.tag :div, class: "gx-3 justify-content-center justify-content-md-start row" do
              h.tag :div, class: "col-4 col-md-6" do
                h.tag :span, class: "me-1 small" do
                  h.text t("noun.character")
                end
              end

              h.tag :div, class: "col-4 col-md-6" do
                h.html Labels::RatingLabelComponent.new(view_context, rating: @work_record.rating_character_state).render
              end
            end
          end

          if @work_record.rating_story_state
            h.tag :div, class: "gx-3 justify-content-center justify-content-md-start row" do
              h.tag :div, class: "col-4 col-md-6" do
                h.tag :span, class: "me-1 small" do
                  h.text t("noun.story")
                end
              end

              h.tag :div, class: "col-4 col-md-6" do
                h.html Labels::RatingLabelComponent.new(view_context, rating: @work_record.rating_story_state).render
              end
            end
          end

          if @work_record.rating_music_state
            h.tag :div, class: "gx-3 justify-content-center justify-content-md-start row" do
              h.tag :div, class: "col-4 col-md-6" do
                h.tag :span, class: "me-1 small" do
                  h.text t("noun.music")
                end
              end

              h.tag :div, class: "col-4 col-md-6" do
                h.html Labels::RatingLabelComponent.new(view_context, rating: @work_record.rating_music_state).render
              end
            end
          end
        end
      end
    end
  end
end
