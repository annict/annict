# frozen_string_literal: true

module Collapses
  class DeprecatedRatingsCollapseComponent < ApplicationV6Component
    def initialize(view_context, record:, class_name: "")
      super view_context
      @record = record
      @deprecated_animation_rating = @record.deprecated_animation_rating
      @deprecated_music_rating = @record.deprecated_music_rating
      @deprecated_story_rating = @record.deprecated_story_rating
      @deprecated_character_rating = @record.deprecated_character_rating
      @class_name = class_name
    end

    def render
      return "" unless @record.deprecated_rating_exists?

      build_html do |h|
        h.tag :span, class: "u-cursor-pointer #{@class_name}", data_bs_toggle: "collapse", href: "##{collapse_dom_id}" do
          h.tag :i, class: "far fa-ellipsis-h"
        end

        h.tag :div, class: "collapse p-2", id: collapse_dom_id do
          h.tag :div, class: "gap-1 vstack" do
            if @deprecated_animation_rating
              h.tag :div, class: "gx-3 row", style: "width: 250px;" do
                h.tag :div, class: "col-6" do
                  h.tag :span, class: "me-1 small text-muted" do
                    h.text t("noun.animation")
                  end
                end

                h.tag :div, class: "col-6" do
                  h.html Labels::RatingLabelComponent.new(view_context, rating: @deprecated_animation_rating).render
                end
              end
            end

            if @deprecated_music_rating
              h.tag :div, class: "gx-3 row", style: "width: 250px;" do
                h.tag :div, class: "col-6" do
                  h.tag :span, class: "me-1 small text-muted" do
                    h.text t("noun.music")
                  end
                end

                h.tag :div, class: "col-6" do
                  h.html Labels::RatingLabelComponent.new(view_context, rating: @deprecated_music_rating).render
                end
              end
            end

            if @deprecated_story_rating
              h.tag :div, class: "gx-3 row", style: "width: 250px;" do
                h.tag :div, class: "col-6" do
                  h.tag :span, class: "me-1 small text-muted" do
                    h.text t("noun.story")
                  end
                end

                h.tag :div, class: "col-6" do
                  h.html Labels::RatingLabelComponent.new(view_context, rating: @deprecated_story_rating).render
                end
              end
            end

            if @deprecated_character_rating
              h.tag :div, class: "gx-3 row", style: "width: 250px;" do
                h.tag :div, class: "col-6" do
                  h.tag :span, class: "me-1 small text-muted" do
                    h.text t("noun.character")
                  end
                end

                h.tag :div, class: "col-6" do
                  h.html Labels::RatingLabelComponent.new(view_context, rating: @deprecated_character_rating).render
                end
              end
            end
          end

          h.tag :div, class: "mt-2 text-muted u-very-small" do
            h.tag :i, class: "far fa-info-circle me-1"
            h.text t("messages._components.deprecated_ratings_collapse.hint")
          end
        end
      end
    end

    private

    def collapse_dom_id
      "deprecated-rating-labels-#{@record.id}"
    end
  end
end
