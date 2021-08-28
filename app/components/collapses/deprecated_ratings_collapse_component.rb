# frozen_string_literal: true

module Collapses
  class DeprecatedRatingsCollapseComponent < ApplicationV6Component
    def initialize(view_context, record:, class_name: "")
      super view_context
      @record = record
      @work_record = @record.work_record
      @animation_rating = @@work_record.animation_rating
      @music_rating = @@work_record.music_rating
      @story_rating = @@work_record.story_rating
      @character_rating = @@work_record.character_rating
      @class_name = class_name
    end

    def render
      return "" unless @work_record

      build_html do |h|
        h.tag :span, class: "u-cursor-pointer #{@class_name}", data_bs_toggle: "collapse", href: "##{collapse_dom_id}" do
          h.tag :i, class: "far fa-ellipsis-h"
        end

        h.tag :div, class: "collapse p-2", id: collapse_dom_id do
          h.tag :div, class: "gap-1 vstack" do
            if @animation_rating
              h.tag :div, class: "gx-3 row", style: "width: 250px;" do
                h.tag :div, class: "col-6" do
                  h.tag :span, class: "me-1 small text-muted" do
                    h.text t("noun.animation")
                  end
                end

                h.tag :div, class: "col-6" do
                  h.html Labels::RatingLabelComponent.new(view_context, rating: @animation_rating).render
                end
              end
            end

            if @music_rating
              h.tag :div, class: "gx-3 row", style: "width: 250px;" do
                h.tag :div, class: "col-6" do
                  h.tag :span, class: "me-1 small text-muted" do
                    h.text t("noun.music")
                  end
                end

                h.tag :div, class: "col-6" do
                  h.html Labels::RatingLabelComponent.new(view_context, rating: @music_rating).render
                end
              end
            end

            if @story_rating
              h.tag :div, class: "gx-3 row", style: "width: 250px;" do
                h.tag :div, class: "col-6" do
                  h.tag :span, class: "me-1 small text-muted" do
                    h.text t("noun.story")
                  end
                end

                h.tag :div, class: "col-6" do
                  h.html Labels::RatingLabelComponent.new(view_context, rating: @story_rating).render
                end
              end
            end

            if @character_rating
              h.tag :div, class: "gx-3 row", style: "width: 250px;" do
                h.tag :div, class: "col-6" do
                  h.tag :span, class: "me-1 small text-muted" do
                    h.text t("noun.character")
                  end
                end

                h.tag :div, class: "col-6" do
                  h.html Labels::RatingLabelComponent.new(view_context, rating: @character_rating).render
                end
              end
            end
          end

          h.tag :div, class: "mt-2 text-muted u-very-small" do
            h.tag :i, class: "far fa-info-circle me-1"
            h.text t("messages._components.ratings_collapse.hint")
          end
        end
      end
    end

    private

    def collapse_dom_id
      "rating-labels-#{@record.id}"
    end
  end
end
